package action

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/FlowingSPDG/streamdeck"
	sdcontext "github.com/FlowingSPDG/streamdeck/context"
	models "github.com/FlowingSPDG/vmix-go"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
	"github.com/puzpuzpuz/xsync/v3"
	"github.com/samber/lo"
	"golang.org/x/xerrors"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/connection"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/setting"
)

const PreviewActionUUID = "dev.flowingspdg.vmix.preview"

type PreviewAction interface {
	OnWillAppear() streamdeck.EventHandler
	OnWillDisappear() streamdeck.EventHandler
	OnUpdateSettings() streamdeck.EventHandler
	OnKeyDown() streamdeck.EventHandler
	OnSendToPlugin() streamdeck.EventHandler
	OnVMixTally(ctx context.Context, resp *vmixtcp.TallyResponse, addr string, vm vmixtcp.Vmix) error
	OnVMixXML(ctx context.Context, resp *vmixtcp.XMLResponse, addr string, vm vmixtcp.Vmix) error
	OnVMixActs(ctx context.Context, resp *vmixtcp.ActsResponse, addr string, vm vmixtcp.Vmix) error
	OnVMixVersion(ctx context.Context, resp *vmixtcp.VersionResponse, addr string, vm vmixtcp.Vmix) error
	OnVMixSubscribe(ctx context.Context, resp *vmixtcp.SubscribeResponse, addr string, vm vmixtcp.Vmix) error
}

type previewAction struct {
	logger            logger.Logger
	connectionManager *connection.ConnectionManager
	store             setting.SettingStore[*setting.PreviewSetting]
	client            *streamdeck.Client
	inputCache        setting.SettingStore[[]*setting.Input]
	contextTallyMap   *xsync.MapOf[string, tallyStatus]
}

func NewPreviewAction(
	logger logger.Logger,
	connectionManager *connection.ConnectionManager,
	store setting.SettingStore[*setting.PreviewSetting],
	client *streamdeck.Client,
	inputCache setting.SettingStore[[]*setting.Input],
) PreviewAction {
	return &previewAction{
		logger:            logger,
		connectionManager: connectionManager,
		store:             store,
		client:            client,
		inputCache:        inputCache,
		contextTallyMap:   xsync.NewMapOf[string, tallyStatus](),
	}
}

func (p *previewAction) OnWillAppear() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		p.logger.Debug(ctx, "OnWillAppear started. contextID: %s", event.Context)
		defer p.logger.Debug(ctx, "OnWillAppear completed. contextID: %s", event.Context)

		payload := streamdeck.WillAppearPayload[*setting.PreviewSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			p.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			return nil
		}

		if payload.Settings.IsDefault() {
			p.logger.Error(ctx, "PreviewAction OnWillAppear settings is default")
			payload.Settings.Initialize()
		}

		p.store.Store(event.Context, payload.Settings)
		p.connectionManager.AddContext(ctx, payload.Settings.VMixAddress, event.Context, PreviewActionUUID, true)
		p.contextTallyMap.Store(event.Context, tallyStatusUnknown)

		p.logger.Debug(ctx, "previewAction OnWillAppear settings: %#v", payload.Settings)

		if err := p.client.SetSettings(ctx, *payload.Settings); err != nil {
			p.logger.Error(ctx, "Failed to set settings %v", err)
		}

		vmix := p.connectionManager.GetClient(ctx, payload.Settings.VMixAddress)
		if vmix == nil {
			p.logger.Error(ctx, "vMix connection not found on previewAction OnWillAppear event")
			p.connectionManager.AddVMix(ctx, payload.Settings.VMixAddress)
			return nil
		}

		switch payload.Settings.TallyMode {
		case setting.TallyModeTALLY:
			if err := vmix.Tally(); err != nil {
				p.logger.Error(ctx, "Failed to set tally", "error", err)
			}
		case setting.TallyModeACTS:
			funcName := "InputPreview"
			switch payload.Settings.Mix {
			case 0:
				funcName = "InputPreview"
			case 1:
				funcName = "InputPreviewMix2"
			case 2:
				funcName = "InputPreviewMix3"
			case 3:
				funcName = "InputPreviewMix4"
			case 4:
				funcName = "InputPreviewMix5"
			case 5:
				funcName = "InputPreviewMix6"
			case 6:
				funcName = "InputPreviewMix7"
			case 7:
				funcName = "InputPreviewMix8"
			case 8:
				funcName = "InputPreviewMix9"
			case 9:
				funcName = "InputPreviewMix10"
			case 10:
				funcName = "InputPreviewMix11"
			case 11:
				funcName = "InputPreviewMix12"
			case 12:
				funcName = "InputPreviewMix13"
			case 13:
				funcName = "InputPreviewMix14"
			case 14:
				funcName = "InputPreviewMix15"
			case 15:
				funcName = "InputPreviewMix16"
			}
			if err := vmix.Acts(funcName, &payload.Settings.Input); err != nil {
				p.logger.Error(ctx, "Failed to execute InputPreview", "error", err)
			}
		}

		// PropertyInspectorを更新
		if err := p.updatePropertyInspector(ctx, event); err != nil {
			return err
		}

		return nil
	}
}

func (p *previewAction) OnWillDisappear() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		p.logger.Debug(ctx, "OnWillDisappear started. contextID: %s", event.Context)
		defer p.logger.Debug(ctx, "OnWillDisappear completed. contextID: %s", event.Context)

		payload := streamdeck.WillDisappearPayload[setting.PreviewSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			p.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			return nil
		}

		p.connectionManager.RemoveContext(ctx, payload.Settings.VMixAddress, event.Context)
		p.store.Delete(event.Context)
		return nil
	}
}

func (p *previewAction) OnUpdateSettings() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		payload := streamdeck.DidReceiveSettingsPayload[setting.PreviewSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			p.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			return nil
		}

		p.logger.Debug(ctx, "OnUpdateSettings started. contextID: %s settings: %#v", event.Context, payload.Settings)
		defer p.logger.Debug(ctx, "OnUpdateSettings completed. contextID: %s", event.Context)

		oldSettings, ok := p.store.Load(event.Context)
		if !ok {
			p.logger.Error(ctx, "Failed to load old settings on programAction OnUpdateSettings event")
		}
		p.connectionManager.UpdateContext(ctx, oldSettings.VMixAddress, payload.Settings.VMixAddress, event.Context, PreviewActionUUID)
		p.store.Store(event.Context, &payload.Settings)

		vmix := p.connectionManager.GetClient(ctx, payload.Settings.VMixAddress)
		if vmix == nil {
			p.logger.Error(ctx, "vMix connection not found on previewAction OnUpdateSettings event")
			return nil
		}

		p.contextTallyMap.Store(event.Context, tallyStatusUnknown)

		switch payload.Settings.TallyMode {
		case setting.TallyModeTALLY:
			if err := vmix.Tally(); err != nil {
				p.logger.Error(ctx, "Failed to set tally", "error", err)
			}
		case setting.TallyModeACTS:
			funcName := "InputPreview"
			switch payload.Settings.Mix {
			case 0:
				funcName = "InputPreview"
			case 1:
				funcName = "InputPreviewMix2"
			case 2:
				funcName = "InputPreviewMix3"
			case 3:
				funcName = "InputPreviewMix4"
			case 4:
				funcName = "InputPreviewMix5"
			case 5:
				funcName = "InputPreviewMix6"
			case 6:
				funcName = "InputPreviewMix7"
			case 7:
				funcName = "InputPreviewMix8"
			case 8:
				funcName = "InputPreviewMix9"
			case 9:
				funcName = "InputPreviewMix10"
			case 10:
				funcName = "InputPreviewMix11"
			case 11:
				funcName = "InputPreviewMix12"
			case 12:
				funcName = "InputPreviewMix13"
			case 13:
				funcName = "InputPreviewMix14"
			case 14:
				funcName = "InputPreviewMix15"
			case 15:
				funcName = "InputPreviewMix16"
			}
			if err := vmix.Acts(funcName, &payload.Settings.Input); err != nil {
				p.logger.Error(ctx, "Failed to execute InputPreview", "error", err)
			}
		}

		// PropertyInspectorを更新
		if err := p.updatePropertyInspector(ctx, event); err != nil {
			return err
		}

		return nil
	}
}

func (p *previewAction) OnKeyDown() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		p.logger.Debug(ctx, "Execute started")
		defer p.logger.Debug(ctx, "Execute completed")

		payload := streamdeck.KeyDownPayload[setting.PreviewSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			p.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			return nil
		}

		p.logger.Debug(ctx, "OnKeyDown started. contextID: %s settings: %#v", event.Context, payload.Settings)

		// Get vMix client
		vmix := p.connectionManager.GetVMixByContext(ctx, event.Context)
		if vmix == nil {
			err := xerrors.Errorf("vMix connection not found on previewAction OnKeyDown event")
			p.logger.Error(ctx, "Failed to get vMix client: %v", err)
			return err
		}

		// Build query with Mix parameter
		query := fmt.Sprintf("Input=%d", payload.Settings.Input)
		if payload.Settings.Mix > 0 {
			query += fmt.Sprintf("&Mix=%d", payload.Settings.Mix)
		}

		// Execute PreviewInput function with query
		p.logger.Debug(ctx, "Executing PreviewInput with query: %s", query)
		if err := vmix.Function("PreviewInput", query); err != nil {
			p.logger.Error(ctx, "Failed to execute PreviewInput", "error", err)
			return xerrors.Errorf("failed to execute PreviewInput: %w", err)
		}

		p.logger.Info(ctx, "PreviewInput executed successfully: %d", payload.Settings.Input)
		return nil
	}
}

func (p *previewAction) OnSendToPlugin() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		p.logger.Debug(ctx, "OnSendToPlugin started")
		defer p.logger.Debug(ctx, "OnSendToPlugin completed")

		type CommandPayload struct {
			Event   string          `json:"event"`
			Payload json.RawMessage `json:"payload"`
		}

		var command CommandPayload
		if err := json.Unmarshal(event.Payload, &command); err != nil {
			p.logger.Error(ctx, "Failed to unmarshal command payload", "error", err)
			return err
		}

		p.logger.Info(ctx, "OnSendToPlugin received command: %s", command.Event)

		switch command.Event {
		case "property_inspector":
			if err := p.updatePropertyInspector(ctx, event); err != nil {
				return err
			}

		case "connect":
			type ConnectArgs struct {
				Host string `json:"host"`
			}
			var args ConnectArgs
			if err := json.Unmarshal(command.Payload, &args); err != nil {
				p.logger.Error(ctx, "Failed to unmarshal connect args", "error", err)
				return err
			}
			p.logger.Info(ctx, "Connect command received: %s", args.Host)

			p.connectionManager.AddVMix(ctx, args.Host)

			// PropertyInspectorを更新
			if err := p.updatePropertyInspector(ctx, event); err != nil {
				return err
			}

		case "disconnect":
			type DisconnectArgs struct {
				Host string `json:"host"`
			}
			var args DisconnectArgs
			if err := json.Unmarshal(command.Payload, &args); err != nil {
				p.logger.Error(ctx, "Failed to unmarshal disconnect args", "error", err)
				return err
			}
			p.logger.Info(ctx, "Disconnect command received: %s", args.Host)
			p.connectionManager.RemoveVMix(ctx, args.Host)

			// PropertyInspectorのdestinationsを更新
			destinations := p.connectionManager.GetAllVMixAddrs(ctx)
			sdctx := sdcontext.WithContext(ctx, event.Context)
			sdctx = sdcontext.WithAction(sdctx, event.Action)
			sdctx = sdcontext.WithDevice(sdctx, event.Device)
			if err := p.sendDestinations(sdctx, destinations); err != nil {
				p.logger.Error(ctx, "Failed to send destinations to PropertyInspector", "error", err)
				return err
			}

		default:
			p.logger.Error(ctx, "Unknown command received: %s", command.Event)
			return xerrors.Errorf("unknown command: %s", command.Event)
		}

		return nil
	}
}

func (p *previewAction) OnVMixTally(ctx context.Context, resp *vmixtcp.TallyResponse, addr string, vm vmixtcp.Vmix) error {
	p.logger.Debug(ctx, "OnVMixTally started")
	defer p.logger.Debug(ctx, "OnVMixTally completed")

	contextIDs := p.connectionManager.GetContextsByActionType(ctx, addr, PreviewActionUUID)
	p.logger.Debug(ctx, "OnVMixTally addr: %s contextIDs: %v resp: %v", addr, contextIDs, resp.Tally)
	for _, contextID := range contextIDs {
		sdctx := sdcontext.WithContext(ctx, contextID)
		s, ok := p.store.Load(contextID)
		if !ok {
			err := xerrors.Errorf("setting not found on previewAction OnVMixTally event")
			p.logger.Error(sdctx, "Failed to load setting: %v", err)
			continue
		}

		// タリー反映処理
		if s.TallyMode != setting.TallyModeTALLY {
			continue
		}
		if len(resp.Tally) <= s.Input {
			continue
		}

		tallyActive := resp.Tally[s.Input-1] == vmixtcp.Preview
		currentTallyStatus, _ := p.contextTallyMap.LoadOrStore(contextID, tallyStatusUnknown)
		shouldUpdate := true

		shouldUpdate = !(currentTallyStatus == tallyStatusOff && !tallyActive || currentTallyStatus == tallyStatusOn && tallyActive)

		p.logger.Debug(sdctx, "OnVMixTally contextID: %s shouldUpdate: %v", contextID, shouldUpdate)

		if !shouldUpdate {
			continue
		}
		p.logger.Debug(sdctx, "Going to apply tally. setting: %v", s)
		tallyImage := tallyInactive
		currentTallyStatus = tallyStatusOff
		if tallyActive {
			tallyImage = tallyPreview
			currentTallyStatus = tallyStatusOn
		}
		p.contextTallyMap.Store(contextID, currentTallyStatus)
		p.client.SetImage(sdctx, tallyImage, streamdeck.HardwareAndSoftware)
	}

	return nil
}

func (p *previewAction) OnVMixActs(ctx context.Context, resp *vmixtcp.ActsResponse, addr string, vm vmixtcp.Vmix) error {
	p.logger.Debug(ctx, "OnVMixActs started")
	defer p.logger.Debug(ctx, "OnVMixActs completed")

	contextIDs := p.connectionManager.GetContextsByActionType(ctx, addr, PreviewActionUUID)
	p.logger.Debug(ctx, "OnVMixActs addr: %s contextIDs: %v resp: %v", addr, contextIDs, resp.Response)
	for _, contextID := range contextIDs {
		sdctx := sdcontext.WithContext(ctx, contextID)
		s, ok := p.store.Load(contextID)
		if !ok {
			err := xerrors.Errorf("setting not found on previewAction OnVMixActs event")
			p.logger.Error(sdctx, "Failed to load setting: %v", err)
			continue
		}

		p.logger.Debug(sdctx, "OnVMixActs applying tally: contextID: %s setting: %#v", contextID, s)

		// タリー反映処理
		if s.TallyMode != setting.TallyModeACTS {
			continue
		}

		acts := strings.Split(resp.Response, " ")
		if len(acts) != 3 {
			continue
		}
		event := acts[0]
		inputNumber := acts[1]
		isActive := acts[2] == "1"

		if inputNumber != strconv.Itoa(s.Input) {
			continue
		}

		isCorrectEvent := false
		switch event {
		case "InputPreview":
			isCorrectEvent = s.Mix == 0
		case "InputPreviewMix2":
			isCorrectEvent = s.Mix == 1
		case "InputPreviewMix3":
			isCorrectEvent = s.Mix == 2
		case "InputPreviewMix4":
			isCorrectEvent = s.Mix == 3
		case "InputPreviewMix5":
			isCorrectEvent = s.Mix == 4
		case "InputPreviewMix6":
			isCorrectEvent = s.Mix == 5
		case "InputPreviewMix7":
			isCorrectEvent = s.Mix == 6
		case "InputPreviewMix8":
			isCorrectEvent = s.Mix == 7
		case "InputPreviewMix9":
			isCorrectEvent = s.Mix == 8
		case "InputPreviewMix10":
			isCorrectEvent = s.Mix == 9
		case "InputPreviewMix11":
			isCorrectEvent = s.Mix == 10
		case "InputPreviewMix12":
			isCorrectEvent = s.Mix == 11
		case "InputPreviewMix13":
			isCorrectEvent = s.Mix == 12
		case "InputPreviewMix14":
			isCorrectEvent = s.Mix == 13
		case "InputPreviewMix15":
			isCorrectEvent = s.Mix == 14
		case "InputPreviewMix16":
			isCorrectEvent = s.Mix == 15

		default:
			continue
		}

		if !isCorrectEvent {
			continue
		}

		// tally cacheを使用する
		// Tally stateとcached stateが一致していれば更新しない
		// Unknownであれば関係なく更新
		currentTallyStatus, _ := p.contextTallyMap.LoadOrStore(contextID, tallyStatusUnknown)
		if (currentTallyStatus == tallyStatusOff && !isActive) || (currentTallyStatus == tallyStatusOn && isActive) {
			p.logger.Debug(sdctx, "Tally state and cached state are the same. skipping. currentTallyStatus: %v isActive: %v", currentTallyStatus, isActive)
			continue
		}

		p.logger.Debug(sdctx, "Going to apply tally. setting: %#v", s)
		tallyImage := tallyInactive

		if isActive {
			tallyImage = tallyPreview
			currentTallyStatus = tallyStatusOn
		} else {
			currentTallyStatus = tallyStatusOff
		}
		p.contextTallyMap.Store(contextID, currentTallyStatus)
		p.client.SetImage(sdctx, tallyImage, streamdeck.HardwareAndSoftware)
	}

	return nil
}

func (p *previewAction) OnVMixXML(ctx context.Context, resp *vmixtcp.XMLResponse, addr string, vm vmixtcp.Vmix) error {
	p.logger.Debug(ctx, "OnVMixXML started")
	defer p.logger.Debug(ctx, "OnVMixXML completed")

	// PropertyInspectorに SendInputsする
	// ステートフルになるので、HostごとにInputをキャッシュする

	// まずは受け取ったXMLをパースして、Inputを取得する
	inputs := lo.Map(resp.XML.Inputs.Input, func(input models.Input, _ int) *setting.Input {
		return &setting.Input{
			Key:    input.Key,
			Name:   input.Title,
			Number: int(input.Number),
		}
	})
	p.inputCache.Store(addr, inputs)

	// 取得したInputをPropertyInspectorにSendInputsする
	if err := p.updatePropertyInspector(ctx, streamdeck.NewEvent(ctx, "", nil)); err != nil {
		p.logger.Error(ctx, "Failed to update PropertyInspector", "error", err)
		return err
	}

	return nil
}

func (p *previewAction) OnVMixVersion(ctx context.Context, resp *vmixtcp.VersionResponse, addr string, vm vmixtcp.Vmix) error {
	p.logger.Debug(ctx, "OnVMixVersion started")
	defer p.logger.Debug(ctx, "OnVMixVersion completed")

	for _, contextID := range p.connectionManager.GetContexts(ctx, addr) {
		s, ok := p.store.Load(contextID)
		if !ok {
			continue
		}

		funcName := "InputPreview"
		switch s.Mix {
		case 0:
			funcName = "InputPreview"
		case 1:
			funcName = "InputPreviewMix2"
		case 2:
			funcName = "InputPreviewMix3"
		case 3:
			funcName = "InputPreviewMix4"
		case 4:
			funcName = "InputPreviewMix5"
		case 5:
			funcName = "InputPreviewMix6"
		case 6:
			funcName = "InputPreviewMix7"
		case 7:
			funcName = "InputPreviewMix8"
		case 8:
			funcName = "InputPreviewMix9"
		case 9:
			funcName = "InputPreviewMix10"
		case 10:
			funcName = "InputPreviewMix11"
		case 11:
			funcName = "InputPreviewMix12"
		case 12:
			funcName = "InputPreviewMix13"
		case 13:
			funcName = "InputPreviewMix14"
		case 14:
			funcName = "InputPreviewMix15"
		case 15:
			funcName = "InputPreviewMix16"
		}
		if err := vm.Acts(funcName, &s.Input); err != nil {
			p.logger.Error(ctx, "Failed to execute InputPreview", "error", err)
			return err
		}
		p.logger.Info(ctx, "InputPreview executed successfully: %d", s.Input)
	}

	return nil
}

func (p *previewAction) OnVMixSubscribe(ctx context.Context, resp *vmixtcp.SubscribeResponse, addr string, vm vmixtcp.Vmix) error {
	p.logger.Debug(ctx, "OnVMixSubscribe started")
	defer p.logger.Debug(ctx, "OnVMixSubscribe completed")

	return nil
}

// updatePropertyInspector updates both destinations and inputs in the property inspector
func (p *previewAction) updatePropertyInspector(ctx context.Context, event streamdeck.Event) error {
	// Create StreamDeck context with all necessary information
	sdctx := sdcontext.WithContext(ctx, event.Context)
	sdctx = sdcontext.WithAction(sdctx, event.Action)
	sdctx = sdcontext.WithDevice(sdctx, event.Device)

	// Send destinations
	destinations := p.connectionManager.GetAllVMixAddrs(ctx)
	if err := p.sendDestinations(sdctx, destinations); err != nil {
		p.logger.Error(ctx, "Failed to send destinations to PropertyInspector", "error", err)
		return err
	}

	// Send inputs
	destinationToInputs := make(setting.DestinationToInputs)
	p.inputCache.Range(func(key string, value []*setting.Input) bool {
		destinationToInputs[key] = value
		return true
	})
	if err := p.sendInputs(sdctx, destinationToInputs); err != nil {
		p.logger.Error(ctx, "Failed to send inputs to PropertyInspector", "error", err)
		return err
	}

	return nil
}

func (p *previewAction) sendInputs(ctx context.Context, inputs setting.DestinationToInputs) error {
	payload := setting.SendInputsPayload{
		Event:  "inputs",
		Inputs: inputs,
	}
	if err := p.client.SendToPropertyInspector(ctx, payload); err != nil {
		return xerrors.Errorf("failed to send inputs to PropertyInspector: %w", err)
	}
	return nil
}
func (p *previewAction) sendDestinations(ctx context.Context, destinations []connection.VMixConnection) error {
	payload := setting.Destinations{
		Event: "destinations",
		Destinations: lo.Map(destinations, func(destination connection.VMixConnection, _ int) setting.DestinationStatus {
			return setting.DestinationStatus{
				Address:   destination.Address,
				Connected: destination.Connected,
			}
		}),
	}
	if err := p.client.SendToPropertyInspector(ctx, payload); err != nil {
		return xerrors.Errorf("failed to send destinations to PropertyInspector: %w", err)
	}
	return nil
}
