package cost

import (
	"fmt"
	"strings"
	"time"

	"github.com/hergert/ccsl/internal/types"
)

type Session struct {
	CostUSD  float64
	Duration time.Duration
}

func Parse(raw map[string]any) (Session, bool) {
	data, ok := raw["cost"].(map[string]any)
	if !ok {
		return Session{}, false
	}

	s := Session{}
	s.CostUSD, _ = data["total_cost_usd"].(float64)
	if ms, ok := data["total_duration_ms"].(float64); ok && ms > 0 {
		s.Duration = time.Duration(ms) * time.Millisecond
	}

	if s.CostUSD == 0 && s.Duration == 0 {
		return Session{}, false
	}
	return s, true
}

func (s Session) Render() types.Segment {
	var text string
	if s.CostUSD > 0 {
		text = fmt.Sprintf("$%.2f", s.CostUSD)
	}
	if s.Duration > 0 {
		text += s.formatDuration()
	}

	return types.Segment{
		Text:     text,
		Style:    "dim",
		Priority: 40,
	}
}

func (s Session) formatDuration() string {
	total := int(s.Duration.Minutes())
	hours := total / 60
	mins := total % 60

	switch {
	case hours > 0:
		return toSuper(hours) + "ʰ" + toSuper(mins) + "ᵐ"
	case mins > 0:
		return toSuper(mins) + "ᵐ"
	default:
		return toSuper(int(s.Duration.Seconds())) + "ˢ"
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
