package main

import (
	"context"
	"image"
	"image/color"
	"io"
	"log"
	"os"
	"time"

	"github.com/FlowingSPDG/streamdeck"
	sdcontext "github.com/FlowingSPDG/streamdeck/context"
)

const (
	// AppName Streamdeck plugin app name
	AppName = "dev.flowingspdg.vmix.sdPlugin"

	// Action Name
	Action = "dev.flowingspdg.vmix.function"
)

var (
	inputCache = make([]input, 0, 500)

	vMixLaunched = false

	tallyInactive string
	tallyPreview  string
	tallyProgram  string
)

func init() {
	// generate tally data
	width := 512
	height := 512
	img := image.NewRGBA(image.Rect(0, 0, 512, 512))

	// generate inactive(grey) tally
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 128})
		}
	}
	grey, err := streamdeck.Image(img)
	if err != nil {
		panic(err)
	}
	tallyInactive = grey

	// generate Preview(green) tally
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{0, 255, 0, 255})
		}
	}
	green, err := streamdeck.Image(img)
	if err != nil {
		panic(err)
	}
	tallyPreview = green

	// generate Program(red) tally
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	red, err := streamdeck.Image(img)
	if err != nil {
		panic(err)
	}
	tallyProgram = red
}

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
		for range time.Tick(time.Second / 4) {
			if !vMixLaunched {
				continue
			}
			inputs, err := getvMixInputs()
			if err != nil {
				log.Println("Failed to get vMix inputs :", err)
				continue
			}

			// Check if there is any update
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

				p, err := settings.Load(ctxStr)
				if err != nil {
					log.Println("Failed to get PI settings :", err)
					continue
				}

				var t tally
				for _, v := range inputs {
					if p.FunctionInput == v.Key {
						t = v.TallyState
					}
				}

				log.Printf("Set tally for context %s : %d\n", ctxStr, t)
				switch t {
				case Inactive:
					if err := client.SetImage(ctx, tallyInactive, streamdeck.HardwareAndSoftware); err != nil {
						log.Println("Failed to set image :", err)
						continue
					}
				case Preview:
					if err := client.SetImage(ctx, tallyPreview, streamdeck.HardwareAndSoftware); err != nil {
						log.Println("Failed to set image :", err)
						continue
					}

				case Program:
					if err := client.SetImage(ctx, tallyProgram, streamdeck.HardwareAndSoftware); err != nil {
						log.Println("Failed to set image :", err)
						continue
					}

				default:
					if err := client.ShowAlert(ctx); err != nil {
						log.Println("Failed to show alert :", err)
						continue
					}
				}

			}
			// }
			inputCache = inputs
		}
	}()
}
