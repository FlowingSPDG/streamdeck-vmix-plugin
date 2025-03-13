package setting

// DestinationToInputs maps destination addresses to their inputs
type DestinationToInputs = map[string][]*Input

// SendInputsPayload represents the payload for sending inputs to the property inspector
type SendInputsPayload struct {
	Event  string              `json:"event"`
	Inputs DestinationToInputs `json:"inputs"`
}

// Destinations represents the payload for sending destinations to the property inspector
type Destinations struct {
	Event        string   `json:"event"`
	Destinations []string `json:"destinations"`
}
