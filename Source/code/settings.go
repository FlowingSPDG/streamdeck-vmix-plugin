package main

import (
	"fmt"
	"net/url"
	"strconv"
)

// SendFunctionPI Settings for each button to save persistantly on action instance
type SendFunctionPI struct {
	Input   string  `json:"input"`
	Inputs  []input `json:"inputs"`
	Name    string  `json:"name"`
	Queries []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"queries"`
}

// GenerateFunction Generate function query.
func (p SendFunctionPI) GenerateFunction() (string, error) {
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

// PreviewPI Property Inspector info for Preview
type PreviewPI struct {
	Input  string  `json:"input"`
	Inputs []input `json:"inputs"`
	Mix    int     `json:"mix,string,omitempty"`
}

// GenerateFunction Generate function query.
func (p PreviewPI) GenerateFunction() (string, error) {
	q := &url.Values{}
	if p.Input != "" {
		q.Set("Input", p.Input)
	}
	if p.Mix != 0 {
		q.Set("Mix", strconv.Itoa(p.Mix))
	}
	return fmt.Sprintf("%s %s", "PreviewInput", q.Encode()), nil
}

// ProgramPI Property Inspector info for PGM(Cut)
type ProgramPI struct {
	Input  string  `json:"input"`
	Inputs []input `json:"inputs"`
	Mix    int     `json:"mix"`
}
