package lines

import (
	"fmt"

	"github.com/hergert/ccsl/internal/types"
)

type Changes struct {
	Added   int
	Removed int
}

func Parse(raw map[string]any) (Changes, bool) {
	data, ok := raw["cost"].(map[string]any)
	if !ok {
		return Changes{}, false
	}

	added, _ := data["total_lines_added"].(float64)
	removed, _ := data["total_lines_removed"].(float64)
	if added == 0 && removed == 0 {
		return Changes{}, false
	}

	return Changes{Added: int(added), Removed: int(removed)}, true
}

func (c Changes) Render() types.Segment {
	return types.Segment{
		Text:     fmt.Sprintf("+%d-%d", c.Added, c.Removed),
		Style:    "dim",
		Priority: 20,
	}
}
