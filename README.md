[![Build Status](https://github.com/alajmo/mani/workflows/test/badge.svg)](https://github.com/alajmo/mani/actions)
[![Release](https://img.shields.io/github/release-pre/alajmo/mani.svg)](https://github.com/alajmo/mani/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](https://img.shields.io/badge/license-MIT-green)
[![Go Report Card](https://goreportcard.com/badge/github.com/alajmo/mani)](https://goreportcard.com/report/github.com/alajmo/mani)

# mani

`mani` is a tool that helps you manage multiple repositories. It's helpful when you are working with microservices or multi-project system and libraries and want a central place for pulling all repositories and running commands over the different projects. You specify projects and commands in a yaml config and then run the commands over all or a subset of the projects.

## Features

- Clone multiple repositories in one command
- Run commands over multiple projects
- Declarative configuration
- Portable, no dependencies
- Supports auto-completion

## Table of Contents

* [Installation](#installation)
  * [Source](#source)
* [Usage](#usage)
  * [Create a New Mani Repository](#create-a-new-mani-repository)
  * [Add New Project to Mani Repository](#add-new-project-to-mani-repository)
  * [Run commands across multiple projects](#run-commands-across-multiple-projects)
* [Config Structure](#config-structure)
* [License](#license)

## Installation

`mani` is available on `Linux` and `Mac`, with partial support for `Windows`. Executables are available [here](https://github.com/alajmo/mani/releases) page.

### Source

1.  Clone the repo
2.  Add the following command in your go.mod file
    ```text
    replace (
      github.com/alajmo/mani => MY_MANI_CLONED_GIT_REPO
    )
    ```
3.  Build and run the executable
    ```shell
    go build
    ```

## Usage

Checkout the [example](/_example) directory to see how it can be used.

```
mani is a tool used to manage multiple repositories

Usage:
  mani [command]

Available Commands:
  completion  Output shell completion code for bash
  exec        Execute arbitrary commands
  help        Help about any command
  info        Print configuration file path
  init        Initialize a mani repository
  list        List projects, commands and tags
  run         Run commands
  sync        Clone repositories and add to gitignore
  version     Print version/build info

Flags:
  -c, --config string   config file (by default it checks current and all parent directories for mani.yaml|yml)
  -h, --help            help for mani

Use "mani [command] --help" for more information about a command.
```

### Create a New Mani Repository

Run the following command inside a directory to initialize a mani repo:

```sh
$ mani init
```

This will generate two files:

- `mani.yaml`: contains projects and custom commands. Any sub-directory that has a `.git` inside it will be included (add flag `--auto-discovery=false` to turn off this feature)
- `.gitignore`: includes the projects specified in `mani.yaml` file

It can be helpful to initialize the `mani` repository as a git repository, so that anyone can easily download the `mani` repository and run `mani sync` to clone all repositories and get the same project setup as you.

### Add New Project to Mani Repository

Add another project to `mani.yaml` and run `mani sync` to pull the repository and add it to `.gitignore`.

### Run commands across multiple projects

```sh
# Run arbitrary command (list all files for instance)
mani exec --all-projects 'ls -alh'

# Execute a command specified in mani.yaml and run only projects that have the frontend tag
mani run list-files -t frontend

# Execute a command specified in mani.yaml and run only specified projects
mani run list-files -p project-a
```

## Config Structure

The `mani.yaml` config is based on two concepts: __projects__ and __commands__. __Projects__ are simply directories, which may be git repositories, in which case they have an url attribute. __Commands__ are arbitrary shell commands that you write and then run for selected __projects__.

```yaml
projects:
  - name: example
    path: .

  - name: pinto
    path: frontend/pinto
    url: git@github.com:alajmo/pinto
    tags: [frontend]

shell: bash -c # defaults to sh -c

commands:
    shell: node -e
    command: |
      console.log("Hello World")

  - name: multi
    command: | # multiline command
      echo "1st line"
      echo "2nd line"

  - name: checkout
    args:
      branch: master # default value, override with: mani run checkout -a branch=development
    command: git checkout $branch

  # Every command has access to environment variable MANI_CONFIG which prints the current mani configuration path
  - name: print-mani-pwd
    command: echo $MANI_CONFIG
```

## [License](LICENSE)

The MIT License (MIT)

Copyright (c) 2020-2021 Samir Alajmovic
