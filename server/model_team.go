package server

// Team is a collection of players clicking things
type Team struct {
	ID     string `json:"id,omitempty"`
	Clicks int64  `json:"clicks,omitempty"`
}
