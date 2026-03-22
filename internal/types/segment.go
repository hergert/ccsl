package types

type Segment struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Style    string `json:"style"`    // "normal" | "bold" | "dim"
	Priority int    `json:"priority"` // for truncation, default 50
}
