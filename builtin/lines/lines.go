package lines

import (
	"context"
	"fmt"

	"github.com/hergert/ccsl/internal/types"
)

// Render extracts lines added/removed from Claude Code's stdin JSON
func Render(_ context.Context, ctxObj map[string]any) types.Segment {
	costData, ok := ctxObj["cost"].(map[string]any)
	if !ok {
		return types.Segment{}
	}

	added, _ := costData["total_lines_added"].(float64)
	removed, _ := costData["total_lines_removed"].(float64)
	if added == 0 && removed == 0 {
		return types.Segment{}
	}

	return types.Segment{
		Text:     fmt.Sprintf("+%.0f-%.0f", added, removed),
		Style:    "dim",
		Priority: 20,
	}
}
