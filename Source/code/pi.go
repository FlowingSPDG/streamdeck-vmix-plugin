package stdvmix

import (
	"reflect"
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
	Port  int    `json:"port"`
	Input string `json:"input"`
	Mix   int    `json:"mix"`
	Tally bool   `json:"tally"`
}

func (p PreviewPI) IsDefault() bool {
	return reflect.ValueOf(p).IsZero()
}

func (p *PreviewPI) Initialize() {
	p.Host = "localhost"
	p.Port = 8088
	p.Input = "0"
	p.Mix = 1
	p.Tally = false
}

// ProgramPI Property Inspector info for PGM(Cut)
type ProgramPI struct {
	Host       string `json:"host"`
	Port       int    `json:"port,string"`
	Input      string `json:"input"`
	Mix        int    `json:"mix"`
	Tally      bool   `json:"tally"`
	Transition string `json:"transition"`
}

func (p ProgramPI) IsDefault() bool {
	return reflect.ValueOf(p).IsZero()
}

func (p *ProgramPI) Initialize() {
	p.Host = "localhost"
	p.Port = 8088
	p.Input = "0"
	p.Mix = 1
	p.Transition = "Cut"
	p.Tally = false
}
