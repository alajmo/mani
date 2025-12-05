# Copilot Instructions for mani

This document provides guidance for GitHub Copilot when working on the `mani` repository. `mani` is a CLI tool written in Go that helps manage multiple repositories. It's useful for microservices, multi-project systems, or any collection of repositories.

## Project Overview

- **Language**: Go 1.23+
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra)
- **Configuration**: YAML-based configuration files (`mani.yaml`)
- **Key Features**: Repository management, task running across multiple repos, TUI interface

## Repository Structure

```
mani/
├── cmd/              # CLI commands (Cobra commands)
├── core/             # Core business logic
│   ├── dao/          # Data access objects, config parsing
│   ├── exec/         # Command execution logic
│   ├── print/        # Output formatting
│   └── tui/          # Terminal UI components
├── docs/             # Documentation
├── examples/         # Example configurations
├── test/             # Test fixtures and integration tests
│   ├── fixtures/     # Test data
│   ├── integration/  # Integration test files
│   └── scripts/      # Test scripts
└── main.go           # Entry point
```

## Development Commands

### Building

```bash
# Build for local platform
make build

# Build with test mode flags
make build-test

# Build for all platforms (requires goreleaser)
make build-all
```

### Testing

```bash
# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Run all tests (unit + integration)
make test

# Update golden files for integration tests
make update-golden-files
```

### Linting and Formatting

```bash
# Format Go code
make gofmt

# Run linter (requires golangci-lint and deadcode)
make lint

# Update dependencies
make tidy
```

## Coding Standards

### Go Code Style

1. **Follow standard Go conventions**:
   - Use `gofmt` for formatting
   - Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
   - Use meaningful variable and function names

2. **Error handling**:
   - Always check and handle errors explicitly
   - Use the custom error types in `core/errors.go` for consistent error messages
   - Return errors to callers rather than panicking

3. **Package structure**:
   - Keep `cmd/` for CLI command definitions only
   - Business logic goes in `core/`
   - Data structures and parsing go in `core/dao/`

### Example Code Patterns

**CLI Command Pattern** (in `cmd/`):
```go
func exampleCmd(config *dao.Config, configErr *error) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "example",
        Short: "Short description",
        Long:  "Long description of the command",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Check config error first
            if *configErr != nil {
                return *configErr
            }
            // Command implementation
            return nil
        },
    }
    return cmd
}
```

**Configuration Struct Pattern** (in `core/dao/`):
```go
type Example struct {
    // Internal fields (not from YAML)
    internalField string `yaml:"-"`
    
    // YAML fields
    Name        string    `yaml:"name"`
    Description string    `yaml:"desc"`
    Options     yaml.Node `yaml:"options"`
}
```

### Testing Guidelines

1. **Unit tests**: Place in the same package with `_test.go` suffix
2. **Integration tests**: Located in `test/integration/`
3. **Golden files**: Used for integration tests, update with `make update-golden-files`
4. **Test naming**: Use descriptive names like `TestConfig_DuplicateProjectName`

## Adding Features

When adding a new feature:

1. **Understand the existing patterns**:
   - Review similar features in the codebase
   - Follow the established code organization

2. **Implementation steps**:
   - Add data structures in `core/dao/` if needed
   - Implement business logic in `core/`
   - Add CLI command in `cmd/` if needed
   - Add tests (unit and/or integration)

3. **Configuration changes**:
   - Update `mani.yaml` schema if adding new config options
   - Update documentation in `docs/config.md`
   - Consider backward compatibility

4. **Documentation**:
   - Update relevant docs in `docs/`
   - Update README.md if needed
   - Add examples if appropriate

## Fixing Bugs

When fixing a bug:

1. **Reproduce the issue**:
   - Create a minimal test case
   - Understand the root cause

2. **Fix process**:
   - Make the minimal change needed
   - Add a test that would have caught the bug
   - Verify the fix doesn't break existing functionality

3. **Verification**:
   - Run `make test` to ensure all tests pass
   - Run `make lint` to check code quality
   - Test manually with real scenarios

## Platform Considerations

- **Windows support**: The code handles Windows-specific shell behavior
  - Default shell on Windows is `powershell -NoProfile`
  - Default shell on Unix is `bash -c`
- **Cross-platform paths**: Use `filepath.Join()` for paths

## Key Dependencies

- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/gdamore/tcell/v2` - Terminal handling for TUI
- `github.com/rivo/tview` - TUI components
- `github.com/jedib0t/go-pretty/v6` - Table formatting

## Configuration File Structure

The main configuration file is `mani.yaml`:

```yaml
# Global settings
shell: bash -c
sync_remotes: false
sync_gitignore: true

# Environment variables
env:
  KEY: value

# Project definitions
projects:
  project-name:
    path: ./relative/path
    url: git@github.com:user/repo.git
    tags: [tag1, tag2]

# Task definitions
tasks:
  task-name:
    desc: Task description
    cmd: echo "command"
    # Or multi-command:
    commands:
      - task: other-task
      - cmd: echo "inline command"
```

## Pull Request Guidelines

1. Reference the issue number if one exists
2. Provide a clear description of changes
3. Ensure all tests pass (`make test`)
4. Ensure code is formatted (`make gofmt`)
5. Ensure linter passes (`make lint`)
6. Update documentation if needed
