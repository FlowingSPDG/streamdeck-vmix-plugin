package action

import (
	"context"
	"encoding/json"
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

const ActivatorActionUUID = "dev.flowingspdg.vmix.activator"

type ActivatorAction interface {
	OnWillAppear() streamdeck.EventHandler
	OnWillDisappear() streamdeck.EventHandler
	OnUpdateSettings() streamdeck.EventHandler
	OnSendToPlugin() streamdeck.EventHandler
	OnVMixTally(ctx context.Context, resp *vmixtcp.TallyResponse, addr string, vm vmixtcp.Vmix) error
	OnVMixXML(ctx context.Context, resp *vmixtcp.XMLResponse, addr string, vm vmixtcp.Vmix) error
	OnVMixActs(ctx context.Context, resp *vmixtcp.ActsResponse, addr string, vm vmixtcp.Vmix) error
	OnVMixVersion(ctx context.Context, resp *vmixtcp.VersionResponse, addr string, vm vmixtcp.Vmix) error
	OnVMixSubscribe(ctx context.Context, resp *vmixtcp.SubscribeResponse, addr string, vm vmixtcp.Vmix) error
}

type activatorAction struct {
	logger            logger.Logger
	connectionManager *connection.ConnectionManager
	store             setting.SettingStore[*setting.ActivatorSetting]
	client            *streamdeck.Client
	inputCache        setting.SettingStore[[]*setting.Input]
	contextTallyMap   *xsync.MapOf[string, tallyStatus]
}

func NewActivatorAction(
	logger logger.Logger,
	connectionManager *connection.ConnectionManager,
	store setting.SettingStore[*setting.ActivatorSetting],
	client *streamdeck.Client,
	inputCache setting.SettingStore[[]*setting.Input],
) ActivatorAction {
	return &activatorAction{
		logger:            logger,
		connectionManager: connectionManager,
		store:             store,
		client:            client,
		inputCache:        inputCache,
		contextTallyMap:   xsync.NewMapOf[string, tallyStatus](),
	}
}

func (a *activatorAction) OnWillAppear() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		a.logger.Debug(ctx, "OnWillAppear started. contextID: %s", event.Context)
		defer a.logger.Debug(ctx, "OnWillAppear completed. contextID: %s", event.Context)

		payload := streamdeck.WillAppearPayload[*setting.ActivatorSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			a.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			return nil
		}

		if payload.Settings.IsDefault() {
			a.logger.Error(ctx, "ActivatorAction OnWillAppear settings is default")
			payload.Settings.Initialize()
		}

		a.store.Store(event.Context, payload.Settings)
		a.connectionManager.AddContext(ctx, payload.Settings.VMixAddress, event.Context, ActivatorActionUUID)
		a.contextTallyMap.Store(event.Context, tallyStatusUnknown)

		a.logger.Debug(ctx, "activatorAction OnWillAppear settings: %#v", payload.Settings)

		if err := a.client.SetSettings(ctx, *payload.Settings); err != nil {
			a.logger.Error(ctx, "Failed to set settings %v", err)
		}
		if err := a.client.SetImage(ctx, "", streamdeck.HardwareAndSoftware); err != nil {
			a.logger.Error(ctx, "Failed to set image %v", err)
		}

		vmix := a.connectionManager.GetClient(ctx, payload.Settings.VMixAddress)
		if vmix == nil {
			a.logger.Error(ctx, "vMix connection not found on activatorAction OnWillAppear event")
			a.connectionManager.AddVMix(ctx, payload.Settings.VMixAddress)
			return nil
		}

		// PropertyInspectorを更新
		if err := a.updatePropertyInspector(ctx, event); err != nil {
			return err
		}

		return nil
	}
}

func (a *activatorAction) OnWillDisappear() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		a.logger.Debug(ctx, "OnWillDisappear started. contextID: %s", event.Context)
		defer a.logger.Debug(ctx, "OnWillDisappear completed. contextID: %s", event.Context)

		payload := streamdeck.WillDisappearPayload[*setting.ActivatorSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			a.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			return nil
		}

		a.connectionManager.RemoveContext(ctx, payload.Settings.VMixAddress, event.Context)
		a.store.Delete(event.Context)
		return nil
	}
}

func (a *activatorAction) OnUpdateSettings() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		payload := streamdeck.DidReceiveSettingsPayload[*setting.ActivatorSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			a.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			return nil
		}

		a.logger.Debug(ctx, "OnUpdateSettings started. contextID: %s settings: %#v", event.Context, payload.Settings)
		defer a.logger.Debug(ctx, "OnUpdateSettings completed. contextID: %s", event.Context)

		oldSettings, ok := a.store.Load(event.Context)
		if !ok {
			a.logger.Error(ctx, "Failed to load old settings on activatorAction OnUpdateSettings event")
		}
		a.connectionManager.UpdateContext(ctx, oldSettings.VMixAddress, payload.Settings.VMixAddress, event.Context, ActivatorActionUUID)
		a.store.Store(event.Context, payload.Settings)

		if err := a.client.SetImage(ctx, "", streamdeck.HardwareAndSoftware); err != nil {
			a.logger.Error(ctx, "Failed to set image", "error", err)
		}

		// PropertyInspectorを更新
		if err := a.updatePropertyInspector(ctx, event); err != nil {
			return err
		}

		// request ACTS
		vmix := a.connectionManager.GetVMixByContext(ctx, event.Context)
		if vmix == nil {
			a.logger.Error(ctx, "vMix connection not found on activatorAction OnUpdateSettings event")
			return xerrors.Errorf("vMix connection not found on activatorAction OnUpdateSettings event")
		}
		// TODO: Acts自体がstringを受けれるようにする
		actsInput, _ := strconv.Atoi(payload.Settings.ActInput)
		vmix.Acts(payload.Settings.ActsEvent, &actsInput)

		return nil
	}
}

func (a *activatorAction) OnSendToPlugin() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		a.logger.Debug(ctx, "OnSendToPlugin started")
		defer a.logger.Debug(ctx, "OnSendToPlugin completed")

		type CommandPayload struct {
			Event   string          `json:"event"`
			Payload json.RawMessage `json:"payload"`
		}

		var command CommandPayload
		if err := json.Unmarshal(event.Payload, &command); err != nil {
			a.logger.Error(ctx, "Failed to unmarshal command payload", "error", err)
			return xerrors.Errorf("failed to unmarshal command payload: %w", err)
		}

		a.logger.Info(ctx, "OnSendToPlugin received command: %s", command.Event)

		switch command.Event {
		case "property_inspector":
			if err := a.updatePropertyInspector(ctx, event); err != nil {
				return xerrors.Errorf("failed to update PropertyInspector: %w", err)
			}

		case "connect":
			type ConnectArgs struct {
				Host string `json:"host"`
			}
			var args ConnectArgs
			if err := json.Unmarshal(command.Payload, &args); err != nil {
				a.logger.Error(ctx, "Failed to unmarshal connect args", "error", err)
				return xerrors.Errorf("failed to unmarshal connect args: %w", err)
			}
			a.logger.Info(ctx, "Connect command received: %s", args.Host)

			a.connectionManager.AddVMix(ctx, args.Host)

			// PropertyInspectorを更新
			if err := a.updatePropertyInspector(ctx, event); err != nil {
				return xerrors.Errorf("failed to update PropertyInspector: %w", err)
			}

		case "disconnect":
			type DisconnectArgs struct {
				Host string `json:"host"`
			}
			var args DisconnectArgs
			if err := json.Unmarshal(command.Payload, &args); err != nil {
				a.logger.Error(ctx, "Failed to unmarshal disconnect args", "error", err)
				return xerrors.Errorf("failed to unmarshal disconnect args: %w", err)
			}
			a.logger.Info(ctx, "Disconnect command received: %s", args.Host)
			a.connectionManager.RemoveVMix(ctx, args.Host)

			// PropertyInspectorのdestinationsを更新
			destinations := a.connectionManager.GetAllVMixAddrs(ctx)
			sdctx := sdcontext.WithContext(ctx, event.Context)
			sdctx = sdcontext.WithAction(sdctx, event.Action)
			sdctx = sdcontext.WithDevice(sdctx, event.Device)
			if err := a.sendDestinations(sdctx, destinations); err != nil {
				a.logger.Error(ctx, "Failed to send destinations to PropertyInspector", "error", err)
				return xerrors.Errorf("failed to send destinations to PropertyInspector: %w", err)
			}

		default:
			a.logger.Error(ctx, "Unknown command received: %s", command.Event)
			return xerrors.Errorf("unknown command: %s", command.Event)
		}

		return nil
	}
}

func (a *activatorAction) OnVMixTally(ctx context.Context, resp *vmixtcp.TallyResponse, addr string, vm vmixtcp.Vmix) error {
	return nil
}

func (a *activatorAction) OnVMixXML(ctx context.Context, resp *vmixtcp.XMLResponse, addr string, vm vmixtcp.Vmix) error {
	a.logger.Debug(ctx, "OnVMixXML started")
	defer a.logger.Debug(ctx, "OnVMixXML completed")

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
	a.inputCache.Store(addr, inputs)

	// 取得したInputをPropertyInspectorにSendInputsする
	if err := a.updatePropertyInspector(ctx, streamdeck.NewEvent(ctx, "", nil)); err != nil {
		a.logger.Error(ctx, "Failed to update PropertyInspector", "error", err)
		return xerrors.Errorf("failed to update PropertyInspector: %w", err)
	}

	return nil
}

func (a *activatorAction) OnVMixActs(ctx context.Context, resp *vmixtcp.ActsResponse, addr string, vm vmixtcp.Vmix) error {
	a.logger.Debug(ctx, "OnVMixActs started")
	defer a.logger.Debug(ctx, "OnVMixActs completed")

	// 全てのコンテキストを取得
	contexts := a.connectionManager.GetContextsByActionType(ctx, addr, ActivatorActionUUID)
	for _, contextID := range contexts {
		settings, ok := a.store.Load(contextID)
		if !ok {
			a.logger.Error(ctx, "Settings not found for context", "contextID", contextID)
			continue
		}

		// イベントをパース
		parts := strings.Split(resp.Response, " ")
		if len(parts) < 2 {
			a.logger.Error(ctx, "Invalid event format", "event", resp.Response)
			continue
		}
		event := parts[0]
		input := parts[1]
		state := ""
		if len(parts) > 2 {
			state = parts[2]
		}

		a.logger.Debug(ctx, "OnVMixActs event: %s input: %s state: %s", event, input, state)

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
		currentTallyStatus, _ := a.contextTallyMap.LoadOrStore(contextID, tallyStatusUnknown)
		if (currentTallyStatus == tallyStatusOff && !isActive) || (currentTallyStatus == tallyStatusOn && isActive) {
			a.logger.Debug(ctx, "Tally state and cached state are the same. skipping. currentTallyStatus: %v isActive: %v", currentTallyStatus, isActive)
			continue
		}

		a.logger.Debug(ctx, "Going to apply tally. setting: %#v", settings)
		tallyImage := tallyInactive
		activeColor := tallyProgram
		if settings.Color == setting.ActivatorColorRed {
			activeColor = tallyProgram
		} else if settings.Color == setting.ActivatorColorGreen {
			activeColor = tallyPreview
		}

		if isActive {
			tallyImage = activeColor
			currentTallyStatus = tallyStatusOn
		} else {
			currentTallyStatus = tallyStatusOff
		}
		a.contextTallyMap.Store(contextID, currentTallyStatus)
		sdctx := sdcontext.WithContext(ctx, contextID)
		a.client.SetImage(sdctx, tallyImage, streamdeck.HardwareAndSoftware)
	}

	return nil
}

func (a *activatorAction) OnVMixVersion(ctx context.Context, resp *vmixtcp.VersionResponse, addr string, vm vmixtcp.Vmix) error {
	return nil
}

func (a *activatorAction) OnVMixSubscribe(ctx context.Context, resp *vmixtcp.SubscribeResponse, addr string, vm vmixtcp.Vmix) error {
	return nil
}

// updatePropertyInspector updates both destinations and inputs in the property inspector
func (a *activatorAction) updatePropertyInspector(ctx context.Context, event streamdeck.Event) error {
	// Create StreamDeck context with all necessary information
	sdctx := sdcontext.WithContext(ctx, event.Context)
	sdctx = sdcontext.WithAction(sdctx, event.Action)
	sdctx = sdcontext.WithDevice(sdctx, event.Device)

	// Send destinations
	destinations := a.connectionManager.GetAllVMixAddrs(ctx)
	if err := a.sendDestinations(sdctx, destinations); err != nil {
		a.logger.Error(ctx, "Failed to send destinations to PropertyInspector", "error", err)
		return xerrors.Errorf("failed to send destinations to PropertyInspector: %w", err)
	}

	// Send inputs
	destinationToInputs := make(setting.DestinationToInputs)
	a.inputCache.Range(func(key string, value []*setting.Input) bool {
		destinationToInputs[key] = value
		return true
	})
	if err := a.sendInputs(sdctx, destinationToInputs); err != nil {
		a.logger.Error(ctx, "Failed to send inputs to PropertyInspector", "error", err)
		return xerrors.Errorf("failed to send inputs to PropertyInspector: %w", err)
	}

	return nil
}

func (a *activatorAction) sendInputs(ctx context.Context, inputs setting.DestinationToInputs) error {
	payload := setting.SendInputsPayload{
		Event:  "inputs",
		Inputs: inputs,
	}
	if err := a.client.SendToPropertyInspector(ctx, payload); err != nil {
		return xerrors.Errorf("failed to send inputs to PropertyInspector: %w", err)
	}
	return nil
}

func (a *activatorAction) sendDestinations(ctx context.Context, destinations []string) error {
	payload := setting.Destinations{
		Event:        "destinations",
		Destinations: destinations,
	}
	if err := a.client.SendToPropertyInspector(ctx, payload); err != nil {
		return xerrors.Errorf("failed to send destinations to PropertyInspector: %w", err)
	}
	return nil
}
