package stdvmix

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FlowingSPDG/streamdeck"
)

// TODO: 共通の処理を纏めて書く

// SendFuncWillAppearHandler willAppear handler.
func (s *StdVmix) SendFuncWillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.WillAppearPayload[*SendFunctionPI]{}
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
	s.sendFuncPIs.Store(event.Context, p.Settings)

	go s.vMixClients.storeNewCtxstr(p.Settings.Dest, event.Context)
	go s.vMixClients.storeNewVmix(ctx, p.Settings.Dest)
	return nil
}

// PreviewWillAppearHandler willAppear handler.
func (s *StdVmix) PreviewWillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.WillAppearPayload[*PreviewPI]{}
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
	s.previewPIs.Store(event.Context, p.Settings)
	if p.Settings.Tally {
		s.vMixClients.activatorContexts.Store(activatorKey{
			input:         p.Settings.Input,
			activatorName: "InputPreview",
		}, activatorContext{
			ctxStr:         event.Context,
			activatorColor: activatorColorGreen,
		})
	}
	go s.vMixClients.storeNewCtxstr(p.Settings.Dest, event.Context)
	go s.vMixClients.storeNewVmix(ctx, p.Settings.Dest)
	return nil
}

// PreviewWillAppearHandler willAppear handler.
func (s *StdVmix) ProgramWillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.WillAppearPayload[*ProgramPI]{}
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
	s.programPIs.Store(event.Context, p.Settings)
	if p.Settings.Tally {
		s.vMixClients.activatorContexts.Store(activatorKey{
			input:         p.Settings.Input,
			activatorName: "Input",
		}, activatorContext{
			ctxStr:         event.Context,
			activatorColor: activatorColorRed,
		})
	}
	go s.vMixClients.storeNewCtxstr(p.Settings.Dest, event.Context)
	go s.vMixClients.storeNewVmix(ctx, p.Settings.Dest)
	return nil
}

// ActivatorWillAppearHandler willAppear handler.
func (s *StdVmix) ActivatorWillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.WillAppearPayload[*ActivatorPI]{}
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
	s.activatorPIs.Store(event.Context, p.Settings)

	s.vMixClients.activatorContexts.Store(
		activatorKey{
			input:         p.Settings.Input,
			activatorName: p.Settings.Activator,
		},
		activatorContext{
			ctxStr:         event.Context,
			activatorColor: p.Settings.Color,
		},
	)

	s.vMixClients.storeNewCtxstr(p.Settings.Dest, event.Context)
	s.vMixClients.storeNewVmix(ctx, p.Settings.Dest)
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

	if err := s.ExecuteSend(ctx, p.Settings); err != nil {
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

	if err := s.ExecutePreview(ctx, p.Settings); err != nil {
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

	if err := s.ExecuteProgram(ctx, p.Settings); err != nil {
		client.ShowAlert(ctx)
		return err
	}
	return client.ShowOk(ctx)
}

func (s *StdVmix) SendFuncDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.DidReceiveSettingsPayload[*SendFunctionPI]{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}
	s.sendFuncPIs.Store(event.Context, p.Settings)
	return nil
}

func (s *StdVmix) PreviewDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.DidReceiveSettingsPayload[*PreviewPI]{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}
	if !p.Settings.Tally {
		go client.SetImage(ctx, "", streamdeck.HardwareAndSoftware)
		s.vMixClients.activatorContexts.Delete(activatorKey{
			input:         p.Settings.Input,
			activatorName: "InputPreview",
		}, event.Context)
	} else {
		s.vMixClients.activatorContexts.Store(activatorKey{
			input:         p.Settings.Input,
			activatorName: "InputPreview",
		}, activatorContext{
			ctxStr:         event.Context,
			activatorColor: activatorColorGreen,
		})
	}
	s.previewPIs.Store(event.Context, p.Settings)
	return nil
}

func (s *StdVmix) ProgramDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.DidReceiveSettingsPayload[*ProgramPI]{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}
	if !p.Settings.Tally {
		go client.SetImage(ctx, "", streamdeck.HardwareAndSoftware)
		s.vMixClients.activatorContexts.Delete(activatorKey{
			input:         p.Settings.Input,
			activatorName: "Input",
		}, event.Context)
	} else {
		s.vMixClients.activatorContexts.Store(activatorKey{
			input:         p.Settings.Input,
			activatorName: "Input",
		}, activatorContext{
			ctxStr:         event.Context,
			activatorColor: activatorColorRed,
		})
	}
	s.programPIs.Store(event.Context, p.Settings)
	return nil
}

func (s *StdVmix) ActivatorDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.DidReceiveSettingsPayload[*ActivatorPI]{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	// Reset off tally
	client.SetImage(ctx, tallyInactive, streamdeck.HardwareAndSoftware)

	// Cleanup previous
	s.vMixClients.activatorContexts.DeleteByContext(event.Context)
	s.vMixClients.activatorContexts.Store(activatorKey{
		input:         p.Settings.Input,
		activatorName: p.Settings.Activator,
	}, activatorContext{
		ctxStr:         event.Context,
		activatorColor: p.Settings.Color,
	})
	return nil
}
