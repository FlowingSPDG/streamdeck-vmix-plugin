package pool

import (
	"context"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/solver"
	sdcontext "github.com/FlowingSPDG/streamdeck/context"

	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
	"github.com/puzpuzpuz/xsync/v3"
)

// VMixPool はvMixへの接続プールを自動管理し、再接続処理などを担うインターフェース
type VMixPool interface {
	Add(ctx context.Context, host string)
	Remove(ctx context.Context)
}

type vmixPool struct {
	// logger
	logger logger.Logger

	// solver
	solver solver.Solver

	// vMix instances
	vs *xsync.MapOf[string, *vmixInstance]
}

type vmixInstance struct {
	vmix   vmixtcp.Vmix
	ctx    context.Context
	cancel context.CancelFunc
}

// NewVMixPool はVMixPoolを生成する
func NewVMixPool(
	logger logger.Logger,
	solver solver.Solver,
) VMixPool {
	return &vmixPool{
		logger: logger,
		solver: solver,
		vs:     xsync.NewMapOf[string, *vmixInstance](),
	}
}

func (v *vmixPool) Add(ctx context.Context, host string) {
	v.logger.LogMessage(ctx, "add vmix host: %s", host)

	ctxStr := sdcontext.Context(ctx)
	if ctxStr == "" {
		panic("context is not registered")
	}

	// 紐づけ登録をする
	v.solver.AddHost(host, ctxStr)

	// canceler
	cctx, cancel := context.WithCancel(context.Background())

	// vMixのインスタンスを保持する
	vmix := vmixtcp.New(host)
	v.vs.Store(host, &vmixInstance{
		vmix:   vmix,
		ctx:    cctx,
		cancel: cancel,
	})
	v.startRetry(cctx, host)
}

func (v *vmixPool) Remove(ctx context.Context) {
	ctxStr := sdcontext.Context(ctx)
	removed := v.solver.RemoveContext(ctxStr)
	if removed {
		// vMixのインスタンスを削除する
		vi, ok := v.vs.Load(ctxStr)
		if ok {
			panic("SOMETHING WENT WRONG!!")
		}

		vi.cancel()
		v.vs.Delete(ctxStr)
	}
}
