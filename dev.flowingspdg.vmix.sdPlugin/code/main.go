package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"time"

	"github.com/FlowingSPDG/streamdeck"
)

const (
	// AppName Streamdeck plugin app name
	AppName = "dev.flowingspdg.vmix.sdPlugin"

	// Action Name
	Action = "dev.flowingspdg.vmix.function"
)

func main() {
	logfile, err := os.OpenFile("./streamdeck-vmix-plugin.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannnot open log:" + err.Error())
	}
	defer logfile.Close()
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))
	log.SetFlags(log.Ldate | log.Ltime)

	ctx := context.Background()
	log.Println("Starting...")
	if err := run(ctx); err != nil {
		log.Fatalf("%v\n", err)
	}
}

func run(ctx context.Context) error {
	params, err := streamdeck.ParseRegistrationParams(os.Args)
	if err != nil {
		return err
	}

	client := streamdeck.NewClient(ctx, params)
	setup(client)

	return client.Run()
}

func setup(client *streamdeck.Client) {
	action := client.Action(Action)

	action.RegisterHandler(streamdeck.WillAppear, WillAppearHandler)
	action.RegisterHandler(streamdeck.WillDisappear, WillDisappearHandler)
	action.RegisterHandler(streamdeck.KeyDown, KeyDownHandler)
	action.RegisterHandler(streamdeck.ApplicationDidLaunch, ApplicationDidLaunchHandler)
	action.RegisterHandler(streamdeck.SetSettings, SetSettingsHandler)

	go func() {
		for range time.Tick(time.Second / 2) {
			inputs, err := getvMixInputs()
			if err != nil {
				return
			}
			b, err := json.Marshal(inputs)
			if err != nil {
				return
			}
			client.SendToPropertyInspector(context.TODO(), string(b))
		}
	}()
}
