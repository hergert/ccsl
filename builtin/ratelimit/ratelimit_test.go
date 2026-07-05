package ratelimit

import (
	"strings"
	"testing"
	"time"

	"github.com/hergert/ccsl/internal/palette"
)

func TestParseAbsent(t *testing.T) {
	if _, ok := Parse(map[string]any{}); ok {
		t.Error("expected ok=false without rate_limits")
	}
	if _, ok := Parse(map[string]any{"rate_limits": map[string]any{}}); ok {
		t.Error("expected ok=false with empty rate_limits")
	}
}

func TestParseWindows(t *testing.T) {
	l, ok := Parse(map[string]any{
		"rate_limits": map[string]any{
			"five_hour": map[string]any{"used_percentage": 12.0, "resets_at": 1738425600.0},
			"seven_day": map[string]any{"used_percentage": 31.0, "resets_at": 1738857600.0},
		},
	})
	if !ok {
		t.Fatal("Parse returned ok=false")
	}
	if l.FiveHour == nil || l.FiveHour.UsedPct != 12 || l.FiveHour.Label != "⁵ʰ" {
		t.Errorf("FiveHour = %+v, want 12%% ⁵ʰ", l.FiveHour)
	}
	if l.FiveHour.ResetsAt.Unix() != 1738425600 {
		t.Errorf("FiveHour.ResetsAt = %v, want epoch 1738425600", l.FiveHour.ResetsAt)
	}
	if l.SevenDay == nil || l.SevenDay.UsedPct != 31 || l.SevenDay.Label != "⁷ᵈ" {
		t.Errorf("SevenDay = %+v, want 31%% ⁷ᵈ", l.SevenDay)
	}
	if l.SevenDay != nil && l.SevenDay.ResetsAt.Unix() != 1738857600 {
		t.Errorf("SevenDay.ResetsAt = %v, want epoch 1738857600", l.SevenDay.ResetsAt)
	}
}

func TestRenderJoinsWindows(t *testing.T) {
	l := Limits{
		FiveHour: &Window{UsedPct: 12, Label: "⁵ʰ"},
		SevenDay: &Window{UsedPct: 31, Label: "⁷ᵈ"},
	}
	seg := l.Render(false)
	if seg.Text != "12%⁵ʰ 31%⁷ᵈ" {
		t.Errorf("Text = %q, want %q", seg.Text, "12%⁵ʰ 31%⁷ᵈ")
	}
	if seg.Style != "dim" || seg.Priority != 30 {
		t.Errorf("Style/Priority = %q/%d, want dim/30", seg.Style, seg.Priority)
	}
}

func TestCountdownGate(t *testing.T) {
	cases := []struct {
		name string
		pct  float64
		in   time.Duration
		want string
	}{
		{"below 70 no countdown", 69, 100 * time.Minute, "69%⁵ʰ"},
		{"70+ hours and minutes", 72, 107*time.Minute + 30*time.Second, "72%⁵ʰ↻1h47m"},
		{"70+ minutes only", 72, 45*time.Minute + 30*time.Second, "72%⁵ʰ↻45m"},
		{"past reset no countdown", 90, -5 * time.Minute, "90%⁵ʰ"},
		{"70+ days and hours", 82, 51*time.Hour + 30*time.Minute, "82%⁵ʰ↻2d3h"},
		{"70+ whole days", 82, 48*time.Hour + 30*time.Second, "82%⁵ʰ↻2d"},
		{"70+ one day one hour", 82, 25*time.Hour + 30*time.Second, "82%⁵ʰ↻1d1h"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := Limits{FiveHour: &Window{UsedPct: tc.pct, Label: "⁵ʰ", ResetsAt: time.Now().Add(tc.in)}}
			if got := l.Render(false).Text; got != tc.want {
				t.Errorf("Text = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestANSISeverity(t *testing.T) {
	cases := []struct {
		pct  float64
		code string
	}{
		{95, palette.Red},
		{72, palette.Yellow},
	}
	for _, tc := range cases {
		l := Limits{FiveHour: &Window{UsedPct: tc.pct, Label: "⁵ʰ"}}
		got := l.Render(true).Text
		if !strings.Contains(got, tc.code) {
			t.Errorf("at %v%% expected color %q in %q", tc.pct, tc.code, got)
		}
	}
	l := Limits{FiveHour: &Window{UsedPct: 30, Label: "⁵ʰ"}}
	if got := l.Render(true).Text; strings.Contains(got, "\x1b[") {
		t.Errorf("at 30%% expected no color, got %q", got)
	}
}
