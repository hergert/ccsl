# ccusage Integration

The `ccsl-ccusage` connector brings real-time usage metrics to your Claude Code status line, powered by [ccusage](https://ccusage.com).

## What it shows

A comprehensive usage segment with:

- Model and session cost
- Today's total cost
- Current 5-hour block cost and time remaining
- Burn rate (cost per hour)
- Context usage (percentage of context window)

Example output:

```
Opus | $0.23 session / $1.23 today / $0.45 block (2h 45m left) | $0.12/hr | 25,000 (12%)
```

## Quick setup

1. **Prerequisites**: Install `bun` (recommended) or `node` for npx

2. **Enable in config** (`~/.config/ccsl/config.toml`):

```toml
[plugins]
order = ["model", "cwd", "agent", "git", "ccusage", "prompt"]

[plugin.ccusage]
type = "exec"
command = "ccsl-ccusage"
timeout_ms = 250
cache_ttl_ms = 1500
cache_key_from = "transcript_path"  # Optional: cache per session
```

3. **Restart Claude Code**

## Environment variables

| Variable                   | Default | Description                                               |
| -------------------------- | ------- | --------------------------------------------------------- |
| `CCSL_CCUSAGE_ONLINE`      | `0`     | Set to `1` to fetch latest pricing online                 |
| `CCSL_CCUSAGE_COST_SOURCE` | `auto`  | Cost source: `auto`, `ccusage`, `cc`, or `both`           |
| `CCSL_CCUSAGE_BURN`        | `off`   | Visual burn rate: `off`, `emoji`, `text`, or `emoji-text` |
| `CCSL_CCUSAGE_CTX_LOW`     | -       | Low context threshold (0-100)                             |
| `CCSL_CCUSAGE_CTX_MED`     | -       | Medium context threshold (0-100)                          |
| `CCSL_CCUSAGE_TOKEN_LIMIT` | -       | Block token limit (number or `max` for historical)        |
| `CCSL_CCUSAGE_TIMEOUT_MS`  | `250`   | Connector timeout in milliseconds                         |
| `CCSL_CCUSAGE_TTL_MS`      | `1500`  | Cache TTL in milliseconds                                 |
| `CCSL_CCUSAGE_FLAGS`       | -       | Raw flags to pass to ccusage                              |
| `CLAUDE_CONFIG_DIR`        | -       | Custom Claude config path(s), comma-separated             |

## How it works

1. The `ccsl-ccusage` connector receives Claude's JSON payload from stdin
2. It locates a JS runtime (`bun x`, `npx -y`, or direct `ccusage`)
3. Runs `ccusage statusline` with your configured options
4. Returns the formatted output as a ccsl segment

The connector:

- Defaults to offline mode for speed (no network requests)
- Respects ccsl's ANSI color settings
- Has a 250ms timeout safeguard
- Caches results for 1.5 seconds to avoid process spam

## Examples

### Show both Claude Code and ccusage costs

```bash
export CCSL_CCUSAGE_COST_SOURCE=both
```

### Add burn rate indicators

```bash
export CCSL_CCUSAGE_BURN=emoji
```

### Track block usage percentage

```bash
# Against a fixed limit
export CCSL_CCUSAGE_TOKEN_LIMIT=500000

# Against your historical maximum
export CCSL_CCUSAGE_TOKEN_LIMIT=max
```

### Force monochrome output

```bash
export CCSL_ANSI=0
```

## Credits & License

Usage metrics are provided by **[ccusage](https://ccusage.com)** (MIT License), © 2024 [ryoppippi](https://github.com/ryoppippi).

See the [ccusage repository](https://github.com/ryoppippi/ccusage) for detailed documentation and advanced configuration options.
