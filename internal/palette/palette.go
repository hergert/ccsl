package palette

import (
	"context"
	"strings"

	"ccsl/internal/config"
	"ccsl/internal/types"
)

const (
	Reset = "\x1b[0m"
	Bold  = "\x1b[1m"
	Dim   = "\x1b[2m"
)

type Palette struct {
	ansi bool
}

func From(cfg *config.Config, _ map[string]any) *Palette {
	return &Palette{ansi: cfg.Theme.ANSI}
}

func (p *Palette) Apply(text, style string) string {
	if !p.ansi || style == "" || style == "normal" {
		return text
	}

	switch style {
	case "bold":
		return Bold + text + Reset
	case "dim":
		return Dim + text + Reset
	default:
		if strings.HasPrefix(style, "\x1b[") {
			return style + text + Reset
		}
		return text
	}
}

func IconsEnabled(ctx context.Context) bool {
	if cfg, ok := ctx.Value(types.CtxKeyConfig).(*config.Config); ok {
		return cfg.Theme.Icons
	}
	return true
}
