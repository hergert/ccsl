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
	Icons bool `toml:"icons"`
	ANSI  bool `toml:"ansi"`
}

type PluginsConfig struct {
	Order []string `toml:"order"` // optional override, otherwise derived from template
}

type PluginConfig struct {
	Type      string   `toml:"type"`    // builtin | exec
	Command   string   `toml:"command"` // for exec type
	Args      []string `toml:"args"`
	TimeoutMS int      `toml:"timeout_ms"`
}

type LimitsConfig struct {
	PerPluginTimeoutMS int `toml:"per_plugin_timeout_ms"`
	TotalBudgetMS      int `toml:"total_budget_ms"`
}

// Load reads configuration from standard locations with env overrides
func Load() *Config {
	cfg := defaultConfig()

	// Try loading from config files
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		xdg = filepath.Join(os.Getenv("HOME"), ".config")
	}
	paths := []string{
		filepath.Join(xdg, "ccsl", "config.toml"),
		filepath.Join(os.Getenv("HOME"), ".claude", "ccsl.toml"),
	}

	for _, path := range paths {
		if data, err := os.ReadFile(path); err == nil {
			toml.Decode(string(data), cfg)
			break
		}
	}

	// Env overrides
	if os.Getenv("CCSL_ANSI") == "0" {
		cfg.Theme.ANSI = false
	}
	if os.Getenv("CCSL_ICONS") == "0" {
		cfg.Theme.Icons = false
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
			Template: "{model}{ctx?prefix= }{cost?prefix= } {cwd}{git?prefix=:}",
			Truncate: 120,
		},
		Theme: ThemeConfig{
			Icons: true,
			ANSI:  true,
		},
		Plugins: PluginsConfig{
			Order: nil, // derive from template by default
		},
		Plugin: map[string]PluginConfig{
			"model": {Type: "builtin", TimeoutMS: 10},
			"cwd":   {Type: "builtin", TimeoutMS: 10},
			"git":   {Type: "builtin", TimeoutMS: 80},
			"ctx":   {Type: "builtin", TimeoutMS: 10},
			"cost":  {Type: "builtin", TimeoutMS: 10},
		},
		Limits: LimitsConfig{
			PerPluginTimeoutMS: 100,
			TotalBudgetMS:      200,
		},
	}
}
