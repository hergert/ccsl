# CLAUDE.md

## Guidelines
- Concise commit messages, no co-author or attribution tags
- Keep it simple - minimal dependencies, clear logic
- Test before commit - `go test ./...` and `make demo`
- No commits without explicit request
- Performance first - keep total budget under 220ms
- Statusline JSON contract: `doctorInput` in cmd/ccsl/main.go mirrors
  https://code.claude.com/docs/en/statusline - re-verify it when Claude Code
  moves major versions

## Development
- **Build**: `make build` or `go build ./cmd/ccsl`
- **Test**: `make test` or `go test -v ./...`
- **Demo**: `make demo` to test with sample data

## Releasing
- Update CHANGELOG.md, then tag: `git tag vX.Y.Z && git push origin vX.Y.Z`.
  goreleaser builds darwin/linux binaries, publishes the GitHub release with
  checksums, and updates the Homebrew tap
- One-time setup before the first release: create the `hergert/homebrew-tap`
  repo and add a repo-scoped PAT as the `HOMEBREW_TAP_TOKEN` actions secret
  on hergert/ccsl
