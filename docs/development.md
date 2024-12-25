# Development

## Build instructions

### Prerequisites

- [go 1.23 or above](https://golang.org/doc/install)
- [goreleaser](https://goreleaser.com/install/)

### Building

```bash
# Build mani for your platform target
make build

# Build mani binaries and archives for all platforms using goreleaser
make build-all

# Generate Manpage
make gen-man
```

## Developing

```bash
# Format code
make gofmt

# Manage dependencies (download/remove unused)
make tidy

# Lint code
make lint

# Build mani and get an interactive docker shell with completion
make build-exec

# Standing in _example directory you can run the following to debug faster
(cd .. && make build-and-link && cd - && ../dist/mani run multi -p template-generator)
```

## Releasing

The following workflow is used for releasing a new `mani` version:

1. Create pull request with changes
2. Verify build works (especially windows build)
   - `make build`
   - `make build-all`
3. Pass all integration and unit tests locally
   - `make test-integration`
   - `make test-unit`
4. Update `config.man` and `config.md` if any config changes and generate manpage
   - `make gen-man`
5. Update `Makefile` and `CHANGELOG.md` with correct version, and add all changes to `CHANGELOG.md`
6. Squash-merge to main with `Release vx.y.z` and description of changes
7. Run `make release`, which will:
   - Create a git tag with release notes
   - Trigger a build in Github that builds cross-platform binaries and generates release notes of changes between current and previous tag

## Dependency Graph

Create SVG dependency graphs using graphviz and [goda](https://github.com/loov/goda)

```bash
goda graph "github.com/alajmo/mani/..." | dot -Tsvg -o res/graph.svg
goda graph "github.com/alajmo/mani:all" | dot -Tsvg -o res/graph-full.svg
```
