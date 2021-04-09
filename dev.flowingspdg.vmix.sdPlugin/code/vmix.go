package main

import vmixgo "github.com/FlowingSPDG/vmix-go"

func getvMixInputs() ([]vmixgo.Input, error) {
	v, err := vmixgo.NewVmix("http://localhost:8088/api")
	if err != nil {
		return nil, err
	}
	return v.Inputs.Input, nil
}
