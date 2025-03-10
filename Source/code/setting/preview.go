package setting

type PreviewSetting struct {
	VMixAddress string `json:"dest"`
	Input       int    `json:"input"`
	Mix         *int   `json:"mix"`
	Tally       bool   `json:"tally"`
	ContextID   string `json:"-"`
}
