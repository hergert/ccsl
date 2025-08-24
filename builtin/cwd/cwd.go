package cwd

import (
	"context"
	"os"
	"path/filepath"

	"ccsl/internal/palette"
	"ccsl/internal/types"
)

// Render shows the basename of the current working directory
func Render(ctx context.Context, ctxObj map[string]any) types.Segment {
	var currentDir string

	// Try to get from Claude context first
	if workspace, ok := ctxObj["workspace"].(map[string]any); ok {
		if dir, ok := workspace["current_dir"].(string); ok {
			currentDir = dir
		}
	}

	// Fallback to actual cwd
	if currentDir == "" {
		if dir, err := os.Getwd(); err == nil {
			currentDir = dir
		}
	}

	if currentDir == "" {
		return types.Segment{}
	}

	dirName := filepath.Base(currentDir)
	if dirName == "" || dirName == "/" {
		dirName = currentDir
	}

	icon := ""
	if palette.IconsEnabled(ctx) {
		icon = "📁 "
	}
	return types.Segment{
		Text:     icon + dirName,
		Priority: 80, // high priority
	}
}
