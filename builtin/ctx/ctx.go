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

	exceeds200k, _ := ctxObj["exceeds_200k_tokens"].(bool)

	text := fmt.Sprintf("%.0f%%", pct)
	style := "dim"

	switch {
	case pct >= 90 || exceeds200k:
		text += "!"
		style = "red"
	case pct >= 75:
		style = "red"
	case pct >= 50:
		style = "yellow"
	}

	return types.Segment{
		Text:     text,
		Style:    style,
		Priority: 45,
	}
}
