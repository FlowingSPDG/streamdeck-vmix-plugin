package actions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"

	"github.com/FlowingSPDG/streamdeck"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/connections"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/settings"

	"golang.org/x/xerrors"
)

type SendFuncAction struct {
	settings settings.SettingStore[SendFunctionPI]
	vc       *connections.VMixCommunicators
}

func NewSendFuncAction(vc *connections.VMixCommunicators) *SendFuncAction {
	return &SendFuncAction{
		settings: settings.NewSettingStore[SendFunctionPI](),
		vc:       vc,
	}
}

type Query struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Queries []Query

func (qs Queries) ToString() string {
	u := &url.URL{}
	q := u.Query()
	for _, query := range qs {
		q.Add(query.Key, query.Value)
	}
	return q.Encode()
}

// SendFunctionPI Settings for each button to save persistantly on action instance
// TODO: Support ACT Tally
type SendFunctionPI struct {
	Dest    string  `json:"dest"`
	Input   int  `json:"input"`
	Name    string  `json:"name"`
	Queries Queries `json:"queries"`
}

func (p SendFunctionPI) IsDefault() bool {
	return reflect.ValueOf(p).IsZero()
}

func (p *SendFunctionPI) Initialize() {
	p.Dest = "localhost"
	p.Input = 0
	p.Name = "PreviewInput"
	p.Queries = []Query{}
}

func (p SendFunctionPI) ToQuery() string {
	p.Queries = append(p.Queries, Query{
		Key:   "Input",
		Value: fmt.Sprint(p.Input),
	})
	return p.Queries.ToString()
}

func (s *SendFuncAction) Execute(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	setting, ok := s.settings.Load(event.Context)
	if !ok {
		return errors.New("failed to get settings for context " + event.Context)
	}

	vc, found := s.vc.FindByContext(event.Context)
	if !found {
		return errors.New("failed to get vMix for context " + event.Context)
	}
	raw := vc.GetRaw()
	if err := raw.Function(setting.Name, setting.ToQuery()); err != nil {
		return xerrors.Errorf("failed to execute function : %w", err)
	}

	return nil
}

func (s *SendFuncAction) WillAppear(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	p := streamdeck.WillAppearPayload[*SendFunctionPI]{}
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

func (s *SendFuncAction) WillDisappear(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	s.settings.Delete(event.Context)
	s.vc.RemovevMixByContext(ctx, event.Context)
	return nil
}

// TODO: PI側の更新をバックエンドに反映する

func (s *SendFuncAction) GetSetting(ctxStr string) (*SendFunctionPI, bool) {
	return s.settings.Load(ctxStr)
}
