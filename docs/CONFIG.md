# ccsl Configuration

## Config File Location

ccsl looks for configuration in these locations (in order):

1. `~/.config/ccsl/config.toml`
2. `~/.claude/ccsl.toml`

## Full Configuration Example

```toml
[ui]
template = "{model}  {cwd}{agent?prefix=  }{git?prefix=  }{prompt?prefix= — 🗣 }"
truncate = 120
padding = 0

[theme]
mode = "auto"    # auto | light | dark
icons = true
ansi = true

[plugins]
order = ["model", "cwd", "agent", "git", "prompt"]

# Built-in configurations
[plugin.model]
type = "builtin"
timeout_ms = 10

[plugin.cwd]
type = "builtin"
timeout_ms = 10

[plugin.agent]
type = "builtin"
timeout_ms = 20
cache_ttl_ms = 100

[plugin.git]
type = "builtin"
style = "dim"
timeout_ms = 90
cache_ttl_ms = 300

[plugin.prompt]
type = "builtin"
timeout_ms = 30
cache_ttl_ms = 50

# External plugin example
[plugin.cost]
type = "exec"
command = "ccsl-cost"
timeout_ms = 80
cache_ttl_ms = 500
cache_key_from = "transcript_path"
only_if = "has(cost.total_cost_usd)"

[limits]
per_plugin_timeout_ms = 120
total_budget_ms = 220
```

## Configuration Sections

### [ui]

- **template**: Template string with `{segment}` placeholders
- **truncate**: Maximum line length before truncation
- **padding**: Claude Code statusLine padding

### [theme]

- **mode**: Color theme mode
- **icons**: Enable/disable emoji icons
- **ansi**: Enable/disable ANSI styling

### [plugins]

- **order**: List of plugins in display order

### [plugin.NAME]

- **type**: "builtin" or "exec"
- **command**: Command to execute (for exec type)
- **args**: Command arguments array (for exec type)
- **style**: Default styling ("normal", "bold", "dim")
- **timeout_ms**: Plugin-specific timeout
- **cache_ttl_ms**: Cache duration in milliseconds
- **cache_key_from**: Dot path in context for cache key (e.g., "transcript_path")
- **only_if**: Condition to run plugin (simple expressions)

### [limits]

- **per_plugin_timeout_ms**: Default timeout per plugin
- **total_budget_ms**: Total time budget for all plugins

## Template Syntax

Templates use `{segment}` placeholders with optional modifiers:

```toml
template = "{model}  {cwd}{agent?prefix=  }{git?prefix=  }{prompt?prefix= — 🗣 }"
```

Modifiers:

- `?prefix=TEXT`: Add prefix only if segment exists
- `?suffix=TEXT`: Add suffix only if segment exists

## Environment Variables

Override config with environment variables:

- `CCSL_PROMPT_MAX`: Override prompt truncation length
- `CCSL_ANSI`: Enable/disable ANSI (0/1)
- `CCSL_ICONS`: Enable/disable icons (0/1)

## Profiles (Future)

```toml
[profile.minimal]
plugins = ["model", "cwd"]

[profile.rich]
plugins = ["model", "cwd", "agent", "git", "cost", "uptime", "prompt"]
```
