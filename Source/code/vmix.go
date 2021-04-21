package main

import vmixgo "github.com/FlowingSPDG/vmix-go"

type tally int

const (
	// Inactive tally status inactive(GREY)
	Inactive tally = iota
	// Preview tally status Preview(GREEN)
	Preview
	// Program tally status Program(RED)
	Program
)

type input struct {
	vmixgo.Input
	TallyPreview bool
	TallyProgram bool
}

// getvMixInputs get inputs,active key, preview key, error.
func getvMixInputs() ([]input, error) {
	vm, err := vmixgo.NewVmix("http://localhost:8088")
	if err != nil {
		return nil, err
	}
	inputs := make([]input, len(vm.Inputs.Input))
	for k, v := range vm.Inputs.Input {
		inputs[k] = input{
			Input:        v,
			TallyPreview: false,
			TallyProgram: false,
		}
		if vm.Preview == v.Number {
			inputs[k].TallyPreview = true
		}
		if vm.Active == v.Number {
			inputs[k].TallyProgram = true
		}
	}
	return inputs, nil
}
