package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FlowingSPDG/streamdeck"
)

// WillAppearHandler willAppear handler.
func WillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.WillAppearPayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	s := PropertyInspector{}
	if err := json.Unmarshal(p.Settings, &s); err != nil {
		return err
	}
	s.Inputs = inputs
	settings.Store(event.Context, &s)
	client.SetSettings(ctx, s)
	msg := fmt.Sprintf("WillAppearHandler:%v\nPI:%v\n", p, s)
	client.LogMessage(msg)
	return nil
}

// KeyDownHandler keyDown handler
func KeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	if !vMixLaunched {
		return client.ShowAlert(ctx)
	}

	s, err := settings.Load(event.Context)
	if err != nil {
		return fmt.Errorf("couldn't find settings for context %v", event.Context)
	}
	client.LogMessage("KeyDownHandler")
	client.LogMessage(fmt.Sprintf("settings for this context:%v\n", s))

	query, err := s.GenerateFunction()
	if err != nil {
		client.LogMessage(fmt.Sprintf("Failed to gemerate function query:%v\n", err))
		client.ShowAlert(ctx)
		return err
	}
	client.LogMessage(fmt.Sprintln("Generated Query:", query))
	if err := vMix.FUNCTION(query); err != nil {
		client.LogMessage(fmt.Sprintln("Failed to send vMix FUNCTION:", err))
		client.ShowAlert(ctx)
		return err
	}

	return client.ShowOk(ctx)
}

// ApplicationDidLaunchHandler applicationDidLaunch handler
func ApplicationDidLaunchHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.ApplicationDidLaunchPayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		client.LogMessage(fmt.Sprintln("Failed to unmarshal ApplicationDidLaunchPayload payload:", err))
		return err
	}
	client.LogMessage(fmt.Sprintf("ApplicationDidLaunchHandler:%s\n", p))
	if p.Application == "vMix64.exe" || p.Application == "vMix.exe" {
		vMixLaunched = true
	}
	return nil
}

// ApplicationDidTerminateHandler applicationDidTerminate handler
func ApplicationDidTerminateHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.ApplicationDidTerminatePayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		client.LogMessage(fmt.Sprintln("Failed to unmarshal ApplicationDidTerminatePayload payload:", err))
		return err
	}
	client.LogMessage(fmt.Sprintln("ApplicationDidTerminateHandler:", p))
	if p.Application == "vMix64.exe" || p.Application == "vMix.exe" {
		vMixLaunched = false
		vMix.Close()
		vMix = nil
	}

	return nil
}

// DidReceiveSettingsHandler didReceiveSettings Handler
func DidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.DidReceiveSettingsPayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		client.LogMessage(fmt.Sprintln("Failed to unmarshal DidReceiveSettingsPayload payload:", err))
		return err
	}

	s := &PropertyInspector{}
	if err := json.Unmarshal(p.Settings, s); err != nil {
		client.LogMessage(fmt.Sprintln("Failed to unmarshal PropertyInspector:", err))
		return err
	}
	s.Inputs = inputs
	client.LogMessage(fmt.Sprintf("DidReceiveSettingsHandler:%v\n", s))
	settings.Store(event.Context, s)
	return nil
}
