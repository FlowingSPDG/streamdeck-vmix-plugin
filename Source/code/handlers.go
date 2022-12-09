package stdvmix

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FlowingSPDG/streamdeck"
)

// SendFuncKeyDownHandler keyDown handler
func (s *StdVmix) SendFuncKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	if !s.vMixLaunched {
		return client.ShowAlert(ctx)
	}

	p := streamdeck.KeyDownPayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}
	sp := SendFunctionPI{}
	if err := json.Unmarshal(p.Settings, &sp); err != nil {
		return err
	}

	client.LogMessage("KeyDownHandler")
	client.LogMessage(fmt.Sprintf("settings for this context:%v\n", s))

	query, err := sp.GenerateFunction()
	if err != nil {
		client.LogMessage(fmt.Sprintf("Failed to gemerate function query:%v\n", err))
		client.ShowAlert(ctx)
		return err
	}
	client.LogMessage(fmt.Sprintln("Generated Query:", query))
	if err := s.v.FUNCTION(query); err != nil {
		client.LogMessage(fmt.Sprintln("Failed to send vMix FUNCTION:", err))
		client.ShowAlert(ctx)
		return err
	}

	return client.ShowOk(ctx)
}

// PreviewKeyDownHandler keyDown handler
func (s *StdVmix) PreviewKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	if !s.vMixLaunched {
		return client.ShowAlert(ctx)
	}
	p := streamdeck.KeyDownPayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}
	pp := PreviewPI{}
	if err := json.Unmarshal(p.Settings, &pp); err != nil {
		return err
	}

	client.LogMessage("KeyDownHandler")
	client.LogMessage(fmt.Sprintf("settings for this context:%v\n", s))

	query, err := pp.GenerateFunction()
	if err != nil {
		client.LogMessage(fmt.Sprintf("Failed to gemerate function query:%v\n", err))
		client.ShowAlert(ctx)
		return err
	}
	client.LogMessage(fmt.Sprintln("Generated Query:", query))
	if err := s.v.FUNCTION(query); err != nil {
		client.LogMessage(fmt.Sprintln("Failed to send vMix FUNCTION:", err))
		client.ShowAlert(ctx)
		return err
	}

	return client.ShowOk(ctx)
}

// ApplicationDidLaunchHandler applicationDidLaunch handler
func (s *StdVmix) ApplicationDidLaunchHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.ApplicationDidLaunchPayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		client.LogMessage(fmt.Sprintln("Failed to unmarshal ApplicationDidLaunchPayload payload:", err))
		return err
	}
	client.LogMessage(fmt.Sprintf("ApplicationDidLaunchHandler:%s\n", p))
	if p.Application == "vMix64.exe" || p.Application == "vMix.exe" {
		s.vMixLaunched = true
	}
	return nil
}

// ApplicationDidTerminateHandler applicationDidTerminate handler
func (s *StdVmix) ApplicationDidTerminateHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.ApplicationDidTerminatePayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		client.LogMessage(fmt.Sprintln("Failed to unmarshal ApplicationDidTerminatePayload payload:", err))
		return err
	}
	client.LogMessage(fmt.Sprintln("ApplicationDidTerminateHandler:", p))
	if p.Application == "vMix64.exe" || p.Application == "vMix.exe" {
		s.vMixLaunched = true
		s.v.Close()
		s.v = nil
	}

	return nil
}

// SendFuncDidReceiveSettingsHandler didReceiveSettings Handler
func (s *StdVmix) SendFuncDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.DidReceiveSettingsPayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		client.LogMessage(fmt.Sprintln("Failed to unmarshal DidReceiveSettingsPayload payload:", err))
		return err
	}

	sfpi := SendFunctionPI{}
	if err := json.Unmarshal(p.Settings, &sfpi); err != nil {
		client.LogMessage(fmt.Sprintln("Failed to unmarshal PropertyInspector:", err))
		return err
	}
	client.LogMessage(fmt.Sprintf("DidReceiveSettingsHandler:%v\n", s))

	// inputsを更新
	sfpi.Inputs = s.inputs
	client.SetSettings(ctx, s)

	return nil
}

// PreviewDidReceiveSettingsHandler didReceiveSettings Handler
func (s *StdVmix) PreviewDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.DidReceiveSettingsPayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		client.LogMessage(fmt.Sprintln("Failed to unmarshal DidReceiveSettingsPayload payload:", err))
		return err
	}

	ppi := &PreviewPI{}
	if err := json.Unmarshal(p.Settings, ppi); err != nil {
		client.LogMessage(fmt.Sprintln("Failed to unmarshal PropertyInspector:", err))
		return err
	}
	client.LogMessage(fmt.Sprintf("DidReceiveSettingsHandler:%v\n", s))

	// inputsを更新
	ppi.Inputs = s.inputs
	client.SetSettings(ctx, s)

	for _, input := range s.inputs {
		if ppi.Input != input.Key {
			continue
		}
		if input.TallyPreview {
			client.SetImage(ctx, tallyPreview, streamdeck.HardwareAndSoftware)
		} else {
			client.SetImage(ctx, tallyInactive, streamdeck.HardwareAndSoftware)
		}
	}

	return nil
}

// ProgramKeyDownHandler keyDown handler
func (s *StdVmix) ProgramKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	if !s.vMixLaunched {
		return client.ShowAlert(ctx)
	}
	p := streamdeck.KeyDownPayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}
	pp := ProgramPI{}
	if err := json.Unmarshal(p.Settings, &pp); err != nil {
		return err
	}

	client.LogMessage("KeyDownHandler")
	client.LogMessage(fmt.Sprintf("settings for this context:%v\n", s))

	query, err := pp.GenerateFunction()
	if err != nil {
		client.LogMessage(fmt.Sprintf("Failed to gemerate function query:%v\n", err))
		client.ShowAlert(ctx)
		return err
	}
	client.LogMessage(fmt.Sprintln("Generated Query:", query))
	if err := s.v.FUNCTION(query); err != nil {
		client.LogMessage(fmt.Sprintln("Failed to send vMix FUNCTION:", err))
		client.ShowAlert(ctx)
		return err
	}

	return client.ShowOk(ctx)
}

// ProgramDidReceiveSettingsHandler didReceiveSettings Handler
func (s *StdVmix) ProgramDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.DidReceiveSettingsPayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		client.LogMessage(fmt.Sprintln("Failed to unmarshal DidReceiveSettingsPayload payload:", err))
		return err
	}

	ppi := &ProgramPI{}
	if err := json.Unmarshal(p.Settings, ppi); err != nil {
		client.LogMessage(fmt.Sprintln("Failed to unmarshal PropertyInspector:", err))
		return err
	}
	client.LogMessage(fmt.Sprintf("DidReceiveSettingsHandler:%v\n", s))

	// inputsを更新
	ppi.Inputs = s.inputs
	client.SetSettings(ctx, s)

	for _, input := range s.inputs {
		if ppi.Input != input.Key {
			continue
		}
		if input.TallyProgram {
			client.SetImage(ctx, tallyProgram, streamdeck.HardwareAndSoftware)
		} else {
			client.SetImage(ctx, tallyInactive, streamdeck.HardwareAndSoftware)
		}
	}

	return nil
}
