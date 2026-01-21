package ctx

import (
	"context"
	"fmt"

	"ccsl/internal/types"
)

// Render extracts context window usage from Claude Code's stdin JSON
func Render(_ context.Context, ctxObj map[string]any) types.Segment {
	cw, ok := ctxObj["context_window"].(map[string]any)
	if !ok {
		return types.Segment{}
	}

	// Try used_percentage first
	if pct, ok := cw["used_percentage"].(float64); ok && pct > 0 {
		return types.Segment{
			Text:     fmt.Sprintf("%.0f%%", pct),
			Style:    "dim",
			Priority: 45,
		}
	}

	// Fallback: calculate from tokens
	input, _ := cw["total_input_tokens"].(float64)
	output, _ := cw["total_output_tokens"].(float64)
	size, _ := cw["context_window_size"].(float64)

	if size > 0 {
		pct := ((input + output) / size) * 100
		return types.Segment{
			Text:     fmt.Sprintf("%.0f%%", pct),
			Style:    "dim",
			Priority: 45,
		}
	}

	return types.Segment{}
}
