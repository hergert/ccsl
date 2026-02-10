package duration

import (
	"context"
	"fmt"

	"github.com/hergert/ccsl/internal/types"
)

// Render extracts total duration from Claude Code's stdin JSON
func Render(ctx context.Context, ctxObj map[string]any) types.Segment {
	costData, ok := ctxObj["cost"].(map[string]any)
	if !ok {
		return types.Segment{}
	}

	ms, ok := costData["total_duration_ms"].(float64)
	if !ok || ms < 60000 {
		return types.Segment{}
	}

	minutes := int(ms / 60000)
	hours := minutes / 60
	mins := minutes % 60

	var text string
	if hours > 0 {
		text = fmt.Sprintf("%dh%dm", hours, mins)
	} else {
		text = fmt.Sprintf("%dm", mins)
	}

	return types.Segment{
		Text:     text,
		Style:    "dim",
		Priority: 25,
	}
}
