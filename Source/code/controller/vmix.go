package controller

import (
	"context"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/action"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/adapter/adapters"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/controller/controllers"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/solver"

	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
	"golang.org/x/xerrors"
)

type vmixController struct {
	// adapters
	vMixAdapter       adapters.VMixAdapter
	streamDeckAdapter adapters.StreamDeckContextAdapter

	// internal
	solver solver.Solver

	// actions
	previewAction action.PreviewAction
}

func NewVMixController(
	vmixAdapter adapters.VMixAdapter,
	streamDeckAdapter adapters.StreamDeckContextAdapter,
	solver solver.Solver,
	previewAction action.PreviewAction,
) controllers.VMixController {
	return &vmixController{
		vMixAdapter:       vmixAdapter,
		streamDeckAdapter: streamDeckAdapter,
		solver:            solver,
		previewAction:     previewAction,
	}
}

func (v *vmixController) OnTally(ctx context.Context, host string, tally *vmixtcp.TallyResponse) error {
	if err := v.previewAction.Tally(ctx, host, tally); err != nil {
		return xerrors.Errorf("failed to trigger tally: %w", err)
	}
	return nil
}
