package model

import (
	"context"
	"strings"

	"ccsl/internal/palette"
	"ccsl/internal/types"
)

// Render extracts and formats the model information
func Render(ctx context.Context, ctxObj map[string]any) types.Segment {
	model, ok := ctxObj["model"].(map[string]any)
	if !ok {
		return types.Segment{}
	}

	modelID := getStringValue(model, "id")
	modelName := getStringValue(model, "display_name")
	
	if modelName == "" && modelID == "" {
		return types.Segment{}
	}

	// Prefer display name, fallback to ID
	text := modelName
	if text == "" {
		text = modelID
	}

	// If both exist and ID not contained in name, show both
	if modelID != "" && modelName != "" && !strings.Contains(strings.ToLower(modelName), strings.ToLower(modelID)) {
		text = modelName + " (" + modelID + ")"
	}

	icon := ""
	if palette.IconsEnabled(ctx) { icon = "🤖 " }
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