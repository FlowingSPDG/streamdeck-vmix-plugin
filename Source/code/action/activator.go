package action

import (
	"context"
	"encoding/json"

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

type ActivatorSetting struct {
	VMixAddress string            `json:"dest"`
	Function    string            `json:"function"`
	TallyMode   setting.TallyMode `json:"tally_mode"`
	EventName   string            `json:"event_name"`
	Arg         string            `json:"arg"`
	State       string            `json:"state"`
}

func (s *ActivatorSetting) IsDefault() bool {
	return s.VMixAddress == "" && s.Function == "" && s.TallyMode == 0
}

func (s *ActivatorSetting) Initialize() {
	s.VMixAddress = "localhost"
	s.Function = ""
	s.TallyMode = setting.TallyModeACTS
	s.EventName = ""
	s.Arg = ""
	s.State = ""
}

func (s *ActivatorSetting) GetVMixAddress() string {
	return s.VMixAddress
}

func (s *ActivatorSetting) GetTallyMode() setting.TallyMode {
	return s.TallyMode
}

func (s *ActivatorSetting) GetMix() int {
	return 0
}

func (s *ActivatorSetting) GetInput() *int {
	return nil
}

type ActivatorAction interface {
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

type activatorAction struct {
	logger            logger.Logger
	connectionManager *connection.ConnectionManager
	store             setting.SettingStore[*ActivatorSetting]
	client            *streamdeck.Client
	inputCache        setting.SettingStore[[]*setting.Input]
	contextTallyMap   *xsync.MapOf[string, tallyStatus]
}

func NewActivatorAction(
	logger logger.Logger,
	connectionManager *connection.ConnectionManager,
	store setting.SettingStore[*ActivatorSetting],
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

		payload := streamdeck.WillAppearPayload[*ActivatorSetting]{}
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

		payload := streamdeck.WillDisappearPayload[ActivatorSetting]{}
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
		payload := streamdeck.DidReceiveSettingsPayload[ActivatorSetting]{}
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
		a.store.Store(event.Context, &payload.Settings)

		if err := a.client.SetImage(ctx, "", streamdeck.HardwareAndSoftware); err != nil {
			a.logger.Error(ctx, "Failed to set image", "error", err)
		}

		// PropertyInspectorを更新
		if err := a.updatePropertyInspector(ctx, event); err != nil {
			return err
		}

		return nil
	}
}

func (a *activatorAction) OnKeyDown() streamdeck.EventHandler {
	return func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		a.logger.Debug(ctx, "Execute started")
		defer a.logger.Debug(ctx, "Execute completed")

		payload := streamdeck.KeyDownPayload[*ActivatorSetting]{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			a.logger.Error(ctx, "Failed to unmarshal payload", "error", err)
			return nil
		}

		vmix := a.connectionManager.GetVMixByContext(ctx, event.Context)
		if vmix == nil {
			err := xerrors.Errorf("vMix connection not found on activatorAction OnKeyDown event")
			a.logger.Error(ctx, "Failed to get vMix client: %v", err)
			return err
		}

		if err := vmix.Function(payload.Settings.Function, ""); err != nil {
			a.logger.Error(ctx, "Failed to execute function", "error", err)
			return err
		}

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
			return err
		}

		a.logger.Info(ctx, "OnSendToPlugin received command: %s", command.Event)

		switch command.Event {
		case "property_inspector":
			if err := a.updatePropertyInspector(ctx, event); err != nil {
				return err
			}

		case "connect":
			type ConnectArgs struct {
				Host string `json:"host"`
			}
			var args ConnectArgs
			if err := json.Unmarshal(command.Payload, &args); err != nil {
				a.logger.Error(ctx, "Failed to unmarshal connect args", "error", err)
				return err
			}
			a.logger.Info(ctx, "Connect command received: %s", args.Host)

			a.connectionManager.AddVMix(ctx, args.Host)

			// PropertyInspectorを更新
			if err := a.updatePropertyInspector(ctx, event); err != nil {
				return err
			}

		case "disconnect":
			type DisconnectArgs struct {
				Host string `json:"host"`
			}
			var args DisconnectArgs
			if err := json.Unmarshal(command.Payload, &args); err != nil {
				a.logger.Error(ctx, "Failed to unmarshal disconnect args", "error", err)
				return err
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
				return err
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
		return err
	}

	return nil
}

func (a *activatorAction) OnVMixActs(ctx context.Context, resp *vmixtcp.ActsResponse, addr string, vm vmixtcp.Vmix) error {
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
		return err
	}

	// Send inputs
	destinationToInputs := make(setting.DestinationToInputs)
	a.inputCache.Range(func(key string, value []*setting.Input) bool {
		destinationToInputs[key] = value
		return true
	})
	if err := a.sendInputs(sdctx, destinationToInputs); err != nil {
		a.logger.Error(ctx, "Failed to send inputs to PropertyInspector", "error", err)
		return err
	}

	a.client.SetImage(sdctx, "", streamdeck.HardwareAndSoftware)

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
