package git

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"ccsl/internal/types"
)

// Render provides git status information with branch, dirty state, and upstream tracking
func Render(ctx context.Context, ctxObj map[string]any) types.Segment {
	// Quick check if we're in a git repo
	if !isGitRepo(ctx) {
		return types.Segment{}
	}

	var parts []string
	
	// Get branch name
	branch := getBranch(ctx)
	if branch == "" {
		return types.Segment{}
	}
	parts = append(parts, branch)
	
	// Check for dirty state
	if isDirty(ctx) {
		parts[len(parts)-1] += "*"
	}
	
	// Get upstream tracking info
	upstream := getUpstreamInfo(ctx)
	if upstream != "" {
		parts = append(parts, upstream)
	}
	
	text := strings.Join(parts, " ")
	
	return types.Segment{
		Text:     text,
		Style:    "dim",
		Priority: 60,
	}
}

func isGitRepo(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree")
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) == "true"
}

func getBranch(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	
	// Try symbolic ref first
	cmd := exec.CommandContext(ctx, "git", "symbolic-ref", "--short", "HEAD")
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output))
	}
	
	// Fallback to short hash for detached HEAD
	cmd = exec.CommandContext(ctx, "git", "rev-parse", "--short", "HEAD")
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output))
	}
	
	return ""
}

func isDirty(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "git", "diff", "--quiet", "--ignore-submodules", "--")
	return cmd.Run() != nil // non-zero exit means dirty
}

func getUpstreamInfo(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, 80*time.Millisecond)
	defer cancel()
	
	// Check if upstream exists
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "@{u}")
	if _, err := cmd.Output(); err != nil {
		return "" // no upstream
	}
	
	// Get ahead/behind counts
	cmd = exec.CommandContext(ctx, "git", "rev-list", "--left-right", "--count", "@{u}...HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	counts := strings.Fields(strings.TrimSpace(string(output)))
	if len(counts) != 2 {
		return ""
	}
	
	var parts []string
	if counts[1] != "0" { // ahead
		parts = append(parts, "↑"+counts[1])
	}
	if counts[0] != "0" { // behind
		parts = append(parts, "↓"+counts[0])
	}
	
	return strings.Join(parts, " ")
}