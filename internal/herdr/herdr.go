package herdr

import (
	"encoding/json"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	ctxbuiltin "github.com/hergert/ccsl/builtin/ctx"
	"github.com/hergert/ccsl/builtin/model"
	"github.com/hergert/ccsl/builtin/ratelimit"
)

const (
	// herdr normalizes custom_status to 32 chars; pre-trim so oversized input
	// degrades instead of being rejected.
	maxStatusRunes = 32
	// Backstop so a dead session's numbers age out of the sidebar.
	ttlMS = 6 * 60 * 60 * 1000

	dialTimeout = 150 * time.Millisecond
	ioDeadline  = 250 * time.Millisecond
)

// Status derives the sidebar strings from the statusline JSON: compact usage
// ("10% 18%⁵ʰ 23%⁷ᵈ") and the lowercased model name for the agent label slot.
func Status(raw map[string]any) (customStatus, displayAgent string) {
	var parts []string
	if c, ok := ctxbuiltin.Parse(raw); ok {
		parts = append(parts, c.Render().Text)
	}
	if l, ok := ratelimit.Parse(raw); ok {
		parts = append(parts, l.Render(false).Text)
	}
	return capRunes(strings.Join(parts, " "), maxStatusRunes), compactModel(model.Parse(raw).DisplayName)
}

func compactModel(name string) string {
	if i := strings.IndexByte(name, '('); i >= 0 {
		name = name[:i]
	}
	return strings.ToLower(strings.TrimSpace(name))
}

func capRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

// Report sends pane metadata to the herdr server over its local socket.
// Display-only (pane.report_metadata carries no lifecycle authority) and
// best-effort: silent no-op outside herdr, on CCSL_HERDR=0, or on any error.
func Report(raw map[string]any) {
	if os.Getenv("HERDR_ENV") != "1" || os.Getenv("CCSL_HERDR") == "0" {
		return
	}
	paneID := os.Getenv("HERDR_PANE_ID")
	socketPath := os.Getenv("HERDR_SOCKET_PATH")
	if paneID == "" || socketPath == "" {
		return
	}

	customStatus, displayAgent := Status(raw)
	if customStatus == "" && displayAgent == "" {
		return
	}

	seq := time.Now().UnixNano()
	params := map[string]any{
		"pane_id": paneID,
		"source":  "ccsl",
		"seq":     seq,
		"ttl_ms":  ttlMS,
	}
	if customStatus != "" {
		params["custom_status"] = customStatus
	}
	if displayAgent != "" {
		params["display_agent"] = displayAgent
	}
	req, err := json.Marshal(map[string]any{
		"id":     "ccsl:" + strconv.FormatInt(seq, 10),
		"method": "pane.report_metadata",
		"params": params,
	})
	if err != nil {
		return
	}

	conn, err := net.DialTimeout("unix", socketPath, dialTimeout)
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(ioDeadline))
	if _, err := conn.Write(append(req, '\n')); err != nil {
		return
	}
	buf := make([]byte, 512)
	_, _ = conn.Read(buf)
}
