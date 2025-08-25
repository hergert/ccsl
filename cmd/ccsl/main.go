package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"ccsl/internal/config"
	"ccsl/internal/palette"
	"ccsl/internal/render"
	"ccsl/internal/runner"
)

func main() {
	// Subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "setup":
			runSetup()
			return
		case "doctor":
			runDoctor(os.Args[2:])
			return
		}
	}

	// Default mode: read Claude JSON on stdin and print one line
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(1)
	}
	var ctxObj map[string]any
	if err := json.Unmarshal(raw, &ctxObj); err != nil {
		// Print nothing on parse errors - graceful degradation
		os.Exit(0)
	}

	cfg := config.Load()
	totalBudget := time.Duration(cfg.Limits.TotalBudgetMS) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), totalBudget)
	defer cancel()

	segs := runner.Collect(ctx, raw, cfg)
	line := render.Line(cfg.UI.Template, segs, palette.From(cfg, ctxObj), cfg.UI.Truncate)
	fmt.Println(line)
}

func runDoctor(args []string) {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	jsonPath := fs.String("json", "", "Path to a Claude JSON fixture to test with")
	noAnsi := fs.Bool("no-ansi", false, "Render without ANSI styling")
	fs.Parse(args)

	var raw []byte
	if *jsonPath != "" {
		b, err := os.ReadFile(*jsonPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "doctor: cannot read %s: %v\n", *jsonPath, err)
			os.Exit(2)
		}
		raw = b
	} else {
		raw = defaultFixture()
	}

	// Parse for palette & info
	var ctxObj map[string]any
	_ = json.Unmarshal(raw, &ctxObj)

	cfg := config.Load()
	if *noAnsi {
		cfg.Theme.ANSI = false
	}

	// Time-bounded run like the real statusline
	totalBudget := time.Duration(cfg.Limits.TotalBudgetMS) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), totalBudget)
	defer cancel()

	segs, diags := runner.CollectWithDiag(ctx, raw, cfg)
	line := render.Line(cfg.UI.Template, segs, palette.From(cfg, ctxObj), cfg.UI.Truncate)

	// Print report
	fmt.Println("ccsl doctor")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("Template : %q\n", cfg.UI.Template)
	fmt.Printf("Truncate : %d\n", cfg.UI.Truncate)
	fmt.Printf("Theme    : mode=%s ansi=%v icons=%v\n", cfg.Theme.Mode, cfg.Theme.ANSI, cfg.Theme.Icons)
	fmt.Printf("Order    : %s\n", strings.Join(cfg.Plugins.Order, ", "))
	fmt.Printf("Limits   : per_plugin=%dms total=%dms\n", cfg.Limits.PerPluginTimeoutMS, cfg.Limits.TotalBudgetMS)

	// Show plugin configs in the same order
	fmt.Println("\nPlugins:")
	type row struct {
		id, kind     string
		timeout, ttl int
		onlyIf       string
	}
	var rows []row
	for _, id := range cfg.Plugins.Order {
		if pcfg, ok := cfg.Plugin[id]; ok {
			rows = append(rows, row{id: id, kind: pcfg.Type, timeout: effTimeout(pcfg, cfg), ttl: pcfg.CacheTTLMS, onlyIf: pcfg.OnlyIf})
		}
	}
	w := func(s string, n int) string {
		if len(s) >= n {
			return s[:n]
		}
		return s + strings.Repeat(" ", n-len(s))
	}
	fmt.Println("  ID            TYPE     TIMEOUT  TTL   ONLY_IF")
	for _, r := range rows {
		fmt.Printf("  %s  %s  %7d  %4d  %s\n",
			w(r.id, 12), w(r.kind, 7), r.timeout, r.ttl, r.onlyIf)
	}

	// Timings
	fmt.Println("\nPlugin timings:")
	fmt.Println("  ID            TYPE     TIME(ms)  TIMEOUT  STATUS")
	// diags are already sorted by order
	for _, d := range diags {
		status := "ok"
		if d.Skipped {
			status = "skipped"
		}
		if d.CacheHit {
			status = "cache"
		}
		if d.Error != "" {
			status = "error: " + truncate(d.Error, 40)
		}
		fmt.Printf("  %-12s %-7s %8d  %7d  %s\n",
			d.ID, d.Kind, d.DurationMS, d.TimeoutMS, status)
	}

	// Deterministic output: show segments we got (IDs)
	var segIDs []string
	for _, s := range segs {
		segIDs = append(segIDs, s.ID)
	}
	sort.Strings(segIDs)
	fmt.Printf("\nSegments returned: %s\n", strings.Join(segIDs, ", "))

	fmt.Println("\nRendered line:")
	fmt.Println("  " + line)
}

func defaultFixture() []byte {
	cwd, _ := os.Getwd()
	return []byte(fmt.Sprintf(`{
  "model": {"display_name": "Doctor Test"},
  "workspace": {"current_dir": %q, "project_dir": %q}
}`, cwd, cwd))
}

func effTimeout(pcfg config.PluginConfig, cfg *config.Config) int {
	if pcfg.TimeoutMS > 0 {
		return pcfg.TimeoutMS
	}
	return cfg.Limits.PerPluginTimeoutMS
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
