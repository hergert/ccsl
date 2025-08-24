# ccsl Plugin Protocol

## Overview

ccsl uses a simple stdin/stdout protocol for plugins. Any executable on your PATH prefixed with `ccsl-` is automatically discovered as a plugin.

## Input

Plugins receive Claude Code's statusLine JSON on stdin. Example:

```json
{
  "model": {
    "id": "claude-3-5-sonnet-20241022",
    "display_name": "Sonnet 3.5"
  },
  "workspace": {
    "current_dir": "/Users/user/project",
    "project_dir": "/Users/user/project"
  },
  "transcript_path": "/path/to/transcript.jsonl",
  "cost": {
    "total_cost_usd": 0.025,
    "total_duration_ms": 1500
  }
}
```

## Output

Plugins can return either:

### Plain Text (Simple)

Just print the segment text and exit 0:

```bash
echo "🔥 hot"
```

### Structured JSON (Advanced)

Return a JSON object with full control:

```json
{
  "text": "🚀 v1.2.3",
  "style": "bold",
  "align": "left",
  "priority": 50,
  "cache_ttl_ms": 800,
  "cache_key": "optional-scope-key"
}
```

### Fields

- **text**: The string to be displayed.
- **style**: "normal", "bold", "dim", or raw ANSI escape codes
- **align**: (Optional) `left` or `right`. Default is `left`. (Right alignment is reserved for future use).
- **priority**: (Optional) A number used for truncation. Higher priority segments are kept. Default is `50`.
- **cache_ttl_ms**: How long ccsl should cache this result
- **cache_key**: Optional string to scope the cache (e.g., include transcript path or current commit)
  If omitted, ccsl scopes by `plugin_id|project_dir|current_dir`.

## Contract Rules

1. **Timeout**: Plugins must complete within their configured timeout (default 120ms)
2. **Silent failure**: If a plugin times out or errors, it's silently skipped
3. **Max output**: Output is limited to prevent runaway segments
4. **No shells**: ccsl executes plugins directly, no shell interpolation
5. **Read stdin**: Always consume stdin to prevent blocking, even if unused

## Discovery

ccsl runs plugins **explicitly listed** in your config under `[plugins].order`.
Executables starting with `ccsl-` on your PATH are **discoverable** (e.g., via `ccsl list`) but are not executed unless listed in the config.

## Examples

### Python with uv

```python
#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
import json, sys

data = json.load(sys.stdin)
# ... process data ...
print(json.dumps({"text": "result", "style": "dim"}))
```

### Bash

```bash
#!/bin/bash
cat > /dev/null  # consume stdin
echo '{"text": "🔋 85%", "priority": 30}'
```

### Go

```go
package main
// ... read from os.Stdin, process, write to os.Stdout
```

## Configuration

Configure plugins in `~/.config/ccsl/config.toml`:

```toml
[plugins]
order = ["model", "cwd", "agent", "git", "cost", "uptime", "prompt"]

[plugin.cost]
type = "exec"
command = "ccsl-cost"
timeout_ms = 80
cache_ttl_ms = 500
only_if = "has(cost.total_cost_usd)"
```

## Testing

Test your plugin:

```bash
echo '{"model":{"display_name":"Test"}}' | ccsl-myplugin
```

## Best Practices

1. **Fast**: Complete in <100ms
2. **Robust**: Handle missing/malformed input gracefully
3. **Cacheable**: Use cache_ttl_ms for expensive operations
4. **Informative**: Use priority to indicate importance
5. **Quiet**: No debug output except on errors
