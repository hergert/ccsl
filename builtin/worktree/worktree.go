package worktree

import "github.com/hergert/ccsl/internal/types"

type Worktree struct {
	Name   string
	Branch string
}

// Only present when running in a --worktree session.
func Parse(raw map[string]any) (Worktree, bool) {
	data, ok := raw["worktree"].(map[string]any)
	if !ok {
		return Worktree{}, false
	}

	w := Worktree{}
	w.Name, _ = data["name"].(string)
	w.Branch, _ = data["branch"].(string)

	if w.Name == "" && w.Branch == "" {
		return Worktree{}, false
	}
	return w, true
}

func (w Worktree) DisplayName() string {
	if w.Name != "" {
		return w.Name
	}
	return w.Branch
}

func (w Worktree) Render() types.Segment {
	return types.Segment{
		Text:     w.DisplayName(),
		Style:    "bold",
		Priority: 83,
	}
}
