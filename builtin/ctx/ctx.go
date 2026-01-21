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

	// Try used_percentage first
	if p, ok := cw["used_percentage"].(float64); ok && p > 0 {
		pct = p
	} else {
		// Fallback: calculate from tokens
		input, _ := cw["total_input_tokens"].(float64)
		output, _ := cw["total_output_tokens"].(float64)
		size, _ := cw["context_window_size"].(float64)
		if size > 0 {
			pct = ((input + output) / size) * 100
		}
	}

	if pct == 0 {
		return types.Segment{}
	}

	return types.Segment{
		Text:     fmt.Sprintf("%.0f%%", pct),
		Style:    styleForPercent(pct),
		Priority: 45,
	}
}

// styleForPercent returns color based on context usage thresholds
func styleForPercent(pct float64) string {
	switch {
	case pct >= 75:
		return "red"
	case pct >= 50:
		return "yellow"
	default:
		return "dim"
	}
}
