# Development

## Build instructions

### Prerequisites

- [go 1.16 or above](https://golang.org/doc/install)
- [goreleaser](https://goreleaser.com/install/)

### Building

```sh
# Build mani for your platform target
make build

# Build mani binaries and archives for all platforms using goreleaser
make build-all

# Build mani and get an interactive docker shell with completion
make build-exec
```

### Releasing

The following workflow is used for releasing a new `mani` version:

1. Create pull request with changes
2. Pass all tests
3. Squash-merge to main with descriptive commit message
4. Update `Makefile` and `CHANGELOG.md` with correct version, and add all changes to `CHANGELOG.md`
5. Push with commit message `Release vx.y.z`
6. Run `make release`, which will:
   1. Create a git tag with release notes
   2. Trigger a build in Github that builds cross-platform binaries and generates release notes of changes between current and previous tag
