package di

import (
	"context"
	"os"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/action"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/adapter"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/controller"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/controller/controllers"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/setting"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/solver"

	"github.com/FlowingSPDG/streamdeck"
)

func InitializeStreamDeckClient(ctx context.Context) *streamdeck.Client {
	params, err := streamdeck.ParseRegistrationParams(os.Args)
	if err != nil {
		panic(err)
	}
	streamDeckClient := streamdeck.NewClient(ctx, params)
	return streamDeckClient
}

func InitializePreviewActionController(ctx context.Context, streamDeckClient *streamdeck.Client) controllers.PreviewActionController {
	// misc
	logger := logger.NewStreamDeckLogger(streamDeckClient)
	solver := solver.NewSolver()
	store := setting.NewSettingStore[setting.PreviewSetting]()

	// adapters
	vMixAdapter := adapter.NewVMixAdapter(logger, solver)
	streamDeckAdapter := adapter.NewStreamDeckContextAdapter(streamDeckClient)

	// action
	previewAction := action.NewPreviewAction(logger, streamDeckAdapter, vMixAdapter, solver, store)

	// controller
	previewController := controller.NewPreviewActionController(previewAction)
	vMixController := controller.NewVMixController(vMixAdapter, streamDeckAdapter, solver, previewAction)

	// register callbacks
	vMixAdapter.OnTally(vMixController.OnTally)

	return previewController
}
