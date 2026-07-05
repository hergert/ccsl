package effort

import "github.com/hergert/ccsl/internal/types"

type Effort struct {
	Level string
}

// Present only when the model supports the effort parameter.
func Parse(raw map[string]any) (Effort, bool) {
	data, ok := raw["effort"].(map[string]any)
	if !ok {
		return Effort{}, false
	}
	level, _ := data["level"].(string)
	if level == "" {
		return Effort{}, false
	}
	return Effort{Level: level}, true
}

func (e Effort) Render() types.Segment {
	return types.Segment{
		Text:     e.Level,
		Style:    "dim",
		Priority: 32,
	}
}
