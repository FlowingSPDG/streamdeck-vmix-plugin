package setting

type PreviewSetting struct {
	VMixAddress string    `json:"dest"`
	Input       int       `json:"input"`
	Mix         int       `json:"mix"`
	TallyMode   TallyMode `json:"tally_mode"`
}

func (p *PreviewSetting) IsDefault() bool {
	return (p.VMixAddress == "" &&
		p.Input == 0 &&
		p.Mix == 0 &&
		p.TallyMode == 0)
}

func (p *PreviewSetting) Initialize() {
	p.VMixAddress = "localhost"
	p.Input = 1
	p.Mix = 0
	p.TallyMode = TallyModeACTS
}

func (p *PreviewSetting) GetVMixAddress() string {
	return p.VMixAddress
}

func (p *PreviewSetting) GetTallyMode() TallyMode {
	return p.TallyMode
}

func (p *PreviewSetting) GetMix() int {
	return p.Mix
}

func (p *PreviewSetting) GetInput() *int {
	return &p.Input
}
