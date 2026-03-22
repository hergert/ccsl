package cwd

import (
	"os"
	"path/filepath"

	"github.com/hergert/ccsl/internal/types"
)

type Dir struct {
	Path string
}

func Parse(raw map[string]any) Dir {
	if ws, ok := raw["workspace"].(map[string]any); ok {
		if dir, ok := ws["current_dir"].(string); ok && dir != "" {
			return Dir{Path: dir}
		}
	}
	if dir, err := os.Getwd(); err == nil {
		return Dir{Path: dir}
	}
	return Dir{}
}

func (d Dir) Render() types.Segment {
	if d.Path == "" {
		return types.Segment{}
	}

	name := filepath.Base(d.Path)
	if name == "" || name == "/" {
		name = d.Path
	}

	return types.Segment{
		Text:     name,
		Priority: 80,
	}
}
