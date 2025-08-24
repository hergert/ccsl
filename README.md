# ccsl — Claude Code status line

`ccsl` renders a fast, configurable status line in Claude Code. It reads a small JSON blob from Claude on stdin and prints a single line on stdout.

- Minimal defaults: model, folder, agent, git, last prompt
- Config via `~/.config/ccsl/config.toml` and a few env vars
- Plugins: small executables on your PATH

## Requirements

- macOS or Linux
- Go (only if building from source)
- Claude Code

## Install

One‑liner (builds from source and wires Claude settings):

```bash
curl -fsSL https://raw.githubusercontent.com/hergert/ccsl/main/scripts/install.sh | bash
```

Or manually:

```bash
git clone https://github.com/hergert/ccsl.git
cd ccsl && ./scripts/install.sh
```

The installer places the binary at `~/.local/bin/ccsl` and updates `~/.claude/settings.json` to use it for `statusLine`. Restart Claude Code after install.

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

Run `ccsl` with a small fixture:

```bash
echo '{"model":{"display_name":"Test"},"workspace":{"current_dir":"'$PWD'","project_dir":"'$PWD'"}}' | ccsl
```

You should see a single rendered line on stdout.

## Configure (quick)

Location: `~/.config/ccsl/config.toml`. A sensible default is created at install.

Minimal example:

```toml
[ui]
# Use conditional prefixes so gaps disappear when a segment is empty.
template = "{model}  {cwd}{agent?prefix=  }{git?prefix=  }{prompt?prefix= — 🗣 }"
truncate = 120

[theme]
mode = "auto"     # auto | light | dark
icons = true
ansi  = true

[plugins]
order = ["model", "cwd", "agent", "git", "prompt"]

# Builtins (timeouts are ms; cache_ttl_ms is a hint for internal caching)
[plugin.model]  type="builtin"  timeout_ms=10
[plugin.cwd]    type="builtin"  timeout_ms=10
[plugin.agent]  type="builtin"  timeout_ms=20  cache_ttl_ms=100
[plugin.git]    type="builtin"  timeout_ms=90  cache_ttl_ms=300  style="dim"
[plugin.prompt] type="builtin"  timeout_ms=30  cache_ttl_ms=50

[limits]
per_plugin_timeout_ms = 120
total_budget_ms       = 220
```

## Env overrides

These can be set in your shell, or in Claude's settings env (project or user scope).

- `CCSL_TEMPLATE` – override the entire template string
- `CCSL_ORDER` – comma list of plugin IDs (e.g., `model,cwd,git,prompt`)
- `CCSL_DISABLE` – comma list to remove from order (e.g., `prompt,git`)
- `CCSL_THEME` – `auto|light|dark`
- `CCSL_ANSI` – `0/1`
- `CCSL_ICONS` – `0/1`
- `CCSL_PROMPT_MAX` – max characters for the prompt segment

Example (shell):

```bash
export CCSL_TEMPLATE='{model}  {cwd}{git?prefix=  }'
export CCSL_PROMPT_MAX=80
```

## Template syntax

Use `{segment}` placeholders. Add conditional glue with `?prefix=` / `?suffix=` (only applied when the segment exists).

```toml
# shows " — 🗣 …" only when prompt segment is non‑empty
template = "{model}  {cwd}{git?prefix=  }{prompt?prefix= — 🗣 }"
```

## Plugins (overview)

- Define in config (`[plugins].order` and `[plugin.NAME]` blocks)
- Two kinds: `builtin` and `exec`
- `exec` plugins are external executables; ccsl runs them without a shell: `command` plus `args` array
- Each plugin receives Claude's JSON on stdin and must print a single line to stdout
- Output can be plain text or a small JSON object (see [Plugin Guide](docs/PLUGIN_PROTOCOL.md))

Add a plugin:

```toml
[plugins]
order = ["model", "cwd", "agent", "git", "uptime", "prompt"]

[plugin.uptime]
type        = "exec"
command     = "ccsl-uptime"  # must be on PATH and executable
timeout_ms  = 80
cache_ttl_ms= 500
only_if     = ""             # optional guard, see guide
```

See `plugins/bash/ccsl-uptime` and `plugins/python/ccsl-cost` for working examples.

## Troubleshooting

- **Nothing shows**: pipe valid JSON; ccsl prints nothing on JSON parse errors.
- **Gaps in output**: make sure your template uses conditional prefixes (see default).
- **Plugin doesn't run**: confirm it's listed in `[plugins].order`, executable, and on `$PATH`.
- **Slow line**: increase `cache_ttl_ms` for expensive plugins and check timeouts.
- **Git looks stale**: your repo may have no upstream or the timeout is too tight; raise `[plugin.git].timeout_ms`.

## Diagnostics

Run `ccsl doctor` to debug your setup:

```bash
# With a fixture JSON
ccsl doctor -json ~/.claude/last_session.json

# Or use the built‑in minimal fixture
ccsl doctor
```

This prints your active template, plugin order, timeouts, per‑plugin timings
(skipped/cache/ran/error), and the final rendered line. Use it to spot slow or
misconfigured plugins quickly.

## Notes

- Exec plugin stdout is capped (~4 KiB). Stderr is discarded.
- Caching is scoped by plugin id + project/current directory. `cache_ttl_ms` controls duration.  
  (Plugins may include a `cache_key` in JSON; it's safe to emit but ccsl may ignore it.)
- `agent` segment is hidden when it equals "main".

## License

MIT
