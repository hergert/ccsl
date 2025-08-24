package model

import (
	"context"
	"strings"

	"ccsl/internal/palette"
	"ccsl/internal/types"
)

// Render extracts the model's display name
func Render(ctx context.Context, ctxObj map[string]any) types.Segment {
	text := "Claude" // default
	if model, ok := ctxObj["model"].(map[string]any); ok {
		if name, ok := model["display_name"].(string); ok && name != "" {
			text = strings.TrimSpace(name)
		}
	}
	icon := ""
	if palette.IconsEnabled(ctx) {
		icon = "🤖 "
	}
	return types.Segment{
		Text:     icon + text,
		Style:    "bold",
		Priority: 90, // high priority
	}
}

func getStringValue(m map[string]any, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
