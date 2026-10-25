# Development

## Build instructions

### Prerequisites

- [go 1.25 or above](https://golang.org/doc/install)
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

## Project config loading

Project loading has three stages. Other resource types retain their existing
loading code.

1. `core/dao/project_definition.go` parses YAML into `projectDefinition` values.
   Each scalar records its typed value, whether it was set, and its source file
   and line. Collection entries retain their own sources. `project_yaml.go`
   handles aliases and YAML merge keys only at this parsing boundary.
2. `core/dao/project_merge.go` groups definitions by project name and merges
   them. `projectCandidates` applies the same precedence rule to fields,
   collection entries and path ownership. `project_precedence.go` derives that
   rule from the import graph, including the user config's fallback priority.
   The merge stage never compares serialized YAML or executes commands.
3. `core/dao/project_resolve.go` resolves the surviving paths, checks path
   ambiguity and applies defaults to produce runtime `Project` values.
   `ReadConfig` finishes validating task references before the final project
   env evaluation pass. That pass evaluates only surviving entries and uses
   their original source locations for errors.

Paths retain their declaring file and `project_dir` through merging. Independent
path candidates are compared after resolution, so two spellings of the same
absolute path agree. Project order and collection order start with the definition
that supplies the project's path; overlays append their additional entries.

The project tests cover this pipeline through `ReadConfig`, including explicit
empty/false overrides, independent conflicts, diamond imports, path ownership,
aliases, and ensuring invalid configs cannot execute project env commands.

## Releasing

The following workflow is used for releasing a new `mani` version:

1. Create pull request with changes
2. Verify build works (especially windows build)
   - `make build`
   - `make build-all`
3. Pass all integration and unit tests locally
   - `make test-unit`
   - `make test-integration`
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
