package action

import (
	"context"
	"errors"
	"slices"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/adapter/adapters"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger/loggers"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/setting"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/solver"

	"github.com/FlowingSPDG/streamdeck"
	sdcontext "github.com/FlowingSPDG/streamdeck/context"
	vmixhttp "github.com/FlowingSPDG/vmix-go/http"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

const PreviewActionUUID = "dev.flowingspdg.vmix.preview"

type PreviewAction interface {
	Appear(ctx context.Context, setting *setting.PreviewSetting) error
	Disappear(ctx context.Context, setting *setting.PreviewSetting) error
	UpdateSetting(ctx context.Context, setting *setting.PreviewSetting) error
	Execute(ctx context.Context) error
	Tally(ctx context.Context, host string, tally *vmixtcp.TallyResponse) error
}

type previewAction struct {
	logger loggers.Logger

	// adapters
	streamDeckAdapter adapters.StreamDeckContextAdapter
	vmixAdapter       adapters.VMixAdapter

	// solver
	solver solver.Solver

	// internal
	store setting.SettingStore[setting.PreviewSetting]
}

func NewPreviewAction(
	logger loggers.Logger,
	streamDeckAdapter adapters.StreamDeckContextAdapter,
	vmixAdapter adapters.VMixAdapter,
	solver solver.Solver,
	store setting.SettingStore[setting.PreviewSetting],
) PreviewAction {
	return &previewAction{
		logger:            logger,
		streamDeckAdapter: streamDeckAdapter,
		vmixAdapter:       vmixAdapter,
		solver:            solver,
		store:             store,
	}
}

func (p *previewAction) Appear(ctx context.Context, setting *setting.PreviewSetting) error {
	if err := p.logger.LogMessage(ctx, "preview action appeared"); err != nil {
		return xerrors.Errorf("failed to log message: %w", err)
	}

	p.storeNewVmix(ctx, setting)

	return nil
}

func (p *previewAction) Disappear(ctx context.Context, setting *setting.PreviewSetting) error {
	if err := p.logger.LogMessage(ctx, "preview action disappeared"); err != nil {
		return xerrors.Errorf("failed to log message: %w", err)
	}

	ctxStr := sdcontext.Context(ctx)
	if ctxStr == "" {
		return errors.New("failed to get context")
	}

	p.store.Delete(ctxStr)
	p.vmixAdapter.RemoveVMix(ctx)

	return nil
}

func (p *previewAction) UpdateSetting(ctx context.Context, setting *setting.PreviewSetting) error {
	if err := p.logger.LogMessage(ctx, "preview action received updated setting: %v", setting); err != nil {
		return xerrors.Errorf("failed to log message: %w", err)
	}

	// 既存の設定と新しい設定を比較して、設定が変更されていれば更新する
	ctxStr := sdcontext.Context(ctx)
	if ctxStr == "" {
		return errors.New("failed to get context")
	}
	if err := p.logger.LogMessage(ctx, "got context: %s", ctxStr); err != nil {
		return xerrors.Errorf("failed to log message: %w", err)
	}

	// ここで古いvMixのインスタンスを削除する
	// p.vmixAdapter.RemoveVMix(ctx) // ??
	p.storeNewVmix(ctx, setting)

	// PIに新しいInputsを送信する
	vc, err := vmixhttp.NewClient(setting.Host, 8088)
	if err != nil {
		return xerrors.Errorf("failed to create vmix http client: %w", err)
	}

	inputs := map[string]adapters.Input{}

	for _, i := range vc.Inputs.Input {
		inputs[i.Key] = adapters.Input{
			Name:   i.Title,
			Number: int(i.Number),
			Key:    i.Key,
		}
	}
	if err := p.logger.LogMessage(ctx, "got inputs: %v", inputs); err != nil {
		return xerrors.Errorf("failed to log message: %w", err)
	}

	p.streamDeckAdapter.SendInputs(ctx, setting.Host, inputs)

	return nil
}

func (p *previewAction) Execute(ctx context.Context) error {
	if err := p.logger.LogMessage(ctx, "preview action executing"); err != nil {
		return xerrors.Errorf("failed to log message: %w", err)
	}

	ctxStr := sdcontext.Context(ctx)
	if ctxStr == "" {
		return errors.New("failed to get context")
	}

	s, ok := p.store.Load(ctxStr)
	if !ok {
		return errors.New("failed to get settings for context " + ctxStr)
	}

	if err := p.vmixAdapter.PreviewInput(s.Host, s.Input); err != nil {
		return err
	}

	return nil
}

func (p *previewAction) Tally(ctx context.Context, host string, tally *vmixtcp.TallyResponse) error {
	if err := p.logger.LogMessage(ctx, "preview action received tally signal"); err != nil {
		return xerrors.Errorf("failed to log message: %w", err)
	}

	contextStrs, ok := p.solver.SolveByHost(ctx, host)
	if !ok {
		return xerrors.Errorf("unknown host detected: %s", host)
	}

	contextStrs = slices.DeleteFunc(contextStrs, func(s string) bool {
		setting, ok := p.store.Load(s)
		if !ok {
			return false
		}
		return !setting.Tally || setting.Host != host || setting.Input == 0 || setting.Input > len(tally.Tally)
	})

	eg := errgroup.Group{}
	for _, contextStr := range contextStrs {
		eg.Go(func() error {
			if err := p.logger.LogMessage(ctx, "Applying tally for context %s", contextStr); err != nil {
				return xerrors.Errorf("failed to log message: %w", err)
			}

			cctx := sdcontext.WithContext(ctx, contextStr)

			// PIの設定を読み出す
			s, ok := p.store.Load(contextStr)
			if !ok {
				return xerrors.Errorf("failed to get settings for context %s", contextStr)
			}

			// 設定情報をもとに、Tallyをセットする
			target := s.Input - 1
			if tally.Tally[target] == vmixtcp.Preview {
				if err := p.streamDeckAdapter.SetPreviewColor(cctx, streamdeck.HardwareAndSoftware); err != nil {
					return xerrors.Errorf("failed to set tally: %w", err)
				}
				return nil
			}

			if err := p.logger.LogMessage(ctx, "Applying Preview tally for context %s", contextStr); err != nil {
				return xerrors.Errorf("failed to log message: %w", err)
			}
			if err := p.streamDeckAdapter.SetInactiveColor(cctx, streamdeck.HardwareAndSoftware); err != nil {
				return xerrors.Errorf("failed to set tally: %w", err)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return xerrors.Errorf("failed to apply tally: %w", err)
	}
	return nil
}

func (p *previewAction) storeNewVmix(ctx context.Context, setting *setting.PreviewSetting) error {
	ctxStr := sdcontext.Context(ctx)
	if ctxStr == "" {
		return errors.New("failed to get context")
	}

	p.store.Store(ctxStr, setting)
	p.vmixAdapter.AddVMix(ctx, setting.Host)
	return nil
}
