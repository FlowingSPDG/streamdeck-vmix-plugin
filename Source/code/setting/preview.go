package setting

type TallyMode int

const (
	_ TallyMode = iota
	TallyModeTALLY
	TallyModeACTS
	TallyModeDisabled
)

type PreviewSetting struct {
	VMixAddress string    `json:"dest"`
	Input       int       `json:"input"`
	Mix         *int      `json:"mix"`
	TallyMode   TallyMode `json:"tally_mode"`
}

func (p *PreviewSetting) IsDefault() bool {
	return (p.VMixAddress == "" ||
		p.Input == 0 ||
		(p.Mix != nil && *p.Mix == 0) ||
		p.TallyMode == TallyModeDisabled)
}

func (p *PreviewSetting) Initialize() {
	p.VMixAddress = "localhost"
	p.Input = 1
	p.Mix = nil
	p.TallyMode = TallyModeTALLY
}
