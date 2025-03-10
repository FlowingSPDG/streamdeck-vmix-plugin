package connection

import (
	"context"
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
}

type vMixConnection struct {
	mu          sync.RWMutex
	client      vmixtcp.Vmix
	contexts    map[string]struct{}
	retryCancel context.CancelFunc
	active      bool
}

func NewConnectionManager(logger logger.Logger) *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*vMixConnection),
		contextMap:  make(map[string]string),
		logger:      logger,
	}
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

func (cm *ConnectionManager) AddContext(ctx context.Context, vmixAddr string, contextID string) {
	cm.safeCall(ctx, func() {
		cm.logger.Info(ctx, "AddContext started. vmixAddr: %s, contextID: %s", vmixAddr, contextID)
		defer cm.logger.Info(ctx, "AddContext completed. vmixAddr: %s, contextID: %s", vmixAddr, contextID)

		cm.mu.Lock()
		defer cm.mu.Unlock()

		conn, exists := cm.connections[vmixAddr]
		if !exists {
			conn = &vMixConnection{
				contexts: make(map[string]struct{}),
				active:   true,
			}
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
	cm.safeCall(ctx, func() {
		cm.logger.Info(ctx, "UpdateContext started. vmixAddr: %s, contextID: %s", vmixAddr, contextID)
		defer cm.logger.Info(ctx, "UpdateContext completed. vmixAddr: %s, contextID: %s", vmixAddr, contextID)

		cm.mu.Lock()
		defer cm.mu.Unlock()

		// Remove from old connection if exists
		if oldAddr, exists := cm.contextMap[contextID]; exists {
			if oldConn, connExists := cm.connections[oldAddr]; connExists {
				oldConn.mu.Lock()
				delete(oldConn.contexts, contextID)
				oldConn.mu.Unlock()

				if len(oldConn.contexts) == 0 {
					oldConn.mu.Lock()
					if oldConn.retryCancel != nil {
						oldConn.retryCancel()
					}
					oldConn.active = false
					oldConn.mu.Unlock()
					delete(cm.connections, oldAddr)
				}
			}
			delete(cm.contextMap, contextID)
		}

		// Add to new connection
		conn, exists := cm.connections[vmixAddr]
		if !exists {
			conn = &vMixConnection{
				contexts: make(map[string]struct{}),
				active:   true,
			}
			cm.connections[vmixAddr] = conn
			go cm.manageConnection(ctx, vmixAddr, conn)
		}

		conn.mu.Lock()
		defer conn.mu.Unlock()
		conn.contexts[contextID] = struct{}{}
		cm.contextMap[contextID] = vmixAddr
	}, "UpdateContext")
}

func (cm *ConnectionManager) RemoveContext(ctx context.Context, vmixAddr string, contextID string) {
	cm.safeCall(ctx, func() {
		cm.logger.Info(ctx, "RemoveContext started. vmixAddr: %s, contextID: %s", vmixAddr, contextID)
		defer cm.logger.Info(ctx, "RemoveContext completed. vmixAddr: %s, contextID: %s", vmixAddr, contextID)

		cm.mu.Lock()
		defer cm.mu.Unlock()

		conn, exists := cm.connections[vmixAddr]
		if !exists {
			return
		}

		conn.mu.Lock()
		delete(conn.contexts, contextID)
		conn.mu.Unlock()

		delete(cm.contextMap, contextID)
		if len(conn.contexts) == 0 {
			conn.mu.Lock()
			if conn.retryCancel != nil {
				conn.retryCancel()
			}
			conn.active = false
			conn.mu.Unlock()
			delete(cm.connections, vmixAddr)
		}
	}, "RemoveContext")
}

func (cm *ConnectionManager) GetVMixByContext(ctx context.Context, contextID string) vmixtcp.Vmix {
	var client vmixtcp.Vmix
	cm.safeCall(ctx, func() {
		cm.logger.Info(ctx, "GetVMixByContext started. contextID: %s", contextID)
		defer cm.logger.Info(ctx, "GetVMixByContext completed. contextID: %s", contextID)

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
	defer func() {
		if r := recover(); r != nil {
			cm.logger.Error(parentCtx, "PANIC in manageConnection: %v\nStack Trace:\n%s",
				r, string(debug.Stack()))
		}
	}()

	cm.logger.Info(parentCtx, "manageConnection started. addr: %s", addr)
	defer cm.logger.Info(parentCtx, "manageConnection completed. addr: %s", addr)

	ctx, cancel := context.WithCancel(parentCtx)
	conn.mu.Lock()
	conn.retryCancel = cancel
	conn.mu.Unlock()

	for {
		cm.logger.Info(ctx, "retry loop started. addr: %s", addr)
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			// Check if connection is still active
			cm.logger.Info(ctx, "retry loop checking connection status. addr: %s", addr)

			conn.mu.RLock()
			active := conn.active
			conn.mu.RUnlock()
			if !active {
				cancel()
				return
			}

			cm.logger.Info(ctx, "retry loop connecting to vmix. addr: %s", addr)
			client := vmixtcp.New(addr)
			if err := client.Connect(ctx, 5*time.Second); err != nil {
				cm.logger.Error(ctx, "Failed to connect to vmix: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}
			conn.mu.Lock()
			conn.client = client
			conn.mu.Unlock()
			cm.logger.Info(ctx, "Connection established: vmixAddr=%s", addr)

			// Monitor connection status
			go func() {
				defer client.Close()

				for {
					select {
					case <-ctx.Done():
						return
					case <-time.After(1 * time.Second):
						conn.mu.RLock()
						active := conn.active
						conn.mu.RUnlock()

						if !active {
							return
						}

						// Check connection health
						// VERSION のがいい気がする
						if err := client.Tally(); err != nil {
							cm.logger.Error(ctx, "Connection lost: %v", err)
							conn.mu.Lock()
							conn.client = nil
							conn.mu.Unlock()
							return
						}
					}
				}
			}()

			// Wait for context cancellation or connection loss
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}
	}
}

func (cm *ConnectionManager) GetClient(ctx context.Context, vmixAddr string) vmixtcp.Vmix {
	var client vmixtcp.Vmix
	cm.safeCall(ctx, func() {
		cm.logger.Info(ctx, "GetClient started. vmixAddr: %s", vmixAddr)
		defer cm.logger.Info(ctx, "GetClient completed. vmixAddr: %s", vmixAddr)

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
	var contexts []string
	cm.safeCall(ctx, func() {
		cm.logger.Info(ctx, "GetContexts started. vmixAddr: %s", vmixAddr)
		defer cm.logger.Info(ctx, "GetContexts completed. vmixAddr: %s", vmixAddr)

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
