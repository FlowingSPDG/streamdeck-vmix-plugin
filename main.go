package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/samwho/streamdeck"

	vmixgo "github.com/FlowingSPDG/vmix-go"
)

const (
	actionPreview = "dev.flowingspdg.vmix.preview"
	actionProgram = "dev.flowingspdg.vmix.program"
	actionCustom  = "dev.flowingspdg.vmix.custom"
)

var (
	vmix *vmixgo.Vmix
)

// PreviewPI Property Inspector JSON structure for "Preview" action.
type PreviewPI struct {
	InputKey string `json:"inputKey,omitempty"`
}

// ProgramPI Property Inspector JSON structure for "Program" action.
type ProgramPI struct {
	InputKey string `json:"inputKey,omitempty"`
}

func init() {
	var err error
	vmix, err = vmixgo.NewVmix("localhost")
	if err != nil {
		log.Println("Failed to initialize vMix. vmix=nil.")
	}
}

func main() {
	f, err := ioutil.TempFile("/tmp", "streamdeck-vmix.log")
	if err != nil {
		log.Fatalf("error creating tempfile: %v", err)
	}
	defer f.Close()

	log.Println("Log output path :", f.Name())
	log.SetOutput(f)

	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatalf("%v\n", err)
	}
}

func retryVmix() error {
	var err error
	vmix, err = vmixgo.NewVmix("localhost")
	return err
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
	prev := client.Action(actionPreview)

	prevPI := &PreviewPI{}
	//programPI := &ProgramPI{}
	contexts := make(map[string]struct{})

	prev.RegisterHandler(streamdeck.SendToPlugin, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		return json.Unmarshal(event.Payload, prevPI)
	})

	prev.RegisterHandler(streamdeck.WillAppear, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		contexts[event.Context] = struct{}{}
		return nil
	})

	prev.RegisterHandler(streamdeck.KeyDown, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		if vmix == nil {
			if err := retryVmix(); err != nil {
				return err
			}
		}
		return vmix.SendFunction("SetPreview", map[string]string{
			"input": prevPI.InputKey,
		})
	})

	prev.RegisterHandler(streamdeck.WillDisappear, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		delete(contexts, event.Context)
		return nil
	})
}
