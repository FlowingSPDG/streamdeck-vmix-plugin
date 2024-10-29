package controllers

import (
	"context"

	"github.com/FlowingSPDG/streamdeck"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
)

type PreviewActionController interface {
	// register
	RegisterAction(sdAction *streamdeck.Action)

	// handlers
	WillAppearHandler() streamdeck.EventHandler
	WilDisappearHandler() streamdeck.EventHandler
	KeyDownHandler() streamdeck.EventHandler
	DidReceiveSettingsHandler() streamdeck.EventHandler
}

type VMixController interface {
	OnTally(ctx context.Context, host string, tally *vmixtcp.TallyResponse) error
}
