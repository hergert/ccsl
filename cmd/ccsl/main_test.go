package main

import (
	"context"
	"testing"
	"time"

	"ccsl/internal/config"
	"ccsl/internal/palette"
	"ccsl/internal/render"
	"ccsl/internal/runner"
)

func TestBasicFunctionality(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())

	// Create a minimal test configuration
	cfg := config.Load()

	// Sample Claude JSON input
	claudeJSON := []byte(`{
		"model": {
			"id": "claude-test",
			"display_name": "Test Model"
		},
		"workspace": {
			"current_dir": "/tmp/test",
			"project_dir": "/tmp/test"
		}
	}`)

	// Test runner.Collect
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	segments := runner.Collect(ctx, claudeJSON, cfg)

	// Should have at least model and cwd segments
	if len(segments) < 2 {
		t.Errorf("Expected at least 2 segments, got %d", len(segments))
	}

	// Test render.Line
	pal := palette.From(cfg, map[string]any{})
	line := render.Line(cfg.UI.Template, segments, pal, cfg.UI.Truncate)

	if line == "" {
		t.Error("Expected non-empty line output")
	}

	t.Logf("Generated line: %s", line)
}

func TestConfigLoading(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())

	cfg := config.Load()

	if cfg == nil {
		t.Fatal("Config should not be nil")
	}

	if len(cfg.Plugins.Order) == 0 {
		t.Error("Expected default plugin order")
	}

	if cfg.UI.Template == "" {
		t.Error("Expected default template")
	}
}

func TestSegmentGeneration(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())

	testCases := []struct {
		name     string
		input    string
		expected []string // segment IDs we expect
	}{
		{
			name: "basic model and workspace",
			input: `{
				"model": {"display_name": "Test"},
				"workspace": {"current_dir": "/tmp"}
			}`,
			expected: []string{"model", "cwd"},
		},
	}

	cfg := config.Load()
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			segments := runner.Collect(ctx, []byte(tc.input), cfg)

			segmentIDs := make(map[string]bool)
			for _, seg := range segments {
				segmentIDs[seg.ID] = true
			}

			for _, expected := range tc.expected {
				if !segmentIDs[expected] {
					t.Errorf("Expected segment %s not found", expected)
				}
			}
		})
	}
}
