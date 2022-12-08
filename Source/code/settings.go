package main

import (
	"fmt"
	"net/url"
	"sync"
)

// Settings settngs for all buttons/contexts
type Settings struct {
	pi sync.Map `json:"-"`
}

var (
	settings = Settings{
		pi: sync.Map{},
	}
)

// Store store PI
func (s *Settings) Store(ctxStr string, pi *PropertyInspector) {
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

// UpdateInputs update all inputs to PI
func (s *Settings) UpdateInputs() {
	s.pi.Range(func(key, value interface{}) bool {
		v := (value).(*PropertyInspector)
		v.Inputs = inputs
		return true
	})
}

// PropertyInspector Settings for each button to save persistantly on action instance
type PropertyInspector struct {
	Input   string  `json:"input"`
	Inputs  []input `json:"inputs"`
	Name    string  `json:"name"`
	Queries []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"queries"`
	UseTallyPreview bool `json:"use_tally_preview"`
	UseTallyProgram bool `json:"use_tally_program"`
}

// GenerateURL Generate function API URL.
func (p PropertyInspector) GenerateURL() (string, error) {
	if p.Name == "" {
		return "", fmt.Errorf("Empty Function Name")
	}
	vm, _ := url.Parse("http://localhost:8088/api")
	q := vm.Query()
	q.Set("Function", p.Name)
	if p.Input != "" {
		q.Set("Input", p.Input)
	}
	for _, v := range p.Queries {
		q.Set(v.Key, v.Value)
	}
	vm.RawQuery = q.Encode()
	return vm.String(), nil
}

// GenerateFunction Generate function query.
func (p PropertyInspector) GenerateFunction() (string, error) {
	if p.Name == "" {
		return "", fmt.Errorf("Empty Function Name")
	}
	q := &url.Values{}
	if p.Input != "" {
		q.Set("Input", p.Input)
	}
	for _, v := range p.Queries {
		q.Set(v.Key, v.Value)
	}
	return fmt.Sprintf("%s %s", p.Name, q.Encode()), nil
}
