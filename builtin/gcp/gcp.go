package gcp

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"

	"ccsl/internal/palette"
	"ccsl/internal/types"
)

// Render returns GCP identity + project from env vars and gcloud config files.
// No subprocess spawning - reads config files directly for speed.
func Render(ctx context.Context, ctxObj map[string]any) types.Segment {
	account := os.Getenv("CLOUDSDK_CORE_ACCOUNT")
	project := os.Getenv("CLOUDSDK_CORE_PROJECT")
	configName := os.Getenv("CLOUDSDK_ACTIVE_CONFIG_NAME")

	// If not fully specified via env, read from config files
	if account == "" || project == "" {
		fileAccount, fileProject, fileName := readGcloudConfig()
		if account == "" {
			account = fileAccount
		}
		if project == "" {
			project = fileProject
		}
		if configName == "" {
			configName = fileName
		}
	}

	// Nothing to show
	if account == "" && project == "" {
		return types.Segment{}
	}

	// Build output
	var parts []string

	if account != "" {
		parts = append(parts, shortenEmail(account))
	}
	if project != "" {
		if len(parts) > 0 {
			parts = append(parts, "/", project)
		} else {
			parts = append(parts, project)
		}
	}

	text := strings.Join(parts, " ")

	// Add config name if not default
	if configName != "" && configName != "default" {
		text += " (" + configName + ")"
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

// readGcloudConfig reads the active gcloud configuration from disk
func readGcloudConfig() (account, project, configName string) {
	configDir := os.Getenv("CLOUDSDK_CONFIG")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}
		configDir = filepath.Join(home, ".config", "gcloud")
	}

	// Determine active config name
	configName = os.Getenv("CLOUDSDK_ACTIVE_CONFIG_NAME")
	if configName == "" {
		// Read from active_config file
		data, err := os.ReadFile(filepath.Join(configDir, "active_config"))
		if err == nil {
			configName = strings.TrimSpace(string(data))
		}
	}
	if configName == "" {
		configName = "default"
	}

	// Read the config file
	configPath := filepath.Join(configDir, "configurations", "config_"+configName)
	account, project = parseGcloudConfig(configPath)

	return
}

// parseGcloudConfig parses a gcloud config INI file for [core] account and project
func parseGcloudConfig(path string) (account, project string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	inCore := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Section header
		if strings.HasPrefix(line, "[") {
			inCore = strings.HasPrefix(line, "[core]")
			continue
		}

		if !inCore {
			continue
		}

		// Key = value
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])

			switch key {
			case "account":
				account = value
			case "project":
				project = value
			}
		}

		// Early exit if we have both
		if account != "" && project != "" {
			return
		}
	}

	return
}

// shortenEmail shortens an email to just the username part
func shortenEmail(email string) string {
	if idx := strings.Index(email, "@"); idx > 0 {
		return email[:idx]
	}
	return email
}
