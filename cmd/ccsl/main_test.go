package main

import (
	"context"
	"encoding/json"
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

	var ctxObj map[string]any
	json.Unmarshal(claudeJSON, &ctxObj)

	// Test runner.Collect
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	segments := runner.Collect(ctx, ctxObj, claudeJSON, cfg)

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

	// Order can be nil (derive from template) or explicit
	// Just verify we have a valid template to derive from
	if cfg.UI.Template == "" && len(cfg.Plugins.Order) == 0 {
		t.Error("Expected either template or explicit plugin order")
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
			var ctxObj map[string]any
			json.Unmarshal([]byte(tc.input), &ctxObj)
			segments := runner.Collect(ctx, ctxObj, []byte(tc.input), cfg)

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
