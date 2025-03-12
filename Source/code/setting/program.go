package setting

type ProgramSetting struct {
	VMixAddress string    `json:"dest"`
	Input       int       `json:"input"`
	Mix         int       `json:"mix"`
	TallyMode   TallyMode `json:"tally_mode"`
	Transition  string    `json:"transition"`
	Duration    int       `json:"duration"`
}

func (p *ProgramSetting) IsDefault() bool {
	return (p.VMixAddress == "" &&
		p.Input == 0 &&
		p.Mix == 0 &&
		p.TallyMode == 0 &&
		p.Transition == "" &&
		p.Duration == 0)
}

func (p *ProgramSetting) Initialize() {
	p.VMixAddress = "localhost"
	p.Input = 1
	p.Mix = 0
	p.TallyMode = TallyModeTALLY
	p.Transition = "Fade"
	p.Duration = 1000
}
