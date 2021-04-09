package main

import (
	"fmt"
	"net/url"

	vmixgo "github.com/FlowingSPDG/vmix-go"
)

var (
	global   = &GlobalSetting{}
	settings = make(map[string]*PropertyInspector) // settings[event.context]
)

func init() {
	vmixURL, _ := url.Parse("http://localhost:8088/api")
	global = &GlobalSetting{
		VMixAPIURL: vmixURL,
		Inputs:     []vmixgo.Input{},
	}
}

// GlobalSetting Global setting for action instance
type GlobalSetting struct {
	VMixAPIURL *url.URL
	Inputs     []vmixgo.Input
}

// PropertyInspector Settings for each button to save persistantly on action instance
type PropertyInspector struct {
	FunctionInput string `json:"functionInput"`
	FunctionName  string `json:"functionName"`
	Queries       []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"queries"`
}

// GenerateURL Generate function API URL.
func (p PropertyInspector) GenerateURL() (string, error) {
	if p.FunctionName == "" {
		return "", fmt.Errorf("Empty Function Name")
	}
	vm, _ := url.Parse("http://localhost:8088/api")
	vm.Query().Add("Function", p.FunctionName)
	if p.FunctionInput != "" {
		vm.Query().Add("Input", p.FunctionInput)
	}
	for _, v := range p.Queries {
		vm.Query().Add(v.Key, v.Value)
	}
	return vm.String(), nil
}
