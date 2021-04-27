package main

type tally int

const (
	// Inactive tally status inactive(GREY)
	Inactive tally = iota
	// Preview tally status Preview(GREEN)
	Preview
	// Program tally status Program(RED)
	Program
)

type input struct {
	Name         string `xml:",chardata"`
	Key          string `xml:"key,attr"`
	Number       int    `xml:"number,attr"`
	TallyPreview bool   `xml:"-"`
	TallyProgram bool   `xml:"-"`
}
