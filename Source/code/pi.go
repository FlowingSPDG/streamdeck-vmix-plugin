package stdvmix

import (
	"fmt"
	"reflect"

	vmixhttp "github.com/FlowingSPDG/vmix-go/http"
)

type GlobalSettings struct {
	Inputs map[string][]Input `json:"inputs"` // key: host:port value: Inputs
}

// SendFunctionPI Settings for each button to save persistantly on action instance
type SendFunctionPI struct {
	Host    string  `json:"host"`
	Port    int     `json:"port,string"`
	Input   string  `json:"input"`
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
	p.Queries = []Query{}
}

// PreviewPI Property Inspector info for Preview
type PreviewPI struct {
	Host  string `json:"host"`
	Port  int    `json:"port,string"`
	Input string `json:"input"`
	Mix   string `json:"mix"`
	Tally bool   `json:"tally"`
}

func (p PreviewPI) IsDefault() bool {
	return reflect.ValueOf(p).IsZero()
}

func (p *PreviewPI) Initialize() {
	p.Host = "localhost"
	p.Port = 8088
	p.Input = "0"
	p.Mix = ""
	p.Tally = false
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

// ProgramPI Property Inspector info for PGM(Cut)
type ProgramPI struct {
	Host      string `json:"host"`
	Port      int    `json:"port,string"`
	Input     string `json:"input"`
	Mix       string `json:"mix"`
	CutDirect bool   `json:"cut_direct"`
	Tally     bool   `json:"tally"`
}

func (p ProgramPI) IsDefault() bool {
	return reflect.ValueOf(p).IsZero()
}

func (p *ProgramPI) Initialize() {
	p.Host = "localhost"
	p.Port = 8088
	p.Input = "0"
	p.Mix = ""
	p.CutDirect = false
	p.Tally = false
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
	// 以下コードだとpanicが起きそう
	// vc.Inputs.Input[vc.Preview-1].Key == p.Input, nil
	for _, input := range vc.Inputs.Input {
		// 一致するinputがあればtrueを返す
		if input.Key != p.Input {
			continue
		}
		return input.Number == vc.Active, nil
	}
	return false, fmt.Errorf("No input found")
}
