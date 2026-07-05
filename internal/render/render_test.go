package render

import (
	"strings"
	"testing"

	"github.com/hergert/ccsl/internal/config"
	"github.com/hergert/ccsl/internal/palette"
	"github.com/hergert/ccsl/internal/types"
)

func plainPalette() *palette.Palette {
	return palette.From(&config.Config{Theme: config.ThemeConfig{ANSI: false}})
}

func ansiPalette() *palette.Palette {
	return palette.From(&config.Config{Theme: config.ThemeConfig{ANSI: true}})
}

func TestPrefixOnlyWithContent(t *testing.T) {
	segs := []types.Segment{{ID: "a", Text: "A", Priority: 50}}
	got := Line("{a?prefix= }{b?prefix=:}", segs, plainPalette(), 0)
	if got != " A" {
		t.Errorf("Line = %q, want %q", got, " A")
	}
}

func TestMissingSegmentLeavesNoResidue(t *testing.T) {
	got := Line("{a?prefix= }end", nil, plainPalette(), 0)
	if got != "end" {
		t.Errorf("Line = %q, want %q", got, "end")
	}
}

func TestTruncateTrimsLowestPriority(t *testing.T) {
	segs := []types.Segment{
		{ID: "low", Text: "LLLLLLLLLL", Priority: 10},
		{ID: "high", Text: "HHHHHHHHHH", Priority: 90},
	}
	got := Line("{low} {high}", segs, plainPalette(), 15)
	want := "L... HHHHHHHHHH"
	if got != want {
		t.Errorf("Line = %q, want %q", got, want)
	}
}

func TestTruncateCountsVisibleRunesNotANSI(t *testing.T) {
	segs := []types.Segment{
		{ID: "low", Text: "LLLLLLLLLL", Style: "dim", Priority: 10},
		{ID: "high", Text: "HHHHHHHHHH", Style: "bold", Priority: 90},
	}
	got := Line("{low} {high}", segs, ansiPalette(), 15)
	stripped := strings.NewReplacer(palette.Dim, "", palette.Bold, "", palette.Reset, "").Replace(got)
	if len(stripped) > 15 {
		t.Errorf("visible length %d > 15 in %q", len(stripped), got)
	}
	if !strings.Contains(got, palette.Bold) {
		t.Errorf("expected ANSI preserved in %q", got)
	}
}

func TestEffectiveMaxLen(t *testing.T) {
	cases := []struct {
		name       string
		configured int
		columns    string
		want       int
	}{
		{"no columns keeps config", 120, "", 120},
		{"columns narrower wins", 120, "80", 80},
		{"columns wider keeps config", 120, "200", 120},
		{"no config uses columns", 0, "80", 80},
		{"garbage columns keeps config", 120, "wide", 120},
		{"zero columns keeps config", 120, "0", 120},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := EffectiveMaxLen(tc.configured, tc.columns); got != tc.want {
				t.Errorf("EffectiveMaxLen(%d, %q) = %d, want %d", tc.configured, tc.columns, got, tc.want)
			}
		})
	}
}

func TestNoTruncationWhenDisabled(t *testing.T) {
	segs := []types.Segment{{ID: "a", Text: strings.Repeat("x", 300), Priority: 50}}
	got := Line("{a}", segs, plainPalette(), 0)
	if len(got) != 300 {
		t.Errorf("len = %d, want 300 untouched", len(got))
	}
}
