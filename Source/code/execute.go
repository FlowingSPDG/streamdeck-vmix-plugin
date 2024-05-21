package stdvmix

import (
	"context"
	"fmt"
)

func (s *StdVmix) ExecuteSend(ctx context.Context, pi SendFunctionPI) error {
	v, err := s.vMixClients.loadOrStore(ctx, pi.Dest)
	if err != nil {
		return err
	}

	return v.Function(pi.Name, pi.Queries.ToString())
}

func (s *StdVmix) ExecutePreview(ctx context.Context, pi PreviewPI) error {
	v, err := s.vMixClients.loadOrStore(ctx, pi.Dest)
	if err != nil {
		return err
	}
	query := fmt.Sprintf("Input=%s", pi.Input)
	if pi.Mix != 1 {
		query = fmt.Sprintf("%s&Mix=%d", query, pi.Mix)
	}
	return v.Function("PreviewInput", query)
}

func (s *StdVmix) ExecuteProgram(ctx context.Context, pi ProgramPI) error {
	v, err := s.vMixClients.loadOrStore(ctx, pi.Dest)
	if err != nil {
		return err
	}
	query := fmt.Sprintf("Input=%s", pi.Input)
	if pi.Mix != 1 {
		query = fmt.Sprintf("%s&Mix=%d", query, pi.Mix)
	}
	return v.Function(pi.Transition, query)
}
