package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"ccsl/builtin/cost"
	ctxbuiltin "ccsl/builtin/ctx"
	"ccsl/builtin/cwd"
	"ccsl/builtin/git"
	"ccsl/builtin/model"
	"ccsl/internal/config"
	"ccsl/internal/types"
)

var segmentRe = regexp.MustCompile(`\{([-\w:.]+)`)

// ParseSegments extracts segment IDs from a template string
func ParseSegments(template string) []string {
	matches := segmentRe.FindAllStringSubmatch(template, -1)
	seen := make(map[string]bool)
	var result []string
	for _, m := range matches {
		id := m[1]
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	return result
}

// Collect runs segments derived from template in parallel
func Collect(ctx context.Context, claudeJSON []byte, cfg *config.Config) []types.Segment {
	var ctxObj map[string]any
	_ = json.Unmarshal(claudeJSON, &ctxObj)

	// Derive segments from template (or use explicit order if set)
	segmentIDs := cfg.Plugins.Order
	if len(segmentIDs) == 0 {
		segmentIDs = ParseSegments(cfg.UI.Template)
	}

	var wg sync.WaitGroup
	results := make(chan types.Segment, len(segmentIDs))

	for _, id := range segmentIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			pcfg := cfg.Plugin[id]
			timeout := time.Duration(pcfg.TimeoutMS) * time.Millisecond
			if timeout == 0 {
				timeout = time.Duration(cfg.Limits.PerPluginTimeoutMS) * time.Millisecond
			}

			pctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			pctx = context.WithValue(pctx, types.CtxKeyConfig, cfg)

			var seg types.Segment
			if pcfg.Type == "exec" && pcfg.Command != "" {
				seg = runExec(pctx, pcfg, claudeJSON)
			} else {
				seg = runBuiltin(pctx, id, ctxObj)
			}

			seg.ID = id
			if seg.Priority == 0 {
				seg.Priority = 50
			}
			results <- seg
		}(id)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var segments []types.Segment
	for seg := range results {
		if seg.Text != "" {
			segments = append(segments, seg)
		}
	}
	return segments
}

func runBuiltin(ctx context.Context, id string, ctxObj map[string]any) types.Segment {
	switch id {
	case "model":
		return model.Render(ctx, ctxObj)
	case "cwd":
		return cwd.Render(ctx, ctxObj)
	case "git":
		return git.Render(ctx, ctxObj)
	case "cost":
		return cost.Render(ctx, ctxObj)
	case "ctx":
		return ctxbuiltin.Render(ctx, ctxObj)
	default:
		return types.Segment{}
	}
}

func runExec(ctx context.Context, pcfg config.PluginConfig, claudeJSON []byte) types.Segment {
	cmd := exec.CommandContext(ctx, pcfg.Command, pcfg.Args...)
	cmd.Stdin = bytes.NewReader(claudeJSON)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		return types.Segment{}
	}

	raw := strings.TrimSpace(buf.String())
	if raw == "" {
		return types.Segment{}
	}
	if i := strings.IndexByte(raw, '\n'); i >= 0 {
		raw = raw[:i]
	}

	var resp types.Segment
	if json.Unmarshal([]byte(raw), &resp) == nil && resp.Text != "" {
		return resp
	}
	return types.Segment{Text: raw}
}
