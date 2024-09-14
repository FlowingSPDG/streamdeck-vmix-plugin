package stdvmix

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

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

	go s.sendFuncPIs.Store(event.Context, p.Settings)
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

	if p.Settings.Tally {
		actName := "InputPreview"
		if p.Settings.Mix != nil {
			actName = fmt.Sprintf("%s%d", "InputPreviewMix", *p.Settings.Mix) //2~16
		}
		s.vMixClients.activatorContexts.Store(event.Context, activatorContext{
			destination: p.Settings.Dest,
			onAct: func(args []string) bool {
				if len(args) < 3 {
					return false
				}
				return args[0] == actName && args[1] == strconv.Itoa(p.Settings.Input)
			},
			activatorColor: activatorColorGreen,
		})
	}

	go s.previewPIs.Store(event.Context, p.Settings)
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

	if p.Settings.Tally {
		actName := "Input"
		if p.Settings.Mix != nil {
			actName = fmt.Sprintf("%s%d", "InputMix", *p.Settings.Mix) //2~16
		}
		s.vMixClients.activatorContexts.Store(event.Context, activatorContext{
			destination: p.Settings.Dest,
			onAct: func(args []string) bool {
				if len(args) < 3 {
					return false
				}
				return args[0] == actName && args[1] == strconv.Itoa(p.Settings.Input)
			},
			activatorColor: activatorColorGreen,
		})
	}
	go s.programPIs.Store(event.Context, p.Settings)
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

	var handler func(args []string) bool
	switch p.Settings.ActivatorName {
	case "InputPreview":
		handler = NewInputPreviewHandler(p.Settings.Input)
	}

	s.vMixClients.activatorContexts.Store(event.Context, activatorContext{
		destination:    p.Settings.Dest,
		onAct:          handler,
		activatorColor: p.Settings.Color,
	},
	)
	go s.activatorPIs.Store(event.Context, p.Settings)
	go s.vMixClients.storeNewCtxstr(p.Settings.Dest, event.Context)
	go s.vMixClients.storeNewVmix(ctx, p.Settings.Dest)
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
	return nil
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
	return nil
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

	// Get old setting
	oldVal, ok := s.previewPIs.Load(event.Context)
	if ok {
		// If destination is changed, delete old destination and store new destination
		if oldVal.Dest != p.Settings.Dest {
			s.logger.Printf("Destination host changed. Old:%s, New:%s\n", oldVal.Dest, p.Settings.Dest)
			s.vMixClients.unregisterDestinationForCtx(event.Context)
			go s.vMixClients.storeNewCtxstr(p.Settings.Dest, event.Context)
			go s.vMixClients.storeNewVmix(ctx, p.Settings.Dest)
		}
	}

	if !p.Settings.Tally {
		// Set default image
		go client.SetImage(ctx, "", streamdeck.HardwareAndSoftware)
		s.vMixClients.activatorContexts.Delete(event.Context)
	} else {
		// Set inactive tally
		client.SetImage(ctx, tallyInactive, streamdeck.HardwareAndSoftware)
		handler := NewInputPreviewHandler(p.Settings.Input)

		s.vMixClients.activatorContexts.Store(event.Context, activatorContext{
			destination:    p.Settings.Dest,
			onAct:          handler,
			activatorColor: activatorColorGreen,
		})
	}
	go s.previewPIs.Store(event.Context, p.Settings)
	return nil
}

func (s *StdVmix) PreviewSendToPluginHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
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

func (s *StdVmix) ProgramSendToPluginHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
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

func (s *StdVmix) ProgramDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.DidReceiveSettingsPayload[*ProgramPI]{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	// Get old setting
	oldVal, ok := s.programPIs.Load(event.Context)
	if ok {
		// If destination is changed, delete old destination and store new destination
		if oldVal.Dest != p.Settings.Dest {
			s.logger.Printf("Destination host changed. Old:%s, New:%s\n", oldVal.Dest, p.Settings.Dest)
			s.vMixClients.unregisterDestinationForCtx(event.Context)
			go s.vMixClients.storeNewCtxstr(p.Settings.Dest, event.Context)
			go s.vMixClients.storeNewVmix(ctx, p.Settings.Dest)
		}
	}

	if !p.Settings.Tally {
		// Set default image
		go client.SetImage(ctx, "", streamdeck.HardwareAndSoftware)
		s.vMixClients.activatorContexts.Delete(event.Context)
	} else {
		// Set inactive tally
		client.SetImage(ctx, tallyInactive, streamdeck.HardwareAndSoftware)
		handler := NewInputPreviewHandler(p.Settings.Input)
		s.vMixClients.activatorContexts.Store(event.Context, activatorContext{
			destination:    p.Settings.Dest,
			onAct:          handler,
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

	// Get old setting
	oldVal, ok := s.activatorPIs.Load(event.Context)
	if ok {
		// If destination is changed, delete old destination and store new destination
		if oldVal.Dest != p.Settings.Dest {
			s.logger.Printf("Destination host changed. Old:%s, New:%s\n", oldVal.Dest, p.Settings.Dest)
			s.vMixClients.unregisterDestinationForCtx(event.Context)
			go s.vMixClients.storeNewCtxstr(p.Settings.Dest, event.Context)
			go s.vMixClients.storeNewVmix(ctx, p.Settings.Dest)
		}
	}

	handler := NewInputPreviewHandler(p.Settings.Input)
	s.vMixClients.activatorContexts.Store(event.Context, activatorContext{
		destination:    p.Settings.Dest,
		onAct:          handler,
		activatorColor: p.Settings.Color,
	})
	go s.activatorPIs.Store(event.Context, p.Settings)

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
