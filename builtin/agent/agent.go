package agent

import "github.com/hergert/ccsl/internal/types"

type Agent struct {
	Name string
}

func Parse(raw map[string]any) (Agent, bool) {
	data, ok := raw["agent"].(map[string]any)
	if !ok {
		return Agent{}, false
	}
	name, ok := data["name"].(string)
	if !ok || name == "" {
		return Agent{}, false
	}
	return Agent{Name: name}, true
}

func (a Agent) Render() types.Segment {
	return types.Segment{
		Text:     a.Name,
		Style:    "bold",
		Priority: 85,
	}
}
