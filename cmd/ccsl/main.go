package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/hergert/ccsl/internal/config"
	"github.com/hergert/ccsl/internal/palette"
	"github.com/hergert/ccsl/internal/render"
	"github.com/hergert/ccsl/internal/runner"
)

func main() {
	// Simple doctor subcommand for debugging
	if len(os.Args) > 1 && os.Args[1] == "doctor" {
		runDoctor()
		return
	}

	// Read Claude JSON from stdin
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(1)
	}

	var ctxObj map[string]any
	if json.Unmarshal(raw, &ctxObj) != nil {
		os.Exit(0) // graceful degradation
	}

	// Extract project_dir for project-local config
	var projectDir string
	if ws, ok := ctxObj["workspace"].(map[string]any); ok {
		if dir, ok := ws["project_dir"].(string); ok {
			projectDir = dir
		}
	}

	cfg := config.Load(projectDir)
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(cfg.Limits.TotalBudgetMS)*time.Millisecond)
	defer cancel()

	segs := runner.Collect(ctx, ctxObj, raw, cfg)
	line := render.Line(cfg.UI.Template, segs, palette.From(cfg, ctxObj), cfg.UI.Truncate)
	fmt.Println(line)
}

func runDoctor() {
	cwd, _ := os.Getwd()
	raw := []byte(fmt.Sprintf(`{
  "model": {"display_name": "Opus"},
  "agent": {"name": "task"},
  "workspace": {"current_dir": %q},
  "context_window": {"used_percentage": 91},
  "exceeds_200k_tokens": true,
  "cost": {"total_cost_usd": 1.23, "total_duration_ms": 498000}
}`, cwd))

	var ctxObj map[string]any
	json.Unmarshal(raw, &ctxObj)

	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(cfg.Limits.TotalBudgetMS)*time.Millisecond)
	defer cancel()

	start := time.Now()
	segs := runner.Collect(ctx, ctxObj, raw, cfg)
	elapsed := time.Since(start)

	line := render.Line(cfg.UI.Template, segs, palette.From(cfg, ctxObj), cfg.UI.Truncate)

	fmt.Printf("template: %s\n", cfg.UI.Template)
	fmt.Printf("segments: %v\n", runner.ParseSegments(cfg.UI.Template))
	fmt.Printf("elapsed:  %dms\n", elapsed.Milliseconds())
	fmt.Printf("output:   %s\n", line)
}
