package ctx

import "testing"

func parse(t *testing.T, raw map[string]any) ContextWindow {
	t.Helper()
	c, ok := Parse(raw)
	if !ok {
		t.Fatalf("Parse returned ok=false for %v", raw)
	}
	return c
}

func TestParseUsesProvidedPercentage(t *testing.T) {
	c := parse(t, map[string]any{
		"context_window": map[string]any{"used_percentage": 42.0, "context_window_size": 200000.0},
	})
	if c.UsedPct != 42 {
		t.Errorf("UsedPct = %v, want 42", c.UsedPct)
	}
}

func TestParseFallsBackToTokenCounts(t *testing.T) {
	c := parse(t, map[string]any{
		"context_window": map[string]any{
			"total_input_tokens":  50000.0,
			"total_output_tokens": 10000.0,
			"context_window_size": 200000.0,
		},
	})
	if c.UsedPct != 30 {
		t.Errorf("UsedPct = %v, want 30 from (50k+10k)/200k", c.UsedPct)
	}
}

func TestParseRejectsMissingOrEmpty(t *testing.T) {
	if _, ok := Parse(map[string]any{}); ok {
		t.Error("expected ok=false without context_window")
	}
	if _, ok := Parse(map[string]any{
		"context_window": map[string]any{"used_percentage": 0.0, "context_window_size": 200000.0},
	}); ok {
		t.Error("expected ok=false at 0% usage")
	}
}

func TestSeverityThresholds(t *testing.T) {
	cases := []struct {
		name      string
		pct       float64
		size      float64
		exceeds   bool
		wantStyle string
		wantText  string
	}{
		{"standard below warn", 49, 200000, false, "dim", "49%"},
		{"standard warn", 50, 200000, false, "yellow", "50%"},
		{"standard high warn", 84, 200000, false, "yellow", "84%"},
		{"standard red", 85, 200000, false, "red", "85%!"},
		{"exceeds 200k forces red", 10, 200000, true, "red", "10%!"},
		{"large below warn", 29, 1000000, false, "dim", "29%"},
		{"large warn", 30, 1000000, false, "yellow", "30%"},
		{"large still warn", 59, 1000000, false, "yellow", "59%"},
		{"large red", 60, 1000000, false, "red", "60%!"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := map[string]any{
				"context_window": map[string]any{
					"used_percentage":     tc.pct,
					"context_window_size": tc.size,
				},
				"exceeds_200k_tokens": tc.exceeds,
			}
			seg := parse(t, raw).Render()
			if seg.Style != tc.wantStyle {
				t.Errorf("Style = %q, want %q", seg.Style, tc.wantStyle)
			}
			if seg.Text != tc.wantText {
				t.Errorf("Text = %q, want %q", seg.Text, tc.wantText)
			}
		})
	}
}
