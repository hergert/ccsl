package ratelimit

import (
	"fmt"
	"strings"
	"time"

	"github.com/hergert/ccsl/internal/palette"
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

	l := Limits{
		FiveHour: parseWindow(rl, "five_hour", "⁵ʰ"),
		SevenDay: parseWindow(rl, "seven_day", "⁷ᵈ"),
	}

	if l.FiveHour == nil && l.SevenDay == nil {
		return Limits{}, false
	}
	return l, true
}

func parseWindow(rl map[string]any, key, label string) *Window {
	m, ok := rl[key].(map[string]any)
	if !ok {
		return nil
	}
	pct, ok := m["used_percentage"].(float64)
	if !ok {
		return nil
	}
	w := &Window{UsedPct: pct, Label: label}
	if resetAt, ok := m["resets_at"].(float64); ok {
		w.ResetsAt = time.Unix(int64(resetAt), 0)
	}
	return w
}

func (w *Window) severity() string {
	switch {
	case w.UsedPct >= 90:
		return "red"
	case w.UsedPct >= 70:
		return "yellow"
	default:
		return ""
	}
}

func (w *Window) format(ansi bool) string {
	s := fmt.Sprintf("%.0f%%", w.UsedPct) + w.Label
	if remaining := time.Until(w.ResetsAt); remaining > 0 && w.UsedPct >= 70 {
		m := int(remaining.Minutes())
		switch {
		case m >= 1440 && (m%1440)/60 == 0:
			s += fmt.Sprintf("↻%dd", m/1440)
		case m >= 1440:
			s += fmt.Sprintf("↻%dd%dh", m/1440, (m%1440)/60)
		case m >= 60:
			s += fmt.Sprintf("↻%dh%dm", m/60, m%60)
		default:
			s += fmt.Sprintf("↻%dm", m)
		}
	}
	if ansi {
		switch w.severity() {
		case "red":
			s = palette.Red + s + palette.Reset
		case "yellow":
			s = palette.Yellow + s + palette.Reset
		}
	}
	return s
}

func (l Limits) Render(ansi bool) types.Segment {
	var parts []string
	if l.FiveHour != nil {
		parts = append(parts, l.FiveHour.format(ansi))
	}
	if l.SevenDay != nil {
		parts = append(parts, l.SevenDay.format(ansi))
	}

	return types.Segment{
		Text:     strings.Join(parts, " "),
		Style:    "dim",
		Priority: 30,
	}
}
