# CLAUDE.md

## Guidelines
- Concise commit messages, no co-author or attribution tags
- Keep it simple - minimal dependencies, clear logic
- Test before commit - `go test ./...` and `make demo`
- No commits without explicit request
- Performance first - keep total budget under 220ms

## Development
- **Build**: `make build` or `go build ./cmd/ccsl`
- **Test**: `make test` or `go test -v ./...`
- **Demo**: `make demo` to test with sample data
