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

type Status struct {
	Branch   string
	Dirty    bool
	Ahead    int
	Behind   int
	HasStash bool
}

func (s Status) Render(ansi bool) types.Segment {
	text := s.Branch
	if s.Dirty {
		text += "*"
	}
	if s.Ahead > 0 {
		ind := fmt.Sprintf("⇡%d", s.Ahead)
		if ansi {
			ind = palette.Yellow + ind + palette.Reset
		}
		text += ind
	}
	if s.Behind > 0 {
		ind := fmt.Sprintf("⇣%d", s.Behind)
		if ansi {
			ind = palette.Red + ind + palette.Reset
		}
		text += ind
	}
	if s.HasStash {
		ind := "≡"
		if ansi {
			ind = palette.Dim + ind + palette.Reset
		}
		text += ind
	}

	return types.Segment{
		Text:     text,
		Style:    "dim",
		Priority: 60,
	}
}

func Collect(ctx context.Context, cfg *config.Config) (Status, bool) {
	args := []string{"status", "--porcelain=v2", "--branch", "--untracked-files=no"}
	if pcfg, exists := cfg.Plugin["git"]; exists && pcfg.Untracked {
		args = []string{"status", "--porcelain=v2", "--branch"}
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := cmd.Output()
	if err != nil {
		return Status{}, false
	}

	s := Status{}
	for _, line := range strings.Split(string(out), "\n") {
		switch {
		case strings.HasPrefix(line, "# branch.head "):
			s.Branch = strings.TrimPrefix(line, "# branch.head ")
		case strings.HasPrefix(line, "# branch.ab "):
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.HasPrefix(p, "+") {
					fmt.Sscanf(p, "+%d", &s.Ahead)
				} else if strings.HasPrefix(p, "-") {
					fmt.Sscanf(p, "-%d", &s.Behind)
				}
			}
		case len(line) > 0 && line[0] != '#':
			s.Dirty = true
		}
	}

	if s.Branch == "" || s.Branch == "(detached)" {
		cmd := exec.CommandContext(ctx, "git", "rev-parse", "--short", "HEAD")
		if out, err := cmd.Output(); err == nil {
			s.Branch = strings.TrimSpace(string(out))
		}
	}

	if s.Branch == "" {
		return Status{}, false
	}

	if gitDir := findGitDir(ctx); gitDir != "" {
		if _, err := os.Stat(filepath.Join(gitDir, "refs", "stash")); err == nil {
			s.HasStash = true
		}
	}

	return s, true
}

func findGitDir(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
