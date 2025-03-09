package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/adapter/adapters"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger/loggers"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/solver"

	sdcontext "github.com/FlowingSPDG/streamdeck/context"
	"github.com/FlowingSPDG/vmix-go/common/models"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/xerrors"
)

type vMixAdapter struct {
	// callbacks
	onTally func(ctx context.Context, host string, tally *vmixtcp.TallyResponse) error
	// onVersion func(ctx context.Context, host string, version *vmixtcp.VersionResponse) error
	onXML func(ctx context.Context, host string, xml *vmixtcp.XMLResponse) error

	// logger
	logger loggers.Logger

	// solver
	solver solver.Solver

	// vMix instances
	// TODO: solverへ共通化できるかも？
	vs *xsync.MapOf[string, *vmixInstance]
}

type vmixInstance struct {
	vmix   vmixtcp.Vmix
	inputs []models.Input
	ctx    context.Context
	cancel context.CancelFunc
}

func NewVMixAdapter(
	logger loggers.Logger,
	solver solver.Solver,
) adapters.VMixAdapter {
	return &vMixAdapter{
		logger: logger,
		solver: solver,
		vs:     xsync.NewMapOf[string, *vmixInstance](),
	}
}

func (v *vMixAdapter) PreviewInput(ctx context.Context, destination string, input int) error {
	vi, ok := v.vs.Load(destination)
	if !ok {
		return xerrors.New("failed to load vmix instance")
	}

	if err := vi.vmix.Function("PreviewInput", fmt.Sprintf("Input=%d", input)); err != nil {
		return xerrors.Errorf("failed to preview input: %w", err)
	}

	return nil
}

func (v *vMixAdapter) AddVMix(ctx context.Context, destination string) {
	v.logger.LogMessage(ctx, "add vmix host: %s", destination)

	ctxStr := sdcontext.Context(ctx)
	if ctxStr == "" {
		panic("context is not registered")
	}

	// 紐づけ登録をする
	v.solver.AddHost(ctx, destination, ctxStr)

	// canceler
	cctx, cancel := context.WithCancel(context.Background())

	// vMixのインスタンスを保持する
	vmix := vmixtcp.New(destination)
	v.vs.Store(destination, &vmixInstance{
		vmix:   vmix,
		ctx:    cctx,
		cancel: cancel,
	})
	v.startRetry(cctx, cancel, destination)
}

func (v *vMixAdapter) RemoveVMix(ctx context.Context) {
	ctxStr := sdcontext.Context(ctx)
	removed := v.solver.RemoveContext(ctx, ctxStr)
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

func (v *vMixAdapter) startRetry(ctx context.Context, cancel context.CancelFunc, host string) {
	v.logger.LogMessage(ctx, "start retry for %s", host)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				if err := v.retry(ctx, cancel, host); err != nil {
					v.logger.LogMessage(ctx, "failed to retry: %v", err)
				}
			}
		}
	}()

}

func (v *vMixAdapter) retry(ctx context.Context, cancel context.CancelFunc, host string) error {
	v.logger.LogMessage(ctx, "retrying for %s", host)

	// vMixのインスタンスが削除されている場合、再接続処理を行わない
	vi, ok := v.vs.Load(host)
	if !ok {
		v.logger.LogMessage(ctx, "destination %s is probably deleted. Abort!", host)
		cancel()
		return nil
	}

	if vi.vmix.IsConnected() {
		v.logger.LogMessage(ctx, "destination %s is already connected. Abort!", host)
		cancel()
		return nil
	}

	v.logger.LogMessage(ctx, "Trying to connect destination %s ...", host)

	// 1.接続処理を行う
	if err := vi.vmix.Connect(vi.ctx, time.Second); err != nil {
		return xerrors.Errorf("failed to connect to vMix: %w", err)
	}
	v.logger.LogMessage(ctx, "connected to vMix destination %s. Register callbacks...", host)

	// 2: コールバックを登録する
	vi.vmix.OnVersion(func(vr *vmixtcp.VersionResponse, err error) {
		if err != nil {
			panic(err)
		}
		// バージョン情報を受け取ったときの処理
		v.logger.LogMessage(ctx, "VersionResponse: %v", vr)
		if err := vi.vmix.Subscribe(vmixtcp.EventTally, ""); err != nil {
			panic(err)
		}

		if err := vi.vmix.XML(); err != nil {
			panic(err)
		}
	})
	vi.vmix.OnTally(func(tr *vmixtcp.TallyResponse, err error) {
		if err != nil {
			panic(err)
		}
		// Tally情報を受け取ったときの処理
		v.logger.LogMessage(ctx, "TallyResponse: %v", tr)
		if err := vi.vmix.XML(); err != nil {
			panic(err)
		}
		v.onTally(ctx, host, tr)
	})
	vi.vmix.OnXML(func(xr *vmixtcp.XMLResponse, err error) {
		if err != nil {
			panic(err)
		}
		// XML情報を受け取ったときの処理
		v.logger.LogMessage(ctx, "XMLResponse: %v", xr)
		vi.inputs = xr.XML.Inputs.Input
		v.onXML(ctx, host, xr)
	})
	// 追加でACTSにSUBSCRIBEする場合、設定項目からSUBSCRIBE対象を取得する

	v.logger.LogMessage(ctx, "Registered callbacks %s", host)

	// 3: 接続に成功したら、vMixの状態を監視する
	if err := vi.vmix.Run(vi.ctx); err != nil {
		return xerrors.Errorf("failed to run vMix: %w", err)
	}

	return nil
}

func (v *vMixAdapter) OnTally(f func(ctx context.Context, host string, tally *vmixtcp.TallyResponse) error) {
	v.onTally = f
}

func (v *vMixAdapter) OnXML(f func(ctx context.Context, host string, xml *vmixtcp.XMLResponse) error) {
	v.onXML = f
}
