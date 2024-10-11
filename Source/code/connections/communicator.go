package connections

import (
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
)

type vMixCommunicator struct {
	dest       string
	contexts   []string
	connection vmixtcp.Vmix
}

func (vc *vMixCommunicator) GetRaw() vmixtcp.Vmix {
	return vc.connection
}

func (vc *vMixCommunicator) Contexts() []string {
	return vc.contexts
}
