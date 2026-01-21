package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"ccsl/internal/types"
)

// Render provides git status using a single command
func Render(ctx context.Context, ctxObj map[string]any) types.Segment {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain=v2", "--branch")
	out, err := cmd.Output()
	if err != nil {
		return types.Segment{}
	}

	var branch string
	var ahead, behind int
	dirty := false

	for _, line := range strings.Split(string(out), "\n") {
		switch {
		case strings.HasPrefix(line, "# branch.head "):
			branch = strings.TrimPrefix(line, "# branch.head ")
		case strings.HasPrefix(line, "# branch.ab "):
			// format: # branch.ab +N -M
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.HasPrefix(p, "+") {
					fmt.Sscanf(p, "+%d", &ahead)
				} else if strings.HasPrefix(p, "-") {
					fmt.Sscanf(p, "-%d", &behind)
				}
			}
		case len(line) > 0 && line[0] != '#':
			dirty = true
		}
	}

	if branch == "" || branch == "(detached)" {
		// Try to get short hash for detached HEAD
		cmd := exec.CommandContext(ctx, "git", "rev-parse", "--short", "HEAD")
		if out, err := cmd.Output(); err == nil {
			branch = strings.TrimSpace(string(out))
		}
	}

	if branch == "" {
		return types.Segment{}
	}

	text := branch
	if dirty {
		text += "*"
	}

	if ahead > 0 {
		text += fmt.Sprintf("↑%d", ahead)
	}
	if behind > 0 {
		text += fmt.Sprintf("↓%d", behind)
	}

	return types.Segment{
		Text:     text,
		Style:    "dim",
		Priority: 60,
	}
}
