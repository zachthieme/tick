# Changelog

## v0.1.0 — 2026-04-05

First tagged release.

### Features
- Terminal dashboard with large figlet numbers visible from across the room
- O(1) weekday calculation with DST-safe UTC normalization
- `--once` mode for one-liner output
- `--json` mode for scripting
- `--hosts-file` with automatic re-read on each tick (TUI mode)
- `--today` override for planning and testing
- Graceful degradation: `(stale)` indicator on file-read errors, plain-number fallback on narrow terminals
- Weekend detection with "next work night is Monday" message

### Build & CI
- GitHub Actions: lint, test with race detector, 80% coverage gate, tag-gated goreleaser releases
- Cross-platform binaries: linux/darwin x amd64/arm64
- Nix flake with git-rev version injection

### Code quality
- 95.8% test coverage (calc 100%, tui 96.4%, run 97.1%)
- Package doc comments, table-driven tests, full TUI lifecycle tests
- Signal handling (SIGTERM, SIGHUP) for clean shutdown
