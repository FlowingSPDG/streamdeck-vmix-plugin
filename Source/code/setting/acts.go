package setting

type ActivatorColor int

const (
	_ ActivatorColor = iota
	ActivatorColorRed
	ActivatorColorGreen
)

type ActivatorSetting struct {
	VMixAddress string         `json:"dest"`
	Color       ActivatorColor `json:"color"`

	// Acts
	ActsEvent         string `json:"acts_event"`
	ActInput          string `json:"acts_input"`
	ActsActiveState   string `json:"acts_active_state"`
	ActsInactiveState string `json:"acts_inactive_state"`
}

func (s *ActivatorSetting) IsDefault() bool {
	return s.VMixAddress == "" && s.Color == 0 && s.ActsEvent == "" && s.ActInput == "" && s.ActsActiveState == "" && s.ActsInactiveState == ""
}

func (s *ActivatorSetting) Initialize() {
	s.VMixAddress = "localhost"
	s.Color = 1
	s.ActsEvent = "Input"
	s.ActInput = "1"
	s.ActsActiveState = "1"
	s.ActsInactiveState = "0"
}

func (s *ActivatorSetting) GetVMixAddress() string {
	return s.VMixAddress
}

func (s *ActivatorSetting) GetMix() int {
	return 0
}
