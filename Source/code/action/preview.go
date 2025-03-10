package action

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FlowingSPDG/streamdeck"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/connection"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/setting"
	"golang.org/x/xerrors"
)

const PreviewActionUUID = "dev.flowingspdg.vmix.preview"

type PreviewAction interface {
	OnWillAppear() streamdeck.EventHandler
	OnWillDisappear() streamdeck.EventHandler
	OnUpdateSettings() streamdeck.EventHandler
	Execute(ctx context.Context) error
	OnVMixTally(ctx context.Context) error
}

type previewAction struct {
	logger            logger.Logger
	connectionManager *connection.ConnectionManager
	store             setting.SettingStore[setting.PreviewSetting]
}

func (p *previewAction) OnWillAppear() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		payload := streamdeck.WillAppearPayload[setting.PreviewSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			p.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			return xerrors.Errorf("failed to unmarshal payload: %w", err)
		}

		p.logger.Info(ctx, "OnWillAppear started. contextID: %s", event.Context)
		defer p.logger.Info(ctx, "OnWillAppear completed. contextID: %s", event.Context)

		p.store.Store(event.Context, &payload.Settings)
		p.connectionManager.AddContext(ctx, payload.Settings.VMixAddress, event.Context)

		return nil
	}
}

func (p *previewAction) OnWillDisappear() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		payload := streamdeck.WillDisappearPayload[setting.PreviewSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			p.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			return xerrors.Errorf("failed to unmarshal payload: %w", err)
		}

		p.logger.Info(ctx, "OnWillDisappear started. contextID: %s", event.Context)
		defer p.logger.Info(ctx, "OnWillDisappear completed. contextID: %s", event.Context)

		p.store.Delete(event.Context)
		p.connectionManager.RemoveContext(ctx, payload.Settings.VMixAddress, event.Context)
		return nil
	}
}

func (p *previewAction) OnUpdateSettings() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		payload := streamdeck.DidReceiveSettingsPayload[setting.PreviewSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			p.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			return xerrors.Errorf("failed to unmarshal payload: %w", err)
		}

		p.logger.Info(ctx, "OnUpdateSettings started. contextID: %s", event.Context)
		defer p.logger.Info(ctx, "OnUpdateSettings completed. contextID: %s", event.Context)

		p.store.Store(event.Context, &payload.Settings)
		p.connectionManager.UpdateContext(ctx, payload.Settings.VMixAddress, event.Context)

		return nil
	}
}

func (p *previewAction) Execute(ctx context.Context) error {
	p.logger.Info(ctx, "Execute started")
	defer p.logger.Info(ctx, "Execute completed")

	// Get current setting from context
	ctxID := ctx.Value("contextID").(string)
	setting, ok := p.store.Load(ctxID)
	if !ok {
		err := fmt.Errorf("setting not found")
		p.logger.Error(ctx, "Failed to load setting", "error", err)
		return err
	}

	// Get vMix client
	client := p.connectionManager.GetVMixByContext(ctx, setting.ContextID)
	if client == nil {
		err := fmt.Errorf("vMix connection not found")
		p.logger.Error(ctx, "Failed to get vMix client", "error", err)
		return err
	}

	// Execute PreviewInput function
	if err := client.Function("PreviewInput", fmt.Sprintf("Input=%d", setting.Input)); err != nil {
		p.logger.Error(ctx, "Failed to execute PreviewInput", "error", err)
		return fmt.Errorf("failed to execute PreviewInput: %w", err)
	}

	p.logger.Info(ctx, "PreviewInput executed successfully", "input", setting.Input)
	return nil
}

func (p *previewAction) OnVMixTally(ctx context.Context) error {
	p.logger.Info(ctx, "OnVMixTally started")
	defer p.logger.Info(ctx, "OnVMixTally completed")

	// Get current setting from context
	ctxID := ctx.Value("contextID").(string)
	setting, ok := p.store.Load(ctxID)
	if !ok {
		err := fmt.Errorf("setting not found")
		p.logger.Error(ctx, "Failed to load setting", "error", err)
		return err
	}

	// Get vMix client
	client := p.connectionManager.GetVMixByContext(ctx, setting.ContextID)
	if client == nil {
		err := fmt.Errorf("vMix connection not found")
		p.logger.Error(ctx, "Failed to get vMix client", "error", err)
		return err
	}

	// TODO: Implement tally status handling
	p.logger.Info(ctx, "Tally status handling not yet implemented")
	return nil
}

func NewPreviewAction(logger logger.Logger, connectionManager *connection.ConnectionManager, store setting.SettingStore[setting.PreviewSetting]) PreviewAction {
	return &previewAction{
		logger:            logger,
		connectionManager: connectionManager,
		store:             store,
	}
}
