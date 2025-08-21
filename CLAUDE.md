# CLAUDE.md

## Commit Guidelines
- **Clear and succinct** - Describe what changed, not attribution
- **No generated-with tags** - Skip Claude Code attribution lines
- **Focus on functionality** - What the change does for users
- **Use bullet points** for multiple changes

## Code Standards  
- **Keep it simple** - Minimal dependencies, clear logic
- **Cross-platform** - Support macOS/Linux/Windows
- **Non-destructive** - Always backup before changes
- **Go conventions** - Follow standard Go formatting and idioms
- **Plugin compatibility** - Maintain stable plugin protocol

## Repository Rules
- **No commits without explicit request**
- **Test before commit** - Run `go test ./...` and `make demo`
- **Clean git history** - Useful commit messages only
- **Performance first** - Keep total budget under 220ms

## Development Workflow
- **Build**: `make build` or `go build ./cmd/ccsl`
- **Test**: `make test` or `go test -v ./...`
- **Demo**: `make demo` to test with sample data
- **Install**: `make install` for local testing

## Plugin Development
- **Protocol**: Simple stdin/stdout JSON contract
- **Timeouts**: Hard limits, fail gracefully
- **Caching**: Use cache_ttl_ms for expensive operations
- **Testing**: Test plugins independently before integration