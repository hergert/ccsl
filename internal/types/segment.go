package types

// Segment represents a rendered piece of the statusline
type Segment struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	Style      string `json:"style"`    // "normal" | "bold" | "dim" | raw ANSI
	Align      string `json:"align"`    // "left" | "right"
	Priority   int    `json:"priority"` // for truncation, default 50
	CacheTTLMS int    `json:"cache_ttl_ms,omitempty"`
	CacheKey   string `json:"cache_key,omitempty"`
}

// PluginResponse represents the structured output from a plugin
type PluginResponse struct {
	Text       string `json:"text"`
	Style      string `json:"style"`
	Align      string `json:"align"`
	Priority   int    `json:"priority"`
	CacheTTLMS int    `json:"cache_ttl_ms"`
	CacheKey   string `json:"cache_key,omitempty"`
}

// Context keys
type ctxKey string
const CtxKeyConfig ctxKey = "ccsl_cfg"