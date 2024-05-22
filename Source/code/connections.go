package stdvmix

import (
	"context"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/FlowingSPDG/streamdeck"
	sdcontext "github.com/FlowingSPDG/streamdeck/context"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
	"github.com/puzpuzpuz/xsync/v3"
)

type vMixConnections struct {
	// TODO: StdVmixに処理を纏めることを検討する
	logger            *log.Logger
	sd                *streamdeck.Client
	activatorContexts *activatorContexts
	// TODO: まとめる?
	connections *xsync.MapOf[string, vmixtcp.Vmix]        // key:dest value:vmix
	sdContexts  *xsync.MapOf[string, map[string]struct{}] // key:dest value:sdcontexts(key only) TODO: xsync.MapOf[string, xsync.MapOf[string, struct{}]]
	vmInputs    *xsync.MapOf[string, []Input]             // key:dest value:inputs
}

func newVMixConnections(logger *log.Logger, sd *streamdeck.Client) *vMixConnections {
	return &vMixConnections{
		logger:            logger,
		sd:                sd,
		activatorContexts: newActivatorContexts(),
		connections:       xsync.NewMapOf[string, vmixtcp.Vmix](),
		sdContexts:        xsync.NewMapOf[string, map[string]struct{}](),
		vmInputs:          xsync.NewMapOf[string, []Input](),
	}
}

func (vc *vMixConnections) newVmix(ctx context.Context, dest string) error {
	// 既に接続済みの場合は何もしない
	if _, ok := vc.connections.Load(dest); ok {
		return nil
	}
	vc.logger.Printf("Connecting to vMix instance. dest: %s\n", dest)

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
		// vc.logger.Printf("Acts: %s\n", resp.Response)
		s := strings.Split(resp.Response, " ")
		if len(s) != 3 {
			return
		}

		// parse input number
		activeInputNumber, err := strconv.Atoi(s[1])
		if err != nil {
			// Some Activator response is in float32 etc. So just ignore it.
			return
		}
		activatorName := s[0]
		isActive := s[2] == "1"

		vc.activatorContexts.contextKeys.Range(func(key string, c activatorContext) bool {
			if c.destination != dest || c.input != activeInputNumber || c.activatorName != activatorName {
				return true
			}
			// vc.logger.Printf("Processing tally for PI: %s input:%d destination:%s activator:%s \n", key, activeInputNumber, dest, activatorName)
			sdctx := sdcontext.WithContext(ctx, key)
			tallyColor := tallyInactive
			switch c.activatorColor {
			case activatorColorRed:
				tallyColor = tallyProgram
			case activatorColorGreen:
				tallyColor = tallyPreview
			}
			if isActive {
				go vc.sd.SetImage(sdctx, tallyColor, streamdeck.HardwareAndSoftware)
			} else {
				go vc.sd.SetImage(sdctx, tallyInactive, streamdeck.HardwareAndSoftware)
			}
			return true
		})

		// Call XML to retrieve latest input list
		vmix.XML()
	})

	vmix.OnXML(func(xml *vmixtcp.XMLResponse) {
		// vc.logger.Printf("Processing XML for %s\n", dest)

		// Initialize input slice
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
		vc.vmInputs.Store(dest, inputs)

		ctxStrs, ok := vc.sdContexts.Load(dest)
		if !ok {
			vc.logger.Printf("No contexts for %s\n", dest)
			return
		}
		// vc.logger.Printf("Processing %d contexts keys with %d inputs.\n", len(ctxStrs), len(inputs))

		for ctxStr := range ctxStrs {
			// 多重送信になるか？
			sdctx := sdcontext.WithContext(ctx, ctxStr)
			if err := vc.sd.SendToPropertyInspector(sdctx, SendToPropertyInspectorPayload[InputsForPI]{
				Event: "inputs",
				Payload: InputsForPI{
					Inputs: map[string][]Input{
						dest: inputs,
					},
				},
			}); err != nil {
				vc.logger.Printf("Failed to set global settings. dest: %s, err: %v\n", dest, err)
			}
		}
	})

	vc.logger.Printf("Store new vmix client: %s\n", dest)
	vc.connections.Store(dest, vmix)

	vc.logger.Printf("Running new vmix client: %s\n", dest)

	if err := vmix.Connect(); err != nil {
		vc.logger.Printf("Failed to connect to vMix instance. dest: %s, err: %v\n", dest, err)
		return err
	}
	go vmix.Run(ctx)
	vc.logger.Printf("Successfully added new vmix client: %s\n", dest)

	return nil
}

// storeNewVmix stores new vmix client.
func (vc *vMixConnections) storeNewVmix(ctx context.Context, dest string) error {
	vc.newVmix(ctx, dest)
	return nil
}

func (vc *vMixConnections) storeNewCtxstr(dest, ctxStr string) error {
	contexts, _ := vc.sdContexts.LoadOrStore(dest, map[string]struct{}{
		ctxStr: {},
	})
	contexts[ctxStr] = struct{}{}
	vc.sdContexts.Store(dest, contexts)
	return nil
}

func (vc *vMixConnections) deleteDestination(dest string) {
	// Close connection?
	vc.connections.Delete(dest)
	vc.sdContexts.Delete(dest)
	vc.vmInputs.Delete(dest)
}

func (vc *vMixConnections) unregisterDestinationForCtx(ctxStr string) error {
	vc.sdContexts.Range(func(dest string, ctxStrs map[string]struct{}) bool {
		// 子がいない場合は削除
		if len(ctxStrs) == 0 {
			vc.deleteDestination(dest)
			return true
		}
		for c := range ctxStrs {
			if c != ctxStr {
				continue
			}

			vc.logger.Printf("Delete context: %s current contexts length:%d\n", ctxStr, len(ctxStrs))
			// 削除対象なのでmapからdeleteする
			delete(ctxStrs, ctxStr)

			// 指定したdestを使っているのが1contextしかいない場合、vMixの接続自体を削除する
			if len(ctxStrs) == 0 {
				vc.logger.Printf("Deleting vMix connection: %s\n", dest)
				vc.deleteDestination(dest)
			}
			continue

		}
		return true
	})
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

// UpdateVMixes updates vmix clients.
func (vc *vMixConnections) UpdateVMixes() {
	// vc.logger.Printf("Updating %d vMix instances.\n", vc.connections.Size())
	wg := &sync.WaitGroup{}
	vc.connections.Range(func(dest string, value vmixtcp.Vmix) bool {
		// ctxs, _ := vc.sdContexts.Load(dest)
		// vc.logger.Printf("Updating vMix instance: %s for contexts:%v\n", dest, ctxs)
		// 再接続処理
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !value.IsConnected() {
				if err := value.Connect(); err != nil {
					vc.logger.Printf("Failed to reconnect to vMix instance. dest: %s, err: %v Retry on next update.\n", dest, err)
				}
			}
		}()
		return true
	})
	wg.Wait()
}
