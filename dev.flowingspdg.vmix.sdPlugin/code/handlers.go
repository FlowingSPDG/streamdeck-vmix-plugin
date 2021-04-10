package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/FlowingSPDG/streamdeck"
)

var (
	// DefaultvMixAPIURL vMix default API URL
	DefaultvMixAPIURL *url.URL
)

func init() {
	u, err := url.Parse("http://localhost:8088/api")
	if err != nil {
		panic(err)
	}
	DefaultvMixAPIURL = u
	log.Println("init complete.")
}

// WillAppearHandler willAppear handler.
func WillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.WillAppearPayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}
	log.Println("WillAppearHandler:", p)

	// s はボタンの設定オブジェクトのポインタ(変更すると直接反映される)
	s, ok := settings[event.Context]
	if !ok {
		// 存在しなかった場合に初期化
		settings[event.Context] = &PropertyInspector{}
		s = settings[event.Context]
	}
	log.Println("settings for this context:", s)
	// Settingのデータをsに反映
	if err := json.Unmarshal(p.Settings, s); err != nil {
		return err
	}
	return client.SetSettings(ctx, s)
}

// WillDisappearHandler willDisappear handler
func WillDisappearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	log.Println("WillDisappearHandler")
	settings[event.Context] = &PropertyInspector{}
	log.Println("settings for this context:", settings[event.Context])
	return client.SetSettings(ctx, settings[event.Context])
}

// KeyDownHandler keyDown handler
func KeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	log.Println("KeyDownHandler")
	s, ok := settings[event.Context]
	if !ok {
		return fmt.Errorf("couldn't find settings for context %v", event.Context)
	}
	log.Println("settings for this context:", s)

	u, err := s.GenerateURL()
	if err != nil {
		log.Println("ERR:", err)
		client.ShowAlert(ctx)
		return err
	}
	r, err := http.Get(u)
	if err != nil {
		log.Println("ERR:", err)
		client.ShowAlert(ctx)
		return err
	}
	defer r.Body.Close()

	return client.ShowOk(ctx)
}

// ApplicationDidLaunchHandler applicationDidLaunch handler
func ApplicationDidLaunchHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.ApplicationDidLaunchPayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}
	log.Println("ApplicationDidLaunchHandler:", p)
	// vMix64.exe?
	if p.Application == "vMix64.exe" {
		vMixLaunched = true
	}
	return nil
}

// ApplicationDidTerminateHandler applicationDidTerminate handler
func ApplicationDidTerminateHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.ApplicationDidTerminatePayload{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}
	log.Println("ApplicationDidTerminateHandler:", p)
	// vMix64.exe?
	if p.Application == "vMix64.exe" {
		vMixLaunched = false
	}

	return nil
}
