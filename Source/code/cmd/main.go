package main

import (
	"context"

	"github.com/FlowingSPDG/streamdeck"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/action"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/di"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger"
)

func main() {
	ctx := context.Background()

	streamDeckClient := di.InitializeStreamDeckClient(ctx)
	sdLogger := di.InitializeLogger(streamDeckClient, logger.DebugLevel)
	connectionManager := di.InitializeConnectionManager(sdLogger)
	inputCache := di.InitializeSettingStore[[]*action.Input]()

	previewAction := di.InitializePreviewAction(sdLogger, connectionManager, streamDeckClient, inputCache)
	sdPreviewAction := streamDeckClient.Action(action.PreviewActionUUID)
	sdPreviewAction.RegisterHandler(streamdeck.WillAppear, previewAction.OnWillAppear())
	sdPreviewAction.RegisterHandler(streamdeck.WillDisappear, previewAction.OnWillDisappear())
	sdPreviewAction.RegisterHandler(streamdeck.DidReceiveSettings, previewAction.OnUpdateSettings())
	sdPreviewAction.RegisterHandler(streamdeck.KeyDown, previewAction.OnKeyDown())
	connectionManager.SetXMLCallback(func(resp *vmixtcp.XMLResponse, vm vmixtcp.Vmix, addr string) {
		previewAction.OnVMixXML(ctx, resp, addr, vm)
	})
	connectionManager.SetTallyCallback(func(resp *vmixtcp.TallyResponse, vm vmixtcp.Vmix, addr string) {
		previewAction.OnVMixTally(ctx, resp, addr, vm)
	})
	connectionManager.SetActsCallback(func(resp *vmixtcp.ActsResponse, vm vmixtcp.Vmix, addr string) {
		previewAction.OnVMixActs(ctx, resp, addr, vm)
	})
	connectionManager.SetVersionCallback(func(resp *vmixtcp.VersionResponse, vm vmixtcp.Vmix, addr string) {
		previewAction.OnVMixVersion(ctx, resp, addr, vm)
	})
	streamDeckClient.Run(ctx)
}
