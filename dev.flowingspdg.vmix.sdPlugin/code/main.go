package main

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/FlowingSPDG/streamdeck"
	sdcontext "github.com/FlowingSPDG/streamdeck/context"

	vmixgo "github.com/FlowingSPDG/vmix-go"
)

const (
	// AppName Streamdeck plugin app name
	AppName = "dev.flowingspdg.vmix.sdPlugin"

	// Action Name
	Action = "dev.flowingspdg.vmix.function"
)

var (
	inputCache = make([]vmixgo.Input, 0, 200)

	vMixLaunched = false
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
	contexts := make(map[string]struct{})

	client.RegisterNoActionHandler(streamdeck.ApplicationDidLaunch, ApplicationDidLaunchHandler)
	client.RegisterNoActionHandler(streamdeck.ApplicationDidTerminate, ApplicationDidTerminateHandler)

	action := client.Action(Action)

	action.RegisterHandler(streamdeck.WillAppear, WillAppearHandler)
	action.RegisterHandler(streamdeck.WillAppear, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		contexts[event.Context] = struct{}{}
		return nil
	})

	action.RegisterHandler(streamdeck.WillDisappear, WillDisappearHandler)
	action.RegisterHandler(streamdeck.WillDisappear, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		delete(contexts, event.Context)
		return nil
	})
	action.RegisterHandler(streamdeck.KeyDown, KeyDownHandler)

	action.RegisterHandler(streamdeck.DidReceiveSettings, DidReceiveSettingsHandler)
	action.RegisterHandler(streamdeck.SendToPlugin, SendToPluginHandler)

	go func() {
		for range time.Tick(time.Second / 2) {
			if !vMixLaunched {
				continue
			}
			inputs, err := getvMixInputs()
			if err != nil {
				log.Println("Failed to get vMix inputs :", err)
				continue
			}

			// if !reflect.DeepEqual(inputs, inputCache) {
			settings.inputs = inputs

			for ctxStr := range contexts {
				ctx := context.Background()
				ctx = sdcontext.WithContext(ctx, ctxStr)

				if err := client.SendToPropertyInspector(ctx, PropertyInspector{
					Inputs: inputs,
				}); err != nil {
					log.Println("Failed to set global settings :", err)
					continue
				}

				/*
					if err := client.SetTitle(ctx, time.Now().String(), streamdeck.HardwareAndSoftware); err != nil {
						log.Println("Failed to set set title :", err)
						continue
					}
				*/
			}

			// }
			inputCache = inputs
		}
	}()
}
