package ratelimit

import (
	"fmt"
	"strings"
	"time"

	"github.com/hergert/ccsl/internal/types"
)

type Window struct {
	UsedPct  float64
	ResetsAt time.Time
	Label    string // superscript label: "⁵ʰ" or "⁷ᵈ"
}

type Limits struct {
	FiveHour *Window
	SevenDay *Window
}

// Only present for Claude.ai Pro/Max subscribers after first API response.
func Parse(raw map[string]any) (Limits, bool) {
	rl, ok := raw["rate_limits"].(map[string]any)
	if !ok {
		return Limits{}, false
	}

	l := Limits{}
	if fh, ok := rl["five_hour"].(map[string]any); ok {
		if pct, ok := fh["used_percentage"].(float64); ok {
			w := &Window{UsedPct: pct, Label: "⁵ʰ"}
			if resetAt, ok := fh["resets_at"].(float64); ok {
				w.ResetsAt = time.Unix(int64(resetAt), 0)
			}
			l.FiveHour = w
		}
	}
	if sd, ok := rl["seven_day"].(map[string]any); ok {
		if pct, ok := sd["used_percentage"].(float64); ok {
			l.SevenDay = &Window{UsedPct: pct, Label: "⁷ᵈ"}
		}
	}

	if l.FiveHour == nil && l.SevenDay == nil {
		return Limits{}, false
	}
	return l, true
}

func (w *Window) format() string {
	s := fmt.Sprintf("%.0f%%", w.UsedPct) + w.Label
	if remaining := time.Until(w.ResetsAt); remaining > 0 && w.UsedPct >= 70 {
		m := int(remaining.Minutes())
		if m >= 60 {
			s += fmt.Sprintf("↻%dh%dm", m/60, m%60)
		} else {
			s += fmt.Sprintf("↻%dm", m)
		}
	}
	return s
}

func (l Limits) maxPct() float64 {
	var max float64
	if l.FiveHour != nil && l.FiveHour.UsedPct > max {
		max = l.FiveHour.UsedPct
	}
	if l.SevenDay != nil && l.SevenDay.UsedPct > max {
		max = l.SevenDay.UsedPct
	}
	return max
}

func (l Limits) Render() types.Segment {
	var parts []string
	if l.FiveHour != nil {
		parts = append(parts, l.FiveHour.format())
	}
	if l.SevenDay != nil {
		parts = append(parts, l.SevenDay.format())
	}

	style := "dim"
	switch {
	case l.maxPct() >= 90:
		style = "red"
	case l.maxPct() >= 70:
		style = "yellow"
	}

	return types.Segment{
		Text:     strings.Join(parts, " "),
		Style:    style,
		Priority: 30,
	}
}
