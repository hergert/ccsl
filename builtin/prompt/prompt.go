package prompt

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"ccsl/internal/types"
)

// Render reads the last user prompt using session-aware logic
func Render(ctx context.Context, ctxObj map[string]any) types.Segment {
	var projectDir string
	var transcriptPath string
	
	// Extract context information
	if workspace, ok := ctxObj["workspace"].(map[string]any); ok {
		if dir, ok := workspace["project_dir"].(string); ok {
			projectDir = dir
		}
	}
	
	if transcript, ok := ctxObj["transcript_path"].(string); ok {
		transcriptPath = transcript
	}
	
	if projectDir == "" {
		if dir, err := os.Getwd(); err == nil {
			projectDir = dir
		}
	}
	
	// Get prompt max length from environment
	promptMax := 80
	if val := os.Getenv("STATUSLINE_PROMPT_MAX"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			promptMax = n
		}
	}

	// Read session-aware prompt
	lastPrompt := readSessionPrompt(projectDir, transcriptPath)
	if lastPrompt == "" {
		lastPrompt = lastUserPromptFromTranscript(transcriptPath)
	}
	
	if lastPrompt == "" {
		return types.Segment{}
	}

	// Squish whitespace and truncate
	lastPrompt = squish(lastPrompt, promptMax)
	
	return types.Segment{
		Text:     lastPrompt,
		Priority: 20, // lower priority for truncation
	}
}

func readSessionPrompt(projectDir, transcriptPath string) string {
	if projectDir == "" {
		return ""
	}
	
	claudeDir := filepath.Join(projectDir, ".claude")
	
	// Try session-specific prompt first
	if transcriptPath != "" {
		sessionID := filepath.Base(transcriptPath)
		if ext := filepath.Ext(sessionID); ext != "" {
			sessionID = strings.TrimSuffix(sessionID, ext)
		}
		
		sessionPrompt := filepath.Join(claudeDir, "prompts", sessionID+".txt")
		if data, err := os.ReadFile(sessionPrompt); err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	
	// Fallback to shared prompt file
	sharedPrompt := filepath.Join(claudeDir, "last_prompt.txt")
	if data, err := os.ReadFile(sharedPrompt); err == nil {
		return strings.TrimSpace(string(data))
	}
	
	return ""
}

func lastUserPromptFromTranscript(path string) string {
	if path == "" {
		return ""
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	
	// Read tail of file for performance (100KB should be plenty)
	content := string(data)
	if len(content) > 100000 {
		content = content[len(content)-100000:]
	}
	
	lines := strings.Split(content, "\n")
	
	// Process lines in reverse order (newest first)
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		
		// Try to parse as JSONL object
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err == nil {
			if objType, ok := obj["type"].(string); ok && objType == "user" {
				if message, ok := obj["message"].(map[string]any); ok {
					if content, ok := message["content"].(string); ok && content != "" {
						return content
					}
				}
			}
		}
	}
	
	// Fallback: try to parse entire file as JSON array
	var data_parsed any
	if err := json.Unmarshal(data, &data_parsed); err == nil {
		if items, ok := data_parsed.([]any); ok {
			for i := len(items) - 1; i >= 0; i-- {
				if item, ok := items[i].(map[string]any); ok {
					if objType, ok := item["type"].(string); ok && objType == "user" {
						if message, ok := item["message"].(map[string]any); ok {
							if content, ok := message["content"].(string); ok && content != "" {
								return content
							}
						}
					}
				}
			}
		}
	}
	
	return ""
}

func squish(s string, maxLen int) string {
	// Normalize whitespace
	s = strings.Join(strings.Fields(s), " ")
	
	if len(s) <= maxLen {
		return s
	}
	
	return strings.TrimSpace(s[:maxLen])
}