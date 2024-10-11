package connections

import (
	"context"
	"errors"
	"slices"
	"strings"

	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"

	"golang.org/x/xerrors"
)

type VMixCommunicators struct {
	comms       []*vMixCommunicator
	actsSender  chan<- vMixCommunicatorActsSenderPayload
	tallySender chan<- vMixCommunicatorTallySenderPayload
}

type VMixChannelSender struct {
	ActsSender  <-chan vMixCommunicatorActsSenderPayload
	TallySender <-chan vMixCommunicatorTallySenderPayload
}

func NewvMixCommunicators() (*VMixCommunicators, *VMixChannelSender) {
	actsSender := make(chan vMixCommunicatorActsSenderPayload)
	tallySender := make(chan vMixCommunicatorTallySenderPayload)

	return &VMixCommunicators{
			comms:       []*vMixCommunicator{},
			actsSender:  actsSender,
			tallySender: tallySender,
		}, &VMixChannelSender{
			ActsSender:  actsSender,
			TallySender: tallySender,
		}
}

type vMixCommunicatorActsSenderPayload struct {
	Destination string
	Acts        []string
}

type vMixCommunicatorTallySenderPayload struct {
	Destination string
	Tally       []vmixtcp.TallyStatus
}

func (vcs *VMixCommunicators) FindByContext(ctxStr string) (*vMixCommunicator, bool) {
	for _, comm := range vcs.comms {
		for _, c := range comm.contexts {
			if c == ctxStr {
				return comm, true
			}
		}
	}
	return nil, false
}

func (vcs *VMixCommunicators) FindByDestination(dest string) (*vMixCommunicator, bool) {
	for _, comm := range vcs.comms {
		if comm.dest == dest {
			return comm, true
		}
	}
	return nil, false
}

func (vcs *VMixCommunicators) AddvMix(ctx context.Context, dest string, contextStr string) error {
	// TODO: 再接続処理

	// すでに所有している場合何もしない
	if v, exist := vcs.FindByDestination(dest); exist {
		// append
		v.contexts = append(v.contexts, contextStr)
		vcs.comms = append(vcs.comms, v)

		// TODO: 再接続処理
		return nil
	}

	// Initiate
	vc := &vMixCommunicator{
		dest:       dest,
		contexts:   []string{contextStr},
		connection: vmixtcp.New(dest),
	}

	vc.connection.OnVersion(func(resp *vmixtcp.VersionResponse) {
		if err := vc.connection.Subscribe(vmixtcp.EventActs, ""); err != nil {
			// log?
		}
	})

	vc.connection.OnActs(func(resp *vmixtcp.ActsResponse) {
		s := strings.Split(resp.Response, " ")
		vcs.actsSender <- vMixCommunicatorActsSenderPayload{
			Destination: dest,
			Acts:        s,
		}
	})
	vc.connection.OnTally(func(resp *vmixtcp.TallyResponse) {
		vcs.tallySender <- vMixCommunicatorTallySenderPayload{
			Destination: dest,
			Tally:       resp.Tally,
		}
	})

	if err := vc.connection.Connect(); err != nil {
		return xerrors.Errorf("Failed to connect vMix TCP API : %w", err)
	}
	go vc.connection.Run(ctx)

	// TODO: slice lock/mutex
	vcs.comms = append(vcs.comms, vc)

	return nil
}

func (vcs *VMixCommunicators) RemovevMixByContext(ctx context.Context, ctxStr string) error {
	vc, found := vcs.FindByContext(ctxStr)
	if !found {
		return errors.New("not found")
	}

	vc.contexts = slices.DeleteFunc(vc.contexts, func(s string) bool {
		return s == ctxStr
	})

	// TODO: protect map
	if len(vc.contexts) == 0 {
		vcs.comms = slices.DeleteFunc(vcs.comms, func(v *vMixCommunicator) bool {
			return v.dest == vc.dest
		})
	}

	return nil
}
