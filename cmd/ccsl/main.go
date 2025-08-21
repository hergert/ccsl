package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"ccsl/internal/config"
	"ccsl/internal/runner"
	"ccsl/internal/render"
	"ccsl/internal/palette"
)

func main() {
	// 1) Read Claude JSON from stdin
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(1)
	}

	var ctxObj map[string]any
	if err := json.Unmarshal(raw, &ctxObj); err != nil {
		// Print nothing on parse errors - graceful degradation
		os.Exit(0)
	}

	// 2) Load config (env + default fallback)
	cfg := config.Load()

	// 3) Build a time-bounded context for the whole render
	totalBudget := time.Duration(cfg.Limits.TotalBudgetMS) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), totalBudget)
	defer cancel()

	// 4) Collect segments: built-ins + exec plugins (parallel)
	segs := runner.Collect(ctx, raw, cfg)

	// 5) Render with template + theme
	line := render.Line(cfg.UI.Template, segs, palette.From(cfg, ctxObj), cfg.UI.Truncate)

	// 6) Output
	fmt.Println(line)
}