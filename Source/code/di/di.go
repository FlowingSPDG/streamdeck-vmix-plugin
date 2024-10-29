package di

import (
	"context"
	"os"

	"github.com/FlowingSPDG/streamdeck"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/action"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/adapter"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/controller"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/pool"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/setting"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/solver"
)

func InitializeStreamDeckClient(ctx context.Context) *streamdeck.Client {
	params, err := streamdeck.ParseRegistrationParams(os.Args)
	if err != nil {
		panic(err)
	}
	streamDeckClient := streamdeck.NewClient(ctx, params)
	return streamDeckClient
}

func InitializePreviewActionController(ctx context.Context, streamDeckClient *streamdeck.Client) controller.PreviewActionController {
	// misc
	logger := logger.NewStreamDeckLogger(streamDeckClient)
	solver := solver.NewSolver()
	vMixPool := pool.NewVMixPool(logger, solver)
	store := setting.NewSettingStore[setting.PreviewSetting]()

	// adapters
	streamDeckAdapter := adapter.NewStreamDeckContextAdapter(streamDeckClient)
	vMixAdapter := adapter.NewVMixAdapter(vMixPool)

	// action
	previewAction := action.NewPreviewAction(logger, streamDeckAdapter, vMixAdapter, solver, store)

	// controller
	ctrl := controller.NewPreviewActionController(previewAction)

	return ctrl
}
