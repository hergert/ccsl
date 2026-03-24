package gcp

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/hergert/ccsl/internal/types"
)

// No subprocess spawning — reads config files directly for speed.
func Render(raw map[string]any) types.Segment {
	envAccount := os.Getenv("CLOUDSDK_CORE_ACCOUNT")
	envProject := os.Getenv("CLOUDSDK_CORE_PROJECT")
	envConfigName := os.Getenv("CLOUDSDK_ACTIVE_CONFIG_NAME")

	fileAccount, fileProject, fileName, configCount := readGcloudConfig()

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

	if account == "" && project == "" {
		return types.Segment{}
	}

	mismatch := (envAccount != "" && fileAccount != "" && envAccount != fileAccount) ||
		(envProject != "" && fileProject != "" && envProject != fileProject)

	text := "gcp:"
	if project != "" {
		text += project
	} else if account != "" {
		text += shortenEmail(account)
	}

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

func readGcloudConfig() (account, project, configName string, configCount int) {
	configDir := os.Getenv("CLOUDSDK_CONFIG")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}
		configDir = filepath.Join(home, ".config", "gcloud")
	}

	configsDir := filepath.Join(configDir, "configurations")
	if entries, err := os.ReadDir(configsDir); err == nil {
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "config_") {
				configCount++
			}
		}
	}

	configName = os.Getenv("CLOUDSDK_ACTIVE_CONFIG_NAME")
	if configName == "" {
		data, err := os.ReadFile(filepath.Join(configDir, "active_config"))
		if err == nil {
			configName = strings.TrimSpace(string(data))
		}
	}
	if configName == "" {
		configName = "default"
	}

	configPath := filepath.Join(configsDir, "config_"+configName)
	account, project = parseGcloudConfig(configPath)

	return
}

func parseGcloudConfig(path string) (account, project string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	inCore := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "[") {
			inCore = strings.HasPrefix(line, "[core]")
			continue
		}

		if !inCore {
			continue
		}

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

		if account != "" && project != "" {
			return
		}
	}

	return
}

func shortenEmail(email string) string {
	if idx := strings.Index(email, "@"); idx > 0 {
		return email[:idx]
	}
	return email
}
