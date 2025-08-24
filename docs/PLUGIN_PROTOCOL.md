# ccsl Plugin Guide

Small executables that read JSON on **stdin** and write a single line on **stdout**.

- `exec` plugins are declared in config and run in parallel with builtins
- `ccsl` executes them directly (no shell), with optional `args`
- Time budget is tight; default timeout is 120 ms per plugin

---

## Input (stdin)

Claude's statusLine JSON. Fields vary by version; expect keys like:

```json
{
  "model": { "id": "claude-3-5-sonnet-20241022", "display_name": "Sonnet 3.5" },
  "workspace": {
    "current_dir": "/path/project",
    "project_dir": "/path/project"
  },
  "transcript_path": "/path/to/session.jsonl",
  "cost": { "total_cost_usd": 0.025, "total_duration_ms": 1500 }
}
```

Always consume stdin (even if you ignore the data) to avoid blocking.

## Output (stdout)

**Option A — Plain text**

Just print the segment text:

```bash
echo "🔋 85%"
```

**Option B — JSON object**

Return a small JSON with formatting and hints:

```json
{
  "text": "⏱ 2m17s",
  "style": "dim",
  "align": "left",
  "priority": 40,
  "cache_ttl_ms": 1500,
  "cache_key": "optional-scope-key"
}
```

### Fields

- **text** (required): string to display
- **style** (optional): `"normal"` | `"bold"` | `"dim"` (or raw ANSI; ccsl will wrap)
- **align** (optional): `"left"` | `"right"` (reserved for future layout)
- **priority** (optional): integer; higher numbers are kept longer when truncating
- **cache_ttl_ms** (optional): cache duration hint
- **cache_key** (optional): scope hint (safe to include; ccsl may ignore)

## Config reference (`~/.config/ccsl/config.toml`)

Declare plugins under `[plugins].order` and configure per plugin:

```toml
[plugins]
order = ["model", "cwd", "git", "cost", "uptime", "prompt"]

[plugin.cost]
type = "exec"
command = "ccsl-cost"
args = []                # optional; no shell is used
timeout_ms = 80
cache_ttl_ms = 1500
only_if = "has(cost.total_cost_usd)"  # see expressions below

[plugin.uptime]
type = "exec"
command = "ccsl-uptime"
timeout_ms = 80
cache_ttl_ms = 500
```

## Execution contract

- **Time budget**: complete before your `timeout_ms` (default 120 ms).
- **Silent failure**: errors/timeouts are skipped; ccsl continues.
- **Stdout limit**: ~4 KiB, first line is used.
- **No shells**: `command` + `args` are executed directly.
- **Read stdin**: always consume it.

## Only‑if expressions

Skip your plugin unless a condition is met. Supported forms:

- `has(a.b.c)` — path exists
- `eq(a.b, "value")` / `ne(a.b, "value")` — equality/inequality (numbers/bools may be unquoted)
- `a.b.c` — truthy check (non‑empty, not "0", not "false")

Examples:

```toml
[plugin.cost]      only_if = "has(cost.total_cost_usd)"
[plugin.telemetry] only_if = "eq(model.display_name, \"Sonnet 3.5\")"
[plugin.debug]     only_if = "workspace.project_dir"
```

## Examples

### Python (with uv inline script)

```python
#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
import json, sys

data = json.load(sys.stdin)
cost = (data.get("cost") or {}).get("total_cost_usd", 0.0)
dur  = (data.get("cost") or {}).get("total_duration_ms", 0)

if cost > 0:
    text = f"${cost:.3f}" + (f" | {int(dur/1000)}s" if dur else "")
    print(json.dumps({"text": text, "style": "dim", "priority": 40, "cache_ttl_ms": 1500}))
```

Make it executable and on PATH:

```bash
chmod +x ~/.local/bin/ccsl-cost
```

Add to config:

```toml
[plugins]        order = ["model","cwd","git","cost","prompt"]
[plugin.cost]    type="exec" command="ccsl-cost" timeout_ms=80 cache_ttl_ms=1500 only_if="has(cost.total_cost_usd)"
```

### Bash

```bash
#!/bin/bash
# ccsl-uptime
# Consume stdin to avoid blocking
cat >/dev/null
if command -v uptime >/dev/null 2>&1; then
  u=$(uptime | sed -E 's/.*up[[:space:]]+([^,]+).*/\1/' | tr -d ' ')
  echo "{\"text\":\"⏰ $u\",\"style\":\"dim\",\"priority\":30}"
fi
```

## Best practices

- **Finish in <100 ms**; prefer local state over network calls.
- **Always handle missing/malformed fields**.
- **Use `cache_ttl_ms`** for expensive computations.
- **Don't write logs to stdout**; stderr is dropped as well—debug locally with a file or run plugin standalone:

```bash
echo '{"model":{"display_name":"Test"}}' | ccsl-myplugin
```

That's it—small programs in, small strings out.
