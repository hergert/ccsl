package model

import (
	"strings"

	"github.com/hergert/ccsl/internal/types"
)

type Model struct {
	DisplayName string
}

func Parse(raw map[string]any) Model {
	m := Model{DisplayName: "Claude"}
	if data, ok := raw["model"].(map[string]any); ok {
		if name, ok := data["display_name"].(string); ok && name != "" {
			m.DisplayName = strings.TrimSpace(name)
		}
	}
	return m
}

func (m Model) Render() types.Segment {
	return types.Segment{
		Text:     m.DisplayName,
		Style:    "bold",
		Priority: 90,
	}
}
