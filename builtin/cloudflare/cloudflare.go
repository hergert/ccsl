package cloudflare

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"ccsl/internal/palette"
	"ccsl/internal/types"

	"github.com/BurntSushi/toml"
)

// wranglerConfig holds relevant fields from wrangler.toml/json
type wranglerConfig struct {
	Name      string `json:"name" toml:"name"`
	AccountID string `json:"account_id" toml:"account_id"`
}

// Render returns Cloudflare account/project info from env vars and wrangler config.
// No subprocess spawning - reads config files directly for speed.
func Render(ctx context.Context, ctxObj map[string]any) types.Segment {
	// Get workspace directory from Claude's context
	workDir := ""
	if ws, ok := ctxObj["workspace"].(map[string]any); ok {
		if dir, ok := ws["current_dir"].(string); ok {
			workDir = dir
		}
	}

	// Read from environment
	envAccountID := getEnvAccountID()
	envName := os.Getenv("CLOUDFLARE_ENV")

	// Read from wrangler config in workspace
	var cfg wranglerConfig
	if workDir != "" {
		cfg = findWranglerConfig(workDir)
	}

	// Determine what to show
	accountID := envAccountID
	if accountID == "" {
		accountID = cfg.AccountID
	}

	projectName := cfg.Name

	// Nothing to show if no account and no project
	if accountID == "" && projectName == "" {
		return types.Segment{}
	}

	// Build output
	var parts []string

	if accountID != "" {
		parts = append(parts, shortenAccountID(accountID))
	}

	if projectName != "" {
		if len(parts) > 0 {
			parts = append(parts, "/", projectName)
		} else {
			parts = append(parts, projectName)
		}
	}

	text := strings.Join(parts, " ")

	// Add env suffix if set
	if envName != "" {
		text += " @" + envName
	}

	// Check for account mismatch (env vs config) - common failure mode
	mismatch := false
	if envAccountID != "" && cfg.AccountID != "" && envAccountID != cfg.AccountID {
		mismatch = true
	}

	if mismatch {
		text += " ⚠"
	}

	// Add icon if enabled
	if palette.IconsEnabled(ctx) {
		text = "☁️ " + text
	}

	return types.Segment{
		Text:     text,
		Style:    "dim",
		Priority: 35,
	}
}

// getEnvAccountID checks both new and deprecated env var names
func getEnvAccountID() string {
	if id := os.Getenv("CLOUDFLARE_ACCOUNT_ID"); id != "" {
		return id
	}
	// Deprecated but still supported
	return os.Getenv("CF_ACCOUNT_ID")
}

// findWranglerConfig looks for wrangler config in the workspace
func findWranglerConfig(workDir string) wranglerConfig {
	// Check for .wrangler/deploy/config.json redirect first
	redirectPath := filepath.Join(workDir, ".wrangler", "deploy", "config.json")
	if data, err := os.ReadFile(redirectPath); err == nil {
		var redirect struct {
			ConfigPath string `json:"configPath"`
		}
		if json.Unmarshal(data, &redirect) == nil && redirect.ConfigPath != "" {
			if cfg := parseWranglerConfig(filepath.Join(workDir, redirect.ConfigPath)); cfg.Name != "" {
				return cfg
			}
		}
	}

	// Try standard config file names
	candidates := []string{
		"wrangler.toml",
		"wrangler.json",
		"wrangler.jsonc",
	}

	for _, name := range candidates {
		path := filepath.Join(workDir, name)
		if cfg := parseWranglerConfig(path); cfg.Name != "" || cfg.AccountID != "" {
			return cfg
		}
	}

	return wranglerConfig{}
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
		// JSON or JSONC - strip comments for JSONC
		cleaned := stripJSONComments(string(data))
		json.Unmarshal([]byte(cleaned), &cfg)
	}

	return cfg
}

// stripJSONComments removes // and /* */ comments for JSONC support
func stripJSONComments(s string) string {
	var result strings.Builder
	i := 0
	inString := false

	for i < len(s) {
		// Track string state
		if s[i] == '"' && (i == 0 || s[i-1] != '\\') {
			inString = !inString
			result.WriteByte(s[i])
			i++
			continue
		}

		// Skip comments only outside strings
		if !inString && i+1 < len(s) {
			// Line comment
			if s[i] == '/' && s[i+1] == '/' {
				for i < len(s) && s[i] != '\n' {
					i++
				}
				continue
			}
			// Block comment
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
