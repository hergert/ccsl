package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	UI      UIConfig                `toml:"ui"`
	Theme   ThemeConfig             `toml:"theme"`
	Plugins PluginsConfig           `toml:"plugins"`
	Plugin  map[string]PluginConfig `toml:"plugin"`
	Limits  LimitsConfig            `toml:"limits"`
}

type UIConfig struct {
	Template string `toml:"template"`
	Truncate int    `toml:"truncate"`
}

type ThemeConfig struct {
	ANSI bool `toml:"ansi"`
}

type PluginsConfig struct {
	Order []string `toml:"order"` // optional override, otherwise derived from template
}

type PluginConfig struct {
	Type      string   `toml:"type"`    // builtin | exec
	Command   string   `toml:"command"` // for exec type
	Args      []string `toml:"args"`
	TimeoutMS int      `toml:"timeout_ms"`
	Untracked bool     `toml:"untracked"` // for git: include untracked files
}

type LimitsConfig struct {
	PerPluginTimeoutMS int `toml:"per_plugin_timeout_ms"`
	TotalBudgetMS      int `toml:"total_budget_ms"`
}

// Load reads configuration from standard locations with env overrides
// Optional projectDir enables project-local config at .claude/ccsl.toml
func Load(projectDir ...string) *Config {
	cfg := defaultConfig()

	// Try loading from config files (first match wins)
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		xdg = filepath.Join(os.Getenv("HOME"), ".config")
	}

	var paths []string
	// Project-local config takes highest priority
	if len(projectDir) > 0 && projectDir[0] != "" {
		paths = append(paths, filepath.Join(projectDir[0], ".claude", "ccsl.toml"))
	}
	// Global config locations
	paths = append(paths,
		filepath.Join(xdg, "ccsl", "config.toml"),
		filepath.Join(os.Getenv("HOME"), ".claude", "ccsl.toml"),
	)

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if _, err := toml.Decode(string(data), cfg); err != nil {
			continue
		}
		break
	}

	// Env overrides
	if os.Getenv("CCSL_ANSI") == "0" {
		cfg.Theme.ANSI = false
	}
	if v := os.Getenv("CCSL_TEMPLATE"); v != "" {
		cfg.UI.Template = v
	}
	if v, ok := os.LookupEnv("CCSL_ORDER"); ok {
		if v == "" {
			cfg.Plugins.Order = nil // empty = derive from template
		} else {
			parts := strings.Split(v, ",")
			var order []string
			for _, p := range parts {
				if s := strings.TrimSpace(p); s != "" {
					order = append(order, s)
				}
			}
			cfg.Plugins.Order = order
		}
	}

	return cfg
}

func defaultConfig() *Config {
	return &Config{
		UI: UIConfig{
			Template: "{model}{effort?prefix= }{agent?prefix= }{worktree?prefix= }{ctx?prefix= }{cost?prefix= }{ratelimit?prefix= · } {cwd}{git?prefix=:}{pr?prefix= }{gcp?prefix= }{cf?prefix= }",
			Truncate: 120,
		},
		Theme: ThemeConfig{
			ANSI: true,
		},
		Plugins: PluginsConfig{
			Order: nil, // derive from template by default
		},
		// Pure JSON parsers need no entries; only segments that do real work
		// (subprocesses, file walks) get their own timeout.
		Plugin: map[string]PluginConfig{
			"git": {Type: "builtin", TimeoutMS: 80},
		},
		Limits: LimitsConfig{
			PerPluginTimeoutMS: 100,
			TotalBudgetMS:      200,
		},
	}
}
