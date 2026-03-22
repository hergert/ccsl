package ratelimit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hergert/ccsl/internal/types"
)

// Render extracts rate limit usage from Claude Code's stdin JSON
// Only present for Claude.ai Pro/Max subscribers after first API response
func Render(_ context.Context, ctxObj map[string]any) types.Segment {
	rl, ok := ctxObj["rate_limits"].(map[string]any)
	if !ok {
		return types.Segment{}
	}

	var parts []string
	var maxPct float64

	if fh, ok := rl["five_hour"].(map[string]any); ok {
		if pct, ok := fh["used_percentage"].(float64); ok {
			s := fmt.Sprintf("5h:%.0f%%", pct)
			if resetAt, ok := fh["resets_at"].(float64); ok {
				if remaining := time.Until(time.Unix(int64(resetAt), 0)); remaining > 0 && pct >= 70 {
					s += fmt.Sprintf("↻%s", fmtDuration(remaining))
				}
			}
			parts = append(parts, s)
			if pct > maxPct {
				maxPct = pct
			}
		}
	}

	if sd, ok := rl["seven_day"].(map[string]any); ok {
		if pct, ok := sd["used_percentage"].(float64); ok {
			parts = append(parts, fmt.Sprintf("7d:%.0f%%", pct))
			if pct > maxPct {
				maxPct = pct
			}
		}
	}

	if len(parts) == 0 {
		return types.Segment{}
	}

	style := "dim"
	switch {
	case maxPct >= 90:
		style = "red"
	case maxPct >= 70:
		style = "yellow"
	}

	return types.Segment{
		Text:     strings.Join(parts, " "),
		Style:    style,
		Priority: 30,
	}
}

func fmtDuration(d time.Duration) string {
	m := int(d.Minutes())
	if m >= 60 {
		return fmt.Sprintf("%dh%dm", m/60, m%60)
	}
	return fmt.Sprintf("%dm", m)
}
