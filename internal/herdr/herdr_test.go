package herdr

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
	"unicode/utf8"
)

// macOS caps unix socket paths at 104 bytes; t.TempDir embeds the test name
// and blows past it, so make a short one.
func shortSocket(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "h")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return filepath.Join(dir, "s.sock")
}

func usageJSON() map[string]any {
	return map[string]any{
		"model": map[string]any{"display_name": "Fable 5"},
		"context_window": map[string]any{
			"used_percentage":     10.0,
			"context_window_size": 200000.0,
		},
		"rate_limits": map[string]any{
			"five_hour": map[string]any{"used_percentage": 18.0},
			"seven_day": map[string]any{"used_percentage": 23.0},
		},
	}
}

func TestStatusComposesUsage(t *testing.T) {
	cs, da := Status(usageJSON())
	if cs != "10% 18%⁵ʰ 23%⁷ᵈ" {
		t.Errorf("custom_status = %q, want %q", cs, "10% 18%⁵ʰ 23%⁷ᵈ")
	}
	if da != "fable 5" {
		t.Errorf("display_agent = %q, want %q", da, "fable 5")
	}
}

func TestStatusPartsAreOptional(t *testing.T) {
	raw := usageJSON()
	delete(raw, "rate_limits")
	if cs, _ := Status(raw); cs != "10%" {
		t.Errorf("ctx only: custom_status = %q, want %q", cs, "10%")
	}

	raw = usageJSON()
	delete(raw, "context_window")
	if cs, _ := Status(raw); cs != "18%⁵ʰ 23%⁷ᵈ" {
		t.Errorf("limits only: custom_status = %q, want %q", cs, "18%⁵ʰ 23%⁷ᵈ")
	}

	raw = usageJSON()
	delete(raw, "context_window")
	delete(raw, "rate_limits")
	if cs, _ := Status(raw); cs != "" {
		t.Errorf("no usage: custom_status = %q, want empty", cs)
	}
}

func TestStatusCapsRunawayInput(t *testing.T) {
	raw := map[string]any{
		"context_window": map[string]any{"used_percentage": 123456.0, "context_window_size": 200000.0},
		"rate_limits": map[string]any{
			"five_hour": map[string]any{"used_percentage": 123456.0},
			"seven_day": map[string]any{"used_percentage": 123456.0},
		},
	}
	cs, _ := Status(raw)
	if n := utf8.RuneCountInString(cs); n > 32 {
		t.Errorf("custom_status %d runes > 32: %q", n, cs)
	}
}

func TestModelCompaction(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Opus 4.8 (1M context)", "opus 4.8"},
		{"Fable 5", "fable 5"},
		{"Opus", "opus"},
	}
	for _, tc := range cases {
		raw := map[string]any{"model": map[string]any{"display_name": tc.in}}
		if _, da := Status(raw); da != tc.want {
			t.Errorf("display_agent for %q = %q, want %q", tc.in, da, tc.want)
		}
	}
}

func herdrEnv(t *testing.T, socket string) {
	t.Helper()
	t.Setenv("HERDR_ENV", "1")
	t.Setenv("HERDR_PANE_ID", "w1:p9")
	t.Setenv("HERDR_SOCKET_PATH", socket)
	t.Setenv("CCSL_HERDR", "")
}

func TestReportSendsMetadataOverSocket(t *testing.T) {
	socket := shortSocket(t)
	ln, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	herdrEnv(t, socket)

	got := make(chan map[string]any, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		line, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			return
		}
		var req map[string]any
		if json.Unmarshal(line, &req) == nil {
			got <- req
		}
		_, _ = conn.Write([]byte(`{"id":"x","result":{"type":"ok"}}` + "\n"))
	}()

	Report(usageJSON())

	select {
	case req := <-got:
		if req["method"] != "pane.report_metadata" {
			t.Errorf("method = %v", req["method"])
		}
		params, _ := req["params"].(map[string]any)
		if params["pane_id"] != "w1:p9" || params["source"] != "ccsl" {
			t.Errorf("params = %v", params)
		}
		if params["custom_status"] != "10% 18%⁵ʰ 23%⁷ᵈ" {
			t.Errorf("custom_status = %v", params["custom_status"])
		}
		if params["display_agent"] != "fable 5" {
			t.Errorf("display_agent = %v", params["display_agent"])
		}
		if seq, ok := params["seq"].(float64); !ok || seq <= 0 {
			t.Errorf("seq = %v", params["seq"])
		}
		if params["ttl_ms"] != float64(21600000) {
			t.Errorf("ttl_ms = %v", params["ttl_ms"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no request arrived on socket")
	}
}

func TestReportGates(t *testing.T) {
	socket := shortSocket(t)
	ln, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()

	accepted := make(chan struct{}, 4)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			accepted <- struct{}{}
			_ = conn.Close()
		}
	}()

	assertNoSend := func(name string) {
		t.Helper()
		Report(usageJSON())
		select {
		case <-accepted:
			t.Errorf("%s: unexpected socket connection", name)
		case <-time.After(100 * time.Millisecond):
		}
	}

	herdrEnv(t, socket)
	t.Setenv("HERDR_ENV", "")
	assertNoSend("HERDR_ENV unset")

	herdrEnv(t, socket)
	t.Setenv("HERDR_PANE_ID", "")
	assertNoSend("no pane id")

	herdrEnv(t, socket)
	t.Setenv("CCSL_HERDR", "0")
	assertNoSend("kill switch")

	herdrEnv(t, socket)
	raw := map[string]any{"model": map[string]any{"display_name": "Fable 5"}}
	delete(raw, "context_window")
	Report(raw) // display_agent alone still reports
	select {
	case <-accepted:
	case <-time.After(2 * time.Second):
		t.Error("model-only payload should still report display_agent")
	}
}

func TestReportSurvivesDeadSocket(t *testing.T) {
	herdrEnv(t, shortSocket(t))
	done := make(chan struct{})
	go func() {
		Report(usageJSON())
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("Report hung on dead socket")
	}
}
