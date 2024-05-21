package stdvmix

import (
	"net/url"
	"reflect"
)

type GlobalSettings struct {
	Inputs map[string][]Input `json:"inputs"` // key: host:port value: Inputs
}

// SendFunctionPI Settings for each button to save persistantly on action instance
type SendFunctionPI struct {
	Dest    string  `json:"dest"`
	Input   string  `json:"input"`
	Name    string  `json:"name"`
	Queries Queries `json:"queries"`
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

func (p SendFunctionPI) IsDefault() bool {
	return reflect.ValueOf(p).IsZero()
}

func (p *SendFunctionPI) Initialize() {
	p.Dest = "localhost"
	p.Input = "0"
	p.Name = "PreviewInput"
	p.Queries = []Query{}
}

// PreviewPI Property Inspector info for Preview
type PreviewPI struct {
	Dest  string `json:"dest"`
	Input int    `json:"input"`
	Mix   int    `json:"mix"`
	Tally bool   `json:"tally"`
}

func (p PreviewPI) IsDefault() bool {
	return reflect.ValueOf(p).IsZero()
}

func (p *PreviewPI) Initialize() {
	p.Dest = "localhost"
	p.Input = 1
	p.Mix = 1
	p.Tally = false
}

// ProgramPI Property Inspector info for PGM
type ProgramPI struct {
	Dest       string `json:"dest"`
	Input      int    `json:"input"`
	Mix        int    `json:"mix"`
	Tally      bool   `json:"tally"`
	Transition string `json:"transition"`
}

func (p ProgramPI) IsDefault() bool {
	return reflect.ValueOf(p).IsZero()
}

func (p *ProgramPI) Initialize() {
	p.Dest = "localhost"
	p.Input = 1
	p.Mix = 1
	p.Transition = "Cut"
	p.Tally = false
}

type TallyPI struct {
	Dest    string `json:"dest"`
	Input   int    `json:"input"`
	Mix     int    `json:"mix"`
	Preview bool   `json:"preview"`
	Program bool   `json:"program"`
}

func (p TallyPI) IsDefault() bool {
	return reflect.ValueOf(p).IsZero()
}

func (p *TallyPI) Initialize() {
	p.Dest = "localhost"
	p.Input = 1
	p.Mix = 1
	p.Preview = false
	p.Program = false
}
