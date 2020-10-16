package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	vmixgo "github.com/FlowingSPDG/vmix-go"
	"github.com/samwho/streamdeck"
)

const (
	actionPreview = "dev.flowingspdg.vmix.preview"
	actionProgram = "dev.flowingspdg.vmix.program"
	actionCustom  = "dev.flowingspdg.vmix.custom"
)

var (
	inputs []vmixgo.Input
)

// PreviewPI Property Inspector JSON structure for "Preview" action.
type PreviewPI struct {
	InputKey string `json:"inputKey,omitempty"`
}

// ProgramPI Property Inspector JSON structure for "Program" action.
type ProgramPI struct {
	InputKey string `json:"inputKey,omitempty"`
}

func main() {
	f, err := ioutil.TempFile("", "streamdeck-vmix.log")
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

func sendFunction(name string, params map[string]string) error {
	v, err := vmixgo.NewVmix("localhost")
	if err != nil {
		return err
	}
	return v.SendFunction(name, params)
}

func getInputs() error {
	v, err := vmixgo.NewVmix("localhost")
	if err != nil {
		return err
	}
	inputs = v.Inputs.Input
	return nil
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
		return sendFunction("SetPreview", map[string]string{
			"input": prevPI.InputKey,
		})
	})

	prev.RegisterHandler(streamdeck.WillDisappear, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		delete(contexts, event.Context)
		return nil
	})
}
