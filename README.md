[![Build Status](https://github.com/samiralajmovic/mani/workflows/build/badge.svg)](https://github.com/samiralajmovic/mani/actions)
[![Release](https://img.shields.io/github/release-pre/samiralajmovic/mani.svg)](https://github.com/samiralajmovic/mani/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](https://img.shields.io/badge/license-MIT-green)
[![Go Report Card](https://goreportcard.com/badge/github.com/samiralajmovic/mani)](https://goreportcard.com/report/github.com/samiralajmovic/mani)

# mani

`mani` is a tool that helps you manage multiple repositories. It's helpful when you are working with microservices or multi-project system and libraries and want a central place for pulling all repositories and running commands over the different projects. You specify projects and commands in a yaml config and then run the commands over all or a subset of the projects.

## Features

- Clone multiple repositories in one command
- Run commands over multiple projects
- Declarative configuration
- Single CLI binary with auto-completion

## Install

`mani` is available on Linux and Mac.

- Binaries are available as tarballs in the [release](https://github.com/samiralajmovic/mani/releases) page.
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

### Auto-completion

#### Bash

Auto-complete requires [bash-completion](https://github.com/scop/bash-completion#installation).

There's two ways to add `mani` auto-completion:

- Source the completion script in your `~/.bashrc` file:

  `echo 'source <(mani completion)' >>~/.bashrc`

or

- Add the completion script to the `/etc/bash_completion.d directory:`

  `mani completion >/etc/bash_completion.d/mani`

#### Zsh

Coming.

## Usage

Checkout the [example](/example) directory to see how it can be used. You can also always run `mani help` to see available commands.

### Create a New Mani Repository

Run the following command inside a directory to initialize a mani repo:

```sh
$ mani init
```

This will generate two files:

- `mani.yaml`: contains projects and custom commands. Any sub-directory that has a `.git` inside it will be included (add flag `--auto-discovery=false` to turn off this feature).
- `.gitignore`: includes the projects specified in `mani.yaml` file.

It can be helpful to initialize the `mani` repository as a git repository, so that your teammates can easily download the `mani` repository and run `mani sync` to clone all repositories and get the same project setup as you.

### Add New Project to Mani file

Add another project to `mani.yaml` and run `mani sync` to pull the repository and add it to the `.gitignore`.

### Run commands across multiple projects

```sh
# Run arbitrary command
mani exec 'ls -alh' --all-projects

# Specify projects using tags flag
mani run list-files -t frontend

# SPecify project using project flag
mani run list-files -p project-a
```

## Config Structure

```yaml
projects:
  - name: example
    path: .

  - name: idetheme
    path: frontend/idetheme
    url: https://github.com/samiralajmovic/idetheme
    tags:
      - frontend

commands:
  - name: multi
    command: | # multiline command
      echo "1st line"
      echo "2nd line"

  - name: checkout
    args:
      branch: master # default value, override with: mani run checkout -a branch=development
    command: git checkout $branch
```

## Roadmap

`mani` is under active development and some of the things I aim to add/fix is:

- [ ] Add completion for zsh
- [ ] Add tests
- [ ] Add package to brew
- [ ] Add package to snap
- [ ] Add package to Ubuntu, Debian
- [ ] Add Windows support
- [ ] Add CRUD methods for project to config via cli (user-input)
- [ ] Add CRUD methods for command to config via cli (user-input)
- [ ] Add `--auto-discovery` flag to sync sub-command
- [ ] Add property for global variables, as well as global variables that are sourced commands (like `date` command to return current date)
- [ ] Remove duplicate flag auto-completion (both and without = showing)
