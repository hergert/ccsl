package ctx

import (
	"fmt"

	"github.com/hergert/ccsl/internal/types"
)

type ContextWindow struct {
	UsedPct       float64
	WindowSize    float64
	Exceeds200k   bool
}

func Parse(raw map[string]any) (ContextWindow, bool) {
	cw, ok := raw["context_window"].(map[string]any)
	if !ok {
		return ContextWindow{}, false
	}

	c := ContextWindow{}
	c.WindowSize, _ = cw["context_window_size"].(float64)
	c.Exceeds200k, _ = raw["exceeds_200k_tokens"].(bool)

	if p, ok := cw["used_percentage"].(float64); ok && p > 0 {
		c.UsedPct = p
	} else {
		input, _ := cw["total_input_tokens"].(float64)
		output, _ := cw["total_output_tokens"].(float64)
		if c.WindowSize > 0 {
			c.UsedPct = ((input + output) / c.WindowSize) * 100
		}
	}

	if c.UsedPct == 0 {
		return ContextWindow{}, false
	}
	return c, true
}

func (c ContextWindow) IsLarge() bool {
	return c.WindowSize >= 500000
}

// Anthropic's MRCR v2 benchmark: Opus drops from 93% (256K) to 76% (1M).
// User reports (github.com/anthropics/claude-code/issues/35296, #34685):
//   20-40% of 1M: degradation starts, wrong approaches
//   40-60%: fabrications, confident false conclusions
//   60%+:   prior facts inaccessible, irrecoverable loops
// Effective useful context is ~200-400K tokens regardless of window size.
// 1M thresholds are therefore more aggressive, not less.
func (c ContextWindow) severity() string {
	if c.IsLarge() {
		// 30% of 1M = 300K tokens (degradation onset)
		// 60% of 1M = 600K tokens (reliability lost)
		switch {
		case c.UsedPct >= 60:
			return "red"
		case c.UsedPct >= 30:
			return "yellow"
		}
	} else {
		switch {
		case c.UsedPct >= 85 || c.Exceeds200k:
			return "red"
		case c.UsedPct >= 50:
			return "yellow"
		}
	}
	return "dim"
}

func (c ContextWindow) Render() types.Segment {
	style := c.severity()
	text := fmt.Sprintf("%.0f%%", c.UsedPct)
	if style == "red" {
		text += "!"
	}

	return types.Segment{
		Text:     text,
		Style:    style,
		Priority: 45,
	}
}
