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
