package cost

import (
	"context"
	"fmt"

	"ccsl/internal/types"
)

// Render extracts cost from Claude Code's stdin JSON
func Render(ctx context.Context, ctxObj map[string]any) types.Segment {
	costData, ok := ctxObj["cost"].(map[string]any)
	if !ok {
		return types.Segment{}
	}

	totalCost, ok := costData["total_cost_usd"].(float64)
	if !ok || totalCost == 0 {
		return types.Segment{}
	}

	return types.Segment{
		Text:     fmt.Sprintf("$%.2f", totalCost),
		Style:    "dim",
		Priority: 40,
	}
}
