package stdvmix

import (
	"fmt"
	"reflect"

	vmixhttp "github.com/FlowingSPDG/vmix-go/http"
)

// SendFunctionPI Settings for each button to save persistantly on action instance
type SendFunctionPI struct {
	Host    string  `json:"host"`
	Port    int     `json:"port,string"`
	Input   string  `json:"input"`
	Inputs  []input `json:"inputs"`
	Name    string  `json:"name"`
	Queries []Query `json:"queries"`
}

type Query struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (p SendFunctionPI) IsDefault() bool {
	return reflect.ValueOf(p).IsZero()
}

func (p *SendFunctionPI) Initialize() {
	p.Host = "localhost"
	p.Port = 8088
	p.Input = "0"
	p.Name = "PreviewInput"
	p.Inputs = []input{}
	p.Queries = []Query{}
}

func (p SendFunctionPI) Execute() error {
	vc, err := vmixhttp.NewClient(p.Host, p.Port)
	if err != nil {
		return err
	}
	params := make(map[string]string)
	for _, query := range p.Queries {
		params[query.Key] = query.Value
	}
	return vc.SendFunction(p.Name, params)
}

// UpdateInputs 自身のInputsを更新する(本当は同じリクエストを何度も送りたくないのでキャッシュしたい)
func (p *SendFunctionPI) UpdateInputs() error {
	if p.Host == "" || p.Port == 0 {
		return nil // HostかPortがゼロ値の場合何もしない
	}
	vc, err := vmixhttp.NewClient(p.Host, p.Port)
	if err != nil {
		return err
	}
	// スライスをリセットして更新
	p.Inputs = make([]input, 0, len(vc.Inputs.Input))
	for _, i := range vc.Inputs.Input {
		p.Inputs = append(p.Inputs, input{
			Name:   i.Name,
			Key:    i.Key,
			Number: int(i.Number),
		})
	}
	return nil
}

// PreviewPI Property Inspector info for Preview
type PreviewPI struct {
	Host   string  `json:"host"`
	Port   int     `json:"port,string"`
	Input  string  `json:"input"`
	Inputs []input `json:"inputs"`
	Mix    string  `json:"mix"`
	Tally  bool    `json:"tally"`
}

func (p PreviewPI) IsDefault() bool {
	return reflect.ValueOf(p).IsZero()
}

func (p *PreviewPI) Initialize() {
	p.Host = "localhost"
	p.Port = 8088
	p.Input = "0"
	p.Inputs = []input{}
	p.Mix = ""
	p.Tally = false
}

func (p PreviewPI) Execute() error {
	vc, err := vmixhttp.NewClient(p.Host, p.Port)
	if err != nil {
		return err
	}
	params := make(map[string]string)
	params["Input"] = p.Input
	params["Mix"] = p.Mix
	return vc.SendFunction("PreviewInput", params)
}

// UpdateTally タリーを更新、点灯する必要がある場合trueが帰る
func (p PreviewPI) UpdateTally() (bool, error) {
	if p.Host == "" || p.Port == 0 {
		return false, nil // HostかPortがゼロ値の場合何もしない
	}
	vc, err := vmixhttp.NewClient(p.Host, p.Port)
	if err != nil {
		return false, err
	}
	// 以下コードだとpanicが起きそう
	// vc.Inputs.Input[vc.Preview-1].Key == p.Input, nil
	for _, input := range vc.Inputs.Input {
		// 一致するinputがあればtrueを返す
		if input.Key != p.Input {
			continue
		}
		return input.Number == vc.Preview, nil
	}
	return false, fmt.Errorf("No input found")
}

// UpdateInputs 自身のInputsを更新する(本当は同じリクエストを何度も送りたくないのでキャッシュしたい)
func (p *PreviewPI) UpdateInputs() error {
	if p.Host == "" || p.Port == 0 {
		return nil // HostかPortがゼロ値の場合何もしない
	}
	vc, err := vmixhttp.NewClient(p.Host, p.Port)
	if err != nil {
		return err
	}
	// スライスをリセットして更新
	p.Inputs = make([]input, 0, len(vc.Inputs.Input))
	for _, i := range vc.Inputs.Input {
		p.Inputs = append(p.Inputs, input{
			Name:   i.Name,
			Key:    i.Key,
			Number: int(i.Number),
		})
	}
	return nil
}

// ProgramPI Property Inspector info for PGM(Cut)
type ProgramPI struct {
	Host      string  `json:"host"`
	Port      int     `json:"port,string"`
	Input     string  `json:"input"`
	Inputs    []input `json:"inputs"`
	Mix       string  `json:"mix"`
	CutDirect bool    `json:"cut_direct"`
	Tally     bool    `json:"tally"`
}

func (p ProgramPI) IsDefault() bool {
	return reflect.ValueOf(p).IsZero()
}

func (p *ProgramPI) Initialize() {
	p.Host = "localhost"
	p.Port = 8088
	p.Input = "0"
	p.Inputs = []input{}
	p.Mix = ""
	p.CutDirect = false
	p.Tally = false
}

func (p ProgramPI) Execute() error {
	vc, err := vmixhttp.NewClient(p.Host, p.Port)
	if err != nil {
		return err
	}
	cut := "Cut"
	if p.CutDirect {
		cut = "CutDirect"
	}
	params := make(map[string]string)
	params["Input"] = p.Input
	params["Mix"] = p.Mix
	return vc.SendFunction(cut, params)
}

// UpdateTally タリーを更新、点灯する必要がある場合trueが帰る
func (p ProgramPI) UpdateTally() (bool, error) {
	if p.Host == "" || p.Port == 0 {
		return false, nil // HostかPortがゼロ値の場合何もしない
	}
	vc, err := vmixhttp.NewClient(p.Host, p.Port)
	if err != nil {
		return false, err
	}
	return vc.Inputs.Input[vc.Active].Key == p.Input, nil
}

func (p *ProgramPI) UpdateInputs() error {
	if p.Host == "" || p.Port == 0 {
		return nil // HostかPortがゼロ値の場合何もしない
	}
	vc, err := vmixhttp.NewClient(p.Host, p.Port)
	if err != nil {
		return err
	}
	// スライスをリセットして更新
	p.Inputs = make([]input, 0, len(vc.Inputs.Input))
	for _, i := range vc.Inputs.Input {
		p.Inputs = append(p.Inputs, input{
			Name:   i.Name,
			Key:    i.Key,
			Number: int(i.Number),
		})
	}
	return nil
}
