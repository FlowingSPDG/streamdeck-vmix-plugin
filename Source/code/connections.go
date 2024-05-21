package stdvmix

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/FlowingSPDG/streamdeck"
	sdcontext "github.com/FlowingSPDG/streamdeck/context"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
	"github.com/puzpuzpuz/xsync/v3"
)

type vMixConnections struct {
	logger         *log.Logger
	sd             *streamdeck.Client
	connections    *xsync.MapOf[string, vmixtcp.Vmix]
	inputs         *xsync.MapOf[string, []Input]
	previewTallies *xsync.MapOf[int, string] // key:InputNumber value:ContextString
	programTallies *xsync.MapOf[int, string] // key:InputNumber value:ContextString
}

func newVMixConnections(logger *log.Logger, sd *streamdeck.Client) *vMixConnections {
	return &vMixConnections{
		logger:         logger,
		sd:             sd,
		connections:    xsync.NewMapOf[string, vmixtcp.Vmix](),
		inputs:         xsync.NewMapOf[string, []Input](),
		previewTallies: xsync.NewMapOf[int, string](),
		programTallies: xsync.NewMapOf[int, string](),
	}
}

func (vc *vMixConnections) newVmix(ctx context.Context, dest string) error {
	// 既に接続済みの場合は何もしない
	if _, ok := vc.connections.Load(dest); ok {
		return nil
	}

	// Initiate
	vmix := vmixtcp.New(dest)
	vmix.OnVersion(func(resp *vmixtcp.VersionResponse) {
		vc.logger.Printf("Connected. Version: %s\n", resp.Version)
		if err := vmix.XML(); err != nil {
			vc.logger.Printf("Failed to send XML() %v\n", err)
		}
		if err := vmix.Subscribe(vmixtcp.EventActs, ""); err != nil {
			vc.logger.Printf("Failed to send Acts() %v\n", err)
		}
	})
	vmix.OnActs(func(resp *vmixtcp.ActsResponse) {
		vc.logger.Printf("Acts: %s\n", resp.Response)
		s := strings.Split(resp.Response, " ")
		if len(s) != 3 {
			return
		}

		// parse input number
		activeInputNumber, _ := strconv.Atoi(s[1])
		isActive := s[2] == "1"
		vc.logger.Printf("Processing tallies for %d PGM contexts, %d PRV contexts\n", vc.programTallies.Size(), vc.previewTallies.Size())

		// TODO: support multiple activators
		switch s[0] {
		case "Input":
			vc.programTallies.Range(func(inputNum int, ctxStr string) bool {
				if activeInputNumber != inputNum {
					return true
				}
				sdctx := sdcontext.WithContext(ctx, ctxStr)
				if isActive {
					go vc.sd.SetImage(sdctx, tallyProgram, streamdeck.HardwareAndSoftware)
				} else {
					go vc.sd.SetImage(sdctx, tallyInactive, streamdeck.HardwareAndSoftware)
				}
				return true
			})
		case "InputPreview":
			vc.previewTallies.Range(func(inputNum int, ctxStr string) bool {
				if activeInputNumber != inputNum {
					return true
				}
				sdctx := sdcontext.WithContext(ctx, ctxStr)
				if isActive {
					go vc.sd.SetImage(sdctx, tallyPreview, streamdeck.HardwareAndSoftware)
				} else {
					go vc.sd.SetImage(sdctx, tallyInactive, streamdeck.HardwareAndSoftware)
				}
				return true
			})
		}

		vmix.XML()
	})

	vmix.OnXML(func(xml *vmixtcp.XMLResponse) {
		// Initialize input slice
		vc.logger.Printf("xml inputs: %#v\n", xml.XML.Inputs)
		inputs := make([]Input, 0, len(xml.XML.Inputs.Input))
		for _, i := range xml.XML.Inputs.Input {
			num, err := strconv.Atoi(i.Number)
			if err != nil {
				continue
			}
			inputs = append(inputs, Input{
				Name:   i.Title, // ?
				Key:    i.Key,
				Number: num,
			})
		}
		vc.inputs.Store(dest, inputs)
		vc.sd.SendToPropertyInspector(ctx, SendToPropertyInspectorPayload[InputsForPI]{
			Event: "inputs",
			Payload: InputsForPI{
				Inputs: inputs,
			},
		})
	})

	if err := vmix.Connect(); err != nil {
		panic(err) // TODO: 後で消す
	}

	vc.logger.Printf("Store new vmix client: %s\n", dest)
	vc.connections.Store(dest, vmix)

	vc.logger.Printf("Running new vmix client: %s\n", dest)

	go vmix.Run(ctx)
	vc.logger.Printf("Successfully added new vmix client: %s\n", dest)

	return nil
}

// storeNewVmix stores new vmix client.
func (vc *vMixConnections) storeNewVmix(ctx context.Context, dest string) error {
	vc.newVmix(ctx, dest)
	return nil
}

// Load loads vmix client.
func (vc *vMixConnections) load(dest string) (vmix vmixtcp.Vmix, ok bool) {
	vm, ok := vc.connections.Load(dest)
	if !ok {
		return nil, false
	}

	return vm, true
}

func (vc *vMixConnections) loadOrStore(ctx context.Context, dest string) (vmixtcp.Vmix, error) {
	vm, ok := vc.load(dest)
	if !ok {
		if err := vc.storeNewVmix(ctx, dest); err != nil {
			return nil, err
		}
		loaded, _ := vc.load(dest)
		return loaded, nil
	}
	return vm, nil
}

func (vc *vMixConnections) StorePreviewContext(inputNumber int, ctxStr string) {
	vc.previewTallies.Store(inputNumber, ctxStr)
}

func (vc *vMixConnections) DeletePreviewContext(inputNumber int) {
	vc.previewTallies.Delete(inputNumber)
}

func (vc *vMixConnections) StoreProgramContext(inputNumber int, ctxStr string) {
	vc.programTallies.Store(inputNumber, ctxStr)
}

func (vc *vMixConnections) DeleteProgramContext(inputNumber int) {
	vc.programTallies.Delete(inputNumber)
}

// UpdateVMixes updates vmix clients.
func (vc *vMixConnections) UpdateVMixes(ctx context.Context, activeVmixDests []string) (before, after int) {
	before = vc.connections.Size()
	vc.connections.Range(func(dest string, value vmixtcp.Vmix) bool {
		// どのContextにも紐づいていないvMixは削除する
		active := false
		for _, activeVmixDest := range activeVmixDests {
			if activeVmixDest == dest {
				active = true
			}
		}
		if !active {
			value.Close()
			vc.connections.Delete(dest)
			return true
		}
		go func() {
			if !value.IsConnected() {
				if err := value.Connect(); err != nil {
					// TODO: err
				}
			}
		}()
		return true
	})
	after = vc.connections.Size()
	return before, after
}
