package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"ccsl/internal/config"
)

func runSetup() {
	fs := flag.NewFlagSet("setup", flag.ExitOnError)
	ask := fs.Bool("ask", false, "interactive prompts (only if TTY)")
	enableCCU := fs.Bool("enable-ccusage", false, "enable ccusage usage segment")
	nonInteractive := fs.Bool("non-interactive", false, "assume defaults; no prompts")
	tokenLimit := fs.String("ccusage-token-limit", "", `token limit for 5h block: "max" or integer; empty to skip`)
	position := fs.String("position", "after:git", `where to insert ccusage: "after:git", "before:prompt", "end"`)
	fs.Parse(os.Args[2:])

	tty := isTTY()
	if !*ask && !*nonInteractive && tty {
		*ask = true
	}

	// Load current config (or defaults if none)
	cfg := config.Load()

	// Ask enablement (or use flags)
	enable := *enableCCU
	if *ask && !*nonInteractive {
		enable = yesNo("Enable ccusage usage segment?", false)
	}

	if !enable {
		fmt.Println("No changes made. You can run `ccsl setup --ask` later.")
		return
	}

	// Ensure plugin block exists
	if cfg.Plugin == nil {
		cfg.Plugin = map[string]config.PluginConfig{}
	}
	pcfg := cfg.Plugin["ccusage"]
	pcfg.Type = "exec"
	pcfg.Command = "ccsl-ccusage"
	if pcfg.TimeoutMS == 0 {
		pcfg.TimeoutMS = 250
	}
	if pcfg.CacheTTLMS == 0 {
		pcfg.CacheTTLMS = 1500
	}
	// Prefer args over env so this is robust in Claude's process env
	// Add token-limit if provided / chosen
	tl := strings.TrimSpace(*tokenLimit)
	if *ask && tl == "" {
		choice := choose("Token limit for 5h block?",
			[]string{"Skip", "Use max from history", "Enter a number"}, 0)
		switch choice {
		case 1:
			tl = "max"
		case 2:
			tl = readLine("Token limit (integer): ")
		}
	}
	if tl != "" {
		pcfg.Args = appendFiltered(pcfg.Args, "--token-limit", tl)
	}
	cfg.Plugin["ccusage"] = pcfg

	// Insert into order if not already there
	cfg.Plugins.Order = ensureInOrder(cfg.Plugins.Order, "ccusage", *position)

	// Write config atomically to ~/.config/ccsl/config.toml
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		xdg = filepath.Join(os.Getenv("HOME"), ".config")
	}
	path := filepath.Join(xdg, "ccsl", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		return
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "config.*.toml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "tempfile: %v\n", err)
		return
	}
	if err := toml.NewEncoder(tmp).Encode(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "encode: %v\n", err)
		tmp.Close()
		os.Remove(tmp.Name())
		return
	}
	tmp.Close()
	if err := os.Rename(tmp.Name(), path); err != nil {
		fmt.Fprintf(os.Stderr, "rename: %v\n", err)
		return
	}

	fmt.Printf("Updated %s\n", path)
	fmt.Println("Done. Restart Claude Code to pick up changes.")
}

func isTTY() bool {
	fi, _ := os.Stdin.Stat()
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func yesNo(prompt string, def bool) bool {
	d := "y/N"
	if def {
		d = "Y/n"
	}
	fmt.Printf("%s [%s]: ", prompt, d)
	in := read()
	s := strings.ToLower(strings.TrimSpace(in))
	if s == "" {
		return def
	}
	return s[0] == 'y'
}

func choose(prompt string, options []string, defIdx int) int {
	fmt.Println(prompt)
	for i, o := range options {
		m := " "
		if i == defIdx {
			m = "*"
		}
		fmt.Printf("  %d) %s %s\n", i, m, o)
	}
	fmt.Printf("Select [default %d]: ", defIdx)
	in := read()
	in = strings.TrimSpace(in)
	if in == "" {
		return defIdx
	}
	for i := range options {
		if fmt.Sprintf("%d", i) == in {
			return i
		}
	}
	return defIdx
}

func readLine(prompt string) string {
	fmt.Print(prompt)
	return strings.TrimSpace(read())
}

func read() string {
	r := bufio.NewReader(os.Stdin)
	s, _ := r.ReadString('\n')
	return s
}

func appendFiltered(args []string, kv ...string) []string {
	// drop any existing keys we set, then append once
	key := kv[0]
	var out []string
	i := 0
	for i < len(args) {
		if args[i] == key {
			i += 2
			continue
		}
		out = append(out, args[i])
		i++
	}
	return append(out, kv...)
}

func ensureInOrder(order []string, id string, position string) []string {
	seen := map[string]bool{}
	var filtered []string
	for _, s := range order {
		if s == id {
			continue
		}
		if !seen[s] {
			filtered = append(filtered, s)
			seen[s] = true
		}
	}
	switch {
	case position == "end":
		return append(filtered, id)
	case strings.HasPrefix(position, "before:"):
		needle := strings.TrimPrefix(position, "before:")
		var out []string
		inserted := false
		for _, s := range filtered {
			if !inserted && s == needle {
				out = append(out, id)
				inserted = true
			}
			out = append(out, s)
		}
		if !inserted {
			out = append(out, id)
		}
		return out
	default: // after:git (default)
		needle := "git"
		if strings.HasPrefix(position, "after:") {
			needle = strings.TrimPrefix(position, "after:")
		}
		var out []string
		inserted := false
		for _, s := range filtered {
			out = append(out, s)
			if !inserted && s == needle {
				out = append(out, id)
				inserted = true
			}
		}
		if !inserted {
			out = append(out, id)
		}
		return out
	}
}
