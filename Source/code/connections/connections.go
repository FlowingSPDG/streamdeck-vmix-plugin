package connections

import (
	"context"
	"errors"
	"iter"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
)

type VMixCommunicators struct {
	logger       logger.Logger
	comms        []*vMixCommunicator
	actsSender   chan<- vMixCommunicatorActsSenderPayload
	tallySender  chan<- vMixCommunicatorTallySenderPayload
	healthSender chan<- vMixCommunicatorHealthSenderPayload
	inputsSender chan<- vMixInputsSenderPayload
}

type VMixChannelSender struct {
	logger       logger.Logger
	ActsSender   <-chan vMixCommunicatorActsSenderPayload
	TallySender  <-chan vMixCommunicatorTallySenderPayload
	HealthSender <-chan vMixCommunicatorHealthSenderPayload
	InputsSender <-chan vMixInputsSenderPayload
}

func NewvMixCommunicators(logger logger.Logger) (*VMixCommunicators, *VMixChannelSender) {
	actsSender := make(chan vMixCommunicatorActsSenderPayload)
	tallySender := make(chan vMixCommunicatorTallySenderPayload)
	healthSender := make(chan vMixCommunicatorHealthSenderPayload)
	inputsSender := make(chan vMixInputsSenderPayload)

	return &VMixCommunicators{
			logger:       logger,
			comms:        []*vMixCommunicator{},
			actsSender:   actsSender,
			tallySender:  tallySender,
			healthSender: healthSender,
			inputsSender: inputsSender,
		}, &VMixChannelSender{
			logger:       logger,
			ActsSender:   actsSender,
			TallySender:  tallySender,
			HealthSender: healthSender,
			InputsSender: inputsSender,
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

type vMixCommunicatorHealthSenderPayload struct {
	Destination string
	Version     string
}

type vMixInputsSenderPayload struct {
	Destination string
	Inputs      []vMixInput
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
	vcs.logger.Log("Adding vMix for destination:%s, context:%s", dest, contextStr)
	// すでに所有している場合、再追加だけする
	if v, exist := vcs.FindByDestination(dest); exist {
		vcs.logger.Log("vMix for destination already found. skip!")

		// append
		v.contexts = append(v.contexts, contextStr)
		vcs.comms = append(vcs.comms, v)

		return nil
	}

	// Initiate
	vc := &vMixCommunicator{
		dest:       dest,
		contexts:   []string{contextStr},
		connection: vmixtcp.New(dest),
	}

	vc.connection.OnVersion(func(resp *vmixtcp.VersionResponse) {
		vcs.logger.Log("vMix for destination received VERSION:%v", resp)

		vcs.healthSender <- vMixCommunicatorHealthSenderPayload{
			Destination: dest,
			Version:     resp.Version,
		}
	})

	vc.connection.OnActs(func(resp *vmixtcp.ActsResponse) {
		vcs.logger.Log("vMix for destination received ACTS:%v", resp)

		s := strings.Split(resp.Response, " ")
		vcs.actsSender <- vMixCommunicatorActsSenderPayload{
			Destination: dest,
			Acts:        s,
		}
	})

	vc.connection.OnTally(func(resp *vmixtcp.TallyResponse) {
		vcs.logger.Log("vMix for destination received TALLY:%v", resp)

		vcs.tallySender <- vMixCommunicatorTallySenderPayload{
			Destination: dest,
			Tally:       resp.Tally,
		}
	})

	vc.connection.OnXML(func(resp *vmixtcp.XMLResponse) {
		vcs.logger.Log("vMix for destination received XML:%v", resp)
		
		inputs := make([]vMixInput, 0, len(resp.XML.Inputs.Input))
		for num, input := range resp.XML.Inputs.Input {
			inputs = append(inputs, vMixInput{
				Number: num,
				Name:   input.Title,
				Key:    input.Key,
			})
		}
		vcs.inputsSender <- vMixInputsSenderPayload{
			Destination: dest,
			Inputs:      inputs,
		}
	})

	// TODO: slice lock/mutex
	vcs.comms = append(vcs.comms, vc)

	return nil
}

func (vcs *VMixCommunicators) RemovevMixByContext(ctx context.Context, ctxStr string) error {
	vcs.logger.Log("Removing vMix for context:%s", ctxStr)
	vc, found := vcs.FindByContext(ctxStr)
	if !found {
		return errors.New("not found")
	}

	vc.contexts = slices.DeleteFunc(vc.contexts, func(s string) bool {
		return s == ctxStr
	})

	if len(vc.contexts) == 0 {
		vcs.logger.Log("Destination %s has 0 contexts.", vc.dest)
		vcs.comms = slices.DeleteFunc(vcs.comms, func(v *vMixCommunicator) bool {
			return v.dest == vc.dest
		})
	}

	return nil
}

func (vcs *VMixCommunicators) Iter() iter.Seq[*vMixCommunicator] {
	return func(yield func(*vMixCommunicator) bool) {
		for _, vc := range vcs.comms {
			if !yield(vc) {
				return
			}
		}
	}
}

func (vcs *VMixCommunicators) RunConnection(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// throttling
			if len(vcs.comms) == 0 {
				time.Sleep(time.Second)
				continue
			}
			wg := sync.WaitGroup{}
			for vc := range vcs.Iter() {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := vc.connection.Connect(ctx, time.Second); err != nil {
						return
					}
					if err := vc.connection.Run(ctx); err != nil {
						return
					}
				}()
			}
			wg.Wait()
		}
	}
}
