# ccsl — Claude Code status line

`ccsl` renders a fast, configurable status line for Claude Code. It reads JSON from Claude on stdin and prints a single line on stdout.

- Built-in segments: model, cwd, git, ctx (context %), cost
- Segments derived from template (no explicit ordering needed)
- Config via `~/.config/ccsl/config.toml` and env vars
- External plugins supported via exec

## Requirements

- macOS or Linux
- Go (only if building from source)
- Claude Code

## Install

One-liner (builds from source and wires Claude settings):

```bash
curl -fsSL https://raw.githubusercontent.com/hergert/ccsl/main/scripts/install.sh | bash
```

Or manually:

```bash
git clone https://github.com/hergert/ccsl.git
cd ccsl && ./scripts/install.sh
```

The installer places the binary at `~/.local/bin/ccsl` and updates `~/.claude/settings.json`. Restart Claude Code after install.

## Update

```bash
cd ccsl
git pull
./scripts/install.sh
```

## Uninstall

```bash
curl -fsSL https://raw.githubusercontent.com/hergert/ccsl/main/scripts/uninstall.sh | bash
```

## Verify

```bash
echo '{"model":{"display_name":"Test"},"workspace":{"current_dir":"'$PWD'"}}' | ccsl
```

## Configure

Location: `~/.config/ccsl/config.toml`

Minimal example:

```toml
[ui]
template = "{model} {cwd}{git?prefix= }{ctx?prefix= }{cost?prefix= }"
truncate = 120

[theme]
icons = true
ansi  = true

[plugin.git]
timeout_ms = 80

[limits]
per_plugin_timeout_ms = 100
total_budget_ms       = 200
```

## Built-in Segments

| Segment | Description                                | Source                           |
|---------|--------------------------------------------|---------------------------------|
| model   | Model display name with icon               | `model.display_name`            |
| cwd     | Current directory (basename)               | `workspace.current_dir`         |
| git     | Branch, dirty flag, ahead/behind           | `git status --porcelain=v2`     |
| ctx     | Context window usage percentage            | `context_window.used_percentage`|
| cost    | Session cost in USD                        | `cost.total_cost_usd`           |

## Env Overrides

- `CCSL_TEMPLATE` – override template string
- `CCSL_ORDER` – comma list of segment IDs (overrides template derivation)
- `CCSL_ANSI` – `0` to disable ANSI styling
- `CCSL_ICONS` – `0` to disable icons

Example:

```bash
export CCSL_TEMPLATE='{model} {cwd}{git?prefix= }'
export CCSL_ANSI=0
```

## Template Syntax

Use `{segment}` placeholders with conditional glue:

```toml
template = "{model} {cwd}{git?prefix= }{ctx?prefix= }{cost?prefix= }"
```

- `?prefix=TEXT` – add prefix only if segment has content
- `?suffix=TEXT` – add suffix only if segment has content

## External Plugins

Add custom segments via exec plugins:

```toml
[plugin.uptime]
type       = "exec"
command    = "ccsl-uptime"   # must be on PATH
timeout_ms = 80
```

Plugin receives Claude's JSON on stdin, outputs plain text or JSON:

```json
{"text": "$1.23", "style": "dim", "priority": 40}
```

See `plugins/` for examples.

## Diagnostics

```bash
ccsl doctor
```

Prints template, segments, timing, and rendered output.

## Troubleshooting

- **Nothing shows**: pipe valid JSON; ccsl exits silently on parse errors
- **Gaps in output**: use conditional prefixes in template
- **Git looks stale**: raise `[plugin.git].timeout_ms`

## License

MIT
