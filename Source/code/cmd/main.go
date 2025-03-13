package main

import (
	"context"

	"github.com/FlowingSPDG/streamdeck"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/action"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/di"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/setting"
)

func main() {
	ctx := context.Background()

	streamDeckClient := di.InitializeStreamDeckClient(ctx)
	sdLogger := di.InitializeLogger(streamDeckClient, logger.DebugLevel|logger.ErrorLevel)
	connectionManager := di.InitializeConnectionManager(sdLogger)
	inputCache := di.InitializeSettingStore[[]*setting.Input]()

	previewAction := di.InitializePreviewAction(sdLogger, connectionManager, streamDeckClient, inputCache)
	sdPreviewAction := streamDeckClient.Action(action.PreviewActionUUID)
	sdPreviewAction.RegisterHandler(streamdeck.WillAppear, previewAction.OnWillAppear())
	sdPreviewAction.RegisterHandler(streamdeck.WillDisappear, previewAction.OnWillDisappear())
	sdPreviewAction.RegisterHandler(streamdeck.DidReceiveSettings, previewAction.OnUpdateSettings())
	sdPreviewAction.RegisterHandler(streamdeck.KeyDown, previewAction.OnKeyDown())
	sdPreviewAction.RegisterHandler(streamdeck.SendToPlugin, previewAction.OnSendToPlugin())

	programAction := di.InitializeProgramAction(sdLogger, connectionManager, streamDeckClient, inputCache)
	sdProgramAction := streamDeckClient.Action(action.ProgramActionUUID)
	sdProgramAction.RegisterHandler(streamdeck.WillAppear, programAction.OnWillAppear())
	sdProgramAction.RegisterHandler(streamdeck.WillDisappear, programAction.OnWillDisappear())
	sdProgramAction.RegisterHandler(streamdeck.DidReceiveSettings, programAction.OnUpdateSettings())
	sdProgramAction.RegisterHandler(streamdeck.KeyDown, programAction.OnKeyDown())
	sdProgramAction.RegisterHandler(streamdeck.SendToPlugin, programAction.OnSendToPlugin())

	functionAction := di.InitializeFunctionAction(sdLogger, connectionManager, streamDeckClient, inputCache)
	sdFunctionAction := streamDeckClient.Action(action.FunctionActionUUID)
	sdFunctionAction.RegisterHandler(streamdeck.WillAppear, functionAction.OnWillAppear())
	sdFunctionAction.RegisterHandler(streamdeck.WillDisappear, functionAction.OnWillDisappear())
	sdFunctionAction.RegisterHandler(streamdeck.DidReceiveSettings, functionAction.OnUpdateSettings())
	sdFunctionAction.RegisterHandler(streamdeck.KeyDown, functionAction.OnKeyDown())
	sdFunctionAction.RegisterHandler(streamdeck.SendToPlugin, functionAction.OnSendToPlugin())

	activatorAction := di.InitializeActivatorAction(sdLogger, connectionManager, streamDeckClient, inputCache)
	sdActivatorAction := streamDeckClient.Action(action.ActivatorActionUUID)
	sdActivatorAction.RegisterHandler(streamdeck.WillAppear, activatorAction.OnWillAppear())
	sdActivatorAction.RegisterHandler(streamdeck.WillDisappear, activatorAction.OnWillDisappear())
	sdActivatorAction.RegisterHandler(streamdeck.DidReceiveSettings, activatorAction.OnUpdateSettings())
	sdActivatorAction.RegisterHandler(streamdeck.KeyDown, activatorAction.OnKeyDown())
	sdActivatorAction.RegisterHandler(streamdeck.SendToPlugin, activatorAction.OnSendToPlugin())

	connectionManager.SetXMLCallback(func(resp *vmixtcp.XMLResponse, vm vmixtcp.Vmix, addr string) {
		previewAction.OnVMixXML(ctx, resp, addr, vm)
		programAction.OnVMixXML(ctx, resp, addr, vm)
		functionAction.OnVMixXML(ctx, resp, addr, vm)
		activatorAction.OnVMixXML(ctx, resp, addr, vm)
	})

	connectionManager.SetTallyCallback(func(resp *vmixtcp.TallyResponse, vm vmixtcp.Vmix, addr string) {
		previewAction.OnVMixTally(ctx, resp, addr, vm)
		programAction.OnVMixTally(ctx, resp, addr, vm)
		functionAction.OnVMixTally(ctx, resp, addr, vm)
		activatorAction.OnVMixTally(ctx, resp, addr, vm)
		if err := vm.XML(); err != nil {
			sdLogger.Error(ctx, "Failed to get XML: %v", err)
		}
	})
	connectionManager.SetActsCallback(func(resp *vmixtcp.ActsResponse, vm vmixtcp.Vmix, addr string) {
		previewAction.OnVMixActs(ctx, resp, addr, vm)
		programAction.OnVMixActs(ctx, resp, addr, vm)
		functionAction.OnVMixActs(ctx, resp, addr, vm)
		activatorAction.OnVMixActs(ctx, resp, addr, vm)
	})
	connectionManager.SetVersionCallback(func(resp *vmixtcp.VersionResponse, vm vmixtcp.Vmix, addr string) {
		previewAction.OnVMixVersion(ctx, resp, addr, vm)
		programAction.OnVMixVersion(ctx, resp, addr, vm)
		functionAction.OnVMixVersion(ctx, resp, addr, vm)
		activatorAction.OnVMixVersion(ctx, resp, addr, vm)

		if err := vm.Subscribe(vmixtcp.EventTally, ""); err != nil {
			sdLogger.Error(ctx, "Failed to subscribe TALLY: %v", err)
		}
		if err := vm.Subscribe(vmixtcp.EventActs, ""); err != nil {
			sdLogger.Error(ctx, "Failed to subscribe ACTS: %v", err)
		}

		if err := vm.Tally(); err != nil {
			sdLogger.Error(ctx, "Failed to get tally: %v", err)
		}
	})
	connectionManager.SetSubscribeCallback(func(resp *vmixtcp.SubscribeResponse, vm vmixtcp.Vmix, addr string) {
		previewAction.OnVMixSubscribe(ctx, resp, addr, vm)
		programAction.OnVMixSubscribe(ctx, resp, addr, vm)
		functionAction.OnVMixSubscribe(ctx, resp, addr, vm)
		activatorAction.OnVMixSubscribe(ctx, resp, addr, vm)
	})

	if err := streamDeckClient.Run(ctx); err != nil {
		panic(err)
	}
}
