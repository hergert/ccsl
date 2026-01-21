package cloudflare

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"ccsl/internal/types"

	"github.com/BurntSushi/toml"
)

// wranglerConfig holds relevant fields from wrangler.toml/json
type wranglerConfig struct {
	Name      string                    `json:"name" toml:"name"`
	AccountID string                    `json:"account_id" toml:"account_id"`
	Env       map[string]wranglerEnvCfg `json:"env" toml:"env"`
}

type wranglerEnvCfg struct {
	Name      string `json:"name" toml:"name"`
	AccountID string `json:"account_id" toml:"account_id"`
}

// Render returns Cloudflare account/project info from env vars and wrangler config.
func Render(ctx context.Context, ctxObj map[string]any) types.Segment {
	// Get workspace directories from Claude's context
	var currentDir, projectDir string
	if ws, ok := ctxObj["workspace"].(map[string]any); ok {
		if dir, ok := ws["current_dir"].(string); ok {
			currentDir = dir
		}
		if dir, ok := ws["project_dir"].(string); ok {
			projectDir = dir
		}
	}

	// Find nearest wrangler config (walk current_dir → project_dir)
	cfg, configDir := findWranglerConfig(currentDir, projectDir)

	// Resolve env
	envName := os.Getenv("CLOUDFLARE_ENV")

	// Apply env-specific overrides from config
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

	// Resolve account ID (priority order)
	accountID := resolveAccountID(configDir, cfg.AccountID)

	projectName := cfg.Name

	// Nothing to show if no project
	if projectName == "" {
		return types.Segment{}
	}

	// Build compact output: cf:project or cf:project@env
	text := "cf:" + projectName

	if envName != "" {
		text += "@" + envName
	}

	// Check for mismatch between process env and config
	envAccountID := getEnvAccountID()
	if envAccountID != "" && cfg.AccountID != "" && envAccountID != cfg.AccountID {
		text += "⚠"
	}

	// If we have account but no project, show shortened account
	if projectName == "" && accountID != "" {
		text = "cf:" + shortenAccountID(accountID)
	}

	return types.Segment{
		Text:     text,
		Style:    "dim",
		Priority: 35,
	}
}

// resolveAccountID resolves account ID (no .env reading for security)
func resolveAccountID(configDir, configAccountID string) string {
	// 1. Process env var
	if id := getEnvAccountID(); id != "" {
		return id
	}

	// 2. Config file account_id
	if configAccountID != "" {
		return configAccountID
	}

	// 3. Wrangler account cache (non-secret metadata)
	if configDir != "" {
		if id := readWranglerAccountCache(configDir); id != "" {
			return id
		}
	}

	return ""
}

// getEnvAccountID checks both new and deprecated env var names
func getEnvAccountID() string {
	if id := os.Getenv("CLOUDFLARE_ACCOUNT_ID"); id != "" {
		return id
	}
	return os.Getenv("CF_ACCOUNT_ID")
}

// findWranglerConfig walks from currentDir up to projectDir looking for wrangler config
// Returns the config and the directory where it was found
func findWranglerConfig(currentDir, projectDir string) (wranglerConfig, string) {
	candidates := []string{"wrangler.toml", "wrangler.json", "wrangler.jsonc"}

	dir := currentDir
	for {
		// Check for .wrangler/deploy/config.json redirect first
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

		// Try standard config file names
		for _, name := range candidates {
			path := filepath.Join(dir, name)
			if cfg := parseWranglerConfig(path); cfg.Name != "" || cfg.AccountID != "" {
				return cfg, dir
			}
		}

		// Stop at project_dir boundary
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

// parseWranglerConfig parses a wrangler config file (toml or json/jsonc)
func parseWranglerConfig(path string) wranglerConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		return wranglerConfig{}
	}

	var cfg wranglerConfig
	if strings.HasSuffix(path, ".toml") {
		toml.Decode(string(data), &cfg)
	} else {
		cleaned := stripJSONComments(string(data))
		json.Unmarshal([]byte(cleaned), &cfg)
	}
	return cfg
}

// readWranglerAccountCache reads from .wrangler/account-id cache
func readWranglerAccountCache(configDir string) string {
	// Try node_modules/.cache/wrangler/wrangler-account.json
	cachePaths := []string{
		filepath.Join(configDir, "node_modules", ".cache", "wrangler", "wrangler-account.json"),
		filepath.Join(configDir, ".wrangler", "wrangler-account.json"),
	}

	for _, path := range cachePaths {
		if data, err := os.ReadFile(path); err == nil {
			var cache struct {
				Account struct {
					ID string `json:"id"`
				} `json:"account"`
			}
			if json.Unmarshal(data, &cache) == nil && cache.Account.ID != "" {
				return cache.Account.ID
			}
		}
	}
	return ""
}

// stripJSONComments removes // and /* */ comments for JSONC support
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
				for i+1 < len(s) && !(s[i] == '*' && s[i+1] == '/') {
					i++
				}
				i += 2
				continue
			}
		}

		result.WriteByte(s[i])
		i++
	}

	return result.String()
}

// shortenAccountID shows first 4 and last 4 chars of account ID
func shortenAccountID(id string) string {
	if len(id) <= 10 {
		return id
	}
	return id[:4] + "…" + id[len(id)-4:]
}
