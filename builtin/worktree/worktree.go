package worktree

import "github.com/hergert/ccsl/internal/types"

type Worktree struct {
	Name   string
	Branch string
}

// worktree.* only exists in --worktree sessions; workspace.git_worktree is set
// whenever the cwd is inside any linked git worktree.
func Parse(raw map[string]any) (Worktree, bool) {
	w := Worktree{}
	if data, ok := raw["worktree"].(map[string]any); ok {
		w.Name, _ = data["name"].(string)
		w.Branch, _ = data["branch"].(string)
	}
	if w.Name == "" && w.Branch == "" {
		if ws, ok := raw["workspace"].(map[string]any); ok {
			w.Name, _ = ws["git_worktree"].(string)
		}
	}

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
