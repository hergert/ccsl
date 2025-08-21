# Changelog

All notable changes to ccsl (Claude Code StatusLine) will be documented in this file.

## [Unreleased]

### Added
- Initial release of ccsl - fast Go-based statusline for Claude Code
- Built-in segments: model, cwd, agent, git, prompt  
- Plugin system with stdin/stdout protocol
- TOML configuration with environment variable overrides
- Session-aware prompt tracking (compatible with cc-inline-bar)
- Parallel plugin execution with timeouts and caching
- ANSI theming with auto/light/dark mode support
- Cross-platform installation script
- Example plugins in Python (uv) and Bash
- Comprehensive documentation and plugin protocol spec

### Architecture
- **Core**: Go binary for fast startup and low memory usage
- **Plugins**: External executables following simple contract
- **Config**: TOML-based with smart defaults
- **Safety**: Hard timeouts, graceful degradation, no shells
- **Performance**: <220ms total budget with per-plugin limits

### Migration from cc-inline-bar
- Drop-in replacement for existing statusLine config
- Compatible with existing hook system
- Maintains session-aware prompt functionality
- Faster execution with plugin parallelization