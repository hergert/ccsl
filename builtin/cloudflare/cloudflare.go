package cloudflare

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/hergert/ccsl/internal/types"

	"github.com/BurntSushi/toml"
)

type wranglerConfig struct {
	Name      string                    `json:"name" toml:"name"`
	AccountID string                    `json:"account_id" toml:"account_id"`
	Env       map[string]wranglerEnvCfg `json:"env" toml:"env"`
}

type wranglerEnvCfg struct {
	Name      string `json:"name" toml:"name"`
	AccountID string `json:"account_id" toml:"account_id"`
}

func Render(raw map[string]any) types.Segment {
	var currentDir, projectDir string
	if ws, ok := raw["workspace"].(map[string]any); ok {
		if dir, ok := ws["current_dir"].(string); ok {
			currentDir = dir
		}
		if dir, ok := ws["project_dir"].(string); ok {
			projectDir = dir
		}
	}

	cfg, _ := findWranglerConfig(currentDir, projectDir)

	envName := os.Getenv("CLOUDFLARE_ENV")

	if envName != "" && cfg.Env != nil {
		if envCfg, ok := cfg.Env[envName]; ok {
			if envCfg.Name != "" {
				cfg.Name = envCfg.Name
			}
			if envCfg.AccountID != "" {
				cfg.AccountID = envCfg.AccountID
			}
		}
	}

	projectName := cfg.Name

	if projectName == "" {
		return types.Segment{}
	}

	text := "cf:" + projectName

	if envName != "" {
		text += "@" + envName
	}

	envAccountID := getEnvAccountID()
	if envAccountID != "" && cfg.AccountID != "" && envAccountID != cfg.AccountID {
		text += "⚠"
	}

	return types.Segment{
		Text:     text,
		Style:    "dim",
		Priority: 35,
	}
}

// Checks both current and deprecated env var names.
func getEnvAccountID() string {
	if id := os.Getenv("CLOUDFLARE_ACCOUNT_ID"); id != "" {
		return id
	}
	return os.Getenv("CF_ACCOUNT_ID")
}

func findWranglerConfig(currentDir, projectDir string) (wranglerConfig, string) {
	candidates := []string{"wrangler.toml", "wrangler.json", "wrangler.jsonc"}

	dir := currentDir
	for {
		// .wrangler/deploy/config.json can redirect to the actual config
		redirectPath := filepath.Join(dir, ".wrangler", "deploy", "config.json")
		if data, err := os.ReadFile(redirectPath); err == nil {
			var redirect struct {
				ConfigPath string `json:"configPath"`
			}
			if json.Unmarshal(data, &redirect) == nil && redirect.ConfigPath != "" {
				cfgPath := filepath.Join(dir, redirect.ConfigPath)
				if cfg := parseWranglerConfig(cfgPath); cfg.Name != "" {
					return cfg, filepath.Dir(cfgPath)
				}
			}
		}

		for _, name := range candidates {
			path := filepath.Join(dir, name)
			if cfg := parseWranglerConfig(path); cfg.Name != "" || cfg.AccountID != "" {
				return cfg, dir
			}
		}

		if dir == projectDir || dir == "/" || dir == "." {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return wranglerConfig{}, ""
}

func parseWranglerConfig(path string) wranglerConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		return wranglerConfig{}
	}

	var cfg wranglerConfig
	if strings.HasSuffix(path, ".toml") {
		if _, err := toml.Decode(string(data), &cfg); err != nil {
			return wranglerConfig{}
		}
	} else {
		cleaned := stripJSONComments(string(data))
		if err := json.Unmarshal([]byte(cleaned), &cfg); err != nil {
			return wranglerConfig{}
		}
	}
	return cfg
}

func stripJSONComments(s string) string {
	var result strings.Builder
	i := 0
	inString := false

	for i < len(s) {
		if s[i] == '"' && (i == 0 || s[i-1] != '\\') {
			inString = !inString
			result.WriteByte(s[i])
			i++
			continue
		}

		if !inString && i+1 < len(s) {
			if s[i] == '/' && s[i+1] == '/' {
				for i < len(s) && s[i] != '\n' {
					i++
				}
				continue
			}
			if s[i] == '/' && s[i+1] == '*' {
				i += 2
				for i+1 < len(s) && (s[i] != '*' || s[i+1] != '/') {
					i++
				}
				if i+1 < len(s) {
					i += 2
				}
				continue
			}
		}

		result.WriteByte(s[i])
		i++
	}

	return result.String()
}
