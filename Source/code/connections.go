package stdvmix

import (
	"fmt"

	vmixhttp "github.com/FlowingSPDG/vmix-go/http"
	"github.com/puzpuzpuz/xsync/v3"
)

type vMixKey struct {
	host string
	port int
}

type vMixConnections struct {
	connections *xsync.MapOf[vMixKey, *vMix]
}

func newVMixConnections() *vMixConnections {
	return &vMixConnections{
		connections: xsync.NewMapOf[vMixKey, *vMix](),
	}
}

// storeNewVmix stores new vmix client.
func (vc *vMixConnections) storeNewVmix(host string, port int) error {
	key := vMixKey{
		host: host,
		port: port,
	}
	vmix, err := vmixhttp.NewClient(key.host, key.port)
	if err != nil {
		return fmt.Errorf("failed to create vmix client: %w", err)
	}

	// Initialize input slice
	inputs := make([]Input, 0, len(vmix.Inputs.Input))
	for _, i := range vmix.Inputs.Input {
		inputs = append(inputs, Input{
			Name:   i.Name,
			Key:    i.Key,
			Number: int(i.Number),
		})
	}
	vm := &vMix{
		client: vmix,
		inputs: inputs,
	}
	vc.connections.Store(key, vm)
	return nil
}

// Load loads vmix client.
func (vc *vMixConnections) load(host string, port int) (vmix *vMix, ok bool) {
	key := vMixKey{
		host: host,
		port: port,
	}
	vm, ok := vc.connections.Load(key)
	if !ok {
		return nil, false
	}

	// TODO: Add "Preview" and "Active" for slice
	return vm, true
}

func (vc *vMixConnections) loadOrStore(host string, port int) (*vMix, error) {
	vm, ok := vc.load(host, port)
	if !ok {
		if err := vc.storeNewVmix(host, port); err != nil {
			return nil, err
		}
		loaded, _ := vc.load(host, port)
		return loaded, nil
	}
	return vm, nil
}

// UpdateVMixes updates vmix clients.
func (vc *vMixConnections) UpdateVMixes() {
	vc.connections.Range(func(key vMixKey, value *vMix) bool {
		go func() {
			newvMix, err := vmixhttp.NewClient(key.host, key.port)
			if err != nil {
				return
			}
			value.client = newvMix
		}()
		return true
	})
}
