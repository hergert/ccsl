package agent

import (
	"context"

	"github.com/hergert/ccsl/internal/types"
)

// Render extracts the agent name from Claude Code's stdin JSON
func Render(ctx context.Context, ctxObj map[string]any) types.Segment {
	agentData, ok := ctxObj["agent"].(map[string]any)
	if !ok {
		return types.Segment{}
	}

	name, ok := agentData["name"].(string)
	if !ok || name == "" {
		return types.Segment{}
	}

	return types.Segment{
		Text:     name,
		Style:    "bold",
		Priority: 85,
	}
}
