package config

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
)

type Config struct {
	UI      UIConfig         `toml:"ui"`
	Theme   ThemeConfig      `toml:"theme"`
	Plugins PluginsConfig    `toml:"plugins"`
	Plugin  map[string]PluginConfig `toml:"plugin"`
	Limits  LimitsConfig     `toml:"limits"`
}

type UIConfig struct {
	Template string `toml:"template"`
	Truncate int    `toml:"truncate"`
	Padding  int    `toml:"padding"`
}

type ThemeConfig struct {
	Mode  string `toml:"mode"`  // auto | light | dark
	Icons bool   `toml:"icons"`
	ANSI  bool   `toml:"ansi"`
}

type PluginsConfig struct {
	Order []string `toml:"order"`
}

type PluginConfig struct {
	Type        string `toml:"type"`         // builtin | exec
	Command     string `toml:"command"`      // for exec type
	Style       string `toml:"style"`        // normal | bold | dim
	TimeoutMS   int    `toml:"timeout_ms"`
	CacheTTLMS  int    `toml:"cache_ttl_ms"`
	OnlyIf      string `toml:"only_if"`      // simple condition
}

type LimitsConfig struct {
	PerPluginTimeoutMS int `toml:"per_plugin_timeout_ms"`
	TotalBudgetMS      int `toml:"total_budget_ms"`
}

// Load reads configuration from standard locations with env overrides
func Load() *Config {
	cfg := defaultConfig()
	
	// Try loading from config files
	configPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".config", "ccsl", "config.toml"),
		filepath.Join(os.Getenv("HOME"), ".claude", "ccsl.toml"),
	}
	
	for _, path := range configPaths {
		if data, err := os.ReadFile(path); err == nil {
			if _, err := toml.Decode(string(data), cfg); err == nil {
				break
			}
		}
	}
	
	// Apply environment variable overrides
	if val := os.Getenv("CCSL_PROMPT_MAX"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.UI.Truncate = n
		}
	}
	
	if val := os.Getenv("CCSL_ANSI"); val != "" {
		cfg.Theme.ANSI = val != "0" && val != "false"
	}
	
	if val := os.Getenv("CCSL_ICONS"); val != "" {
		cfg.Theme.Icons = val != "0" && val != "false"
	}
	
	return cfg
}

func defaultConfig() *Config {
	return &Config{
		UI: UIConfig{
			Template: "{model}  {cwd}  {agent}  {git?prefix=  }{prompt?prefix= — 🗣 }",
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