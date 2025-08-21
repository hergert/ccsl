package agent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"ccsl/internal/types"
)

// Render reads the active agent from .claude/state.json
func Render(ctx context.Context, ctxObj map[string]any) types.Segment {
	agent := "main" // default
	
	// Try to get project directory from Claude context
	var projectDir string
	if workspace, ok := ctxObj["workspace"].(map[string]any); ok {
		if dir, ok := workspace["project_dir"].(string); ok {
			projectDir = dir
		} else if dir, ok := workspace["current_dir"].(string); ok {
			projectDir = dir
		}
	}
	
	// Fallback to current working directory
	if projectDir == "" {
		if dir, err := os.Getwd(); err == nil {
			projectDir = dir
		}
	}
	
	if projectDir == "" {
		return types.Segment{
			Text:     "⚙ " + agent,
			Priority: 70,
		}
	}

	// Read from .claude/state.json
	stateFile := filepath.Join(projectDir, ".claude", "state.json")
	if data, err := os.ReadFile(stateFile); err == nil {
		var state map[string]any
		if err := json.Unmarshal(data, &state); err == nil {
			if activeAgent, ok := state["active_agent"].(string); ok && activeAgent != "" {
				agent = activeAgent
			}
		}
	}

	return types.Segment{
		Text:     "⚙ " + agent,
		Priority: 70,
	}
}