package setting

type PreviewSetting struct {
	Host  string `json:"dest"`
	Input int    `json:"input"`
	Mix   *int   `json:"mix"`
	Tally bool   `json:"tally"`
}
