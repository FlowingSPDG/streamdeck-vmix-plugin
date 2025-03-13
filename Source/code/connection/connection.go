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
		retryCancel:   nil,
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

func (cm *ConnectionManager) setupCallbacks(ctx context.Context, client vmixtcp.Vmix, conn *vMixConnection) {
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

func (cm *ConnectionManager) AddContext(ctx context.Context, vmixAddr string, contextID string, actionType string, isInitialization bool) {
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

	conn, exists := cm.connections.Load(vmixAddr)
	if !exists && isInitialization {
		conn = cm.newVMixConnection()
		cm.connections.Store(vmixAddr, conn)
		go cm.manageConnection(ctx, vmixAddr, conn)
	}

	conn.contexts.Store(contextID, struct{}{})
}

func (cm *ConnectionManager) UpdateContext(ctx context.Context, oldVmixAddr, newVmixAddr string, contextID string, actionType string) {
	cm.RemoveContext(ctx, oldVmixAddr, contextID)
	cm.AddContext(ctx, newVmixAddr, contextID, actionType, false)
}

func (cm *ConnectionManager) RemoveContext(ctx context.Context, vmixAddr string, contextID string) {
	cm.contextMap.Delete(contextID)

	addrMap, exists := cm.actionTypeMap.Load(vmixAddr)
	if exists {
		addrMap.Range(func(actionType string, actionMap *xsync.MapOf[string, struct{}]) bool {
			actionMap.Delete(contextID)
			return true
		})
	}

	conn, exists := cm.connections.Load(vmixAddr)
	if exists {
		conn.contexts.Delete(contextID)
	}
}

func (cm *ConnectionManager) RemoveVMix(ctx context.Context, vmixAddr string) {
	conn, exists := cm.connections.Load(vmixAddr)
	if !exists {
		return
	}

	cm.handleConnectionCleanup(ctx, conn, vmixAddr)
	cm.connections.Delete(vmixAddr)
}

func (cm *ConnectionManager) AddVMix(ctx context.Context, vmixAddr string) {
	if strings.TrimSpace(vmixAddr) == "" {
		return
	}

	conn, exists := cm.connections.Load(vmixAddr)
	if exists {
		return
	}

	conn = cm.newVMixConnection()
	cm.connections.Store(vmixAddr, conn)
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

	defer func() {
		if r := recover(); r != nil {
			cm.logger.Error(ctx, "PANIC in manageConnection: %v\nStack Trace:\n%s",
				r, string(debug.Stack()))
		}
		cancel()
	}()

	go cm.handleAllMessages(ctx, conn, addr)

	const (
		connectionCheckInterval = 10 * time.Second
		stateCheckInterval      = 5 * time.Second
	)

	immediate := make(chan struct{}, 1)
	immediate <- struct{}{}

	for {
		select {
		case <-ctx.Done():
			cm.logger.Debug(ctx, "Context canceled. %s", ctx.Err())
			return
		case <-immediate:
			if conn.client != nil {
				if conn.client.IsConnected() {
					continue
				}
				continue
			}

			cm.logger.Info(ctx, "Connecting to vmix: %s", addr)
			client := vmixtcp.New(addr)
			if err := client.Connect(ctx, 5*time.Second); err != nil {
				cm.logger.Warn(ctx, "Failed to connect to vmix: %v", err)
				continue
			}

			cm.setupCallbacks(ctx, client, conn)
			cm.logger.Info(ctx, "Connected to vmix: %s", addr)

			conn.client = client

			go func() {
				defer func() {
					if r := recover(); r != nil {
						cm.logger.Error(ctx, "PANIC in vmix run: %v\nStack Trace:\n%s",
							r, string(debug.Stack()))
					}
				}()
				cm.logger.Debug(ctx, "Running vmix for %s", addr)
				if err := client.Run(ctx); err != nil {
					cm.logger.Error(ctx, "Failed to run vmix: %v", err)
					if !errors.Is(err, vmixtcp.ErrDisconnected) {
						client.Close()
					}
					conn.client = nil
				}
			}()

			go cm.monitorConnectionState(ctx, conn, client, stateCheckInterval)
		case <-time.After(connectionCheckInterval):
			if conn.client != nil {
				if conn.client.IsConnected() {
					continue
				}
				continue
			}

			cm.logger.Info(ctx, "Connecting to vmix: %s", addr)
			client := vmixtcp.New(addr)
			if err := client.Connect(ctx, 5*time.Second); err != nil {
				cm.logger.Warn(ctx, "Failed to connect to vmix: %v", err)
				continue
			}

			cm.setupCallbacks(ctx, client, conn)
			cm.logger.Info(ctx, "Connected to vmix: %s", addr)

			conn.client = client

			go func() {
				defer func() {
					if r := recover(); r != nil {
						cm.logger.Error(ctx, "PANIC in vmix run: %v\nStack Trace:\n%s",
							r, string(debug.Stack()))
					}
				}()
				cm.logger.Debug(ctx, "Running vmix for %s", addr)
				if err := client.Run(ctx); err != nil {
					cm.logger.Error(ctx, "Failed to run vmix: %v", err)
					if !errors.Is(err, vmixtcp.ErrDisconnected) {
						client.Close()
					}
					conn.client = nil
				}
			}()

			go cm.monitorConnectionState(ctx, conn, client, stateCheckInterval)
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

func (cm *ConnectionManager) monitorConnectionState(ctx context.Context, conn *vMixConnection, client vmixtcp.Vmix, interval time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			cm.logger.Error(ctx, "PANIC in monitorConnectionState: %v\nStack Trace:\n%s",
				r, string(debug.Stack()))
		}
	}()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer client.Close()

	var lastState bool
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if client == nil {
				return
			}
			currentState := client.IsConnected()
			if currentState == lastState {
				continue
			}
			cm.logger.Debug(ctx, "Connection state changed to %v", currentState)
			lastState = currentState
			if !currentState {
				conn.client = nil
				return
			}
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
