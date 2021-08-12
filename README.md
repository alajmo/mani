[![Build Status](https://github.com/alajmo/mani/workflows/test/badge.svg)](https://github.com/alajmo/mani/actions)
[![Release](https://img.shields.io/github/release-pre/alajmo/mani.svg)](https://github.com/alajmo/mani/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](https://img.shields.io/badge/license-MIT-green)
[![Go Report Card](https://goreportcard.com/badge/github.com/alajmo/mani)](https://goreportcard.com/report/github.com/alajmo/mani)

# Mani

<img src="./res/logo-1.png" align="right"/>

`mani` is a many-repo tool that helps you manage multiple repositories or plain directories. It's useful when you are working with microservices, multi-project systems, and libraries or just a bunch of repositories and want a central place for pulling all repositories and running commands over the different repositories.

You specify repository and commands in a config file and then run the commands over all or a subset of the projects.

![demo](res/output.gif)

## Features

- Clone multiple repositories in one command
- Run custom or ad-hoc commands over multiple repositories
- Flexible filtering
- Declarative configuration
- Portable, no dependencies
- Supports auto-completion

## Table of Contents

<!-- vim-markdown-toc GFM -->

* [Installation](#installation)
  * [Building From Source](#building-from-source)
* [Usage](#usage)
  * [Create a New Mani Repository](#create-a-new-mani-repository)
  * [Common Commands](#common-commands)
  * [Documentation](#documentation)
* [License](#license)

<!-- vim-markdown-toc -->

## Installation

`mani` is available on Linux and Mac, with partial support for Windows.


* Binaries are available on the [release](https://github.com/alajmo/mani/releases) page

* Via GO install
    ```sh
    go get -u github.com/alajmo/mani
    ```

Auto-completion is available via `mani completion bash|zsh|fish|powershell`.

### Building From Source

1. Clone the repo
2. Build and run the executable
    ```sh
    make build && ./dist/mani
    ```

## Usage

### Create a New Mani Repository

Run the following command inside a directory containing your `git` repositories, to initialize a mani repo:

```sh
$ mani init
```

This will generate two files:

- `mani.yaml`: contains projects and custom tasks. Any sub-directory that has a `.git` inside it will be included (add the flag `--auto-discovery=false` to turn off this feature)
- `.gitignore`: includes the projects specified in `mani.yaml` file

It can be helpful to initialize the `mani` repository as a git repository so that anyone can easily download the `mani` repository and run `mani sync` to clone all repositories and get the same project setup as you.

### Common Commands

```sh
# Run arbitrary command (list all files for instance)
mani exec --all-projects 'ls -alh'

# List all repositories
mani list projects

# List repositories in a tree-like format
mani tree

# Describe available tasks
mani describe tasks

# Run task specified in mani.yaml and run only projects that have the frontend tag
mani run list-files -t frontend

# Run task specified in mani.yaml and run only specified projects
mani run list-files -p project-a

# Open up mani.yaml in your preferred editor
mani edit
```
### Documentation

Checkout the following to learn more about mani:

- [Examples](_example)
- [Documentation](docs/DOCUMENTATION.md)
- [List of Useful Git Commands](docs/COMMANDS.md)
- [Project Background](docs/PROJECT-BACKGROUND.md)

## [License](LICENSE)

The MIT License (MIT)

Copyright (c) 2020-2021 Samir Alajmovic
