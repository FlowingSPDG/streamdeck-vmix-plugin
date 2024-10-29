package adapters

import (
	"context"

	"github.com/FlowingSPDG/streamdeck"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
)

type StreamDeckContextAdapter interface {
	// TALLY
	SetInactiveColor(ctx context.Context, target streamdeck.Target) error
	SetPreviewColor(ctx context.Context, target streamdeck.Target) error
	SetProgramColor(ctx context.Context, target streamdeck.Target) error
}

type VMixAdapter interface {
	AddVMix(ctx context.Context, destination string)
	RemoveVMix(ctx context.Context)
	PreviewInput(destination string, input int) error
	OnTally(f func(ctx context.Context, host string, tally *vmixtcp.TallyResponse) error)
	// OnVersion(f func(ctx context.Context, host string, version *vmixtcp.VersionResponse) error)
}
