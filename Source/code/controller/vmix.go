package controller

import (
	"context"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/action"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/adapter"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/solver"
	"golang.org/x/xerrors"
)

type VMixController interface {
	OnTally(ctx context.Context, host string, input int) error
}

type vmixController struct {
	// adapters
	vMixAdapter       adapter.VMixAdapter
	streamDeckAdapter adapter.StreamDeckContextAdapter

	// internal
	solver solver.Solver

	// actions
	previewAction action.PreviewAction
}

func NewVMixController(
	vmixAdapter adapter.VMixAdapter,
	streamDeckAdapter adapter.StreamDeckContextAdapter,
	solver solver.Solver,
	previewAction action.PreviewAction,
) VMixController {
	return &vmixController{
		vMixAdapter:       vmixAdapter,
		streamDeckAdapter: streamDeckAdapter,
		solver:            solver,
		previewAction:     previewAction,
	}
}

func (v *vmixController) OnTally(ctx context.Context, host string, input int) error {
	if err := v.previewAction.Tally(ctx, host, input); err != nil {
		return xerrors.Errorf("failed to trigger tally: %w", err)
	}
	return nil
}
