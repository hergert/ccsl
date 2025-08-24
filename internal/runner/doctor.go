// internal/runner/doctor.go
package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"ccsl/internal/config"
	"ccsl/internal/types"
)

type PluginDiag struct {
	ID         string
	Kind       string // "builtin" | "exec"
	TimeoutMS  int
	CacheHit   bool
	Skipped    bool   // only_if prevented run
	Ran        bool   // actually spawned/executed (or builtin executed)
	DurationMS int64  // ms spent (0 when cache hit or skipped)
	Error      string // "timeout" or trimmed error message
}

// CollectWithDiag mirrors Collect but also returns per-plugin diagnostics.
func CollectWithDiag(ctx context.Context, claudeJSON []byte, cfg *config.Config) ([]types.Segment, []PluginDiag) {
	var ctxObj map[string]any
	_ = json.Unmarshal(claudeJSON, &ctxObj)

	var (
		wg     sync.WaitGroup
		segCh  = make(chan types.Segment, len(cfg.Plugins.Order))
		diagCh = make(chan PluginDiag, len(cfg.Plugins.Order))
	)

	for _, pluginID := range cfg.Plugins.Order {
		pluginCfg, exists := cfg.Plugin[pluginID]
		if !exists {
			// Unknown plugin in order; record as skipped for clarity.
			diagCh <- PluginDiag{ID: pluginID, Kind: "unknown", Skipped: true}
			continue
		}

		wg.Add(1)
		go func(id string, pcfg config.PluginConfig) {
			defer wg.Done()

			diag := PluginDiag{
				ID:        id,
				Kind:      pcfg.Type,
				TimeoutMS: effTimeoutMS(pcfg, cfg),
			}

			// only_if expression check
			if pcfg.OnlyIf != "" && !shouldRun(ctxObj, pcfg.OnlyIf) {
				diag.Skipped = true
				diagCh <- diag
				return
			}

			// Cache check (same keying as normal Collect)
			defaultKey := buildDefaultCacheKey(id, ctxObj)
			if seg, hit := getCached(defaultKey); hit {
				diag.CacheHit = true
				segCh <- seg
				diagCh <- diag
				return
			}

			timeout := time.Duration(diag.TimeoutMS) * time.Millisecond
			pctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			// Make cfg available to builtins via context (future use)
			pctx = context.WithValue(pctx, types.CtxKeyConfig, cfg)

			start := time.Now()
			var seg types.Segment
			var err error

			switch pcfg.Type {
			case "builtin":
				diag.Ran = true
				seg = runBuiltin(pctx, id, ctxObj)
			case "exec":
				diag.Ran = true
				seg, err = runExecWithErr(pctx, pcfg, claudeJSON)
			default:
				diag.Skipped = true
			}

			diag.DurationMS = time.Since(start).Milliseconds()

			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) || pctx.Err() == context.DeadlineExceeded {
					diag.Error = "timeout"
				} else {
					// trim noisy errors
					msg := err.Error()
					if len(msg) > 200 {
						msg = msg[:200] + "..."
					}
					diag.Error = msg
				}
			}

			seg.ID = id
			if seg.Priority == 0 {
				seg.Priority = 50
			}

			// Cache the result (same precedence: plugin response TTL > config)
			key := defaultKey
			if seg.CacheKey != "" {
				key = id + "|" + seg.CacheKey
			}
			ttl := pcfg.CacheTTLMS
			if seg.CacheTTLMS > 0 {
				ttl = seg.CacheTTLMS
			}
			if ttl > 0 && seg.Text != "" {
				setCached(key, seg, time.Duration(ttl)*time.Millisecond)
			}

			if seg.Text != "" {
				segCh <- seg
			}
			diagCh <- diag
		}(pluginID, pluginCfg)
	}

	// Drain
	go func() {
		wg.Wait()
		close(segCh)
		close(diagCh)
	}()

	var segs []types.Segment
	for s := range segCh {
		segs = append(segs, s)
	}

	var diags []PluginDiag
	for d := range diagCh {
		diags = append(diags, d)
	}

	// Order diags according to config.plugins.order for stable output
	idx := make(map[string]int, len(cfg.Plugins.Order))
	for i, id := range cfg.Plugins.Order {
		idx[id] = i
	}
	sort.SliceStable(diags, func(i, j int) bool {
		return idx[diags[i].ID] < idx[diags[j].ID]
	})

	return segs, diags
}

func effTimeoutMS(pcfg config.PluginConfig, cfg *config.Config) int {
	if pcfg.TimeoutMS > 0 {
		return pcfg.TimeoutMS
	}
	return cfg.Limits.PerPluginTimeoutMS
}

// Copy of runExec with error surfaced for diagnostics; keeps stdout size bounded.
func runExecWithErr(ctx context.Context, pcfg config.PluginConfig, claudeJSON []byte) (types.Segment, error) {
	cmd := exec.CommandContext(ctx, pcfg.Command, pcfg.Args...)
	cmd.Stdin = strings.NewReader(string(claudeJSON))
	var buf bytes.Buffer
	cmd.Stdout = &limitedWriter{w: &buf, n: 4096}
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		return types.Segment{}, err
	}

	result := strings.TrimSpace(buf.String())
	if result == "" {
		return types.Segment{}, nil
	}

	var resp types.PluginResponse
	if err := json.Unmarshal([]byte(result), &resp); err == nil {
		return types.Segment{
			Text:       resp.Text,
			Style:      resp.Style,
			Align:      resp.Align,
			Priority:   resp.Priority,
			CacheTTLMS: resp.CacheTTLMS,
			CacheKey:   resp.CacheKey,
		}, nil
	}
	return types.Segment{Text: result}, nil
}
