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

const ProgramActionUUID = "dev.flowingspdg.vmix.program"

type ProgramAction interface {
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

type programAction struct {
	logger            logger.Logger
	connectionManager *connection.ConnectionManager
	store             setting.SettingStore[*setting.ProgramSetting]
	client            *streamdeck.Client
	inputCache        setting.SettingStore[[]*Input]
	contextTallyMap   *xsync.MapOf[string, tallyStatus]
}

func NewProgramAction(
	logger logger.Logger,
	connectionManager *connection.ConnectionManager,
	store setting.SettingStore[*setting.ProgramSetting],
	client *streamdeck.Client,
	inputCache setting.SettingStore[[]*Input],
) ProgramAction {
	return &programAction{
		logger:            logger,
		connectionManager: connectionManager,
		store:             store,
		client:            client,
		inputCache:        inputCache,
		contextTallyMap:   xsync.NewMapOf[string, tallyStatus](),
	}
}

func (p *programAction) OnWillAppear() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		p.logger.Debug(ctx, "OnWillAppear started. contextID: %s", event.Context)
		defer p.logger.Debug(ctx, "OnWillAppear completed. contextID: %s", event.Context)

		payload := streamdeck.WillAppearPayload[*setting.ProgramSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			p.logger.Error(ctx, "Failed to unmarshal payload %v", err)
			// エラーを返すと後続のイベント処理が止まるっぽいので一旦return nilしてみる
			return nil
		}

		p.store.Store(event.Context, payload.Settings)
		p.connectionManager.AddContext(ctx, payload.Settings.VMixAddress, event.Context, ProgramActionUUID)
		p.contextTallyMap.Store(event.Context, tallyStatusUnknown)

		p.logger.Debug(ctx, "programAction OnWillAppear settings: %#v", payload.Settings)

		if err := p.client.SetSettings(ctx, *payload.Settings); err != nil {
			p.logger.Error(ctx, "Failed to set settings %v", err)
		}
		if err := p.client.SetImage(ctx, "", streamdeck.HardwareAndSoftware); err != nil {
			p.logger.Error(ctx, "Failed to set image %v", err)
		}

		vmix := p.connectionManager.GetClient(ctx, payload.Settings.VMixAddress)
		if vmix == nil {
			p.logger.Error(ctx, "vMix connection not found on previewAction OnWillAppear event")
			// vMixへの接続処理
			p.connectionManager.AddVMix(ctx, payload.Settings.VMixAddress)
			return nil
		}
		switch payload.Settings.TallyMode {
		case setting.TallyModeTALLY:
			if err := vmix.Tally(); err != nil {
				p.logger.Error(ctx, "Failed to set tally", "error", err)
			}
		case setting.TallyModeACTS:
			funcName := "Input"
			switch payload.Settings.Mix {
			case 0:
				funcName = "Input"
			case 1:
				funcName = "InputMix2"
			case 2:
				funcName = "InputMix3"
			case 3:
				funcName = "InputMix4"
			case 4:
				funcName = "InputMix5"
			case 5:
				funcName = "InputMix6"
			case 6:
				funcName = "InputMix7"
			case 7:
				funcName = "InputMix8"
			case 8:
				funcName = "InputMix9"
			case 9:
				funcName = "InputMix10"
			case 10:
				funcName = "InputMix11"
			case 11:
				funcName = "InputMix12"
			case 12:
				funcName = "InputMix13"
			case 13:
				funcName = "InputMix14"
			case 14:
				funcName = "InputMix15"
			case 15:
				funcName = "InputMix16"
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

func (p *programAction) OnWillDisappear() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		p.logger.Debug(ctx, "OnWillDisappear started. contextID: %s", event.Context)
		defer p.logger.Debug(ctx, "OnWillDisappear completed. contextID: %s", event.Context)

		payload := streamdeck.WillDisappearPayload[setting.ProgramSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			p.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			// エラーを返すと後続のイベント処理が止まるっぽいので一旦return nilしてみる
			return nil
		}

		p.connectionManager.RemoveContext(ctx, payload.Settings.VMixAddress, event.Context)

		p.store.Delete(event.Context)
		return nil
	}
}

func (p *programAction) OnUpdateSettings() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		payload := streamdeck.DidReceiveSettingsPayload[setting.ProgramSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			p.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			// エラーを返すと後続のイベント処理が止まるっぽいので一旦return nilしてみる
			return nil
		}

		p.logger.Debug(ctx, "OnUpdateSettings started. contextID: %s settings: %#v", event.Context, payload.Settings)
		defer p.logger.Debug(ctx, "OnUpdateSettings completed. contextID: %s", event.Context)

		oldSettings, ok := p.store.Load(event.Context)
		if !ok {
			p.logger.Error(ctx, "Failed to load old settings on programAction OnUpdateSettings event")
		}
		p.connectionManager.UpdateContext(ctx, oldSettings.VMixAddress, payload.Settings.VMixAddress, event.Context, ProgramActionUUID)
		p.store.Store(event.Context, &payload.Settings)

		if err := p.client.SetImage(ctx, "", streamdeck.HardwareAndSoftware); err != nil {
			p.logger.Error(ctx, "Failed to set image", "error", err)
		}

		vmix := p.connectionManager.GetClient(ctx, payload.Settings.VMixAddress)
		if vmix == nil {
			p.logger.Error(ctx, "vMix connection not found on programAction OnUpdateSettings event")
			return nil
		}

		p.contextTallyMap.Store(event.Context, tallyStatusUnknown)

		switch payload.Settings.TallyMode {
		case setting.TallyModeTALLY:
			if err := vmix.Tally(); err != nil {
				p.logger.Error(ctx, "Failed to set tally", "error", err)
			}
		case setting.TallyModeACTS:
			funcName := "Input"
			switch payload.Settings.Mix {
			case 0:
				funcName = "Input"
			case 1:
				funcName = "InputMix2"
			case 2:
				funcName = "InputMix3"
			case 3:
				funcName = "InputMix4"
			case 4:
				funcName = "InputMix5"
			case 5:
				funcName = "InputMix6"
			case 6:
				funcName = "InputMix7"
			case 7:
				funcName = "InputMix8"
			case 8:
				funcName = "InputMix9"
			case 9:
				funcName = "InputMix10"
			case 10:
				funcName = "InputMix11"
			case 11:
				funcName = "InputMix12"
			case 12:
				funcName = "InputMix13"
			case 13:
				funcName = "InputMix14"
			case 14:
				funcName = "InputMix15"
			case 15:
				funcName = "InputMix16"
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

func (p *programAction) OnKeyDown() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		p.logger.Debug(ctx, "Execute started")
		defer p.logger.Debug(ctx, "Execute completed")

		payload := streamdeck.KeyDownPayload[setting.ProgramSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			p.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			// エラーを返すと後続のイベント処理が止まるっぽいので一旦return nilしてみる
			return nil
		}

		// Get vMix client
		vmix := p.connectionManager.GetVMixByContext(ctx, event.Context)
		if vmix == nil {
			err := xerrors.Errorf("vMix connection not found on programAction OnKeyDown event")
			p.logger.Error(ctx, "Failed to get vMix client: %v", err)
			return err
		}

		// Build query with Mix parameter
		query := fmt.Sprintf("Input=%d", payload.Settings.Input)
		if payload.Settings.Mix > 0 {
			query += fmt.Sprintf("&Mix=%d", payload.Settings.Mix)
		}
		if payload.Settings.Duration > 0 {
			query += fmt.Sprintf("&Duration=%d", payload.Settings.Duration)
		}

		// Execute PreviewInput function with query
		p.logger.Debug(ctx, "Executing PreviewInput with query: %s", query)
		if err := vmix.Function(payload.Settings.Transition, query); err != nil {
			p.logger.Error(ctx, "Failed to execute PreviewInput", "error", err)
			return fmt.Errorf("failed to execute PreviewInput: %w", err)
		}

		p.logger.Info(ctx, "PreviewInput executed successfully: %d", payload.Settings.Input)
		return nil
	}
}

func (p *programAction) OnSendToPlugin() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		p.logger.Debug(ctx, "OnSendToPlugin started")
		defer p.logger.Debug(ctx, "OnSendToPlugin completed")

		type CommandPayload struct {
			Event   string          `json:"event"`
			Payload json.RawMessage `json:"payload"`
		}

		// まずCommandPayloadとしてパース
		var command CommandPayload
		if err := json.Unmarshal(event.Payload, &command); err != nil {
			p.logger.Error(ctx, "Failed to unmarshal command payload", "error", err)
			return err
		}

		p.logger.Info(ctx, "OnSendToPlugin received command: %s", command.Event)

		// コマンド名によってパースするpayloadを分岐
		switch command.Event {
		case "property_inspector":
			// TODO: inputの表示をリロードなしで実施する
			if err := p.updatePropertyInspector(ctx, event); err != nil {
				return err
			}

		case "connect":
			// 接続コマンドの場合
			type ConnectArgs struct {
				Host string `json:"host"`
			}
			var args ConnectArgs
			if err := json.Unmarshal(command.Payload, &args); err != nil {
				p.logger.Error(ctx, "Failed to unmarshal connect args", "error", err)
				return err
			}
			p.logger.Info(ctx, "Connect command received: %s", args.Host)

			// vMixへの接続処理
			p.connectionManager.AddVMix(ctx, args.Host)

			// vMixへの接続後、PropertyInspectorを更新
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
			return fmt.Errorf("unknown command: %s", command.Event)
		}

		return nil
	}
}

func (p *programAction) OnVMixTally(ctx context.Context, resp *vmixtcp.TallyResponse, addr string, vm vmixtcp.Vmix) error {
	p.logger.Debug(ctx, "OnVMixTally started")
	defer p.logger.Debug(ctx, "OnVMixTally completed")

	contextIDs := p.connectionManager.GetContextsByActionType(ctx, addr, ProgramActionUUID)
	// TODO: Program actionのみ取得する
	p.logger.Debug(ctx, "OnVMixTally addr: %s contextIDs: %v resp: %v", addr, contextIDs, resp.Tally)
	for _, contextID := range contextIDs {
		sdctx := sdcontext.WithContext(ctx, contextID)
		s, ok := p.store.Load(contextID)
		if !ok {
			err := xerrors.Errorf("setting not found on programAction OnVMixTally event")
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

		tallyActive := resp.Tally[s.Input-1] == vmixtcp.Program
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
			tallyImage = tallyProgram
			currentTallyStatus = tallyStatusOn
		}
		p.contextTallyMap.Store(contextID, currentTallyStatus)
		p.client.SetImage(sdctx, tallyImage, streamdeck.HardwareAndSoftware)
	}

	return nil
}

func (p *programAction) OnVMixXML(ctx context.Context, resp *vmixtcp.XMLResponse, addr string, vm vmixtcp.Vmix) error {
	p.logger.Debug(ctx, "OnVMixXML started")
	defer p.logger.Debug(ctx, "OnVMixXML completed")

	// PropertyInspectorに SendInputsする
	// ステートフルになるので、HostごとにInputをキャッシュする

	// まずは受け取ったXMLをパースして、Inputを取得する
	inputs := lo.Map(resp.XML.Inputs.Input, func(input models.Input, _ int) *Input {
		return &Input{
			Key:    input.Key,
			Name:   input.Title,
			Number: input.Number,
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

func (p *programAction) OnVMixActs(ctx context.Context, resp *vmixtcp.ActsResponse, addr string, vm vmixtcp.Vmix) error {
	p.logger.Debug(ctx, "OnVMixActs started")
	defer p.logger.Debug(ctx, "OnVMixActs completed")

	contextIDs := p.connectionManager.GetContextsByActionType(ctx, addr, ProgramActionUUID)
	p.logger.Debug(ctx, "OnVMixActs addr: %s contextIDs: %v resp: %v", addr, contextIDs, resp.Response)
	for _, contextID := range contextIDs {
		sdctx := sdcontext.WithContext(ctx, contextID)
		s, ok := p.store.Load(contextID)
		if !ok {
			err := xerrors.Errorf("setting not found on programAction OnVMixActs event")
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
		case "Input":
			isCorrectEvent = s.Mix == 0
		case "InputMix2":
			isCorrectEvent = s.Mix == 1
		case "InputMix3":
			isCorrectEvent = s.Mix == 2
		case "InputMix4":
			isCorrectEvent = s.Mix == 3
		case "InputMix5":
			isCorrectEvent = s.Mix == 4
		case "InputMix6":
			isCorrectEvent = s.Mix == 5
		case "InputMix7":
			isCorrectEvent = s.Mix == 6
		case "InputMix8":
			isCorrectEvent = s.Mix == 7
		case "InputMix9":
			isCorrectEvent = s.Mix == 8
		case "InputMix10":
			isCorrectEvent = s.Mix == 9
		case "InputMix11":
			isCorrectEvent = s.Mix == 10
		case "InputMix12":
			isCorrectEvent = s.Mix == 11
		case "InputMix13":
			isCorrectEvent = s.Mix == 12
		case "InputMix14":
			isCorrectEvent = s.Mix == 13
		case "InputMix15":
			isCorrectEvent = s.Mix == 14
		case "InputMix16":
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
			tallyImage = tallyProgram
			currentTallyStatus = tallyStatusOn
		} else {
			currentTallyStatus = tallyStatusOff
		}
		p.contextTallyMap.Store(contextID, currentTallyStatus)
		p.client.SetImage(sdctx, tallyImage, streamdeck.HardwareAndSoftware)
	}

	return nil
}

func (p *programAction) OnVMixVersion(ctx context.Context, resp *vmixtcp.VersionResponse, addr string, vm vmixtcp.Vmix) error {
	p.logger.Debug(ctx, "OnVMixVersion started")
	defer p.logger.Debug(ctx, "OnVMixVersion completed")

	for _, contextID := range p.connectionManager.GetContexts(ctx, addr) {
		s, ok := p.store.Load(contextID)
		if !ok {
			continue
		}

		funcName := "Input"
		switch s.Mix {
		case 0:
			funcName = "Input"
		case 1:
			funcName = "InputMix2"
		case 2:
			funcName = "InputMix3"
		case 3:
			funcName = "InputMix4"
		case 4:
			funcName = "InputMix5"
		case 5:
			funcName = "InputMix6"
		case 6:
			funcName = "InputMix7"
		case 7:
			funcName = "InputMix8"
		case 8:
			funcName = "InputMix9"
		case 9:
			funcName = "InputMix10"
		case 10:
			funcName = "InputMix11"
		case 11:
			funcName = "InputMix12"
		case 12:
			funcName = "InputMix13"
		case 13:
			funcName = "InputMix14"
		case 14:
			funcName = "InputMix15"
		case 15:
			funcName = "InputMix16"
		}
		if err := vm.Acts(funcName, &s.Input); err != nil {
			p.logger.Error(ctx, "Failed to execute Input", "error", err)
			return err
		}
		p.logger.Info(ctx, "Input executed successfully: %d", s.Input)
	}

	return nil
}

func (p *programAction) OnVMixSubscribe(ctx context.Context, resp *vmixtcp.SubscribeResponse, addr string, vm vmixtcp.Vmix) error {
	p.logger.Debug(ctx, "OnVMixSubscribe started")
	defer p.logger.Debug(ctx, "OnVMixSubscribe completed")

	return nil
}

// updatePropertyInspector updates both destinations and inputs in the property inspector
func (p *programAction) updatePropertyInspector(ctx context.Context, event streamdeck.Event) error {
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
	destinationToInputs := make(DestinationToInputs)
	p.inputCache.Range(func(key string, value []*Input) bool {
		destinationToInputs[key] = value
		return true
	})
	if err := p.sendInputs(sdctx, destinationToInputs); err != nil {
		p.logger.Error(ctx, "Failed to send inputs to PropertyInspector", "error", err)
		return err
	}

	p.client.SetImage(sdctx, "", streamdeck.HardwareAndSoftware)

	return nil
}

func (p *programAction) sendInputs(ctx context.Context, inputs DestinationToInputs) error {
	payload := SendInputsPayload{
		Event:  "inputs",
		Inputs: inputs,
	}
	if err := p.client.SendToPropertyInspector(ctx, payload); err != nil {
		return fmt.Errorf("failed to send inputs to PropertyInspector: %w", err)
	}
	return nil
}

func (p *programAction) sendDestinations(ctx context.Context, destinations []string) error {
	payload := Destinations{
		Event:        "destinations",
		Destinations: destinations,
	}
	if err := p.client.SendToPropertyInspector(ctx, payload); err != nil {
		return fmt.Errorf("failed to send destinations to PropertyInspector: %w", err)
	}
	return nil
}
