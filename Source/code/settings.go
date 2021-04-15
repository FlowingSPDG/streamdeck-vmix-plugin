package main

import (
	"fmt"
	"net/url"
	"sync"
)

// Settings settngs for all buttons/contexts
type Settings struct {
	sync.Mutex `json:"-"`
	inputs     []input                       `json:"-"`
	pi         map[string]*PropertyInspector `json:"-"`
}

var (
	settings = Settings{
		inputs: make([]input, 0, 500),
		pi:     make(map[string]*PropertyInspector),
	}
)

// Save save setting with sd context
func (s *Settings) Save(ctxStr string, pi *PropertyInspector) {
	s.Lock()
	defer s.Unlock()
	s.pi[ctxStr] = pi
	pi.Inputs = s.inputs
}

// Load setting with specified context
func (s *Settings) Load(ctxStr string) (*PropertyInspector, error) {
	s.Lock()
	defer s.Unlock()
	b, ok := s.pi[ctxStr]
	if !ok {
		return nil, fmt.Errorf("Setting not found for this context")
	}
	return b, nil
}

// PropertyInspector Settings for each button to save persistantly on action instance
type PropertyInspector struct {
	FunctionInput string `json:"functionInput,omitempty"`
	FunctionName  string `json:"functionName,omitempty"`
	Queries       []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"queries,omitempty"`
	Inputs []input `json:"inputs"`
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
