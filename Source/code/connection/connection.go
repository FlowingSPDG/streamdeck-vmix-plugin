package connection

import (
	"context"
	"errors"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
)

type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]*vMixConnection
	contextMap  map[string]string // contextID -> vmixAddr
	logger      logger.Logger

	// callbacks. string is vMixAddr.
	xmlCallback       func(*vmixtcp.XMLResponse, vmixtcp.Vmix, string)
	tallyCallback     func(*vmixtcp.TallyResponse, vmixtcp.Vmix, string)
	actsCallback      func(*vmixtcp.ActsResponse, vmixtcp.Vmix, string)
	versionCallback   func(*vmixtcp.VersionResponse, vmixtcp.Vmix, string)
	subscribeCallback func(*vmixtcp.SubscribeResponse, vmixtcp.Vmix, string)
}

type vMixConnection struct {
	mu          sync.RWMutex
	client      vmixtcp.Vmix
	contexts    map[string]struct{}
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
		connections: make(map[string]*vMixConnection),
		contextMap:  make(map[string]string),
		logger:      logger,

		xmlCallback:       func(*vmixtcp.XMLResponse, vmixtcp.Vmix, string) {},
		tallyCallback:     func(*vmixtcp.TallyResponse, vmixtcp.Vmix, string) {},
		actsCallback:      func(*vmixtcp.ActsResponse, vmixtcp.Vmix, string) {},
		versionCallback:   func(*vmixtcp.VersionResponse, vmixtcp.Vmix, string) {},
		subscribeCallback: func(*vmixtcp.SubscribeResponse, vmixtcp.Vmix, string) {},
	}
}

const (
	// チャネルバッファサイズの最適化
	defaultBufferSize = 50
)

func (cm *ConnectionManager) newVMixConnection() *vMixConnection {
	return &vMixConnection{
		contexts:      make(map[string]struct{}),
		xmlChan:       make(chan *vmixtcp.XMLResponse, defaultBufferSize),
		tallyChan:     make(chan *vmixtcp.TallyResponse, defaultBufferSize),
		actsChan:      make(chan *vmixtcp.ActsResponse, defaultBufferSize),
		versionChan:   make(chan *vmixtcp.VersionResponse, defaultBufferSize),
		subscribeChan: make(chan *vmixtcp.SubscribeResponse, defaultBufferSize),
	}
}

func (cm *ConnectionManager) handleConnectionCleanup(ctx context.Context, conn *vMixConnection, addr string) {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.retryCancel != nil {
		cm.logger.Debug(ctx, "Cancelling retry connection for %s", addr)
		conn.retryCancel()
	}
	if conn.client != nil {
		conn.client.Close()
		conn.client = nil
	}
}

func (cm *ConnectionManager) setupCallbacks(ctx context.Context, client vmixtcp.Vmix, conn *vMixConnection) {
	// XMLコールバック
	client.OnXML(func(resp *vmixtcp.XMLResponse, err error) {
		if err != nil {
			cm.logger.Error(ctx, "XML callback error: %v", err)
			return
		}
		select {
		case conn.xmlChan <- resp:
		default:
			cm.logger.Warn(ctx, "XML channel buffer full, dropping message")
		}
	})

	// Tallyコールバック
	client.OnTally(func(resp *vmixtcp.TallyResponse, err error) {
		if err != nil {
			cm.logger.Error(ctx, "Tally callback error: %v", err)
			return
		}
		select {
		case conn.tallyChan <- resp:
		default:
			cm.logger.Warn(ctx, "Tally channel buffer full, dropping message")
		}
	})

	// Actsコールバック
	client.OnActs(func(resp *vmixtcp.ActsResponse, err error) {
		if err != nil {
			cm.logger.Error(ctx, "Acts callback error: %v", err)
			return
		}
		select {
		case conn.actsChan <- resp:
		default:
			cm.logger.Warn(ctx, "Acts channel buffer full, dropping message")
		}
	})

	// Versionコールバック
	client.OnVersion(func(resp *vmixtcp.VersionResponse, err error) {
		if err != nil {
			cm.logger.Error(ctx, "Version callback error: %v", err)
			return
		}
		select {
		case conn.versionChan <- resp:
		default:
			cm.logger.Warn(ctx, "Version channel buffer full, dropping message")
		}
	})

	// Subscribeコールバック
	client.OnSubscribe(func(resp *vmixtcp.SubscribeResponse, err error) {
		if err != nil {
			cm.logger.Error(ctx, "Subscribe callback error: %v", err)
			return
		}
		select {
		case conn.subscribeChan <- resp:
		default:
			cm.logger.Warn(ctx, "Subscribe channel buffer full, dropping message")
		}
	})
}

func (cm *ConnectionManager) AddContext(ctx context.Context, vmixAddr string, contextID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.contextMap[contextID] = vmixAddr

	conn, exists := cm.connections[vmixAddr]
	if !exists {
		conn = cm.newVMixConnection()
		cm.connections[vmixAddr] = conn
		go cm.manageConnection(ctx, vmixAddr, conn)
	}

	conn.mu.Lock()
	defer conn.mu.Unlock()
	conn.contexts[contextID] = struct{}{}
}

func (cm *ConnectionManager) UpdateContext(ctx context.Context, vmixAddr string, contextID string) {
	// Remove from old connection
	cm.RemoveContext(ctx, vmixAddr, contextID)

	// Add to new connection
	cm.AddContext(ctx, vmixAddr, contextID)
}

func (cm *ConnectionManager) RemoveContext(ctx context.Context, vmixAddr string, contextID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.contextMap, contextID)

	conn, exists := cm.connections[vmixAddr]
	if !exists {
		return
	}

	conn.mu.Lock()
	defer conn.mu.Unlock()
	delete(conn.contexts, contextID)
}

func (cm *ConnectionManager) RemoveVMix(ctx context.Context, vmixAddr string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.handleConnectionCleanup(ctx, cm.connections[vmixAddr], vmixAddr)
	delete(cm.connections, vmixAddr)
}

func (cm *ConnectionManager) AddVMix(ctx context.Context, vmixAddr string) {
	if strings.TrimSpace(vmixAddr) == "" {
		return
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.connections[vmixAddr]; exists {
		return
	}

	conn := cm.newVMixConnection()
	cm.connections[vmixAddr] = conn
	go cm.manageConnection(ctx, vmixAddr, conn)
}

func (cm *ConnectionManager) GetVMixByContext(ctx context.Context, contextID string) vmixtcp.Vmix {
	var client vmixtcp.Vmix
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	vmixAddr, exists := cm.contextMap[contextID]
	if !exists {
		return nil
	}

	if conn, exists := cm.connections[vmixAddr]; exists {
		conn.mu.RLock()
		defer conn.mu.RUnlock()
		client = conn.client
	}
	return client
}

func (cm *ConnectionManager) GetAllVMixAddrs(ctx context.Context) []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	addrs := make([]string, 0, len(cm.connections))
	for addr := range cm.connections {
		addrs = append(addrs, addr)
	}
	return addrs
}

func (cm *ConnectionManager) manageConnection(parentCtx context.Context, addr string, conn *vMixConnection) {
	ctx, cancel := context.WithCancel(parentCtx)
	conn.mu.Lock()
	conn.retryCancel = cancel
	conn.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			cm.logger.Error(ctx, "PANIC in manageConnection: %v\nStack Trace:\n%s",
				r, string(debug.Stack()))
		}
		cancel()
	}()

	// 単一のメッセージハンドラーを開始
	go cm.handleAllMessages(ctx, conn, addr)

	const (
		connectionCheckInterval = 10 * time.Second
		stateCheckInterval      = 5 * time.Second
	)

	for {
		select {
		case <-ctx.Done():
			cm.logger.Debug(ctx, "Context canceled. %s", ctx.Err())
			return
		case <-time.After(connectionCheckInterval):
			conn.mu.RLock()
			if conn.client != nil {
				if conn.client.IsConnected() {
					conn.mu.RUnlock()
					continue
				}
				conn.mu.RUnlock()
				continue
			}
			conn.mu.RUnlock()

			cm.logger.Info(ctx, "Connecting to vmix: %s", addr)
			client := vmixtcp.New(addr)
			if err := client.Connect(ctx, 5*time.Second); err != nil {
				cm.logger.Warn(ctx, "Failed to connect to vmix: %v", err)
				continue
			}

			cm.setupCallbacks(ctx, client, conn)
			cm.logger.Info(ctx, "Connected to vmix: %s", addr)

			conn.mu.Lock()
			conn.client = client
			conn.mu.Unlock()

			go func() {
				if err := client.Run(ctx); err != nil {
					cm.logger.Error(ctx, "Failed to run vmix: %v", err)
					if !errors.Is(err, vmixtcp.ErrDisconnected) {
						client.Close()
					}
					conn.mu.Lock()
					conn.client = nil
					conn.mu.Unlock()
				}
			}()

			// 状態監視の最適化
			go cm.monitorConnectionState(ctx, conn, client, stateCheckInterval)
		}
	}
}

// 統合されたメッセージハンドラー
func (cm *ConnectionManager) handleAllMessages(ctx context.Context, conn *vMixConnection, addr string) {
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

// 最適化された接続状態監視
func (cm *ConnectionManager) monitorConnectionState(ctx context.Context, conn *vMixConnection, client vmixtcp.Vmix, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer client.Close()

	var lastState bool
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			currentState := client.IsConnected()
			if currentState != lastState {
				lastState = currentState
				if !currentState {
					conn.mu.Lock()
					conn.client = nil
					conn.mu.Unlock()
					return
				}
			}
		}
	}
}

func (cm *ConnectionManager) GetClient(ctx context.Context, vmixAddr string) vmixtcp.Vmix {
	var client vmixtcp.Vmix
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if conn, exists := cm.connections[vmixAddr]; exists {
		conn.mu.RLock()
		defer conn.mu.RUnlock()
		client = conn.client
	}
	return client
}

func (cm *ConnectionManager) GetContexts(ctx context.Context, vmixAddr string) []string {
	var contexts []string
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if conn, exists := cm.connections[vmixAddr]; exists {
		conn.mu.RLock()
		defer conn.mu.RUnlock()
		contexts = make([]string, 0, len(conn.contexts))
		for ctx := range conn.contexts {
			contexts = append(contexts, ctx)
		}
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
	ctxs := make([]string, 0, len(cm.contextMap))
	for ctx := range cm.contextMap {
		ctxs = append(ctxs, ctx)
	}
	return ctxs
}
