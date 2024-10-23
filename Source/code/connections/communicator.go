package connections

import (
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
)

type vMixCommunicator struct {
	dest       string
	contexts   []string
	connection vmixtcp.Vmix
	inputs     []vMixInput
}

type vMixInput struct {
	Number int
	Name   string
	Key    string
}

func (vc *vMixCommunicator) GetRaw() vmixtcp.Vmix {
	return vc.connection
}

func (vc *vMixCommunicator) Contexts() []string {
	return vc.contexts
}

func (vc *vMixCommunicator) SetInputs(inputs []vMixInput) {
	vc.inputs = inputs
}
