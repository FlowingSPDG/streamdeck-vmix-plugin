package controller

import (
	"context"
	"encoding/json"

	"github.com/FlowingSPDG/streamdeck"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/action"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/setting"
	"golang.org/x/xerrors"
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

type previewActionController struct {
	// actions
	previewAction action.PreviewAction
}

func NewPreviewActionController(previewAction action.PreviewAction) PreviewActionController {
	return &previewActionController{
		previewAction: previewAction,
	}
}

func (s *previewActionController) RegisterAction(sdAction *streamdeck.Action) {
	sdAction.RegisterHandler(streamdeck.WillAppear, s.WillAppearHandler())
	sdAction.RegisterHandler(streamdeck.WillDisappear, s.WilDisappearHandler())
	sdAction.RegisterHandler(streamdeck.KeyDown, s.KeyDownHandler())
	sdAction.RegisterHandler(streamdeck.DidReceiveSettings, s.DidReceiveSettingsHandler())
}

func (s *previewActionController) WillAppearHandler() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		p := streamdeck.WillAppearPayload[setting.PreviewSetting]{}
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return xerrors.Errorf("failed to unmarshal setting from payload: %w", err)
		}
		if err := s.previewAction.Appear(ctx, &p.Settings); err != nil {
			return xerrors.Errorf("failed to appear: %w", err)
		}
		return nil
	}
}

func (s *previewActionController) WilDisappearHandler() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		p := streamdeck.WillDisappearPayload[setting.PreviewSetting]{}
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return xerrors.Errorf("failed to unmarshal setting from payload: %w", err)
		}
		if err := s.previewAction.Disappear(ctx, &p.Settings); err != nil {
			return xerrors.Errorf("failed to disappear: %w", err)
		}
		return nil
	}
}

func (s *previewActionController) KeyDownHandler() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		if err := s.previewAction.Execute(ctx); err != nil {
			return xerrors.Errorf("failed to execute: %w", err)
		}
		return nil
	}
}

func (s *previewActionController) DidReceiveSettingsHandler() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		p := streamdeck.DidReceiveSettingsPayload[setting.PreviewSetting]{}
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return xerrors.Errorf("failed to unmarshal setting from payload: %w", err)
		}

		if err := s.previewAction.UpdateSetting(ctx, &p.Settings); err != nil {
			return xerrors.Errorf("failed to update setting: %w", err)
		}
		return nil
	}
}
