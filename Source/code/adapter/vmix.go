package adapter

import (
	"context"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/pool"
)

type VMixAdapter interface {
	AddVMix(ctx context.Context, destination string)
	RemoveVMix(ctx context.Context)
	PreviewInput(destination string, input int) error
}

type vMixAdapter struct {
	pool pool.VMixPool
}

func NewVMixAdapter(pool pool.VMixPool) VMixAdapter {
	return &vMixAdapter{
		pool: pool,
	}
}

func (v *vMixAdapter) PreviewInput(destination string, input int) error {
	panic("not implemented") // TODO: Implement
}

func (v *vMixAdapter) AddVMix(ctx context.Context, destination string) {
	v.pool.Add(ctx, destination)
}

func (v *vMixAdapter) RemoveVMix(ctx context.Context) {
	v.pool.Remove(ctx)
}
