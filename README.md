<h1 align="center"><code>mani</code></h1>

<div align="center">
  <a href="https://github.com/alajmo/mani/releases">
    <img src="https://img.shields.io/github/release-pre/alajmo/mani.svg" alt="version">
  </a>

  <a href="https://github.com/alajmo/mani/actions">
    <img src="https://github.com/alajmo/mani/workflows/release/badge.svg" alt="build status">
  </a>

  <a href="https://img.shields.io/badge/license-MIT-green">
    <img src="https://img.shields.io/badge/license-MIT-green" alt="license">
  </a>

  <a href="https://goreportcard.com/report/github.com/alajmo/mani">
    <img src="https://goreportcard.com/badge/github.com/alajmo/mani" alt="Go Report Card">
  </a>

  <a href="https://pkg.go.dev/github.com/alajmo/mani">
    <img src="https://pkg.go.dev/badge/github.com/alajmo/mani.svg" alt="reference">
  </a>
</div>

<br>

<img src="./res/logo.svg" align="right" width="150"/>

`mani` lets you manage multiple repositories and run commands across them.

![demo](res/demo.gif)

Interested in managing your servers in a similar way? Checkout [sake](https://github.com/alajmo/sake)!

## Table of Contents

- [Sponsors](#sponsors)
- [Installation](#installation)
  - [Building From Source](#building-from-source)
- [Usage](#usage)
  - [Initialize Mani](#initialize-mani)
  - [Example Commands](#example-commands)
  - [Documentation](#documentation)
- [License](#license)

## Sponsors

Mani is an MIT-licensed open source project with ongoing development. If you'd like to support their efforts, check out [Tabify](https://chromewebstore.google.com/detail/tabify/bokfkclamoepkmhjncgkdldmhfpgfdmo) - a Chrome extension that enhances your browsing experience with powerful window and tab management, focus-improving site blocking, and numerous features to optimize your browser workflow.

## Installation

[![Packaging status](https://repology.org/badge/vertical-allrepos/mani.svg)](https://repology.org/project/mani/versions)

`mani` is available on Linux and Mac, with partial support for Windows.

* Binaries are available on the [release](https://github.com/alajmo/mani/releases) page

* via cURL (Linux & macOS)
  ```sh
  curl -sfL https://raw.githubusercontent.com/alajmo/mani/main/install.sh | sh
  ```

* via Homebrew
  ```sh
  brew tap alajmo/mani
  brew install mani
  ```

* via MacPorts
  ```sh
  sudo port install mani
  ```

* via Arch (AUR)
  ```sh
  yay -S mani
  ```

* via Nix
  ```sh
  nix-env -iA nixos.mani
  ```

* via Go
  ```sh
  go get -u github.com/alajmo/mani
  ```

Auto-completion is available via `mani completion bash|zsh|fish|powershell` and man page via `mani gen`.

### Building From Source

1. Clone the repo
2. Build and run the executable
    ```sh
    make build && ./dist/mani
    ```

## Usage

### Initialize Mani

Run the following command inside a directory containing your `git` repositories:

```sh
mani init
```

This will generate:

- `mani.yaml`: Contains projects and custom tasks. Any subdirectory that has a `.git` directory will be included (add the flag `--auto-discovery=false` to turn off this feature)
- `.gitignore`: (only when inside a git repo) Includes the projects specified in `mani.yaml` file. To opt out, use `mani init --sync-gitignore=false`.

It can be helpful to initialize the `mani` repository as a git repository so that anyone can easily download the `mani` repository and run `mani sync` to clone all repositories and get the same project setup as you.

### Example Commands

```bash
# List all projects
mani list projects

# Run git status across all projects
mani exec --all git status

# Run git status across all projects in parallel with output in table format
mani exec --all --parallel --output table git status
```

### Documentation

Checkout the following to learn more about mani:

- [Examples](examples)
- [Config](docs/config.md)
- [Commands](docs/commands.md)
- Documentation
  - [Filtering Projects](docs/filtering-projects.md)
  - [Variables](docs/variables.md)
  - [Output](docs/output.md)
- [Changelog](/docs/changelog.md)
- [Roadmap](/docs/roadmap.md)
- [Project Background](docs/project-background.md)
- [Contributing](docs/contributing.md)

## [License](LICENSE)

The MIT License (MIT)

Copyright (c) 2020-2021 Samir Alajmovic
