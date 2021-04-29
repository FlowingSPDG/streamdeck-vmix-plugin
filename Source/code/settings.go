package main

import (
	"fmt"
	"net/url"
	"sync"
)

// Settings settngs for all buttons/contexts
type Settings struct {
	preview int      `json:"-"`
	active  int      `json:"-"`
	Inputs  []input  `json:"-" xml:"inputs"`
	pi      sync.Map `json:"-"`
}

var (
	settings = Settings{
		preview: 0,
		active:  0,
		Inputs:  make([]input, 0, 500),
		pi:      sync.Map{},
	}
)

// Save save setting with sd context
func (s *Settings) Save(ctxStr string, pi *PropertyInspector) {
	pi.Inputs = s.Inputs
	s.pi.Store(ctxStr, pi)
}

// Load setting with specified context
func (s *Settings) Load(ctxStr string) (*PropertyInspector, error) {
	v, ok := s.pi.Load(ctxStr)
	if !ok {
		return nil, fmt.Errorf("Setting not found for this context")
	}

	return (v).(*PropertyInspector), nil
}

// PropertyInspector Settings for each button to save persistantly on action instance
type PropertyInspector struct {
	FunctionInput string `json:"functionInput"`
	FunctionName  string `json:"functionName"`
	Queries       []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"queries"`
	Inputs          []input `json:"inputs"`
	UseTallyPreview bool    `json:"use_tally_preview"`
	UseTallyProgram bool    `json:"use_tally_program"`
}

// GenerateURL Generate function API URL.
func (p PropertyInspector) GenerateURL() (string, error) {
	if p.FunctionName == "" {
		return "", fmt.Errorf("Empty Function Name")
	}
	vm, _ := url.Parse("http://localhost:8088/api")
	q := vm.Query()
	q.Set("Function", p.FunctionName)
	if p.FunctionInput != "" {
		q.Set("Input", p.FunctionInput)
	}
	for _, v := range p.Queries {
		q.Set(v.Key, v.Value)
	}
	vm.RawQuery = q.Encode()
	return vm.String(), nil
}

// GenerateFunction Generate function query.
func (p PropertyInspector) GenerateFunction() (string, error) {
	if p.FunctionName == "" {
		return "", fmt.Errorf("Empty Function Name")
	}
	q := &url.Values{}
	if p.FunctionInput != "" {
		q.Set("Input", p.FunctionInput)
	}
	for _, v := range p.Queries {
		q.Set(v.Key, v.Value)
	}
	return fmt.Sprintf("%s %s", p.FunctionName, q.Encode()), nil
}
