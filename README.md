# ccsl

Fast status line for Claude Code.

```
Opus 4.5 42% $1.23 myproject:main*⇡2≡ gcp:prod@default
```

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/hergert/ccsl/main/scripts/install.sh | bash
```

Requires Go. Restart Claude Code after install.

## Uninstall

```bash
curl -fsSL https://raw.githubusercontent.com/hergert/ccsl/main/scripts/uninstall.sh | bash
```

## Segments

| ID | Shows |
|----|-------|
| `model` | Model name (bold) |
| `cwd` | Current directory |
| `git` | branch*⇡N⇣N≡ (dirty, ahead, behind, stash) |
| `ctx` | Context % (dims yellow→red at 50%→75%) |
| `cost` | Session cost |
| `gcp` | gcp:project@config |
| `cf` | cf:worker@env |

## Config

`~/.config/ccsl/config.toml`

```toml
[ui]
template = "{model}{ctx?prefix= }{cost?prefix= } {cwd}{git?prefix=:}{gcp?prefix= }{cf?prefix= }"
truncate = 120

[limits]
total_budget_ms = 200
```

## Template

`{segment}` — include segment
`{segment?prefix= }` — add space before, only if segment has content

## Debug

```bash
ccsl doctor
```

## License

MIT
