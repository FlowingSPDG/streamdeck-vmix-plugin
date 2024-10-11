package actions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/FlowingSPDG/streamdeck"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/connections"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/settings"

	"golang.org/x/xerrors"
)

type PreviewAction struct {
	settings settings.SettingStore[PreviewPI]
	vc       *connections.VMixCommunicators
}

func NewPreviewAction(vc *connections.VMixCommunicators) *PreviewAction {
	return &PreviewAction{
		settings: settings.NewSettingStore[PreviewPI](),
		vc:       vc,
	}
}

type PreviewPI struct {
	Dest  string `json:"dest"`
	Input int    `json:"input"`
	Tally bool   `json:"tally"`
}

func (p PreviewPI) IsDefault() bool {
	return reflect.ValueOf(p).IsZero()
}

func (p *PreviewPI) Initialize() {
	p.Dest = "localhost"
	p.Input = 0
	p.Tally = true
}

func (s *PreviewAction) Execute(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	setting, ok := s.settings.Load(event.Context)
	if !ok {
		return errors.New("failed to get settings for context " + event.Context)
	}

	vc, found := s.vc.FindByContext(event.Context)
	if !found {
		return errors.New("failed to get vMix for context " + event.Context)
	}
	raw := vc.GetRaw()
	if err := raw.Function("PreviewInput", fmt.Sprintf("Input=%d", setting.Input)); err != nil {
		return xerrors.Errorf("failed to execute function : %w", err)
	}

	return nil
}

func (s *PreviewAction) WillAppear(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.WillAppearPayload[*PreviewPI]{}
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return err
	}

	if p.Settings.IsDefault() {
		p.Settings.Initialize()
		msg := fmt.Sprintf("Forcing Default value:%v", p.Settings)
		client.LogMessage(msg)
		if err := client.SetSettings(ctx, p.Settings); err != nil {
			return xerrors.Errorf("Failed to save settings : %w", err)
		}
	}

	go s.settings.Store(event.Context, p.Settings)
	s.vc.AddvMix(ctx, p.Settings.Dest, event.Context)
	return nil
}

func (s *PreviewAction) WillDisappear(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	s.settings.Delete(event.Context)
	s.vc.RemovevMixByContext(ctx, event.Context)
	return nil
}

// TODO: PI側の更新をバックエンドに反映する

func (s *PreviewAction) GetSetting(ctxStr string) (*PreviewPI, bool) {
	return s.settings.Load(ctxStr)
}
