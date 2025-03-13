package setting

import (
	"context"

	"github.com/FlowingSPDG/streamdeck"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
)

// VMixConnector provides vMix connection management functionality
type VMixConnector interface {
	GetVMixClient(ctx context.Context) vmixtcp.Vmix
	GetClient(ctx context.Context, addr string) vmixtcp.Vmix
	GetAllVMixAddrs(ctx context.Context) []DestinationStatus
	ConnectVMix(ctx context.Context, addr string) error
	DisconnectVMix(ctx context.Context, addr string) error
}

// PropertyInspectorHandler provides property inspector communication functionality
type PropertyInspectorHandler interface {
	UpdatePropertyInspector(ctx context.Context, event streamdeck.Event) error
	SendToPropertyInspector(ctx context.Context, payload interface{}) error
	SetVMixConnector(vmixConn VMixConnector)
}

// TallyHandler provides tally state management functionality
type TallyHandler interface {
	HandleTally(ctx context.Context, resp *vmixtcp.TallyResponse) error
	HandleActs(ctx context.Context, resp *vmixtcp.ActsResponse) error
}

// ShortcutHandler provides vMix shortcut functionality
type ShortcutHandler interface {
	SendShortcut(ctx context.Context, shortcut string) error
}

// SettingsHandler provides settings management functionality
type SettingsHandler interface {
	LoadSettings(ctx context.Context, contextID string) (Setting, bool)
	StoreSettings(ctx context.Context, contextID string, settings Setting)
	DeleteSettings(ctx context.Context, contextID string)
}
