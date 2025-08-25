package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type out struct {
	Text       string `json:"text"`
	Style      string `json:"style,omitempty"`
	Priority   int    `json:"priority,omitempty"`
	CacheTTLMS int    `json:"cache_ttl_ms,omitempty"`
	CacheKey   string `json:"cache_key,omitempty"`
}

func main() {
	// Read the Claude Code statusline JSON from stdin (pass through to ccusage)
	raw, _ := io.ReadAll(os.Stdin)

	// Extra, tight timeout on the connector
	timeoutMS := getenvInt("CCSL_CCUSAGE_TIMEOUT_MS", 250)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMS)*time.Millisecond)
	defer cancel()

	// Choose runner: bun x ccusage -> npx -y ccusage -> ccusage
	var cmd *exec.Cmd
	if has("bun") {
		cmd = exec.CommandContext(ctx, "bun", append([]string{"x", "ccusage", "statusline"}, extraArgs()...)...)
	} else if has("npx") {
		cmd = exec.CommandContext(ctx, "npx", append([]string{"-y", "ccusage", "statusline"}, extraArgs()...)...)
	} else if has("ccusage") {
		cmd = exec.CommandContext(ctx, "ccusage", append([]string{"statusline"}, extraArgs()...)...)
	} else {
		// Silent failure per ccsl plugin contract
		return
	}

	// Environment: offline & quiet by default; color in sync with ccsl theme
	env := os.Environ()
	env = setIfMissing(env, "LOG_LEVEL", "0")
	// Offline by default; users can override with CCSL_CCUSAGE_ONLINE=1
	if os.Getenv("CCSL_CCUSAGE_ONLINE") != "1" {
		env = setIfMissing(env, "CCUSAGE_OFFLINE", "1")
	}
	// Respect ccsl's ANSI setting: CCSL_ANSI=0 -> disable colors
	if v := os.Getenv("CCSL_ANSI"); v == "0" || strings.EqualFold(v, "false") {
		env = setIfMissing(env, "NO_COLOR", "1")
	}
	cmd.Env = env

	cmd.Stdin = bytes.NewReader(raw)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		return // silent
	}

	text := strings.TrimSpace(buf.String())
	if text == "" {
		return
	}

	// Return ccsl plugin JSON wrapper with caching hints
	ttl := getenvInt("CCSL_CCUSAGE_TTL_MS", 1500)
	key := "" // optionally scope by transcript_path if desired
	// We pass through color (if any), so don't add an extra style
	resp := out{Text: text, Priority: 55, CacheTTLMS: ttl, CacheKey: key}

	enc := json.NewEncoder(os.Stdout)
	_ = enc.Encode(resp)
}

func has(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

func getenvInt(k string, def int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	return def
}

func setIfMissing(env []string, key, val string) []string {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return env
		}
	}
	return append(env, prefix+val)
}

func extraArgs() []string {
	// Map simple ccsl envs to ccusage flags.
	var args []string

	if v := os.Getenv("CCSL_CCUSAGE_COST_SOURCE"); v != "" {
		// auto (default) | ccusage | cc | both
		args = append(args, "--cost-source", v)
	}
	if v := os.Getenv("CCSL_CCUSAGE_BURN"); v != "" {
		// off | emoji | text | emoji-text
		args = append(args, "--visual-burn-rate", v)
	}
	// Optional thresholds (integers 0–100)
	if v := os.Getenv("CCSL_CCUSAGE_CTX_LOW"); v != "" {
		args = append(args, "--context-low-threshold", v)
	}
	if v := os.Getenv("CCSL_CCUSAGE_CTX_MED"); v != "" {
		args = append(args, "--context-medium-threshold", v)
	}
	// Optional token-limit for 5h block % (number or "max")
	if v := os.Getenv("CCSL_CCUSAGE_TOKEN_LIMIT"); v != "" {
		args = append(args, "--token-limit", v)
	}
	// Users can pass through raw flags if needed
	if v := os.Getenv("CCSL_CCUSAGE_FLAGS"); v != "" {
		parts := strings.Fields(v)
		args = append(args, parts...)
	}
	return args
}
