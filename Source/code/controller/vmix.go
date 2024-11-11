package controller

import (
	"context"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/action"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/adapter/adapters"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/controller/controllers"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/solver"

	sdcontext "github.com/FlowingSPDG/streamdeck/context"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
	"golang.org/x/sync/errgroup"
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

func (v *vmixController) OnXML(ctx context.Context, host string, xr *vmixtcp.XMLResponse) error {
	ctxStrs, found := v.solver.SolveByHost(host)
	if !found {
		return xerrors.Errorf("failed to solve host %s", host)
	}

	inputs := map[string]adapters.Input{}
	for _, input := range xr.XML.Inputs.Input {
		inputs[input.Key] = adapters.Input{
			Number: int(input.Number),
			Name:   input.Title,
			Key:    input.Key,
		}
	}

	eg := errgroup.Group{}
	for _, ctxStr := range ctxStrs {
		ctx := sdcontext.WithContext(ctx, ctxStr)
		eg.Go(func() error {
			return v.streamDeckAdapter.SendInputs(ctx, inputs)
		})
	}
	if err := eg.Wait(); err != nil {
		return xerrors.Errorf("failed to send inputs: %w", err)
	}
	return nil
}
