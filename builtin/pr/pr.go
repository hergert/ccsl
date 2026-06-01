package pr

import (
	"fmt"
	"strings"

	"github.com/hergert/ccsl/internal/palette"
	"github.com/hergert/ccsl/internal/types"
)

// PR is the open pull request for the current branch, surfaced by Claude Code
// in the status line "pr" field (Claude Code v2.1.145+). Absent until a PR is
// found, and removed once it merges or closes.
type PR struct {
	Number      int
	ReviewState string
}

func Parse(raw map[string]any) (PR, bool) {
	p, ok := raw["pr"].(map[string]any)
	if !ok {
		return PR{}, false
	}

	pr := PR{}
	if n, ok := p["number"].(float64); ok {
		pr.Number = int(n)
	}
	if rs, ok := p["review_state"].(string); ok {
		pr.ReviewState = rs
	}

	if pr.Number == 0 {
		return PR{}, false
	}
	return pr, true
}

func (p PR) Render(ansi bool) types.Segment {
	text := fmt.Sprintf("#%d", p.Number)

	// review_state is optional; match loosely so unknown values degrade silently.
	switch {
	case strings.Contains(p.ReviewState, "approv"):
		text += "✓"
	case strings.Contains(p.ReviewState, "change"):
		ind := "✗"
		if ansi {
			ind = palette.Red + ind + palette.Reset
		}
		text += ind
	}

	return types.Segment{
		Text:     text,
		Style:    "dim",
		Priority: 58, // just below git (60): truncates after the branch
	}
}
