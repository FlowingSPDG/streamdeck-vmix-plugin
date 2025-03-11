package action

type tallyStatus int

const (
	tallyStatusUnknown tallyStatus = iota
	tallyStatusOff
	tallyStatusOn
)
