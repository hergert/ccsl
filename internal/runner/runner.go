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

	"github.com/hergert/ccsl/builtin/agent"
	"github.com/hergert/ccsl/builtin/cloudflare"
	"github.com/hergert/ccsl/builtin/cost"
	ctxbuiltin "github.com/hergert/ccsl/builtin/ctx"
	"github.com/hergert/ccsl/builtin/cwd"
	"github.com/hergert/ccsl/builtin/duration"
	"github.com/hergert/ccsl/builtin/gcp"
	"github.com/hergert/ccsl/builtin/git"
	"github.com/hergert/ccsl/builtin/lines"
	"github.com/hergert/ccsl/builtin/model"
	"github.com/hergert/ccsl/builtin/ratelimit"
	"github.com/hergert/ccsl/builtin/worktree"
	"github.com/hergert/ccsl/internal/config"
	"github.com/hergert/ccsl/internal/types"
)

const maxPluginStdout = 4096

type limitedWriter struct {
	w io.Writer
	n int
}

func (l *limitedWriter) Write(p []byte) (int, error) {
	if l.n <= 0 {
		return len(p), nil // silently discard
	}
	if len(p) > l.n {
		p = p[:l.n]
	}
	n, err := l.w.Write(p)
	l.n -= n
	return n, err
}

var segmentRe = regexp.MustCompile(`\{([-\w:.]+)`)

func parseSegments(template string) []string {
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

func Collect(ctx context.Context, ctxObj map[string]any, claudeJSON []byte, cfg *config.Config) []types.Segment {

	segmentIDs := cfg.Plugins.Order
	if len(segmentIDs) == 0 {
		segmentIDs = parseSegments(cfg.UI.Template)
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

			var seg types.Segment
			if pcfg.Type == "exec" && pcfg.Command != "" {
				seg = runExec(pctx, pcfg, claudeJSON)
			} else {
				seg = runBuiltin(pctx, id, ctxObj, cfg)
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

func runBuiltin(ctx context.Context, id string, raw map[string]any, cfg *config.Config) types.Segment {
	switch id {
	case "model":
		return model.Parse(raw).Render()
	case "cwd":
		return cwd.Parse(raw).Render()
	case "git":
		if s, ok := git.Collect(ctx, cfg); ok {
			return s.Render(cfg.Theme.ANSI)
		}
	case "cost":
		if s, ok := cost.Parse(raw); ok {
			return s.Render()
		}
	case "ctx":
		if c, ok := ctxbuiltin.Parse(raw); ok {
			return c.Render()
		}
	case "gcp":
		return gcp.Render(raw)
	case "cf", "cloudflare":
		return cloudflare.Render(raw)
	case "agent":
		if a, ok := agent.Parse(raw); ok {
			return a.Render()
		}
	case "duration":
		if d, ok := duration.Parse(raw); ok {
			return d.Render()
		}
	case "ratelimit":
		if l, ok := ratelimit.Parse(raw); ok {
			return l.Render(cfg.Theme.ANSI)
		}
	case "worktree":
		if w, ok := worktree.Parse(raw); ok {
			return w.Render()
		}
	case "lines":
		if c, ok := lines.Parse(raw); ok {
			return c.Render()
		}
	}
	return types.Segment{}
}

func runExec(ctx context.Context, pcfg config.PluginConfig, claudeJSON []byte) types.Segment {
	cmd := exec.CommandContext(ctx, pcfg.Command, pcfg.Args...)
	cmd.Stdin = bytes.NewReader(claudeJSON)
	var buf bytes.Buffer
	lw := &limitedWriter{w: &buf, n: maxPluginStdout}
	cmd.Stdout = lw
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
