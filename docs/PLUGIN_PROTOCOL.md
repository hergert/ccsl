# ccsl Plugin Guide

Small executables that read JSON on **stdin** and write a single line on **stdout**.

- `exec` plugins are declared in config and run in parallel with builtins
- `ccsl` executes them directly (no shell), with optional `args`
- Time budget is tight; default timeout is 100 ms per plugin

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
  "context_window": {
    "used_percentage": 42,
    "total_input_tokens": 10000,
    "total_output_tokens": 5000,
    "context_window_size": 200000
  },
  "cost": { "total_cost_usd": 0.025 }
}
```

Always consume stdin (even if you ignore the data) to avoid blocking.

## Output (stdout)

**Option A — Plain text**

Just print the segment text:

```bash
echo "85%"
```

**Option B — JSON object**

Return a small JSON with formatting hints:

```json
{
  "text": "2m17s",
  "style": "dim",
  "priority": 40
}
```

### Fields

- **text** (required): string to display
- **style** (optional): `"normal"` | `"bold"` | `"dim"` (or raw ANSI)
- **priority** (optional): integer; higher numbers are kept longer when truncating

## Config reference

Declare plugins in config. Segments are derived from template, or specify explicit order:

```toml
[ui]
template = "{model} {cwd}{git?prefix= }{uptime?prefix= }"

[plugin.uptime]
type = "exec"
command = "ccsl-uptime"
args = []                # optional; no shell is used
timeout_ms = 80
```

Or explicit order:

```toml
[plugins]
order = ["model", "cwd", "git", "uptime"]

[plugin.uptime]
type = "exec"
command = "ccsl-uptime"
timeout_ms = 80
```

## Execution contract

- **Time budget**: complete before your `timeout_ms` (default 100 ms).
- **Silent failure**: errors/timeouts are skipped; ccsl continues.
- **Stdout limit**: ~4 KiB, first line is used.
- **No shells**: `command` + `args` are executed directly.
- **Read stdin**: always consume it.

## Examples

### Bash

```bash
#!/bin/bash
# ccsl-uptime
cat >/dev/null  # consume stdin
if command -v uptime >/dev/null 2>&1; then
  u=$(uptime | sed -E 's/.*up[[:space:]]+([^,]+).*/\1/' | tr -d ' ')
  echo "{\"text\":\"$u\",\"style\":\"dim\",\"priority\":30}"
fi
```

### Python

```python
#!/usr/bin/env python3
import json, sys

data = json.load(sys.stdin)
cost = (data.get("cost") or {}).get("total_cost_usd", 0.0)

if cost > 0:
    print(json.dumps({"text": f"${cost:.2f}", "style": "dim", "priority": 40}))
```

Make executable and place on PATH:

```bash
chmod +x ~/.local/bin/ccsl-uptime
```

Add to template:

```toml
[ui]
template = "{model} {cwd}{git?prefix= }{uptime?prefix= }"

[plugin.uptime]
type = "exec"
command = "ccsl-uptime"
timeout_ms = 80
```

## Best practices

- **Finish in <100 ms**; prefer local state over network calls.
- **Always handle missing/malformed fields**.
- **Debug locally** by running plugin standalone:

```bash
echo '{"model":{"display_name":"Test"}}' | ccsl-myplugin
```

That's it—small programs in, small strings out.
