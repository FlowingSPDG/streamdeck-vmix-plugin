package connection

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
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
	xmlCallback   func(*vmixtcp.XMLResponse, vmixtcp.Vmix, string)
	tallyCallback func(*vmixtcp.TallyResponse, vmixtcp.Vmix, string)
	actsCallback  func(*vmixtcp.ActsResponse, vmixtcp.Vmix, string)
}

type vMixConnection struct {
	mu          sync.RWMutex
	client      vmixtcp.Vmix
	contexts    map[string]struct{}
	retryCancel context.CancelFunc

	// チャネル for vMix callbacks
	xmlChan   chan *vmixtcp.XMLResponse
	tallyChan chan *vmixtcp.TallyResponse
	actsChan  chan *vmixtcp.ActsResponse
}

func NewConnectionManager(logger logger.Logger) *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*vMixConnection),
		contextMap:  make(map[string]string),
		logger:      logger,

		xmlCallback:   func(*vmixtcp.XMLResponse, vmixtcp.Vmix, string) {},
		tallyCallback: func(*vmixtcp.TallyResponse, vmixtcp.Vmix, string) {},
		actsCallback:  func(*vmixtcp.ActsResponse, vmixtcp.Vmix, string) {},
	}
}

func (cm *ConnectionManager) logMethodEntry(ctx context.Context, method string, args ...interface{}) {
	cm.logger.LogMessage(ctx, "ENTER %s with args: %v", method, args)
}

func (cm *ConnectionManager) logMethodExit(ctx context.Context, method string) {
	cm.logger.LogMessage(ctx, "EXIT %s", method)
}

func (cm *ConnectionManager) safeCall(ctx context.Context, fn func(), operation string) {
	defer func() {
		if r := recover(); r != nil {
			cm.logger.Error(ctx, "PANIC in %s: %v\nStack Trace:\n%s",
				operation, r, string(debug.Stack()))
		}
	}()
	fn()
}

func (cm *ConnectionManager) newVMixConnection() *vMixConnection {
	return &vMixConnection{
		contexts:  make(map[string]struct{}),
		xmlChan:   make(chan *vmixtcp.XMLResponse, 100),
		tallyChan: make(chan *vmixtcp.TallyResponse, 100),
		actsChan:  make(chan *vmixtcp.ActsResponse, 100),
	}
}

func (cm *ConnectionManager) handleConnectionCleanup(ctx context.Context, conn *vMixConnection, addr string) bool {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	isEmpty := len(conn.contexts) == 0
	if isEmpty {
		if conn.retryCancel != nil {
			cm.logger.LogMessage(ctx, "Cancelling retry connection for %s", addr)
			conn.retryCancel()
		}
		if conn.client != nil {
			conn.client.Close()
			conn.client = nil
		}
	}
	return isEmpty
}

func (cm *ConnectionManager) setupCallbacks(ctx context.Context, client vmixtcp.Vmix, conn *vMixConnection) {
	client.OnXML(func(resp *vmixtcp.XMLResponse, err error) {
		if err != nil {
			cm.logger.Error(ctx, "XML callback error: %v", err)
			return
		}
		conn.xmlChan <- resp
	})

	client.OnTally(func(resp *vmixtcp.TallyResponse, err error) {
		if err != nil {
			cm.logger.Error(ctx, "Tally callback error: %v", err)
			return
		}
		conn.tallyChan <- resp
	})

	client.OnActs(func(resp *vmixtcp.ActsResponse, err error) {
		if err != nil {
			cm.logger.Error(ctx, "Acts callback error: %v", err)
			return
		}
		conn.actsChan <- resp
	})
}

func (cm *ConnectionManager) AddContext(ctx context.Context, vmixAddr string, contextID string) {
	cm.logMethodEntry(ctx, "AddContext", fmt.Sprintf("vmixAddr: %s, contextID: %s", vmixAddr, contextID))
	defer cm.logMethodExit(ctx, "AddContext")

	cm.safeCall(ctx, func() {
		cm.mu.Lock()
		defer cm.mu.Unlock()

		conn, exists := cm.connections[vmixAddr]
		if !exists {
			conn = cm.newVMixConnection()
			cm.connections[vmixAddr] = conn
			go cm.manageConnection(ctx, vmixAddr, conn)
		}

		conn.mu.Lock()
		defer conn.mu.Unlock()
		conn.contexts[contextID] = struct{}{}
		cm.contextMap[contextID] = vmixAddr
	}, "AddContext")
}

func (cm *ConnectionManager) UpdateContext(ctx context.Context, vmixAddr string, contextID string) {
	cm.logMethodEntry(ctx, "UpdateContext", fmt.Sprintf("vmixAddr: %s, contextID: %s", vmixAddr, contextID))
	defer cm.logMethodExit(ctx, "UpdateContext")

	cm.safeCall(ctx, func() {
		cm.mu.Lock()
		defer cm.mu.Unlock()

		// Remove from old connection
		cm.RemoveContext(ctx, vmixAddr, contextID)

		// Add to new connection
		cm.AddContext(ctx, vmixAddr, contextID)
	}, "UpdateContext")
}

func (cm *ConnectionManager) RemoveContext(ctx context.Context, vmixAddr string, contextID string) {
	cm.logMethodEntry(ctx, "RemoveContext", fmt.Sprintf("vmixAddr: %s, contextID: %s", vmixAddr, contextID))
	defer cm.logMethodExit(ctx, "RemoveContext")

	cm.safeCall(ctx, func() {
		cm.mu.Lock()
		defer cm.mu.Unlock()

		conn, exists := cm.connections[vmixAddr]
		if !exists {
			return
		}

		conn.mu.Lock()
		delete(conn.contexts, contextID)
		conn.mu.Unlock()

		if cm.handleConnectionCleanup(ctx, conn, vmixAddr) {
			delete(cm.connections, vmixAddr)
		}

		delete(cm.contextMap, contextID)
	}, "RemoveContext")
}

func (cm *ConnectionManager) GetVMixByContext(ctx context.Context, contextID string) vmixtcp.Vmix {
	cm.logMethodEntry(ctx, "GetVMixByContext", fmt.Sprintf("contextID: %s", contextID))
	defer cm.logMethodExit(ctx, "GetVMixByContext")

	var client vmixtcp.Vmix
	cm.safeCall(ctx, func() {
		cm.mu.RLock()
		defer cm.mu.RUnlock()

		vmixAddr, exists := cm.contextMap[contextID]
		if !exists {
			return
		}

		if conn, exists := cm.connections[vmixAddr]; exists {
			conn.mu.RLock()
			defer conn.mu.RUnlock()
			client = conn.client
		}
	}, "GetVMixByContext")
	return client
}

func (cm *ConnectionManager) manageConnection(parentCtx context.Context, addr string, conn *vMixConnection) {
	cm.logMethodEntry(parentCtx, "manageConnection", fmt.Sprintf("addr: %s", addr))
	defer cm.logMethodExit(parentCtx, "manageConnection")

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

	// Start callback handlers
	go cm.handleXMLMessages(ctx, conn, addr)
	go cm.handleTallyMessages(ctx, conn, addr)
	go cm.handleACTSMessages(ctx, conn, addr)

	for {
		select {
		case <-ctx.Done():
			cm.logger.LogMessage(ctx, "Context canceled. %s", ctx.Err())
			return
		case <-time.After(5 * time.Second):
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

			cm.logger.LogMessage(ctx, "Connecting to vmix: %s", addr)
			client := vmixtcp.New(addr)
			if err := client.Connect(ctx, 5*time.Second); err != nil {
				cm.logger.Error(ctx, "Failed to connect to vmix: %v", err)
				continue
			}

			cm.setupCallbacks(ctx, client, conn)
			cm.logger.LogMessage(ctx, "Connected to vmix: %s", addr)

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

			go func() {
				ticker := time.NewTicker(time.Second)
				defer ticker.Stop()
				defer client.Close()

				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						if err := client.Tally(); err != nil {
							conn.mu.Lock()
							conn.client = nil
							conn.mu.Unlock()
							return
						}
					}
				}
			}()
		}
	}
}

func (cm *ConnectionManager) handleXMLMessages(ctx context.Context, conn *vMixConnection, addr string) {
	cm.logMethodEntry(ctx, "handleXMLMessages")
	defer cm.logMethodExit(ctx, "handleXMLMessages")

	for {
		select {
		case <-ctx.Done():
			return
		case resp := <-conn.xmlChan:
			cm.logger.LogMessage(ctx, "Received XML response: %+v", resp)
			cm.xmlCallback(resp, conn.client, addr)
		}
	}
}

func (cm *ConnectionManager) handleTallyMessages(ctx context.Context, conn *vMixConnection, addr string) {
	cm.logMethodEntry(ctx, "handleTallyMessages")
	defer cm.logMethodExit(ctx, "handleTallyMessages")

	for {
		select {
		case <-ctx.Done():
			return
		case resp := <-conn.tallyChan:
			cm.logger.LogMessage(ctx, "Received Tally response: %+v", resp)
			cm.tallyCallback(resp, conn.client, addr)
		}
	}
}

func (cm *ConnectionManager) handleACTSMessages(ctx context.Context, conn *vMixConnection, addr string) {
	cm.logMethodEntry(ctx, "handleACTSMessages")
	defer cm.logMethodExit(ctx, "handleACTSMessages")

	for {
		select {
		case <-ctx.Done():
			return
		case resp := <-conn.actsChan:
			cm.logger.LogMessage(ctx, "Received Acts response: %+v", resp)
			cm.actsCallback(resp, conn.client, addr)
		}
	}
}

func (cm *ConnectionManager) GetClient(ctx context.Context, vmixAddr string) vmixtcp.Vmix {
	cm.logMethodEntry(ctx, "GetClient", fmt.Sprintf("vmixAddr: %s", vmixAddr))
	defer cm.logMethodExit(ctx, "GetClient")

	var client vmixtcp.Vmix
	cm.safeCall(ctx, func() {
		cm.mu.RLock()
		defer cm.mu.RUnlock()

		if conn, exists := cm.connections[vmixAddr]; exists {
			conn.mu.RLock()
			defer conn.mu.RUnlock()
			client = conn.client
		}
	}, "GetClient")
	return client
}

func (cm *ConnectionManager) GetContexts(ctx context.Context, vmixAddr string) []string {
	cm.logMethodEntry(ctx, "GetContexts", fmt.Sprintf("vmixAddr: %s", vmixAddr))
	defer cm.logMethodExit(ctx, "GetContexts")

	var contexts []string
	cm.safeCall(ctx, func() {
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
	}, "GetContexts")
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

func (cm *ConnectionManager) Contexts() []string {
	ctxs := make([]string, 0, len(cm.contextMap))
	for ctx := range cm.contextMap {
		ctxs = append(ctxs, ctx)
	}
	return ctxs
}
