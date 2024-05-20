package stdvmix

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FlowingSPDG/streamdeck"
)

// SendFuncWillAppearHandler willAppear handler.
func (s *StdVmix) SendFuncWillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.WillAppearPayload[SendFunctionPI]{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	if p.Settings.IsDefault() {
		p.Settings.Initialize()
		msg := fmt.Sprintf("Forcing Default value:%v", p.Settings)
		client.LogMessage(msg)
		if err := client.SetSettings(ctx, p.Settings); err != nil {
			return err
		}
	}
	s.sendFuncContexts.Store(event.Context, p.Settings)

	go s.vMixClients.storeNewVmix(p.Settings.Host, p.Settings.Port)
	return nil
}

// PreviewWillAppearHandler willAppear handler.
func (s *StdVmix) PreviewWillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.WillAppearPayload[PreviewPI]{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	if p.Settings.IsDefault() {
		p.Settings.Initialize()
		msg := fmt.Sprintf("Forcing Default value:%v", p.Settings)
		client.LogMessage(msg)
		if err := client.SetSettings(ctx, p.Settings); err != nil {
			return err
		}
	}
	s.previewContexts.Store(event.Context, p.Settings)
	go s.vMixClients.storeNewVmix(p.Settings.Host, p.Settings.Port)
	return nil
}

// PreviewWillAppearHandler willAppear handler.
func (s *StdVmix) ProgramWillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.WillAppearPayload[ProgramPI]{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	if p.Settings.IsDefault() {
		p.Settings.Initialize()
		msg := fmt.Sprintf("Forcing Default value:%v", p.Settings)
		client.LogMessage(msg)
		if err := client.SetSettings(ctx, p.Settings); err != nil {
			return err
		}
	}
	s.programContexts.Store(event.Context, p.Settings)
	go s.vMixClients.storeNewVmix(p.Settings.Host, p.Settings.Port)
	return nil
}

// SendFuncKeyDownHandler keyDown handler
func (s *StdVmix) SendFuncKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.KeyDownPayload[SendFunctionPI]{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		client.ShowAlert(ctx)
		return err
	}

	client.LogMessage("KeyDownHandler")
	client.LogMessage(fmt.Sprintf("settings for this context:%#v", p.Settings))

	if err := s.ExecuteSend(p.Settings); err != nil {
		client.ShowAlert(ctx)
		return err
	}
	return client.ShowOk(ctx)
}

// PreviewKeyDownHandler keyDown handler
func (s *StdVmix) PreviewKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.KeyDownPayload[PreviewPI]{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	client.LogMessage("KeyDownHandler")
	client.LogMessage(fmt.Sprintf("settings for this context:%#v", p.Settings))

	if err := s.ExecutePreview(p.Settings); err != nil {
		client.ShowAlert(ctx)
		return err
	}
	return client.ShowOk(ctx)
}

// ProgramKeyDownHandler keyDown handler
func (s *StdVmix) ProgramKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.KeyDownPayload[ProgramPI]{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	client.LogMessage("KeyDownHandler")
	client.LogMessage(fmt.Sprintf("settings for this context:%#v", p.Settings))

	if err := s.ExecuteProgram(p.Settings); err != nil {
		client.ShowAlert(ctx)
		return err
	}
	return client.ShowOk(ctx)
}

func (s *StdVmix) SendFuncDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.DidReceiveSettingsPayload[SendFunctionPI]{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}
	s.sendFuncContexts.Store(event.Context, p.Settings)
	return nil
}

func (s *StdVmix) PreviewDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.DidReceiveSettingsPayload[PreviewPI]{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}
	if !p.Settings.Tally {
		client.SetImage(ctx, "", streamdeck.HardwareAndSoftware)
	}
	s.previewContexts.Store(event.Context, p.Settings)
	return nil
}

func (s *StdVmix) ProgramDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.DidReceiveSettingsPayload[ProgramPI]{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}
	if !p.Settings.Tally {
		client.SetImage(ctx, "", streamdeck.HardwareAndSoftware)
	}
	s.programContexts.Store(event.Context, p.Settings)
	return nil
}
