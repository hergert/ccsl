package types

// Segment represents a rendered piece of the statusline
type Segment struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Style    string `json:"style"`    // "normal" | "bold" | "dim"
	Priority int    `json:"priority"` // for truncation, default 50
}

// Context keys
type CtxKey string

const CtxKeyConfig CtxKey = "ccsl_cfg"
