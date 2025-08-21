package palette

import (
	"context"
	"strings"

	"ccsl/internal/config"
	"ccsl/internal/types"
)

// ANSI escape codes
const (
	ANSIReset = "\x1b[0m"
	ANSIBold  = "\x1b[1m"
	ANSIDim   = "\x1b[2m"
)

type Palette struct {
	ansiEnabled bool
}

// From creates a palette based on config and Claude context
func From(cfg *config.Config, ctxObj map[string]any) *Palette {
	ansiEnabled := cfg.Theme.ANSI
	
	// Auto-detect theme based on Claude's output_style if available
	if cfg.Theme.Mode == "auto" {
		// TODO: Implement auto-detection based on ctxObj output_style
		// For now, default to enabled
	}

	return &Palette{
		ansiEnabled: ansiEnabled,
	}
}

// Apply applies styling to text based on the style string
func (p *Palette) Apply(text, style string) string {
	if !p.ansiEnabled || style == "" || style == "normal" {
		return text
	}

	switch style {
	case "bold":
		return ANSIBold + text + ANSIReset
	case "dim":
		return ANSIDim + text + ANSIReset
	default:
		// Check if it's already raw ANSI
		if strings.Contains(style, "\x1b[") {
			return style + text + ANSIReset
		}
		return text
	}
}

// IconsEnabled checks if icons should be displayed based on theme config
func IconsEnabled(ctx context.Context) bool {
	if cfg, ok := ctx.Value(types.CtxKeyConfig).(*config.Config); ok {
		return cfg.Theme.Icons
	}
	return true
}