package main

import (
	"context"
	"encoding/xml"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/FlowingSPDG/streamdeck"
	sdcontext "github.com/FlowingSPDG/streamdeck/context"
	"github.com/FlowingSPDG/vmix-go/common/models"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
)

const (
	// AppName Streamdeck plugin app name
	AppName = "dev.flowingspdg.vmix.sdPlugin"

	// Action Name
	Action = "dev.flowingspdg.vmix.function"

	// tally color
	tallyInactive string = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAgAAAAIACAYAAAD0eNT6AAAIZUlEQVR4nOzWQQ0AIRDAwMsF4RjfBBfw6IyCPrtmZn8AQMr/OgAAuM8AAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAgyAAAQZAAAIMgAAECQAQCAIAMAAEEGAACCDAAABBkAAAg6AQAA//8SUgd661coNQAAAABJRU5ErkJggg=="
	tallyPreview  string = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAgAAAAIACAIAAAB7GkOtAAAHIklEQVR4nOzVMREAIAzAwB6Hf8sgo0P+FWTLnTcABJ3tAAB2GABAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBEGQBAlAEARBkAQJQBAEQZAECUAQBE/QAAAP//IpYFAvc6O0oAAAAASUVORK5CYII="
	tallyProgram  string = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAgAAAAIACAIAAAB7GkOtAAAHIUlEQVR4nOzVMREAIAzAQI7Dv2Uqo0P+FWTL+weAorsdAMAOAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAIAoAwCIMgCAKAMAiDIAgCgDAIgyAICoCQAA//8jlQUCXOd0pgAAAABJRU5ErkJggg=="
)

var (
	vMixLaunched = false

	shouldUpdate bool

	vMix *vmixtcp.Vmix
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
	setupClient(client)

	go func() {
		for err := vMixGoroutine(ctx); err != nil; {
			time.Sleep(time.Second)
			if !vMixLaunched {
				continue
			}
			log.Println("RETRY")
			err = vMixGoroutine(ctx)
		}
	}()

	return client.Run()
}

func vMixGoroutine(ctx context.Context) error {
	// reconnect
	var err error
	vMix, err = vmixtcp.New("localhost")
	if err != nil {
		return err
	}

	// re-subscribe
	if err = vMix.SUBSCRIBE(vmixtcp.EVENT_TALLY, ""); err != nil {
		return err
	}

	// We use Tally for checking input added or deleted.
	vMix.Register(vmixtcp.EVENT_TALLY, func(r *vmixtcp.Response) {
		log.Println("TALLY updated. Refreshing... ", r)
		if err := vMix.XML(); err != nil {
			log.Println("Failed to send XMLPATH:", err)
		}
	})

	// If we receive XMLTEXT...
	vMix.Register(vmixtcp.EVENT_XML, func(r *vmixtcp.Response) {
		// log.Println("XML response received:", r)
		x := models.APIXML{}
		if err := xml.Unmarshal([]byte(r.Response), &x); err != nil {
			log.Println("Failed to unmarshal XML:", err)
		}
		newinputs := make([]input, len(x.Inputs.Input))
		for k, v := range x.Inputs.Input {
			num, _ := strconv.Atoi(v.Number)
			newinputs[k] = input{
				Name:         v.Text,
				Key:          v.Key,
				Number:       num,
				TallyPreview: x.Preview == v.Number,
				TallyProgram: x.Active == v.Number,
			}
		}
		settings.Inputs = newinputs
		shouldUpdate = true
	})
	// timeout
	time.Sleep(time.Second)

	// run
	return vMix.Run(ctx)
}

func setupClient(client *streamdeck.Client) {
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
		for {
			// sleep 100ms
			time.Sleep(time.Millisecond * 100)
			if !vMixLaunched {
				continue
			}
			if !shouldUpdate {
				continue
			}

			log.Printf("Should update %d contexts\n", len(contexts))
			wg := &sync.WaitGroup{}
			for ctxStr := range contexts {
				wg.Add(1)
				go func() {
					defer wg.Done()
					ctx := context.Background()
					ctx = sdcontext.WithContext(ctx, ctxStr)

					// get settings
					p, err := settings.Load(ctxStr)
					if err != nil {
						log.Println("Failed to get PI settings :", err)
						return
					}

					p.Inputs = settings.Inputs
					if err := client.SendToPropertyInspector(ctx, p); err != nil {
						log.Println("Failed to send inputs to PI :", err)
						return
					}
					settings.Save(ctxStr, p)
					client.SetSettings(ctx, *p)

					// If tally disabled
					if !p.UseTallyPreview && !p.UseTallyProgram {
						return
					}

					var tallyPRV bool
					var tallyPGM bool
					for _, v := range settings.Inputs {
						if p.FunctionInput == v.Key {
							tallyPRV = v.TallyPreview
							tallyPGM = v.TallyProgram
							break
						}
					}

					// Only PRV
					if p.UseTallyPreview && !p.UseTallyProgram {
						if tallyPRV {
							if err := client.SetImage(ctx, tallyPreview, streamdeck.HardwareAndSoftware); err != nil {
								log.Println("Failed to set image :", err)
								return
							}
						} else {
							if err := client.SetImage(ctx, tallyInactive, streamdeck.HardwareAndSoftware); err != nil {
								log.Println("Failed to set image :", err)
								return
							}
						}
					} else if p.UseTallyProgram && !p.UseTallyPreview { // Only PGM
						if tallyPGM {
							if err := client.SetImage(ctx, tallyProgram, streamdeck.HardwareAndSoftware); err != nil {
								log.Println("Failed to set image :", err)
								return
							}
						} else {
							if err := client.SetImage(ctx, tallyInactive, streamdeck.HardwareAndSoftware); err != nil {
								log.Println("Failed to set image :", err)
								return
							}
						}
					} else if p.UseTallyProgram && p.UseTallyPreview { // Both
						// Inactive
						if !tallyPRV && !tallyPGM {
							if err := client.SetImage(ctx, tallyInactive, streamdeck.HardwareAndSoftware); err != nil {
								log.Println("Failed to set image :", err)
								return
							}
						} else if tallyPRV && !tallyPGM { // Preview
							if err := client.SetImage(ctx, tallyPreview, streamdeck.HardwareAndSoftware); err != nil {
								log.Println("Failed to set image :", err)
								return
							}
						} else if tallyPGM { // Program
							if err := client.SetImage(ctx, tallyProgram, streamdeck.HardwareAndSoftware); err != nil {
								log.Println("Failed to set image :", err)
								return
							}
						}
					}
				}()
				wg.Wait()
			}
			shouldUpdate = false
		}
	}()
}
