package main

import (
	"context"

	"github.com/FlowingSPDG/streamdeck"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/action"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/di"
)

func main() {
	ctx := context.Background()

	streamDeckClient := di.InitializeStreamDeckClient(ctx)
	fileLogger := di.InitializeFileLogger(ctx)
	multiLogger := di.InitializeMultiLogger(fileLogger)
	connectionManager := di.InitializeConnectionManager(multiLogger)
	inputCache := di.InitializeSettingStore[[]*action.Input]()

	previewAction := di.InitializePreviewAction(multiLogger, connectionManager, streamDeckClient, inputCache)
	sdPreviewAction := streamDeckClient.Action(action.PreviewActionUUID)
	sdPreviewAction.RegisterHandler(streamdeck.WillAppear, previewAction.OnWillAppear())
	sdPreviewAction.RegisterHandler(streamdeck.WillDisappear, previewAction.OnWillDisappear())
	sdPreviewAction.RegisterHandler(streamdeck.DidReceiveSettings, previewAction.OnUpdateSettings())
	sdPreviewAction.RegisterHandler(streamdeck.KeyDown, previewAction.OnKeyDown())
	connectionManager.SetXMLCallback(func(resp *vmixtcp.XMLResponse, vm vmixtcp.Vmix, addr string) {
		go previewAction.OnVMixXML(ctx, resp, addr)
	})
	connectionManager.SetTallyCallback(func(resp *vmixtcp.TallyResponse, vm vmixtcp.Vmix, addr string) {
		go previewAction.OnVMixTally(ctx, resp, addr)
		go vm.XML()
	})
	connectionManager.SetActsCallback(func(resp *vmixtcp.ActsResponse, vm vmixtcp.Vmix, addr string) {
		go previewAction.OnVMixActs(ctx, resp, addr)
	})
	streamDeckClient.Run(ctx)
}
