# Changelog

All notable changes to ccsl (Claude Code StatusLine) will be documented in this file.

## [Unreleased]

### Changed
- Simplified architecture: removed caching, onlyif conditions, setup wizard
- Segments now derived from template placeholders (no explicit order needed)
- Git uses single `git status --porcelain=v2 --branch` command (~15ms)
- Removed agent, prompt builtins (not provided by Claude Code)
- Removed ccusage integration (cost now built-in)

### Added
- `ctx` builtin: context window usage percentage from Claude's JSON
- `cost` builtin: session cost from Claude's JSON
- ANSI-aware, UTF-8 safe truncation

### Removed
- Caching system (not useful for repeated CLI invocations)
- Setup wizard
- ccsl-ccusage connector
- agent, prompt builtins
- only_if conditions
- cache_ttl_ms, cache_key_from config options

## [0.1.0] - Initial Release

### Added
- Initial release of ccsl - fast Go-based statusline for Claude Code
- Built-in segments: model, cwd, git
- Plugin system with stdin/stdout protocol
- TOML configuration with environment variable overrides
- Parallel plugin execution with timeouts
- ANSI theming support
- Cross-platform installation script
- Example plugins in Python (uv) and Bash

### Architecture
- **Core**: Go binary for fast startup and low memory usage
- **Plugins**: External executables following simple contract
- **Config**: TOML-based with smart defaults
- **Safety**: Hard timeouts, graceful degradation, no shells
- **Performance**: <200ms total budget with per-plugin limits
