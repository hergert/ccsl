package runner

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"sync"
	"time"

	"ccsl/internal/config"
	"ccsl/internal/types"
	"ccsl/builtin/model"
	"ccsl/builtin/cwd"
	"ccsl/builtin/agent"
	"ccsl/builtin/prompt"
	"ccsl/builtin/git"
)

// Cache for plugin results
type cacheEntry struct {
	segment types.Segment
	expires time.Time
}

var (
	pluginCache = make(map[string]cacheEntry)
	cacheMux    sync.RWMutex
)

// Collect runs all plugins and built-ins to gather segments
func Collect(ctx context.Context, claudeJSON []byte, cfg *config.Config) []types.Segment {
	var wg sync.WaitGroup
	segments := make(chan types.Segment, len(cfg.Plugins.Order))
	
	var ctxObj map[string]any
	json.Unmarshal(claudeJSON, &ctxObj)

	// Process each plugin in the configured order
	for _, pluginID := range cfg.Plugins.Order {
		pluginCfg, exists := cfg.Plugin[pluginID]
		if !exists {
			continue
		}

		wg.Add(1)
		go func(id string, pcfg config.PluginConfig) {
			defer wg.Done()
			
			// Check cache first
			if seg, found := getCached(id); found {
				segments <- seg
				return
			}

			// Create plugin-specific timeout context
			timeout := time.Duration(pcfg.TimeoutMS) * time.Millisecond
			if timeout == 0 {
				timeout = time.Duration(cfg.Limits.PerPluginTimeoutMS) * time.Millisecond
			}
			
			pluginCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			var seg types.Segment
			if pcfg.Type == "builtin" {
				seg = runBuiltin(pluginCtx, id, ctxObj)
			} else if pcfg.Type == "exec" {
				seg = runExec(pluginCtx, pcfg.Command, claudeJSON)
			}
			
			seg.ID = id
			if seg.Priority == 0 {
				seg.Priority = 50 // default priority
			}

			// Cache the result
			if pcfg.CacheTTLMS > 0 {
				setCached(id, seg, time.Duration(pcfg.CacheTTLMS)*time.Millisecond)
			}

			segments <- seg
		}(pluginID, pluginCfg)
	}

	// Close segments channel after all goroutines complete
	go func() {
		wg.Wait()
		close(segments)
	}()

	// Collect all segments
	var result []types.Segment
	for seg := range segments {
		if seg.Text != "" { // only include non-empty segments
			result = append(result, seg)
		}
	}

	return result
}

func runBuiltin(ctx context.Context, id string, ctxObj map[string]any) types.Segment {
	switch id {
	case "model":
		return model.Render(ctx, ctxObj)
	case "cwd":
		return cwd.Render(ctx, ctxObj)
	case "agent":
		return agent.Render(ctx, ctxObj)
	case "prompt":
		return prompt.Render(ctx, ctxObj)
	case "git":
		return git.Render(ctx, ctxObj)
	default:
		return types.Segment{}
	}
}

func runExec(ctx context.Context, command string, claudeJSON []byte) types.Segment {
	cmd := exec.CommandContext(ctx, command)
	cmd.Stdin = strings.NewReader(string(claudeJSON))
	
	output, err := cmd.Output()
	if err != nil {
		return types.Segment{} // silent failure
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return types.Segment{}
	}

	// Try to parse as JSON first
	var resp types.PluginResponse
	if err := json.Unmarshal([]byte(result), &resp); err == nil {
		return types.Segment{
			Text:     resp.Text,
			Style:    resp.Style,
			Align:    resp.Align,
			Priority: resp.Priority,
		}
	}

	// Otherwise treat as plain text
	return types.Segment{
		Text: result,
	}
}

func getCached(id string) (types.Segment, bool) {
	cacheMux.RLock()
	defer cacheMux.RUnlock()
	
	entry, exists := pluginCache[id]
	if !exists || time.Now().After(entry.expires) {
		return types.Segment{}, false
	}
	
	return entry.segment, true
}

func setCached(id string, seg types.Segment, ttl time.Duration) {
	cacheMux.Lock()
	defer cacheMux.Unlock()
	
	pluginCache[id] = cacheEntry{
		segment: seg,
		expires: time.Now().Add(ttl),
	}
}