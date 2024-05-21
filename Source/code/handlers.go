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
		s.vMixClients.activatorContexts.Store(event.Context, activatorContext{
			destination:    p.Settings.Dest,
			input:          p.Settings.Input,
			activatorName:  "InputPreview",
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
		s.vMixClients.activatorContexts.Store(event.Context, activatorContext{
			destination:    p.Settings.Dest,
			input:          p.Settings.Input,
			activatorName:  "Input",
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

	s.vMixClients.activatorContexts.Store(event.Context, activatorContext{
		destination:    p.Settings.Dest,
		input:          p.Settings.Input,
		activatorName:  p.Settings.Activator,
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
		s.vMixClients.activatorContexts.Delete(event.Context)
	} else {
		s.vMixClients.activatorContexts.Store(event.Context, activatorContext{
			destination:    p.Settings.Dest,
			input:          p.Settings.Input,
			activatorName:  "InputPreview",
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
		s.vMixClients.activatorContexts.Delete(event.Context)
	} else {
		s.vMixClients.activatorContexts.Store(event.Context, activatorContext{
			destination:    p.Settings.Dest,
			input:          p.Settings.Input,
			activatorName:  "Input",
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
	s.logger.Printf("ActivatorDidReceiveSettingsHandler. Settings:%#v\n", p.Settings)

	// Reset off tally
	client.SetImage(ctx, tallyInactive, streamdeck.HardwareAndSoftware)

	s.vMixClients.activatorContexts.Store(event.Context, activatorContext{
		destination:    p.Settings.Dest,
		input:          p.Settings.Input,
		activatorName:  p.Settings.Activator,
		activatorColor: p.Settings.Color,
	})
	s.vMixClients.deleteByCtxstr(event.Context)
	s.vMixClients.storeNewCtxstr(p.Settings.Dest, event.Context)

	return nil
}

func (s *StdVmix) ActivatorSendToPluginHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := make(map[string]string)
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}
	ev := p["property_inspector"]
	switch ev {
	case "propertyInspectorConnected":
		payload := InputsForPI{
			Inputs: make(map[string][]Input, s.vMixClients.vmInputs.Size()),
		}
		s.vMixClients.vmInputs.Range(func(dest string, inputs []Input) bool {
			payload.Inputs[dest] = inputs
			return true
		})
		client.SendToPropertyInspector(ctx, SendToPropertyInspectorPayload[InputsForPI]{
			Event:   "inputs",
			Payload: payload,
		})
	}
	return nil
}
