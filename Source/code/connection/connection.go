package connection

import (
	"context"
	"errors"
	"runtime/debug"
	"strings"
	"time"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
	"github.com/puzpuzpuz/xsync/v3"
)

// ContextInfo holds information about a context
type ContextInfo struct {
	VMixAddr   string
	ActionType string
}

type ConnectionManager struct {
	connections   *xsync.MapOf[string, *vMixConnection]
	contextMap    *xsync.MapOf[string, *ContextInfo]                                         // contextID -> ContextInfo
	actionTypeMap *xsync.MapOf[string, *xsync.MapOf[string, *xsync.MapOf[string, struct{}]]] // vmixAddr -> actionType -> contextIDs
	logger        logger.Logger

	// callbacks. string is vMixAddr.
	xmlCallback       func(*vmixtcp.XMLResponse, vmixtcp.Vmix, string)
	tallyCallback     func(*vmixtcp.TallyResponse, vmixtcp.Vmix, string)
	actsCallback      func(*vmixtcp.ActsResponse, vmixtcp.Vmix, string)
	versionCallback   func(*vmixtcp.VersionResponse, vmixtcp.Vmix, string)
	subscribeCallback func(*vmixtcp.SubscribeResponse, vmixtcp.Vmix, string)
}

type vMixConnection struct {
	client      vmixtcp.Vmix
	contexts    *xsync.MapOf[string, struct{}]
	retryCancel context.CancelFunc

	// チャネル for vMix callbacks
	xmlChan       chan *vmixtcp.XMLResponse
	tallyChan     chan *vmixtcp.TallyResponse
	actsChan      chan *vmixtcp.ActsResponse
	versionChan   chan *vmixtcp.VersionResponse
	subscribeChan chan *vmixtcp.SubscribeResponse
}

func NewConnectionManager(logger logger.Logger) *ConnectionManager {
	return &ConnectionManager{
		connections:   xsync.NewMapOf[string, *vMixConnection](),
		contextMap:    xsync.NewMapOf[string, *ContextInfo](),
		actionTypeMap: xsync.NewMapOf[string, *xsync.MapOf[string, *xsync.MapOf[string, struct{}]]](),
		logger:        logger,

		xmlCallback:       func(*vmixtcp.XMLResponse, vmixtcp.Vmix, string) {},
		tallyCallback:     func(*vmixtcp.TallyResponse, vmixtcp.Vmix, string) {},
		actsCallback:      func(*vmixtcp.ActsResponse, vmixtcp.Vmix, string) {},
		versionCallback:   func(*vmixtcp.VersionResponse, vmixtcp.Vmix, string) {},
		subscribeCallback: func(*vmixtcp.SubscribeResponse, vmixtcp.Vmix, string) {},
	}
}

func (cm *ConnectionManager) initActionTypeMap(vmixAddr, actionType string) {
	addrMap, exists := cm.actionTypeMap.Load(vmixAddr)
	if !exists {
		addrMap = xsync.NewMapOf[string, *xsync.MapOf[string, struct{}]]()
		cm.actionTypeMap.Store(vmixAddr, addrMap)
	}

	actionMap, exists := addrMap.Load(actionType)
	if !exists {
		actionMap = xsync.NewMapOf[string, struct{}]()
		addrMap.Store(actionType, actionMap)
	}
}

const (
	defaultBufferSize = 50
)

func (cm *ConnectionManager) newVMixConnection() *vMixConnection {
	return &vMixConnection{
		client:        nil,
		contexts:      xsync.NewMapOf[string, struct{}](),
		retryCancel:   func() {},
		xmlChan:       make(chan *vmixtcp.XMLResponse, defaultBufferSize),
		tallyChan:     make(chan *vmixtcp.TallyResponse, defaultBufferSize),
		actsChan:      make(chan *vmixtcp.ActsResponse, defaultBufferSize),
		versionChan:   make(chan *vmixtcp.VersionResponse, defaultBufferSize),
		subscribeChan: make(chan *vmixtcp.SubscribeResponse, defaultBufferSize),
	}
}

func (cm *ConnectionManager) handleConnectionCleanup(ctx context.Context, conn *vMixConnection, addr string) {
	cm.logger.Debug(ctx, "Handling connection cleanup for %s", addr)
	defer cm.logger.Debug(ctx, "Connection cleanup completed for %s", addr)

	defer func() {
		if r := recover(); r != nil {
			cm.logger.Error(ctx, "PANIC in handleConnectionCleanup: %v\nStack Trace:\n%s",
				r, string(debug.Stack()))
		}
	}()

	if conn.retryCancel != nil {
		cm.logger.Debug(ctx, "Cancelling retry for %s", addr)
		conn.retryCancel()
		cm.logger.Debug(ctx, "Retry cancelled for %s", addr)
	}
}

func (cm *ConnectionManager) setupCallbacks(ctx context.Context, conn *vMixConnection) {
	// connがnilでないことを確認
	if conn == nil {
		cm.logger.Error(ctx, "setupCallbacks: conn is nil")
		return
	}

	// clientがnilでないことを確認
	if conn.client == nil {
		cm.logger.Error(ctx, "setupCallbacks: client is nil")
		return
	}

	conn.client.OnXML(func(resp *vmixtcp.XMLResponse) {
		select {
		case conn.xmlChan <- resp:
		default:
			cm.logger.Warn(ctx, "XML channel buffer full, dropping message")
		}
	})

	conn.client.OnTally(func(resp *vmixtcp.TallyResponse) {
		select {
		case conn.tallyChan <- resp:
		default:
			cm.logger.Warn(ctx, "Tally channel buffer full, dropping message")
		}
	})

	conn.client.OnActs(func(resp *vmixtcp.ActsResponse) {
		select {
		case conn.actsChan <- resp:
		default:
			cm.logger.Warn(ctx, "Acts channel buffer full, dropping message")
		}
	})

	conn.client.OnVersion(func(resp *vmixtcp.VersionResponse) {
		select {
		case conn.versionChan <- resp:
		default:
			cm.logger.Warn(ctx, "Version channel buffer full, dropping message")
		}
	})

	conn.client.OnSubscribe(func(resp *vmixtcp.SubscribeResponse) {
		select {
		case conn.subscribeChan <- resp:
		default:
			cm.logger.Warn(ctx, "Subscribe channel buffer full, dropping message")
		}
	})
}

func (cm *ConnectionManager) AddContext(ctx context.Context, vmixAddr string, contextID string, actionType string, isInitialization bool) {
	cm.logger.Debug(ctx, "[AddContext START] vmixAddr:%s contextID:%s actionType:%s isInit:%v", vmixAddr, contextID, actionType, isInitialization)
	startTime := time.Now()
	conn, _ := cm.connections.Load(vmixAddr) // 接続情報を事前に取得
	defer func() {
		if conn != nil {
			cm.logger.Debug(ctx, "[AddContext END] contextID:%s duration:%v remaining:%d",
				contextID, time.Since(startTime), conn.contexts.Size())
		} else {
			cm.logger.Debug(ctx, "[AddContext END] contextID:%s duration:%v (no connection)",
				contextID, time.Since(startTime))
		}
	}()

	defer func() {
		if r := recover(); r != nil {
			cm.logger.Error(ctx, "PANIC in AddContext: %v\nStack Trace:\n%s",
				r, string(debug.Stack()))
		}
	}()

	contextInfo := &ContextInfo{
		VMixAddr:   vmixAddr,
		ActionType: actionType,
	}
	cm.contextMap.Store(contextID, contextInfo)

	cm.initActionTypeMap(vmixAddr, actionType)
	if addrMap, exists := cm.actionTypeMap.LoadOrStore(vmixAddr, xsync.NewMapOf[string, *xsync.MapOf[string, struct{}]]()); exists {
		if actionMap, exists := addrMap.LoadOrStore(actionType, xsync.NewMapOf[string, struct{}]()); exists {
			actionMap.Store(contextID, struct{}{})
		}
	}

	cm.AddVMix(ctx, vmixAddr)

	if conn == nil {
		cm.logger.Error(ctx, "Connection is nil for %s after creation, contextID=%s", vmixAddr, contextID)
		return
	}

	conn.contexts.Store(contextID, struct{}{})
	cm.logger.Debug(ctx, "Context added: vmixAddr=%s, contextID=%s", vmixAddr, contextID)
}

func (cm *ConnectionManager) UpdateContext(ctx context.Context, oldVmixAddr, newVmixAddr string, contextID string, actionType string) {
	cm.logger.Debug(ctx, "UpdateContext: oldAddr=%s, newAddr=%s, contextID=%s, actionType=%s",
		oldVmixAddr, newVmixAddr, contextID, actionType)

	defer func() {
		if r := recover(); r != nil {
			cm.logger.Error(ctx, "PANIC in UpdateContext: %v\nStack Trace:\n%s",
				r, string(debug.Stack()))
		}
	}()

	// 古いアドレスと新しいアドレスが同じ場合は早期リターン
	if oldVmixAddr == newVmixAddr {
		cm.logger.Debug(ctx, "UpdateContext: oldAddr equals newAddr, no update needed")
		return
	}

	cm.RemoveContext(ctx, oldVmixAddr, contextID)

	// すでに接続が存在するか確認
	_, exists := cm.connections.Load(newVmixAddr)

	// 新しいアドレスに接続が既に存在する場合は、初期化フラグをfalseに
	// 存在しない場合は初期化フラグをtrueにして接続管理を開始
	isInitialization := !exists

	cm.AddContext(ctx, newVmixAddr, contextID, actionType, isInitialization)
	cm.logger.Debug(ctx, "UpdateContext completed: contextID=%s, newAddr=%s", contextID, newVmixAddr)
}

func (cm *ConnectionManager) RemoveContext(ctx context.Context, vmixAddr string, contextID string) {
	cm.logger.Debug(ctx, "RemoveContext: vmixAddr=%s, contextID=%s", vmixAddr, contextID)

	defer func() {
		if r := recover(); r != nil {
			cm.logger.Error(ctx, "PANIC in RemoveContext: %v\nStack Trace:\n%s",
				r, string(debug.Stack()))
		}
	}()

	// コンテキストをコンテキストマップから削除
	if info, exists := cm.contextMap.LoadAndDelete(contextID); exists {
		cm.logger.Debug(ctx, "Removed contextID=%s from contextMap (addr=%s, actionType=%s)",
			contextID, info.VMixAddr, info.ActionType)
	} else {
		cm.logger.Debug(ctx, "contextID=%s not found in contextMap", contextID)
	}

	// アクションタイプマップからコンテキストを削除
	addrMap, exists := cm.actionTypeMap.Load(vmixAddr)
	if exists {
		cm.logger.Debug(ctx, "Processing actionTypeMap for vmixAddr=%s", vmixAddr)
		addrMap.Range(func(actionType string, actionMap *xsync.MapOf[string, struct{}]) bool {
			if _, deleted := actionMap.LoadAndDelete(contextID); deleted {
				cm.logger.Debug(ctx, "Removed contextID=%s from actionType=%s", contextID, actionType)
			}
			return true
		})
	}

	// コネクションからコンテキストを削除
	conn, exists := cm.connections.Load(vmixAddr)
	if !exists {
		cm.logger.Debug(ctx, "Connection not found for vmixAddr=%s", vmixAddr)
		return
	}

	if conn == nil {
		cm.logger.Warn(ctx, "Connection is nil for vmixAddr=%s", vmixAddr)
		return
	}

	// この下のどこかがpanicに起因している
	conn.contexts.Delete(contextID)
	cm.logger.Debug(ctx, "Removed contextID=%s from connection.contexts, remaining=%d",
		contextID, conn.contexts.Size())

	// コンテキストがなくなった場合はvMixを削除
	if conn.contexts.Size() == 0 {
		cm.logger.Debug(ctx, "Removing vMix %s because it has no contexts", vmixAddr)
		// cm.RemoveVMix(ctx, vmixAddr) // ここがpanicの間接的な原因となっていそう
	}
}

func (cm *ConnectionManager) RemoveVMix(ctx context.Context, vmixAddr string) {
	cm.logger.Debug(ctx, "RemoveVMix: vmixAddr=%s", vmixAddr)
	defer func() {
		cm.logger.Debug(ctx, "RemoveVMix completed: vmixAddr=%s remaining connections: %d", vmixAddr, cm.connections.Size())
	}()

	defer func() {
		if r := recover(); r != nil {
			cm.logger.Error(ctx, "PANIC in RemoveVMix: %v\nStack Trace:\n%s",
				r, string(debug.Stack()))
		}
	}()

	// 接続を取得して削除（アトミックな操作）
	conn, exists := cm.connections.LoadAndDelete(vmixAddr)
	if !exists {
		cm.logger.Debug(ctx, "Connection not found for vmixAddr=%s, nothing to remove", vmixAddr)
		return
	}

	if conn == nil {
		cm.logger.Warn(ctx, "Connection is nil for vmixAddr=%s", vmixAddr)
		return
	}

	// 接続に関連するコンテキストをログに出力
	contextCount := conn.contexts.Size()
	cm.logger.Debug(ctx, "Found %d context(s) for vmixAddr=%s before cleanup", contextCount, vmixAddr)

	// 接続のクリーンアップ処理
	cm.handleConnectionCleanup(ctx, conn, vmixAddr) // ここの先の処理でpanic
}

func (cm *ConnectionManager) AddVMix(ctx context.Context, vmixAddr string) {
	if strings.TrimSpace(vmixAddr) == "" {
		return
	}

	// 既存の接続を取得
	conn, exists := cm.connections.Load(vmixAddr)
	if exists && conn != nil {
		// 接続が存在し、かつ有効な場合は何もしない
		return
	}

	// 新しい接続を作成
	conn = cm.newVMixConnection()

	// 接続を保存（既存の接続がある場合は上書き）
	cm.connections.Store(vmixAddr, conn)

	// 接続管理を開始
	go cm.manageConnection(ctx, vmixAddr, conn)
}

func (cm *ConnectionManager) GetVMixByContext(ctx context.Context, contextID string) vmixtcp.Vmix {
	var client vmixtcp.Vmix
	contextInfo, exists := cm.contextMap.Load(contextID)
	if !exists {
		return nil
	}

	conn, exists := cm.connections.Load(contextInfo.VMixAddr)
	if !exists {
		return nil
	}

	client = conn.client

	return client
}

type VMixConnection struct {
	Address   string
	Connected bool
}

func (cm *ConnectionManager) GetAllVMixAddrs(ctx context.Context) []VMixConnection {
	addrs := make([]VMixConnection, 0, cm.connections.Size())
	cm.connections.Range(func(key string, value *vMixConnection) bool {
		isConnected := false
		if value.client != nil {
			isConnected = value.client.IsConnected()
		}
		addrs = append(addrs, VMixConnection{
			Address:   key,
			Connected: isConnected,
		})
		return true
	})
	return addrs
}

func (cm *ConnectionManager) manageConnection(parentCtx context.Context, addr string, conn *vMixConnection) {
	ctx, cancel := context.WithCancel(parentCtx)
	conn.retryCancel = cancel

	cm.logger.Debug(ctx, "Starting connection management for %s", addr)

	defer func() {
		if r := recover(); r != nil {
			cm.logger.Error(ctx, "PANIC in manageConnection: %v\nStack Trace:\n%s",
				r, string(debug.Stack()))
		}
		cm.logger.Debug(ctx, "Connection management ended for %s", addr)
	}()

	// メッセージハンドラーを開始
	messageDone := make(chan struct{})
	go func() {
		// cm.handleAllMessages(ctx, conn, addr)
		close(messageDone)
	}()

	const (
		connectionCheckInterval = 10 * time.Second
		stateCheckInterval      = 5 * time.Second
	)

	// 初回接続用のチャネル
	immediate := make(chan struct{}, 1)
	immediate <- struct{}{}

	// 接続試行中フラグ (重複接続試行防止用)
	var connecting bool

	// 接続処理を行う関数
	connectToVMix := func() bool {
		if ctx.Err() != nil {
			cm.logger.Debug(ctx, "Context already canceled, skipping connection attempt")
			return false
		}

		if conn.client != nil {
			if conn.client.IsConnected() {
				return true
			}
		}

		// すでに接続試行中なら何もしない
		if connecting {
			cm.logger.Debug(ctx, "Connection attempt already in progress for %s", addr)
			return false
		}

		connecting = true
		defer func() { connecting = false }()

		cm.logger.Info(ctx, "Connecting to vmix: %s", addr)
		client := vmixtcp.New(addr)
		if err := client.Connect(ctx, 5*time.Second); err != nil {
			cm.logger.Warn(ctx, "Failed to connect to vmix: %v", err)
			return false
		}

		cm.setupCallbacks(ctx, conn)
		cm.logger.Info(ctx, "Connected to vmix: %s", addr)

		conn.client = client

		// クライアント実行
		go func() {
			defer func() {
				if r := recover(); r != nil {
					cm.logger.Error(ctx, "PANIC in vmix run: %v\nStack Trace:\n%s",
						r, string(debug.Stack()))
				}
			}()
			cm.logger.Debug(ctx, "Running vmix for %s", addr)
			if client == nil {
				cm.logger.Error(ctx, "Client is nil for %s", addr)
				return
			}
			// panic発生個所疑惑
			if err := client.Run(ctx); err != nil {
				cm.logger.Error(ctx, "Failed to run vmix: %v", err)
				if !errors.Is(err, vmixtcp.ErrDisconnected) {
					if client == nil {
						cm.logger.Error(ctx, "Client is nil for %s", addr)
						return
					} else {
						if err := client.Close(); err != nil {
							cm.logger.Error(ctx, "Failed to close vmix: %v", err)
						}
					}
				}
			}
		}()
		return true
	}

	// メインループ
	ticker := time.NewTicker(connectionCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			cm.logger.Debug(ctx, "Context canceled: %s", ctx.Err())
			return

		case <-immediate:
			connectToVMix()

		case <-ticker.C:
			connectToVMix()

		case <-messageDone:
			cm.logger.Debug(ctx, "Message handler ended for %s", addr)
			return
		}
	}
}

func (cm *ConnectionManager) handleAllMessages(ctx context.Context, conn *vMixConnection, addr string) {
	defer func() {
		if r := recover(); r != nil {
			cm.logger.Error(ctx, "PANIC in handleAllMessages: %v\nStack Trace:\n%s",
				r, string(debug.Stack()))
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case resp := <-conn.xmlChan:
			cm.xmlCallback(resp, conn.client, addr)
		case resp := <-conn.tallyChan:
			cm.tallyCallback(resp, conn.client, addr)
		case resp := <-conn.actsChan:
			cm.actsCallback(resp, conn.client, addr)
		case resp := <-conn.versionChan:
			cm.versionCallback(resp, conn.client, addr)
		case resp := <-conn.subscribeChan:
			cm.subscribeCallback(resp, conn.client, addr)
		}
	}
}

func (cm *ConnectionManager) GetClient(ctx context.Context, vmixAddr string) vmixtcp.Vmix {
	var client vmixtcp.Vmix
	conn, exists := cm.connections.Load(vmixAddr)
	if !exists {
		return nil
	}

	client = conn.client

	return client
}

func (cm *ConnectionManager) GetContexts(ctx context.Context, vmixAddr string) []string {
	conn, exists := cm.connections.Load(vmixAddr)
	if !exists {
		return nil
	}

	contexts := make([]string, 0, conn.contexts.Size())
	conn.contexts.Range(func(key string, value struct{}) bool {
		contexts = append(contexts, key)
		return true
	})

	return contexts
}

func (cm *ConnectionManager) GetContextsByActionType(ctx context.Context, vmixAddr string, actionType string) []string {
	contexts := make([]string, 0, cm.contextMap.Size())
	for _, contextID := range cm.Contexts() {
		contextInfo, exists := cm.contextMap.Load(contextID)
		if !exists {
			continue
		}

		if contextInfo.VMixAddr != vmixAddr {
			continue
		}

		if contextInfo.ActionType != actionType {
			continue
		}
		contexts = append(contexts, contextID)
	}

	return contexts
}

func (cm *ConnectionManager) SetXMLCallback(callback func(*vmixtcp.XMLResponse, vmixtcp.Vmix, string)) {
	cm.xmlCallback = callback
}

func (cm *ConnectionManager) SetTallyCallback(callback func(*vmixtcp.TallyResponse, vmixtcp.Vmix, string)) {
	cm.tallyCallback = callback
}

func (cm *ConnectionManager) SetActsCallback(callback func(*vmixtcp.ActsResponse, vmixtcp.Vmix, string)) {
	cm.actsCallback = callback
}

func (cm *ConnectionManager) SetVersionCallback(callback func(*vmixtcp.VersionResponse, vmixtcp.Vmix, string)) {
	cm.versionCallback = callback
}

func (cm *ConnectionManager) SetSubscribeCallback(callback func(*vmixtcp.SubscribeResponse, vmixtcp.Vmix, string)) {
	cm.subscribeCallback = callback
}

func (cm *ConnectionManager) Contexts() []string {
	ctxs := make([]string, 0, cm.contextMap.Size())
	cm.contextMap.Range(func(key string, value *ContextInfo) bool {
		ctxs = append(ctxs, key)
		return true
	})
	return ctxs
}
