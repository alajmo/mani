# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

mani is a CLI tool written in Go that helps manage multiple repositories. It's useful for microservices, multi-project systems, or any collection of repositories where you need a central place to pull repositories and run commands across them.

## Build & Development Commands

```bash
make build              # Build for current platform (output: dist/mani)
make build-all          # Cross-platform builds using goreleaser
make test               # Run unit + integration tests
make test-unit          # Run unit tests only (go test ./core/dao/...)
make test-integration   # Run integration tests (requires Docker)
make update-golden-files # Update golden test outputs
make gofmt              # Format Go code
make lint               # Run golangci-lint and deadcode
make tidy               # Update dependencies
```

## Architecture

**Entry point flow:**
```
main.go -> cmd.Execute() (cmd/root.go) -> Cobra command routing
```

**Core layers:**
- `cmd/` - Cobra CLI command definitions (exec, run, sync, list, describe, tui, etc.)
- `core/dao/` - Data structures, YAML config parsing, project/task/spec definitions
- `core/exec/` - Task execution engine, SSH/exec clients, cloning logic
- `core/print/` - Output formatters (table, stream, tree, html, markdown)
- `core/tui/` - Terminal UI using tview/tcell

**Configuration:** YAML-based (`mani.yaml`) with projects, tasks, specs, and themes.

## Key Patterns

- CLI commands in `cmd/` delegate to business logic in `core/`
- Custom error types in `core/errors.go` - use `core.CheckIfError()` for consistency
- Config lookup: `mani.yaml`, `mani.yml`, `.mani.yaml`, `.mani.yml`
- Shell handling: bash on Unix, powershell on Windows
- Use `filepath.Join()` for cross-platform paths

## Testing

- Unit tests in `core/dao/*_test.go`
- Integration tests in `test/integration/` using Docker and golden files
- Golden files in `test/integration/golden/` serve as expected outputs
- Test script: `./test/scripts/test` (supports --debug, --build, --run, --update, --count, --clean)

## Adding Features

1. Add data structures in `core/dao/` if needed
2. Implement business logic in `core/`
3. Add CLI command in `cmd/` if needed
4. Update config schema in `docs/config.md` if adding new options
5. Add tests (unit and/or integration)
