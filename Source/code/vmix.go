package stdvmix

type input struct {
	Name         string `json:"name" xml:",chardata"`
	Key          string `json:"key" xml:"key,attr"`
	Number       int    `json:"number" xml:"number,attr"`
	TallyPreview bool   `json:"tally_preview" xml:"-"`
	TallyProgram bool   `json:"tally_program" xml:"-"`
}
