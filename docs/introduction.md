---
slug: /
---

# Introduction

`mani` is a CLI tool that helps you manage multiple repositories. It's useful when you are working with microservices, multi-project systems, multiple libraries, or just a collection of repositories and want a central place for pulling all repositories and running commands across them.

`mani` has many features:

- Declarative configuration
- Clone multiple repositories with a single command
- Run custom or ad-hoc commands across multiple repositories
- Built-in TUI
- Flexible filtering
- Customizable theme
- Auto-completion support
- Portable, no dependencies

## Demo

![demo](/img/demo.gif)

## Example

You specify repositories and commands in a configuration file:

```yaml title="mani.yaml"
projects:
  pinto:
    url: https://github.com/alajmo/pinto.git
    desc: A vim theme editor
    tags: [frontend, node]

  template-generator:
    url: https://github.com/alajmo/template-generator.git
    desc: A simple bash script used to manage boilerplates
    tags: [cli, bash]
    env:
      branch: dev

tasks:
  git-status:
    desc: Show working tree status
    cmd: git status

  git-create:
    desc: Create branch
    spec:
      output: text
    env:
      branch: main
    cmd: git checkout -b $branch
```

Run `mani sync` to clone the repositories:

```bash
$ mani sync
✓ pinto
✓ dashgrid

All projects synced
```

Then run commands across all or a subset of the repositories:

```bash
# Target repositories that have the tag 'node'
$ mani run git-status --tags node

┌──────────┬─────────────────────────────────────────────────┐
│ Project  │ git-status                                      │
├──────────┼─────────────────────────────────────────────────┤
│ pinto    │ On branch master                                │
│          │ Your branch is up to date with 'origin/master'. │
│          │                                                 │
│          │ nothing to commit, working tree clean           │
└──────────┴─────────────────────────────────────────────────┘

# Target project 'pinto'
$ mani run git-create --projects pinto branch=dev

[pinto] TASK [git-create: create branch] *******************

Switched to a new branch 'dev'
```
