package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hergert/ccsl/internal/config"
	"github.com/hergert/ccsl/internal/palette"
	"github.com/hergert/ccsl/internal/render"
	"github.com/hergert/ccsl/internal/runner"
)

func isolatedConfig(t *testing.T) *config.Config {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	return config.Load()
}

func TestBasicFunctionality(t *testing.T) {
	cfg := isolatedConfig(t)

	claudeJSON := []byte(`{
		"model": {"id": "claude-test", "display_name": "Test Model"},
		"workspace": {"current_dir": "/tmp/test", "project_dir": "/tmp/test"}
	}`)

	var ctxObj map[string]any
	_ = json.Unmarshal(claudeJSON, &ctxObj)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	segments := runner.Collect(ctx, ctxObj, claudeJSON, cfg)
	if len(segments) < 2 {
		t.Errorf("Expected at least 2 segments, got %d", len(segments))
	}

	line := render.Line(cfg.UI.Template, segments, palette.From(cfg), cfg.UI.Truncate)
	if line == "" {
		t.Error("Expected non-empty line output")
	}
	t.Logf("Generated line: %s", line)
}

func TestConfigLoading(t *testing.T) {
	cfg := isolatedConfig(t)

	if cfg == nil {
		t.Fatal("Config should not be nil")
		return
	}
	if cfg.UI.Template == "" && len(cfg.Plugins.Order) == 0 {
		t.Error("Expected either template or explicit plugin order")
	}
	if cfg.UI.Template == "" {
		t.Error("Expected default template")
	}
}

func TestSegmentGeneration(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		template string
		expected []string
	}{
		{
			name: "basic model and workspace",
			input: `{
				"model": {"display_name": "Test"},
				"workspace": {"current_dir": "/tmp"}
			}`,
			expected: []string{"model", "cwd"},
		},
		{
			name: "with rate limits",
			input: `{
				"model": {"display_name": "Test"},
				"workspace": {"current_dir": "/tmp"},
				"rate_limits": {"five_hour": {"used_percentage": 72}, "seven_day": {"used_percentage": 15}}
			}`,
			expected: []string{"model", "cwd", "ratelimit"},
		},
		{
			name: "with worktree",
			input: `{
				"model": {"display_name": "Test"},
				"workspace": {"current_dir": "/tmp"},
				"worktree": {"name": "fix-bug", "branch": "wt/fix-bug"}
			}`,
			expected: []string{"model", "cwd", "worktree"},
		},
		{
			name: "with lines changed",
			input: `{
				"model": {"display_name": "Test"},
				"workspace": {"current_dir": "/tmp"},
				"cost": {"total_lines_added": 50, "total_lines_removed": 10}
			}`,
			template: "{model} {cwd}{lines?prefix= }",
			expected: []string{"model", "cwd", "lines"},
		},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := isolatedConfig(t)
			if tc.template != "" {
				cfg.UI.Template = tc.template
			}
			var ctxObj map[string]any
			_ = json.Unmarshal([]byte(tc.input), &ctxObj)
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
