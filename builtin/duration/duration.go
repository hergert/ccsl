package duration

import (
	"fmt"
	"time"

	"github.com/hergert/ccsl/internal/types"
)

type Duration struct {
	Elapsed time.Duration
}

func Parse(raw map[string]any) (Duration, bool) {
	data, ok := raw["cost"].(map[string]any)
	if !ok {
		return Duration{}, false
	}

	ms, ok := data["total_duration_ms"].(float64)
	if !ok || ms <= 0 {
		return Duration{}, false
	}

	return Duration{Elapsed: time.Duration(ms) * time.Millisecond}, true
}

func (d Duration) Render() types.Segment {
	total := int(d.Elapsed.Minutes())
	hours := total / 60
	mins := total % 60

	var text string
	switch {
	case hours > 0:
		text = fmt.Sprintf("%dh%dm", hours, mins)
	case mins > 0:
		text = fmt.Sprintf("%dm", mins)
	default:
		text = fmt.Sprintf("%ds", int(d.Elapsed.Seconds()))
	}

	return types.Segment{
		Text:     text,
		Style:    "dim",
		Priority: 25,
	}
}
