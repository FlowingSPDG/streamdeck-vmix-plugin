package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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
	s.Inputs = settings.Inputs
	// log.Printf("WillAppearHandler %s:%v\n", event.Context, s)

	settings.Save(event.Context, &s)
	client.SetSettings(ctx, s)

	if err := client.SendToPropertyInspector(ctx, s); err != nil {
		log.Println("Failed to send PI settings :", err)
		return err
	}

	// log.Printf("settings for context %s context:%#v\n", event.Context, s)
	return nil
}

// WillDisappearHandler willDisappear handler
func WillDisappearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	log.Println("WillDisappearHandler")
	settings.Save(event.Context, &PropertyInspector{})
	log.Println("Refreshing settings for this context:", event.Context)
	s, _ := settings.Load(event.Context)
	return client.SetSettings(ctx, s)
}

// KeyDownHandler keyDown handler
func KeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	log.Println("KeyDownHandler")
	s, err := settings.Load(event.Context)
	if err != nil {
		return fmt.Errorf("couldn't find settings for context %v", event.Context)
	}
	// log.Println("settings for this context:", s)

	if !vMixLaunched {
		return client.ShowAlert(ctx)
	}

	query, err := s.GenerateFunction()
	if err != nil {
		log.Println("ERR:", err)
		client.ShowAlert(ctx)
		return err
	}
	log.Println("Generated Query:", query)
	if err := vMix.FUNCTION(query); err != nil {
		log.Println("ERR:", err)
		client.ShowAlert(ctx)
		return err
	}

	return client.ShowOk(ctx)
}

// ApplicationDidLaunchHandler applicationDidLaunch handler
func ApplicationDidLaunchHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.ApplicationDidLaunchPayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		log.Println("ERR:", err)
		return err
	}
	log.Println("ApplicationDidLaunchHandler:", p)
	if p.Application == "vMix64.exe" {
		vMixLaunched = true
	}
	return nil
}

// ApplicationDidTerminateHandler applicationDidTerminate handler
func ApplicationDidTerminateHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.ApplicationDidTerminatePayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		log.Println("ERR:", err)
		return err
	}
	log.Println("ApplicationDidTerminateHandler:", p)
	if p.Application == "vMix64.exe" {
		vMixLaunched = false
	}

	return nil
}

// DidReceiveSettingsHandler didReceiveSettings Handler
func DidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.DidReceiveSettingsPayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		log.Println("ERR:", err)
		return err
	}
	// log.Println("DidReceiveSettingsHandler:", p)

	s := &PropertyInspector{}
	if err := json.Unmarshal(p.Settings, s); err != nil {
		log.Println("ERR:", err)
		return err
	}
	s.Inputs = settings.Inputs
	settings.Save(event.Context, s)

	return nil
}

// SendToPluginHandler SendToPlugin Handler
func SendToPluginHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	s := PropertyInspector{}
	if err := json.Unmarshal(event.Payload, &s); err != nil {
		log.Println("ERR:", err)
		return err
	}
	log.Println("SendToPluginHandler:", s)

	// If PI disabled tally completely
	if !s.UseTallyPreview && !s.UseTallyProgram {
		client.SetImage(ctx, "", streamdeck.HardwareAndSoftware)
	} else {
		var tallyPRV bool
		var tallyPGM bool
		for _, v := range s.Inputs {
			if s.FunctionInput == v.Key {
				tallyPRV = v.TallyPreview
				tallyPGM = v.TallyProgram
				break
			}
		}

		// Only PRV
		if s.UseTallyPreview && !s.UseTallyProgram {
			if tallyPRV {
				if err := client.SetImage(ctx, tallyPreview, streamdeck.HardwareAndSoftware); err != nil {
					log.Println("Failed to set image :", err)
					return err
				}
			} else {
				if err := client.SetImage(ctx, tallyInactive, streamdeck.HardwareAndSoftware); err != nil {
					log.Println("Failed to set image :", err)
					return err
				}
			}
		} else if s.UseTallyProgram && !s.UseTallyPreview { //// Only PGM
			if tallyPGM {
				if err := client.SetImage(ctx, tallyProgram, streamdeck.HardwareAndSoftware); err != nil {
					log.Println("Failed to set image :", err)
					return err
				}
			} else {
				if err := client.SetImage(ctx, tallyInactive, streamdeck.HardwareAndSoftware); err != nil {
					log.Println("Failed to set image :", err)
					return err
				}
			}
		} else if s.UseTallyProgram && s.UseTallyPreview { // Both
			// Inactive
			if !tallyPRV && !tallyPGM {
				if err := client.SetImage(ctx, tallyInactive, streamdeck.HardwareAndSoftware); err != nil {
					log.Println("Failed to set image :", err)
					return err
				}
			} else if tallyPRV && !tallyPGM { // Preview only
				if err := client.SetImage(ctx, tallyPreview, streamdeck.HardwareAndSoftware); err != nil {
					log.Println("Failed to set image :", err)
					return err
				}
			} else if tallyPGM { // Program
				if err := client.SetImage(ctx, tallyProgram, streamdeck.HardwareAndSoftware); err != nil {
					log.Println("Failed to set image :", err)
					return err
				}
			}
		}
	}

	settings.Save(event.Context, &s)
	return client.SetSettings(ctx, s)
}
