package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hergert/ccsl/internal/config"
	"github.com/hergert/ccsl/internal/palette"
	"github.com/hergert/ccsl/internal/types"
)

// Render provides git status using a single command
func Render(ctx context.Context, ctxObj map[string]any) types.Segment {
	// Check if untracked files should be shown (default: no for speed)
	args := []string{"status", "--porcelain=v2", "--branch", "--untracked-files=no"}
	if cfg, ok := ctx.Value(types.CtxKeyConfig).(*config.Config); ok {
		if pcfg, exists := cfg.Plugin["git"]; exists && pcfg.Untracked {
			args = []string{"status", "--porcelain=v2", "--branch"}
		}
	}

	cmd := exec.CommandContext(ctx, "git", args...)
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

	// Check for stash (fast file stat)
	hasStash := false
	if gitDir := findGitDir(ctx); gitDir != "" {
		if _, err := os.Stat(filepath.Join(gitDir, "refs", "stash")); err == nil {
			hasStash = true
		}
	}

	text := branch
	if dirty {
		text += "*"
	}

	// Sync indicators with meaningful colors
	if ahead > 0 {
		text += palette.Yellow + fmt.Sprintf("⇡%d", ahead) + palette.Reset
	}
	if behind > 0 {
		text += palette.Red + fmt.Sprintf("⇣%d", behind) + palette.Reset
	}
	if hasStash {
		text += palette.Dim + "≡" + palette.Reset
	}

	return types.Segment{
		Text:     text,
		Style:    "dim",
		Priority: 60,
	}
}

// findGitDir returns the .git directory path
func findGitDir(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
