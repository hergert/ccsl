package ctx

import (
	"context"
	"fmt"

	"github.com/hergert/ccsl/internal/types"
)

// Render extracts context window usage from Claude Code's stdin JSON
func Render(_ context.Context, ctxObj map[string]any) types.Segment {
	cw, ok := ctxObj["context_window"].(map[string]any)
	if !ok {
		return types.Segment{}
	}

	var pct float64
	size, _ := cw["context_window_size"].(float64)

	// Try used_percentage first
	if p, ok := cw["used_percentage"].(float64); ok && p > 0 {
		pct = p
	} else {
		// Fallback: calculate from tokens
		input, _ := cw["total_input_tokens"].(float64)
		output, _ := cw["total_output_tokens"].(float64)
		if size > 0 {
			pct = ((input + output) / size) * 100
		}
	}

	if pct == 0 {
		return types.Segment{}
	}

	// Anthropic's MRCR v2 benchmark: Opus drops from 93% (256K) to 76% (1M).
	// User reports (github.com/anthropics/claude-code/issues/35296, #34685):
	//   20-40% of 1M: degradation starts, wrong approaches
	//   40-60%: fabrications, confident false conclusions
	//   60%+:   prior facts inaccessible, irrecoverable loops
	// Effective useful context is ~200-400K tokens regardless of window size.
	// 1M thresholds are therefore more aggressive, not less.
	large := size >= 500000
	exceeds200k, _ := ctxObj["exceeds_200k_tokens"].(bool)

	style := "dim"
	if large {
		// 30% of 1M = 300K tokens (degradation onset)
		// 60% of 1M = 600K tokens (reliability lost)
		switch {
		case pct >= 60:
			style = "red"
		case pct >= 30:
			style = "yellow"
		}
	} else {
		switch {
		case pct >= 85 || exceeds200k:
			style = "red"
		case pct >= 50:
			style = "yellow"
		}
	}

	text := fmt.Sprintf("%.0f%%", pct)
	if large {
		text += "·1M"
	}
	if style == "red" {
		text += "!"
	}

	return types.Segment{
		Text:     text,
		Style:    style,
		Priority: 45,
	}
}
