package main

import (
	"context"

	"github.com/FlowingSPDG/streamdeck"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/action"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/di"
)

func main() {
	ctx := context.Background()

	streamDeckClient := di.InitializeStreamDeckClient(ctx)
	logger := di.InitializeLogger(streamDeckClient)
	connectionManager := di.InitializeConnectionManager(logger)

	previewAction := di.InitializePreviewAction(logger, connectionManager)
	sdPreviewAction := streamDeckClient.Action(action.PreviewActionUUID)
	sdPreviewAction.RegisterHandler(streamdeck.WillAppear, previewAction.OnWillAppear())
	sdPreviewAction.RegisterHandler(streamdeck.WillDisappear, previewAction.OnWillDisappear())
	sdPreviewAction.RegisterHandler(streamdeck.DidReceiveSettings, previewAction.OnUpdateSettings())
	streamDeckClient.Run(ctx)
}
