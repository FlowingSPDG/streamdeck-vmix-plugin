package main

import (
	"context"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/action"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/di"
)

func main() {
	ctx := context.Background()

	streamDeckClient := di.InitializeStreamDeckClient(ctx)
	previewActionController := di.InitializePreviewActionController(ctx, streamDeckClient)
	previewAction := streamDeckClient.Action(action.PreviewActionUUID)
	previewActionController.RegisterAction(previewAction)

	streamDeckClient.Run(ctx)
}
