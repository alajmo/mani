# Development

## Build instructions

### Prerequisites

- [go 1.18 or above](https://golang.org/doc/install)
- [goreleaser](https://goreleaser.com/install/)

### Building

```bash
# Build mani for your platform target
make build

# Build mani binaries and archives for all platforms using goreleaser
make build-all

# Build mani and get an interactive docker shell with completion
make build-exec

# Standing in _example directory you can run the following to debug faster
(cd .. && make build-and-link && cd - && ../dist/mani run multi -p template-generator)
```

## Dependency Graph

Create SVG dependency graphs using graphviz and [goda](https://github.com/loov/goda)

```bash
goda graph "github.com/alajmo/mani/..." | dot -Tsvg -o res/graph.svg
goda graph "github.com/alajmo/mani:all" | dot -Tsvg -o res/graph-full.svg
```
