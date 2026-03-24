# ccsl

Practical status line for Claude Code.

```
Opus 4.6 (1M context) 2% $0.08¹⁶ˢ · 12%⁵ʰ 31%⁷ᵈ  ccsl:main*⇡2≡ gcp:lantern-prod@default
Sonnet 4.6 18% $1.42⁴ʰ²ᵐ · 72%⁵ʰ↻1h48m 15%⁷ᵈ  ccsl:feat/ratelimit⇡1 cf:inbox-worker@staging
```

## Install

Requires Go.

```bash
curl -fsSL https://raw.githubusercontent.com/hergert/ccsl/main/scripts/install.sh | bash
```

The script builds the binary, installs it to `~/.local/bin`, and configures Claude Code's `settings.json`.

## Segments

| ID | Shows |
|----|-------|
| `model` | Model name, e.g. `Opus 4.6 (1M context)` |
| `agent` | Agent name when in a subagent session |
| `worktree` | Worktree name when in a `--worktree` session |
| `ctx` | Context % — yellow→red as usage climbs |
| `cost` | Session cost + superscript duration: `$0.08¹⁶ˢ` |
| `ratelimit` | Rate limit windows: `12%⁵ʰ 31%⁷ᵈ` — yellow at 70%, red at 90%, reset countdown at ≥70% |
| `duration` | Elapsed session time |
| `lines` | Lines changed: `+156-23` |
| `cwd` | Current directory |
| `git` | `branch*⇡N⇣N≡` — dirty, ahead, behind, stash |
| `gcp` | `gcp:project@config` — ⚠ on mismatch |
| `cf` | `cf:worker@env` — ⚠ on mismatch |

## Config

Checked in order: `.claude/ccsl.toml` (project), `~/.claude/ccsl.toml`, `~/.config/ccsl/config.toml`

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
```

**Env overrides:** `CCSL_TEMPLATE`, `CCSL_ORDER`, `CCSL_ANSI=0`, `CCSL_ICONS=0`

Templates use `{segment}` to include a segment and `{segment?prefix= }` to add a space before it only when it has content.

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
