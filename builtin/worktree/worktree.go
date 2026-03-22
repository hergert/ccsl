package worktree

import (
	"context"

	"github.com/hergert/ccsl/internal/types"
)

// Render extracts worktree info from Claude Code's stdin JSON
// Only present when running in a --worktree session
func Render(_ context.Context, ctxObj map[string]any) types.Segment {
	wt, ok := ctxObj["worktree"].(map[string]any)
	if !ok {
		return types.Segment{}
	}

	name, _ := wt["name"].(string)
	if name == "" {
		name, _ = wt["branch"].(string)
	}
	if name == "" {
		return types.Segment{}
	}

	return types.Segment{
		Text:     name,
		Style:    "bold",
		Priority: 83,
	}
}
