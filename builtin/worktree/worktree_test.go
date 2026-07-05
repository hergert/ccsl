package worktree

import "testing"

func TestParseWorktreeSession(t *testing.T) {
	w, ok := Parse(map[string]any{
		"worktree": map[string]any{"name": "fix-auth", "branch": "wt/fix-auth"},
	})
	if !ok || w.DisplayName() != "fix-auth" {
		t.Errorf("Parse = %+v ok=%v, want fix-auth", w, ok)
	}
}

func TestParseBranchFallback(t *testing.T) {
	w, ok := Parse(map[string]any{
		"worktree": map[string]any{"branch": "wt/fix-auth"},
	})
	if !ok || w.DisplayName() != "wt/fix-auth" {
		t.Errorf("Parse = %+v ok=%v, want branch fallback", w, ok)
	}
}

func TestParseAbsent(t *testing.T) {
	if _, ok := Parse(map[string]any{}); ok {
		t.Error("expected ok=false without worktree data")
	}
}

func TestParseGitWorktreeFallback(t *testing.T) {
	w, ok := Parse(map[string]any{
		"workspace": map[string]any{"git_worktree": "feature-xyz"},
	})
	if !ok || w.DisplayName() != "feature-xyz" {
		t.Errorf("Parse = %+v ok=%v, want feature-xyz from workspace.git_worktree", w, ok)
	}
}

func TestWorktreeSessionWinsOverGitWorktree(t *testing.T) {
	w, ok := Parse(map[string]any{
		"worktree":  map[string]any{"name": "fix-auth"},
		"workspace": map[string]any{"git_worktree": "feature-xyz"},
	})
	if !ok || w.DisplayName() != "fix-auth" {
		t.Errorf("Parse = %+v ok=%v, want worktree.* to win", w, ok)
	}
}
