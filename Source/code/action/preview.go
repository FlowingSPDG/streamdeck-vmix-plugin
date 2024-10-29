package action

import (
	"context"
	"errors"

	"github.com/FlowingSPDG/streamdeck"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/adapter"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/setting"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/solver"
	"golang.org/x/xerrors"

	sdcontext "github.com/FlowingSPDG/streamdeck/context"
)

const PreviewActionUUID = "dev.flowingspdg.vmix.preview"

type PreviewAction interface {
	Appear(ctx context.Context, setting *setting.PreviewSetting) error
	Disappear(ctx context.Context, setting *setting.PreviewSetting) error
	UpdateSetting(ctx context.Context, setting *setting.PreviewSetting) error
	Execute(ctx context.Context) error
	Tally(ctx context.Context, host string, input int) error
}

type previewAction struct {
	logger logger.Logger

	// adapters
	streamDeckAdapter adapter.StreamDeckContextAdapter
	vmixAdapter       adapter.VMixAdapter

	// solver
	solver solver.Solver

	// internal
	store setting.SettingStore[setting.PreviewSetting]
}

func NewPreviewAction(
	logger logger.Logger,
	streamDeckAdapter adapter.StreamDeckContextAdapter,
	vmixAdapter adapter.VMixAdapter,
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
	oldSetting, _ := p.store.Load(ctxStr)
	if oldSetting.Host == setting.Host && oldSetting.Input == setting.Input {
		return nil
	}

	// ここで古いvMixのインスタンスを削除する
	p.vmixAdapter.RemoveVMix(ctx)
	p.storeNewVmix(ctx, setting)

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

func (p *previewAction) Tally(ctx context.Context, host string, input int) error {
	if err := p.logger.LogMessage(ctx, "preview action received tally signal"); err != nil {
		return xerrors.Errorf("failed to log message: %w", err)
	}

	contextStrs, ok := p.solver.SolveByHost(host)
	if !ok {
		return xerrors.Errorf("unknown host detected: %s", host)
	}

	for _, contextStr := range contextStrs {
		cctx := sdcontext.WithContext(ctx, contextStr)

		// PIの設定を読み出す
		s, ok := p.store.Load(contextStr)
		if !ok {
			return xerrors.Errorf("failed to get settings for context %s", contextStr)
		}

		// 設定情報をもとに、Tallyをセットする
		if s.Input == input {
			if err := p.streamDeckAdapter.SetPreviewColor(cctx, streamdeck.HardwareAndSoftware); err != nil {
				return xerrors.Errorf("failed to set tally: %w", err)
			}
		} else {
			if err := p.streamDeckAdapter.SetInactiveColor(cctx, streamdeck.HardwareAndSoftware); err != nil {
				return xerrors.Errorf("failed to set tally: %w", err)
			}
		}
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
