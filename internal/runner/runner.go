package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"ccsl/builtin/agent"
	"ccsl/builtin/cwd"
	"ccsl/builtin/git"
	"ccsl/builtin/model"
	"ccsl/builtin/prompt"
	"ccsl/internal/config"
	"ccsl/internal/types"
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
	if err := json.Unmarshal(claudeJSON, &ctxObj); err != nil {
		// Can't do much, but at least don't panic on nil map access
	}

	// Process each plugin in the configured order
	for _, pluginID := range cfg.Plugins.Order {
		pluginCfg, exists := cfg.Plugin[pluginID]
		if !exists {
			continue
		}

		wg.Add(1)
		go func(id string, pcfg config.PluginConfig) {
			defer wg.Done()

			// Evaluate only_if guard if present
			if pcfg.OnlyIf != "" && !shouldRun(ctxObj, pcfg.OnlyIf) {
				return
			}

			// Build a cache key early (project/current dir, and optionally a context field)
			defaultKey := buildDefaultCacheKey(id, ctxObj)
			key := defaultKey
			if pcfg.CacheKeyFrom != "" {
				if v, ok := lookupPath(ctxObj, pcfg.CacheKeyFrom); ok {
					key = id + "|" + stringify(v)
				}
			}

			// Check cache first
			if seg, found := getCached(key); found {
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
			// Make cfg available to builtins via context
			pluginCtx = context.WithValue(pluginCtx, types.CtxKeyConfig, cfg)

			var seg types.Segment
			if pcfg.Type == "builtin" {
				seg = runBuiltin(pluginCtx, id, ctxObj)
			} else if pcfg.Type == "exec" {
				seg = runExec(pluginCtx, pcfg, claudeJSON)
			}

			seg.ID = id
			if seg.Priority == 0 {
				seg.Priority = 50 // default priority
			}

			// Allow plugin response to refine the key after run
			if seg.CacheKey != "" {
				key = id + "|" + seg.CacheKey
			}
			ttl := pcfg.CacheTTLMS
			if seg.CacheTTLMS > 0 {
				ttl = seg.CacheTTLMS
			}
			if ttl > 0 {
				setCached(key, seg, time.Duration(ttl)*time.Millisecond)
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

func runExec(ctx context.Context, pcfg config.PluginConfig, claudeJSON []byte) types.Segment {
	cmd := exec.CommandContext(ctx, pcfg.Command, pcfg.Args...)
	cmd.Stdin = bytes.NewReader(claudeJSON)
	// Limit stdout to 4KB
	var buf bytes.Buffer
	cmd.Stdout = &limitedWriter{w: &buf, n: 4096}
	cmd.Stderr = io.Discard
	err := cmd.Run()
	if err != nil {
		return types.Segment{} // silent failure
	}

	raw := strings.TrimSpace(buf.String())
	if raw == "" {
		return types.Segment{}
	}
	// Keep only the first line; ignore trailing noise or accidental logs.
	result := raw
	if i := strings.IndexByte(raw, '\n'); i >= 0 {
		result = raw[:i]
	}

	// Try to parse as JSON first
	var resp types.PluginResponse
	if err := json.Unmarshal([]byte(result), &resp); err == nil {
		return types.Segment{
			Text:       resp.Text,
			Style:      resp.Style,
			Align:      resp.Align,
			Priority:   resp.Priority,
			CacheTTLMS: resp.CacheTTLMS,
			CacheKey:   resp.CacheKey,
		}
	} else {
		// Otherwise treat as plain text
		return types.Segment{
			Text: result,
		}
	}
}

// buildDefaultCacheKey creates a cache key with workspace context
func buildDefaultCacheKey(id string, ctxObj map[string]any) string {
	var pd, cd string
	if ws, ok := ctxObj["workspace"].(map[string]any); ok {
		if s, ok := ws["project_dir"].(string); ok {
			pd = s
		}
		if s, ok := ws["current_dir"].(string); ok {
			cd = s
		}
	}
	return id + "|" + pd + "|" + cd
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

type limitedWriter struct {
	w       io.Writer
	n       int64
	written int64
}

func (l *limitedWriter) Write(p []byte) (int, error) {
	// Remaining capacity to store
	remain := l.n - l.written
	if remain <= 0 {
		// Discard but *pretend* we consumed everything to keep draining.
		l.written += int64(len(p))
		return len(p), nil
	}
	if int64(len(p)) <= remain {
		n, err := l.w.Write(p)
		l.written += int64(n)
		return n, err
	}
	// Write up to remain, then discard the rest but report full consumption.
	n, err := l.w.Write(p[:remain])
	l.written += int64(n)
	return len(p), err
}
