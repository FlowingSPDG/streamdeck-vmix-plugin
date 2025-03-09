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

	SendInputs(ctx context.Context, host string, inputs map[string]Input) error
}

type Input struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	Number int    `json:"number"`
}

type VMixAdapter interface {
	AddVMix(ctx context.Context, destination string)
	RemoveVMix(ctx context.Context)
	PreviewInput(ctx context.Context, destination string, input int) error
	OnTally(f func(ctx context.Context, host string, tally *vmixtcp.TallyResponse) error)
	// OnVersion(f func(ctx context.Context, host string, version *vmixtcp.VersionResponse) error)
	OnXML(f func(ctx context.Context, host string, xml *vmixtcp.XMLResponse) error)
}
