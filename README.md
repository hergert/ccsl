# ccsl: Claude Code StatusLine

A fast, extensible statusline for Claude Code with a clean plugin architecture.

## Features

- **⚡ Fast**: Go binary with sub-200ms rendering
- **🔧 Extensible**: Plugin system for custom segments
- **🎨 Themeable**: ANSI styling with auto/light/dark modes
- **⚙️ Configurable**: TOML config with environment overrides
- **🔒 Safe**: Timeouts, budgets, graceful degradation

## Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/hergert/ccsl/main/scripts/install.sh | bash
```

Or clone and install:

```bash
git clone https://github.com/hergert/ccsl.git ~/ccsl
cd ~/ccsl && ./scripts/install.sh
```

## What You Get

Default statusline shows:

- **🤖 Model**: Display name (e.g. "Sonnet 3.5")
- **📁 Directory**: Current working directory name
- **⚙ Agent**: Current subagent or "main"
- **Git Status**: Branch, dirty state, upstream tracking
- **🗣 Last Prompt**: Most recent user message

Example output:

```
🤖 Sonnet 3.5  📁 ccsl  ⚙ main  main* ↑2 — 🗣 Add support for custom plugins
```

## Configuration

Config file: `~/.config/ccsl/config.toml`

```toml
[ui]
template = "{model}  {cwd}  {agent}  {git?prefix=  }{prompt?prefix= — 🗣 }"
truncate = 120

[theme]
mode = "auto"    # auto | light | dark
icons = true
ansi = true

[plugins]
order = ["model", "cwd", "agent", "git", "prompt"]

[plugin.git]
type = "builtin"
style = "dim"
timeout_ms = 90
cache_ttl_ms = 300
```

## Built-in Segments

- **model**: Claude model info from context
- **cwd**: Current working directory basename
- **agent**: Active subagent from `.claude/state.json`
- **git**: Branch, dirty state, upstream (with timeout protection)
- **prompt**: Session-aware last user message

## Plugin System

Create executable scripts named `ccsl-*` anywhere on your PATH:

### Python Plugin (with uv)

```python
#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
import json, sys

data = json.load(sys.stdin)
cost = data.get("cost", {}).get("total_cost_usd", 0)

if cost > 0:
    print(json.dumps({
        "text": f"${cost:.3f}",
        "style": "dim",
        "priority": 40
    }))
```

### Bash Plugin

```bash
#!/bin/bash
cat > /dev/null  # consume stdin
echo '{"text": "🔋 85%", "style": "dim"}'
```

See [docs/PLUGIN_PROTOCOL.md](docs/PLUGIN_PROTOCOL.md) for details.

## Environment Variables

- `CCSL_PROMPT_MAX`: Override prompt length limit
- `CCSL_ANSI`: Enable/disable ANSI colors (0/1)
- `CCSL_ICONS`: Enable/disable emoji icons (0/1)
- `CCSL_TEMPLATE`: Override UI template
- `CCSL_ORDER`: Override plugin order (comma-separated)
- `CCSL_THEME`: Override theme mode (auto/light/dark)
- `CCSL_DISABLE`: Disable specific plugins (comma-separated)

## Performance

- **Total budget**: 220ms max for all plugins
- **Per-plugin timeout**: 120ms default
- **Caching**: Plugin results cached with configurable TTL
- **Parallel execution**: All plugins run concurrently

## Diagnostics

Run `ccsl doctor` to debug your setup:

```bash
# With a fixture JSON
ccsl doctor -json ~/.claude/last_session.json

# Or use the built-in minimal fixture
ccsl doctor
```

This prints your active template, plugin order, timeouts, per-plugin timings
(skipped/cache/ran/error), and the final rendered line. Use it to spot slow or
misconfigured plugins quickly.

## Comparison with cc-inline-bar

| Feature      | cc-inline-bar       | ccsl                    |
| ------------ | ------------------- | ----------------------- |
| Startup      | Python + uv (~50ms) | Go binary (~5ms)        |
| Plugins      | Monolithic script   | External processes      |
| Config       | Hardcoded           | TOML with env overrides |
| Caching      | None                | Per-plugin TTL          |
| Distribution | Script install      | Static binary           |

## Documentation

- [Plugin Protocol](docs/PLUGIN_PROTOCOL.md)
- [Configuration](docs/CONFIG.md)
- [Integration Guide](docs/INTEGRATION.md)

## Uninstall

```bash
curl -fsSL https://raw.githubusercontent.com/hergert/ccsl/main/scripts/uninstall.sh | bash
```

## License

MIT
