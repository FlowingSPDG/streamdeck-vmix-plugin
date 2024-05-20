package stdvmix

import "strconv"

func (s *StdVmix) ExecuteSend(pi SendFunctionPI) error {
	v, err := s.vMixClients.loadOrStore(pi.Host, pi.Port)
	if err != nil {
		return err
	}
	params := make(map[string]string, len(pi.Queries))
	for _, query := range pi.Queries {
		params[query.Key] = query.Value
	}
	return v.client.SendFunction(pi.Name, params)
}

func (s *StdVmix) ExecutePreview(pi PreviewPI) error {
	v, err := s.vMixClients.loadOrStore(pi.Host, pi.Port)
	if err != nil {
		return err
	}
	params := make(map[string]string, 2)
	params["Input"] = pi.Input
	params["Mix"] = strconv.Itoa(pi.Mix)
	return v.client.SendFunction("PreviewInput", params)
}

func (s *StdVmix) ExecuteProgram(pi ProgramPI) error {
	v, err := s.vMixClients.loadOrStore(pi.Host, pi.Port)
	if err != nil {
		return err
	}
	params := make(map[string]string, 2)
	params["Input"] = pi.Input
	params["Mix"] = strconv.Itoa(pi.Mix)
	return v.client.SendFunction(pi.Transition, params)
}
