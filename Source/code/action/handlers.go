package action

import (
	"context"

	"github.com/FlowingSPDG/streamdeck"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/xerrors"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/connection"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/setting"
)

// VMixConnectorImpl implements VMixConnector interface
type VMixConnectorImpl struct {
	logger            logger.Logger
	connectionManager *connection.ConnectionManager
	actionUUID        string
}

func NewVMixConnector(logger logger.Logger, connectionManager *connection.ConnectionManager, actionUUID string) *VMixConnectorImpl {
	return &VMixConnectorImpl{
		logger:            logger,
		connectionManager: connectionManager,
		actionUUID:        actionUUID,
	}
}

func (v *VMixConnectorImpl) GetVMixClient(ctx context.Context) vmixtcp.Vmix {
	return v.connectionManager.GetClient(ctx, "")
}

func (v *VMixConnectorImpl) ConnectVMix(ctx context.Context, addr string) error {
	v.connectionManager.AddVMix(ctx, addr)
	return nil
}

func (v *VMixConnectorImpl) DisconnectVMix(ctx context.Context, addr string) error {
	v.connectionManager.RemoveVMix(ctx, addr)
	return nil
}

func (v *VMixConnectorImpl) GetClient(ctx context.Context, addr string) vmixtcp.Vmix {
	return v.connectionManager.GetClient(ctx, addr)
}

func (v *VMixConnectorImpl) GetAllVMixAddrs(ctx context.Context) []string {
	return v.connectionManager.GetAllVMixAddrs(ctx)
}

// PropertyInspectorHandlerImpl implements PropertyInspectorHandler interface
type PropertyInspectorHandlerImpl struct {
	logger   logger.Logger
	client   *streamdeck.Client
	vmixConn setting.VMixConnector
}

func NewPropertyInspectorHandler(logger logger.Logger, client *streamdeck.Client, vmixConn setting.VMixConnector) *PropertyInspectorHandlerImpl {
	return &PropertyInspectorHandlerImpl{
		logger:   logger,
		client:   client,
		vmixConn: vmixConn,
	}
}

func (p *PropertyInspectorHandlerImpl) UpdatePropertyInspector(ctx context.Context, event streamdeck.Event) error {
	// Send destinations
	destinations := p.vmixConn.GetAllVMixAddrs(ctx)
	payload := struct {
		Event        string   `json:"event"`
		Destinations []string `json:"destinations"`
	}{
		Event:        "destinations",
		Destinations: destinations,
	}
	if err := p.client.SendToPropertyInspector(ctx, payload); err != nil {
		return xerrors.Errorf("failed to send destinations to PropertyInspector: %w", err)
	}

	// Send inputs
	inputsMap := make(setting.DestinationToInputs)
	for _, addr := range destinations {
		if vmix := p.vmixConn.GetClient(ctx, addr); vmix != nil {
			// TODO: vmix-goのXMLResponseの構造を確認して実装を更新
			// 一時的な実装としてダミーデータを使用
			inputs := []*setting.Input{
				{
					Key:    "1",
					Name:   "Input 1",
					Number: 1,
				},
				{
					Key:    "2",
					Name:   "Input 2",
					Number: 2,
				},
			}
			inputsMap[addr] = inputs
		}
	}

	inputsPayload := struct {
		Event  string                      `json:"event"`
		Inputs setting.DestinationToInputs `json:"inputs"`
	}{
		Event:  "inputs",
		Inputs: inputsMap,
	}
	if err := p.client.SendToPropertyInspector(ctx, inputsPayload); err != nil {
		return xerrors.Errorf("failed to send inputs to PropertyInspector: %w", err)
	}

	return nil
}

func (p *PropertyInspectorHandlerImpl) SendToPropertyInspector(ctx context.Context, payload interface{}) error {
	return p.client.SendToPropertyInspector(ctx, payload)
}

// TallyHandlerImpl implements TallyHandler interface
type TallyHandlerImpl struct {
	logger          logger.Logger
	contextTallyMap *xsync.MapOf[string, tallyStatus]
	getTallyImage   func(isActive bool) string
}

func NewTallyHandler(logger logger.Logger, getTallyImage func(isActive bool) string) *TallyHandlerImpl {
	return &TallyHandlerImpl{
		logger:          logger,
		contextTallyMap: xsync.NewMapOf[string, tallyStatus](),
		getTallyImage:   getTallyImage,
	}
}

func (t *TallyHandlerImpl) HandleTally(ctx context.Context, resp *vmixtcp.TallyResponse) error {
	if resp == nil {
		return xerrors.Errorf("tally response is nil")
	}

	// TODO: TallyResponseの正しいフィールドを使用して状態を更新
	t.contextTallyMap.Range(func(contextID string, _ tallyStatus) bool {
		// 一時的な実装: ランダムな状態を設定
		status := tallyStatusOff
		t.contextTallyMap.Store(contextID, status)
		return true
	})

	return nil
}

func (t *TallyHandlerImpl) HandleActs(ctx context.Context, resp *vmixtcp.ActsResponse) error {
	if resp == nil {
		return xerrors.Errorf("acts response is nil")
	}

	// TODO: ActsResponseの正しいフィールドを使用して状態を更新
	t.contextTallyMap.Range(func(contextID string, _ tallyStatus) bool {
		// 一時的な実装: ランダムな状態を設定
		status := tallyStatusOff
		t.contextTallyMap.Store(contextID, status)
		return true
	})

	return nil
}

// ShortcutHandlerImpl implements ShortcutHandler interface
type ShortcutHandlerImpl struct {
	logger logger.Logger
	vmix   setting.VMixConnector
}

func NewShortcutHandler(logger logger.Logger, vmix setting.VMixConnector) *ShortcutHandlerImpl {
	return &ShortcutHandlerImpl{
		logger: logger,
		vmix:   vmix,
	}
}

func (s *ShortcutHandlerImpl) SendShortcut(ctx context.Context, shortcut string) error {
	vmix := s.vmix.GetVMixClient(ctx)
	if vmix == nil {
		return xerrors.Errorf("vMix client not found")
	}
	return vmix.Function(shortcut, "")
}

// SettingsHandlerImpl implements SettingsHandler interface
type SettingsHandlerImpl[T setting.Setting] struct {
	logger logger.Logger
	store  setting.SettingStore[T]
}

func NewSettingsHandler[T setting.Setting](logger logger.Logger, store setting.SettingStore[T]) *SettingsHandlerImpl[T] {
	return &SettingsHandlerImpl[T]{
		logger: logger,
		store:  store,
	}
}

func (s *SettingsHandlerImpl[T]) LoadSettings(ctx context.Context, contextID string) (setting.Setting, bool) {
	settings, ok := s.store.Load(contextID)
	return settings, ok
}

func (s *SettingsHandlerImpl[T]) StoreSettings(ctx context.Context, contextID string, settings setting.Setting) {
	s.store.Store(contextID, settings.(T))
}

func (s *SettingsHandlerImpl[T]) DeleteSettings(ctx context.Context, contextID string) {
	s.store.Delete(contextID)
}
