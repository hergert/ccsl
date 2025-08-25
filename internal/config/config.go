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
	Padding  int    `toml:"padding"`
}

type ThemeConfig struct {
	Mode  string `toml:"mode"` // auto | light | dark
	Icons bool   `toml:"icons"`
	ANSI  bool   `toml:"ansi"`
}

type PluginsConfig struct {
	Order []string `toml:"order"`
}

type PluginConfig struct {
	Type         string   `toml:"type"`    // builtin | exec
	Command      string   `toml:"command"` // for exec type
	Args         []string `toml:"args"`
	Style        string   `toml:"style"` // normal | bold | dim
	TimeoutMS    int      `toml:"timeout_ms"`
	CacheTTLMS   int      `toml:"cache_ttl_ms"`
	OnlyIf       string   `toml:"only_if"`        // simple condition
	CacheKeyFrom string   `toml:"cache_key_from"` // dot path in context (e.g., "transcript_path")
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
	configPaths := []string{
		filepath.Join(xdg, "ccsl", "config.toml"),
		filepath.Join(os.Getenv("HOME"), ".claude", "ccsl.toml"),
	}

	for _, path := range configPaths {
		if data, err := os.ReadFile(path); err == nil {
			if _, err := toml.Decode(string(data), cfg); err == nil {
				break
			}
		}
	}

	// NOTE: CCSL_PROMPT_MAX belongs to the prompt segment and is read in the
	// prompt builtin, not here.

	// Apply simple overrides last
	if val := os.Getenv("CCSL_ANSI"); val == "0" {
		cfg.Theme.ANSI = false
	}

	if val := os.Getenv("CCSL_ICONS"); val == "0" {
		cfg.Theme.Icons = false
	}

	if v := os.Getenv("CCSL_TEMPLATE"); v != "" {
		cfg.UI.Template = v
	}
	if v := os.Getenv("CCSL_ORDER"); v != "" {
		parts := strings.Split(v, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		cfg.Plugins.Order = parts
	}
	if v := os.Getenv("CCSL_THEME"); v != "" {
		cfg.Theme.Mode = v
	} // auto|light|dark
	// CCSL_DISABLE: comma-separated segments to drop
	if v := os.Getenv("CCSL_DISABLE"); v != "" {
		drop := map[string]bool{}
		for _, s := range strings.Split(v, ",") {
			drop[strings.TrimSpace(s)] = true
		}
		filtered := make([]string, 0, len(cfg.Plugins.Order))
		for _, id := range cfg.Plugins.Order {
			if !drop[id] {
				filtered = append(filtered, id)
			}
		}
		cfg.Plugins.Order = filtered
	}

	return cfg
}

func defaultConfig() *Config {
	return &Config{
		UI: UIConfig{
			Template: "{model}  {cwd}{agent?prefix=  }{git?prefix=  }{prompt?prefix= — 🗣 }",
			Truncate: 120,
			Padding:  0,
		},
		Theme: ThemeConfig{
			Mode:  "auto",
			Icons: true,
			ANSI:  true,
		},
		Plugins: PluginsConfig{
			Order: []string{"model", "cwd", "agent", "git", "prompt"},
		},
		Plugin: map[string]PluginConfig{
			"git": {
				Type:       "builtin",
				Style:      "dim",
				TimeoutMS:  90,
				CacheTTLMS: 300,
			},
			"model": {
				Type:      "builtin",
				TimeoutMS: 10,
			},
			"cwd": {
				Type:      "builtin",
				TimeoutMS: 10,
			},
			"agent": {
				Type:       "builtin",
				TimeoutMS:  20,
				CacheTTLMS: 100,
			},
			"prompt": {
				Type:       "builtin",
				TimeoutMS:  30,
				CacheTTLMS: 50,
			},
		},
		Limits: LimitsConfig{
			PerPluginTimeoutMS: 120,
			TotalBudgetMS:      220,
		},
	}
}
