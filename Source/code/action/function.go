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

const FunctionActionUUID = "dev.flowingspdg.vmix.function"

type FunctionAction interface {
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

type functionAction struct {
	logger            logger.Logger
	connectionManager *connection.ConnectionManager
	store             setting.SettingStore[*setting.FunctionSetting]
	client            *streamdeck.Client
	inputCache        setting.SettingStore[[]*setting.Input]
	contextTallyMap   *xsync.MapOf[string, tallyStatus]
}

func NewFunctionAction(
	logger logger.Logger,
	connectionManager *connection.ConnectionManager,
	store setting.SettingStore[*setting.FunctionSetting],
	client *streamdeck.Client,
	inputCache setting.SettingStore[[]*setting.Input],
) FunctionAction {
	return &functionAction{
		logger:            logger,
		connectionManager: connectionManager,
		store:             store,
		client:            client,
		inputCache:        inputCache,
		contextTallyMap:   xsync.NewMapOf[string, tallyStatus](),
	}
}

func (f *functionAction) OnWillAppear() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		f.logger.Debug(ctx, "OnWillAppear started. contextID: %s", event.Context)
		defer f.logger.Debug(ctx, "OnWillAppear completed. contextID: %s", event.Context)

		payload := streamdeck.WillAppearPayload[*setting.FunctionSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			f.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			return nil
		}

		if payload.Settings.IsDefault() {
			f.logger.Error(ctx, "FunctionAction OnWillAppear settings is default")
			payload.Settings.Initialize()
		}

		f.store.Store(event.Context, payload.Settings)
		f.connectionManager.AddContext(ctx, payload.Settings.VMixAddress, event.Context, FunctionActionUUID, true)
		f.contextTallyMap.Store(event.Context, tallyStatusUnknown)

		f.logger.Debug(ctx, "functionAction OnWillAppear settings: %#v", payload.Settings)

		if err := f.client.SetSettings(ctx, *payload.Settings); err != nil {
			f.logger.Error(ctx, "Failed to set settings %v", err)
		}

		vmix := f.connectionManager.GetClient(ctx, payload.Settings.VMixAddress)
		if vmix == nil {
			f.logger.Error(ctx, "vMix connection not found on functionAction OnWillAppear event")
			f.connectionManager.AddVMix(ctx, payload.Settings.VMixAddress)
			return nil
		}

		// PropertyInspectorを更新
		if err := f.updatePropertyInspector(ctx, event); err != nil {
			return err
		}

		return nil
	}
}

func (f *functionAction) OnWillDisappear() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		f.logger.Debug(ctx, "OnWillDisappear started. contextID: %s", event.Context)
		defer f.logger.Debug(ctx, "OnWillDisappear completed. contextID: %s", event.Context)

		payload := streamdeck.WillDisappearPayload[*setting.FunctionSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			f.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			return nil
		}

		f.connectionManager.RemoveContext(ctx, payload.Settings.VMixAddress, event.Context)
		f.store.Delete(event.Context)
		return nil
	}
}

func (f *functionAction) OnUpdateSettings() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		payload := streamdeck.DidReceiveSettingsPayload[*setting.FunctionSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			f.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			return nil
		}

		f.logger.Debug(ctx, "OnUpdateSettings started. contextID: %s settings: %#v", event.Context, payload.Settings)
		defer f.logger.Debug(ctx, "OnUpdateSettings completed. contextID: %s", event.Context)

		oldSettings, ok := f.store.Load(event.Context)
		if !ok {
			f.logger.Error(ctx, "Failed to load old settings on functionAction OnUpdateSettings event")
		}
		f.connectionManager.UpdateContext(ctx, oldSettings.VMixAddress, payload.Settings.VMixAddress, event.Context, FunctionActionUUID)
		f.store.Store(event.Context, payload.Settings)

		// PropertyInspectorを更新
		if err := f.updatePropertyInspector(ctx, event); err != nil {
			return err
		}

		// request ACTS
		vmix := f.connectionManager.GetVMixByContext(ctx, event.Context)
		if vmix == nil {
			f.logger.Error(ctx, "vMix connection not found on functionAction OnUpdateSettings event")
			return xerrors.Errorf("vMix connection not found on functionAction OnUpdateSettings event")
		}
		var actInput *int = nil
		if payload.Settings.ActInput != "" {
			actInputInt, err := strconv.Atoi(payload.Settings.ActInput)
			if err != nil {
				f.logger.Error(ctx, "Failed to convert actInput to int", "error", err)
				return xerrors.Errorf("failed to convert actInput to int: %w", err)
			}
			actInput = &actInputInt
		}
		vmix.Acts(payload.Settings.ActsEvent, actInput)

		return nil
	}
}

func (f *functionAction) OnKeyDown() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		f.logger.Debug(ctx, "Execute started")
		defer f.logger.Debug(ctx, "Execute completed")

		payload := streamdeck.KeyDownPayload[*setting.FunctionSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			f.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			return nil
		}

		vmix := f.connectionManager.GetVMixByContext(ctx, event.Context)
		if vmix == nil {
			err := xerrors.Errorf("vMix connection not found on functionAction OnKeyDown event")
			f.logger.Error(ctx, "Failed to get vMix client: %v", err)
			return err
		}

		if payload.Settings.Input != nil {
			payload.Settings.Query = fmt.Sprintf("%s&Input=%d", payload.Settings.Query, *payload.Settings.Input)
		}

		if err := vmix.Function(payload.Settings.Function, payload.Settings.Query); err != nil {
			f.logger.Error(ctx, "Failed to execute function", "error", err)
			return err
		}

		return nil
	}
}

func (f *functionAction) OnSendToPlugin() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		f.logger.Debug(ctx, "OnSendToPlugin started")
		defer f.logger.Debug(ctx, "OnSendToPlugin completed")

		type CommandPayload struct {
			Event   string          `json:"event"`
			Payload json.RawMessage `json:"payload"`
		}

		var command CommandPayload
		if err := json.Unmarshal(event.Payload, &command); err != nil {
			f.logger.Error(ctx, "Failed to unmarshal command payload", "error", err)
			return xerrors.Errorf("failed to unmarshal command payload: %w", err)
		}

		f.logger.Info(ctx, "OnSendToPlugin received command: %s", command.Event)

		switch command.Event {
		case "property_inspector":
			if err := f.updatePropertyInspector(ctx, event); err != nil {
				return xerrors.Errorf("failed to update PropertyInspector: %w", err)
			}

		case "connect":
			type ConnectArgs struct {
				Host string `json:"host"`
			}
			var args ConnectArgs
			if err := json.Unmarshal(command.Payload, &args); err != nil {
				f.logger.Error(ctx, "Failed to unmarshal connect args", "error", err)
				return xerrors.Errorf("failed to unmarshal connect args: %w", err)
			}
			f.logger.Info(ctx, "Connect command received: %s", args.Host)

			f.connectionManager.AddVMix(ctx, args.Host)

			// PropertyInspectorを更新
			if err := f.updatePropertyInspector(ctx, event); err != nil {
				return xerrors.Errorf("failed to update PropertyInspector: %w", err)
			}

		case "disconnect":
			type DisconnectArgs struct {
				Host string `json:"host"`
			}
			var args DisconnectArgs
			if err := json.Unmarshal(command.Payload, &args); err != nil {
				f.logger.Error(ctx, "Failed to unmarshal disconnect args", "error", err)
				return xerrors.Errorf("failed to unmarshal disconnect args: %w", err)
			}
			f.logger.Info(ctx, "Disconnect command received: %s", args.Host)
			f.connectionManager.RemoveVMix(ctx, args.Host)

			// PropertyInspectorのdestinationsを更新
			destinations := f.connectionManager.GetAllVMixAddrs(ctx)
			sdctx := sdcontext.WithContext(ctx, event.Context)
			sdctx = sdcontext.WithAction(sdctx, event.Action)
			sdctx = sdcontext.WithDevice(sdctx, event.Device)
			if err := f.sendDestinations(sdctx, destinations); err != nil {
				f.logger.Error(ctx, "Failed to send destinations to PropertyInspector", "error", err)
				return xerrors.Errorf("failed to send destinations to PropertyInspector: %w", err)
			}

		default:
			f.logger.Error(ctx, "Unknown command received: %s", command.Event)
			return xerrors.Errorf("unknown command: %s", command.Event)
		}

		return nil
	}
}

func (f *functionAction) OnVMixTally(ctx context.Context, resp *vmixtcp.TallyResponse, addr string, vm vmixtcp.Vmix) error {
	return nil
}

func (f *functionAction) OnVMixXML(ctx context.Context, resp *vmixtcp.XMLResponse, addr string, vm vmixtcp.Vmix) error {
	f.logger.Debug(ctx, "OnVMixXML started")
	defer f.logger.Debug(ctx, "OnVMixXML completed")

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
	f.inputCache.Store(addr, inputs)

	// 取得したInputをPropertyInspectorにSendInputsする
	if err := f.updatePropertyInspector(ctx, streamdeck.NewEvent(ctx, "", nil)); err != nil {
		f.logger.Error(ctx, "Failed to update PropertyInspector", "error", err)
		return xerrors.Errorf("failed to update PropertyInspector: %w", err)
	}

	return nil
}

func (f *functionAction) OnVMixActs(ctx context.Context, resp *vmixtcp.ActsResponse, addr string, vm vmixtcp.Vmix) error {
	f.logger.Debug(ctx, "OnVMixActs started")
	defer f.logger.Debug(ctx, "OnVMixActs completed")

	// 全てのコンテキストを取得
	contexts := f.connectionManager.GetContextsByActionType(ctx, addr, FunctionActionUUID)
	for _, contextID := range contexts {
		settings, ok := f.store.Load(contextID)
		if !ok {
			f.logger.Error(ctx, "Settings not found for context", "contextID", contextID)
			continue
		}

		// イベントをパース
		parts := strings.Split(resp.Response, " ")
		if len(parts) < 2 {
			f.logger.Error(ctx, "Invalid event format", "event", resp.Response)
			continue
		}
		event := parts[0]
		input := parts[1]
		state := ""
		if len(parts) > 2 {
			state = parts[2]
		}

		f.logger.Debug(ctx, "OnVMixActs event: %s input: %s state: %s", event, input, state)

		// イベントが一致しない場合はスキップ
		if settings.ActsEvent != event {
			continue
		}

		// 入力が一致しない場合はスキップ
		if settings.ActInput != input {
			continue
		}

		// ステートが一致しない場合はスキップ
		if settings.ActsActiveState != "" && settings.ActsInactiveState != "" {
			if settings.ActsActiveState != state && settings.ActsInactiveState != state {
				continue
			}
		}
		isActive := settings.ActsActiveState == state

		// イベントが一致した場合は、タリーを反映
		// tally cacheを使用する
		// Tally stateとcached stateが一致していれば更新しない
		// Unknownであれば関係なく更新
		currentTallyStatus, _ := f.contextTallyMap.LoadOrStore(contextID, tallyStatusUnknown)
		if (currentTallyStatus == tallyStatusOff && !isActive) || (currentTallyStatus == tallyStatusOn && isActive) {
			f.logger.Debug(ctx, "Tally state and cached state are the same. skipping. currentTallyStatus: %v isActive: %v", currentTallyStatus, isActive)
			continue
		}

		f.logger.Debug(ctx, "Going to apply tally. setting: %#v", settings)
		tallyImage := tallyInactive
		activeColor := tallyProgram
		if settings.ActsColor == setting.ActivatorColorRed {
			activeColor = tallyProgram
		} else if settings.ActsColor == setting.ActivatorColorGreen {
			activeColor = tallyPreview
		}

		if isActive {
			tallyImage = activeColor
			currentTallyStatus = tallyStatusOn
		} else {
			currentTallyStatus = tallyStatusOff
		}
		f.contextTallyMap.Store(contextID, currentTallyStatus)
		sdctx := sdcontext.WithContext(ctx, contextID)
		f.client.SetImage(sdctx, tallyImage, streamdeck.HardwareAndSoftware)
	}

	return nil
}

func (f *functionAction) OnVMixVersion(ctx context.Context, resp *vmixtcp.VersionResponse, addr string, vm vmixtcp.Vmix) error {
	return nil
}

func (f *functionAction) OnVMixSubscribe(ctx context.Context, resp *vmixtcp.SubscribeResponse, addr string, vm vmixtcp.Vmix) error {
	return nil
}

// updatePropertyInspector updates both destinations and inputs in the property inspector
func (f *functionAction) updatePropertyInspector(ctx context.Context, event streamdeck.Event) error {
	// Create StreamDeck context with all necessary information
	sdctx := sdcontext.WithContext(ctx, event.Context)
	sdctx = sdcontext.WithAction(sdctx, event.Action)
	sdctx = sdcontext.WithDevice(sdctx, event.Device)

	// Send destinations
	destinations := f.connectionManager.GetAllVMixAddrs(ctx)
	if err := f.sendDestinations(sdctx, destinations); err != nil {
		f.logger.Error(ctx, "Failed to send destinations to PropertyInspector", "error", err)
		return xerrors.Errorf("failed to send destinations to PropertyInspector: %w", err)
	}

	return nil
}

func (f *functionAction) sendDestinations(ctx context.Context, destinations []connection.VMixConnection) error {
	payload := setting.Destinations{
		Event: "destinations",
		Destinations: lo.Map(destinations, func(destination connection.VMixConnection, _ int) setting.DestinationStatus {
			return setting.DestinationStatus{
				Address:   destination.Address,
				Connected: destination.Connected,
			}
		}),
	}
	if err := f.client.SendToPropertyInspector(ctx, payload); err != nil {
		return xerrors.Errorf("failed to send destinations to PropertyInspector: %w", err)
	}
	return nil
}
