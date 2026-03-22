package cost

import (
	"context"
	"fmt"
	"strings"

	"github.com/hergert/ccsl/internal/types"
)

// Render extracts cost and duration from Claude Code's stdin JSON.
// Duration is shown as superscript suffix on cost: $8.67²ʰ⁴⁵ᵐ
func Render(ctx context.Context, ctxObj map[string]any) types.Segment {
	costData, ok := ctxObj["cost"].(map[string]any)
	if !ok {
		return types.Segment{}
	}

	totalCost, _ := costData["total_cost_usd"].(float64)
	ms, _ := costData["total_duration_ms"].(float64)

	if totalCost == 0 && ms <= 0 {
		return types.Segment{}
	}

	var text string
	if totalCost > 0 {
		text = fmt.Sprintf("$%.2f", totalCost)
	}
	if ms > 0 {
		text += superDuration(ms)
	}

	return types.Segment{
		Text:     text,
		Style:    "dim",
		Priority: 40,
	}
}

var superDigits = [10]rune{'⁰', '¹', '²', '³', '⁴', '⁵', '⁶', '⁷', '⁸', '⁹'}

func toSuper(n int) string {
	s := fmt.Sprintf("%d", n)
	var b strings.Builder
	for _, c := range s {
		b.WriteRune(superDigits[c-'0'])
	}
	return b.String()
}

func superDuration(ms float64) string {
	secs := int(ms / 1000)
	minutes := secs / 60
	hours := minutes / 60
	mins := minutes % 60

	switch {
	case hours > 0:
		return toSuper(hours) + "ʰ" + toSuper(mins) + "ᵐ"
	case minutes > 0:
		return toSuper(mins) + "ᵐ"
	default:
		return toSuper(secs) + "ˢ"
	}
}
