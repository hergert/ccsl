package gcp

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/hergert/ccsl/internal/types"
)

// Render returns GCP identity + project from env vars and gcloud config files.
// No subprocess spawning - reads config files directly for speed.
func Render(ctx context.Context, ctxObj map[string]any) types.Segment {
	envAccount := os.Getenv("CLOUDSDK_CORE_ACCOUNT")
	envProject := os.Getenv("CLOUDSDK_CORE_PROJECT")
	envConfigName := os.Getenv("CLOUDSDK_ACTIVE_CONFIG_NAME")

	// Read from config files
	fileAccount, fileProject, fileName, configCount := readGcloudConfig()

	// Determine effective values
	account := envAccount
	if account == "" {
		account = fileAccount
	}
	project := envProject
	if project == "" {
		project = fileProject
	}
	configName := envConfigName
	if configName == "" {
		configName = fileName
	}

	// Nothing to show
	if account == "" && project == "" {
		return types.Segment{}
	}

	// Check for mismatch (env vars override file config)
	mismatch := false
	if (envAccount != "" && fileAccount != "" && envAccount != fileAccount) ||
		(envProject != "" && fileProject != "" && envProject != fileProject) {
		mismatch = true
	}

	// Build compact output: gcp:project or gcp:project@config
	text := "gcp:"
	if project != "" {
		text += project
	} else if account != "" {
		text += shortenEmail(account)
	}

	// Show config name if multiple configs or non-default
	if configCount > 1 || (configName != "" && configName != "default") {
		text += "@" + configName
	}

	if mismatch {
		text += "⚠"
	}

	return types.Segment{
		Text:     text,
		Style:    "dim",
		Priority: 35,
	}
}

// readGcloudConfig reads the active gcloud configuration from disk
func readGcloudConfig() (account, project, configName string, configCount int) {
	configDir := os.Getenv("CLOUDSDK_CONFIG")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}
		configDir = filepath.Join(home, ".config", "gcloud")
	}

	// Count configurations
	configsDir := filepath.Join(configDir, "configurations")
	if entries, err := os.ReadDir(configsDir); err == nil {
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "config_") {
				configCount++
			}
		}
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
	configPath := filepath.Join(configsDir, "config_"+configName)
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
