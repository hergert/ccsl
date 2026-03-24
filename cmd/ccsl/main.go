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
	if len(os.Args) > 1 && os.Args[1] == "doctor" {
		runDoctor()
		return
	}

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(1)
	}

	var ctxObj map[string]any
	if json.Unmarshal(raw, &ctxObj) != nil {
		os.Exit(0)
	}

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
	fmt.Println(render.Line(cfg.UI.Template, segs, palette.From(cfg), cfg.UI.Truncate))
}

type doctorInput struct {
	Model         struct{ DisplayName string `json:"display_name"` } `json:"model"`
	Agent         struct{ Name string `json:"name"` }               `json:"agent"`
	Worktree struct {
		Name   string `json:"name"`
		Branch string `json:"branch"`
	} `json:"worktree"`
	Workspace     struct{ CurrentDir string `json:"current_dir"` }  `json:"workspace"`
	ContextWindow struct {
		UsedPercentage    float64 `json:"used_percentage"`
		ContextWindowSize float64 `json:"context_window_size"`
	} `json:"context_window"`
	Exceeds200kTokens bool `json:"exceeds_200k_tokens"`
	Cost              struct {
		TotalCostUSD      float64 `json:"total_cost_usd"`
		TotalDurationMS   float64 `json:"total_duration_ms"`
		TotalLinesAdded   float64 `json:"total_lines_added"`
		TotalLinesRemoved float64 `json:"total_lines_removed"`
	} `json:"cost"`
	RateLimits struct {
		FiveHour struct {
			UsedPercentage float64 `json:"used_percentage"`
			ResetsAt       int64   `json:"resets_at"`
		} `json:"five_hour"`
		SevenDay struct{ UsedPercentage float64 `json:"used_percentage"` } `json:"seven_day"`
	} `json:"rate_limits"`
}

func runDoctor() {
	cwd, _ := os.Getwd()
	input := doctorInput{}
	input.Model.DisplayName = "Opus 4.6"
	input.Agent.Name = "task"
	input.Worktree.Name = "fix-auth"
	input.Worktree.Branch = "wt/fix-auth"
	input.Workspace.CurrentDir = cwd
	input.ContextWindow.UsedPercentage = 91
	input.ContextWindow.ContextWindowSize = 1000000
	input.Exceeds200kTokens = true
	input.Cost.TotalCostUSD = 1.23
	input.Cost.TotalDurationMS = 498000
	input.Cost.TotalLinesAdded = 156
	input.Cost.TotalLinesRemoved = 23
	input.RateLimits.FiveHour.UsedPercentage = 72
	input.RateLimits.FiveHour.ResetsAt = time.Now().Add(23 * time.Minute).Unix()
	input.RateLimits.SevenDay.UsedPercentage = 15

	raw, _ := json.Marshal(input)
	var ctxObj map[string]any
	_ = json.Unmarshal(raw, &ctxObj)

	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(cfg.Limits.TotalBudgetMS)*time.Millisecond)
	defer cancel()

	start := time.Now()
	segs := runner.Collect(ctx, ctxObj, raw, cfg)
	elapsed := time.Since(start)

	line := render.Line(cfg.UI.Template, segs, palette.From(cfg), cfg.UI.Truncate)

	fmt.Printf("template: %s\n", cfg.UI.Template)
	fmt.Printf("elapsed:  %dms\n", elapsed.Milliseconds())
	fmt.Printf("output:   %s\n", line)
}
