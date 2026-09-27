package setting

type FunctionSetting struct {
	VMixAddress string `json:"dest"`
	Function    string `json:"function"`
	Query       string `json:"query"`
	Input       *int   `json:"input"`

	// Acts
	ActsEvent         string         `json:"acts_event"`
	ActInput          string         `json:"acts_input"`
	ActsActiveState   string         `json:"acts_active_state"`
	ActsInactiveState string         `json:"acts_inactive_state"`
	ActsColor         ActivatorColor `json:"acts_color"`
}

func (s *FunctionSetting) IsDefault() bool {
	return s.VMixAddress == "" && s.Function == "" && s.ActsEvent == "" && s.ActInput == "" && s.ActsActiveState == "" && s.ActsInactiveState == "" && s.ActsColor == 0
}

func (s *FunctionSetting) Initialize() {
	s.VMixAddress = "localhost"
	s.Function = "PreviewInput"
	s.Query = "Input=1"
	s.ActsEvent = "InputPreview"
	s.ActInput = "1"
	s.ActsActiveState = "1"
	s.ActsInactiveState = "0"
	s.ActsColor = ActivatorColorRed
}

func (s *FunctionSetting) GetVMixAddress() string {
	return s.VMixAddress
}

func (s *FunctionSetting) GetActsEvent() string {
	return s.ActsEvent
}

func (s *FunctionSetting) GetMix() int {
	return 0
}

func (s *FunctionSetting) GetInput() *int {
	return nil
}
