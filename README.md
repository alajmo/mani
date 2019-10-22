# Mani

`mani` is a CLI tool that helps you manage multiple repositories. It's helpful when you are working with microservices and want a central place for pulling all repositories and running common commands over different projects.

---

[![Build Status](https://github.com/samiralajmovic/mani/workflows/build/badge.svg)](https://github.com/samiralajmovic/mani/actions)
[![Release](https://img.shields.io/github/release-pre/samiralajmovic/mani.svg)](https://github.com/samiralajmovic/mani/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](https://img.shields.io/badge/license-MIT-green)
[![Go Report Card](https://goreportcard.com/badge/github.com/samiralajmovic/mani)](https://goreportcard.com/report/github.com/samiralajmovic/mani)

## Dependencies

- [git](https://git-scm.com)

## Install

`mani` is available on Linux and Mac.

- Binaries for Linux and Mac are available as tarballs in the [release](https://github.com/samiralajmovic/loop/releases) page.
- Build from source:
  1.  Clone the repo
  2.  Add the following command in your go.mod file
      ```text
      replace (
        github.com/samiralajmovic/mani => MY_MANI_CLONED_GIT_REPO
      )
      ```
  3.  Build and run the executable
      ```shell
      go build
      ```

## Usage

### 1) Initialize a mani configuration file

Assuming you stand in a directory called `demo` and run `mani init`, the following files will be generated:

`mani.yaml`

```yaml
projects:
  - name: demo
    path: .
    url:

commands:
  - name: hello-world
    description: Print Hello World
    command: echo "Hello World"
```

`.gitignore`

```
demo
```

### 2) Add some projects to your mani.yaml

Add two projects, `project-a` and `project-b`, and 1 new command.

`mani.yaml`

```yaml
version: alpha

projects:
  - name: demo
    path: .

  - name: project-a
    tags:
      - frontend

  - name: project-b
    tags:
      - backend

commands:
  - name: hello-world
    description: Print Hello World
    command: echo "$PWD"

  - name: list-files
    command: ls
```

### 3) Run some command across multiple projects

```
# Select using tags flag
mani run list-files -t frontend

# Select using project flag
mani run list-files -p project-a
```

## Common Commands

```
# List all available CLI options
mani help

# Sync mani
mani sync

# List projects
mani list projects

# Info about which config file is used
mani info
```
