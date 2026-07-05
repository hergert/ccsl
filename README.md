# ccsl

Practical status line for Claude Code.

```
Fable 5 max 2% $0.08¹⁶ˢ · 12%⁵ʰ 31%⁷ᵈ  ccsl:main*⇡2≡ gcp:lantern-prod@default
Opus 4.8 (1M context) 18% $1.42⁴ʰ²ᵐ · 72%⁵ʰ↻1h48m 82%⁷ᵈ↻2d3h  ccsl:feat/ratelimit⇡1 cf:inbox-worker@staging
```

Inside [herdr](https://herdr.dev), ccsl also feeds each agent's sidebar row: `idle · fable 5 · 2% 12%⁵ʰ 31%⁷ᵈ`.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/hergert/ccsl/main/scripts/install.sh | bash
```

Fetches the latest release binary (checksum-verified), installs it to `~/.local/bin`, and points Claude Code's `settings.json` at it, preserving anything else in your `statusLine` block. No Go needed. To build from source instead (requires Go): append `-s -- --build`.

Or Homebrew (macOS):

```bash
brew install --cask hergert/tap/ccsl
```

Brew installs only the binary; add the `statusLine` block yourself or run the installer once.

## Update

`brew upgrade ccsl`, or re-run the installer. `ccsl --version` shows what you have.

## Segments

| ID | Shows |
|----|-------|
| `model` | Model name, e.g. `Opus 4.8 (1M context)` |
| `effort` | Reasoning effort level when set, e.g. `max` |
| `agent` | Agent name when in a subagent session |
| `worktree` | Worktree name (`--worktree` session, or any linked git worktree) |
| `ctx` | Context % — yellow→red as usage climbs |
| `cost` | Session cost + superscript duration: `$0.08¹⁶ˢ` |
| `ratelimit` | Rate limit windows: `12%⁵ʰ 31%⁷ᵈ` — yellow at 70%, red at 90%, reset countdown at ≥70% (`↻1h48m`, `↻2d3h`) |
| `duration` | Elapsed session time |
| `lines` | Lines changed: `+156-23` |
| `cwd` | Current directory |
| `git` | `branch*⇡N⇣N≡` — dirty, ahead, behind, stash |
| `gcp` | `gcp:project@config` — ⚠ on mismatch |
| `cf` | `cf:worker@env` — ⚠ on mismatch |

## Config

Checked in order (first match wins): `.claude/ccsl.toml` (project), `~/.config/ccsl/config.toml`, `~/.claude/ccsl.toml`

**Default template:**
```toml
[ui]
template = "{model}{agent?prefix= }{worktree?prefix= }{ctx?prefix= }{cost?prefix= }{ratelimit?prefix= · } {cwd}{git?prefix=:}{gcp?prefix= }{cf?prefix= }"
```

**Minimal:**
```toml
[ui]
template = "{model} {cwd}{git?prefix=:}"
```

**Other options:**
```toml
[plugin.git]
untracked = true  # include untracked files (slower)

[theme]
ansi = false  # plain text, no colors

[limits]
per_plugin_timeout_ms = 100
total_budget_ms = 200

# external segment: any executable that reads the statusline JSON on stdin
# and prints text (or a segment JSON), see docs/PLUGIN_PROTOCOL.md
[plugin.myseg]
type = "exec"
command = "~/bin/myseg"
```

**Env overrides:** `CCSL_TEMPLATE`, `CCSL_ORDER`, `CCSL_ANSI=0`, `CCSL_HERDR=0`

Templates use `{segment}` to include a segment and `{segment?prefix= }` to add a space before it only when it has content. Lines longer than `truncate` (or the terminal width Claude Code reports via `COLUMNS`, whichever is smaller) trim the lowest-priority segment first.

## herdr

When running inside a [herdr](https://herdr.dev) pane (`HERDR_ENV=1`), ccsl also reports the model and usage to that pane's sidebar row over herdr's local socket (`pane.report_metadata`, display-only). Best-effort with a 6h TTL: herdr down or absent means nothing happens. Disable with `CCSL_HERDR=0`. The installer sets `statusLine.refreshInterval: 60` so idle panes keep their numbers fresh.

## Uninstall

```bash
curl -fsSL https://raw.githubusercontent.com/hergert/ccsl/main/scripts/uninstall.sh | bash
```

## Debug

```bash
ccsl doctor
```

## License

MIT
