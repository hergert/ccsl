# ccsl Configuration

## Config File Location

ccsl looks for configuration in order:

1. `~/.config/ccsl/config.toml`
2. `~/.claude/ccsl.toml`

## Full Configuration Example

```toml
[ui]
template = "{model} {cwd}{git?prefix= }{ctx?prefix= }{cost?prefix= }"
truncate = 120

[theme]
icons = true
ansi = true

[plugin.model]
timeout_ms = 10

[plugin.cwd]
timeout_ms = 10

[plugin.git]
timeout_ms = 80

[plugin.ctx]
timeout_ms = 10

[plugin.cost]
timeout_ms = 10

[limits]
per_plugin_timeout_ms = 100
total_budget_ms = 200
```

## Configuration Sections

### [ui]

- **template**: Template string with `{segment}` placeholders. Segments are derived from this.
- **truncate**: Maximum line length before truncation

### [theme]

- **icons**: Enable/disable emoji icons
- **ansi**: Enable/disable ANSI styling

### [plugins]

- **order**: Optional explicit segment order (if omitted, derived from template)

### [plugin.NAME]

- **type**: "builtin" or "exec"
- **command**: Command to execute (for exec type)
- **args**: Command arguments array (for exec type)
- **timeout_ms**: Plugin-specific timeout

### [limits]

- **per_plugin_timeout_ms**: Default timeout per plugin
- **total_budget_ms**: Total time budget for all plugins

## Template Syntax

Templates use `{segment}` placeholders with optional modifiers:

```toml
template = "{model} {cwd}{git?prefix= }{ctx?prefix= }{cost?prefix= }"
```

Modifiers:
- `?prefix=TEXT`: Add prefix only if segment has content
- `?suffix=TEXT`: Add suffix only if segment has content

## Environment Variables

- `CCSL_TEMPLATE`: Override template
- `CCSL_ORDER`: Comma-separated segment order (e.g., `model,cwd,git`)
- `CCSL_ANSI`: Enable/disable ANSI (`0` or `1`)
- `CCSL_ICONS`: Enable/disable icons (`0` or `1`)

## Built-in Segments

| Segment | Description                        |
|---------|------------------------------------|
| model   | Model display name with icon       |
| cwd     | Current directory basename         |
| git     | Branch, dirty flag, ahead/behind   |
| ctx     | Context window usage percentage    |
| cost    | Session cost in USD                |
