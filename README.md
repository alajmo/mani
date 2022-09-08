[![Build Status](https://github.com/alajmo/mani/workflows/test/badge.svg)](https://github.com/alajmo/mani/actions)
[![Release](https://img.shields.io/github/release-pre/alajmo/mani.svg)](https://github.com/alajmo/mani/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](https://img.shields.io/badge/license-MIT-green)
[![Go Report Card](https://goreportcard.com/badge/github.com/alajmo/mani)](https://goreportcard.com/report/github.com/alajmo/mani)

# Mani

<img src="./res/logo.svg" align="right"/>

`mani` is a CLI tool that helps you manage multiple repositories. It's useful when you are working with microservices, multi-project systems, many libraries or just a bunch of repositories and want a central place for pulling all repositories and running commands over them.

You specify repository and commands in a config file and then run the commands over all or a subset of the repositories.

![demo](res/output.gif)

Interested in managing your servers in a similar way? Checkout [sake](https://github.com/alajmo/sake)!

## Features

- Clone multiple repositories in one command
- Declarative configuration
- Run custom or ad-hoc commands over multiple repositories
- Flexible filtering
- Customizable theme
- Portable, no dependencies
- Supports auto-completion

## Table of Contents

- [Installation](#installation)
  - [Building From Source](#building-from-source)
- [Usage](#usage)
  - [Create a New Mani Repository](#create-a-new-mani-repository)
  - [Command Examples](#run-some-commands)
  - [Documentation](#documentation)
- [Contributing](docs/contributing.md)
- [License](#license)

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

* via Arch
  ```sh
  pacman -S mani
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

### Create a New Mani Repository

Run the following command inside a directory containing your `git` repositories:

```sh
$ mani init
```

This will generate **two** files:

- `mani.yaml`: contains projects and custom tasks. Any sub-directory that has a `.git` inside it will be included (add the flag `--auto-discovery=false` to turn off this feature)
- `.gitignore`: includes the projects specified in `mani.yaml` file. To opt out, use `mani init --vcs=none`.

It can be helpful to initialize the `mani` repository as a git repository so that anyone can easily download the `mani` repository and run `mani sync` to clone all repositories and get the same project setup as you.

### Run Some Commands

```bash
# List all projects
$ mani list projects

# Count number of files in each project in parallel
$ mani exec --all --output table --parallel 'find . -type f | wc -l'
```

### Documentation

Checkout the following to learn more about mani:

- [Examples](examples)
- [Config](docs/config.md)
- [Commands](docs/commands.md)
- [Changelog](/docs/changelog.md)
- [Project Background](docs/project-background.md)

## [License](LICENSE)

The MIT License (MIT)

Copyright (c) 2020-2021 Samir Alajmovic
