# Example

This is an example of how you could use `mani`.

`mani.yaml`
```yaml
projects:
  - name: example
    path: .

  - name: pinto
    path: frontend/pinto
    url: git@github.com:alajmo/pinto
    description: A vim theme editor
    tags:
      - frontend

  - name: dashgrid
    path: frontend/dashgrid
    url: git@github.com:alajmo/dashgrid
    description: A highly customizable drag-and-drop grid
    tags:
      - frontend

  - name: template-generator
    url: git@github.com:alajmo/template-generator
    description: A simple bash script used to manage boilerplates
    tags: [cli, bash]

tasks:
  - name: fetch
    command: git fetch

  - name: status
    command: git status

  - name: checkout
    args:
      branch: master
    command: git checkout $branch

  - name: create-branch
    command: git checkout -b $branch

  - name: multi
    command: | #Multi line command
      echo "2nd line "
      echo "3rd line"
```

Given the above `mani.yaml` we can run commands like:

```sh
# List all projects
$ mani list projects

┌────────────────────┬───────────┬──────────────────────────────────────────────────┐
│ NAME               │ TAGS      │ DESCRIPTION                                      │
├────────────────────┼───────────┼──────────────────────────────────────────────────┤
│ example            │           │                                                  │
│ pinto              │ frontend  │ A vim theme editor                               │
│ dashgrid           │ frontend  │ A highly customizable drag-and-drop grid         │
│ template-generator │ cli, bash │ A simple bash script used to manage boilerplates │
└────────────────────┴───────────┴──────────────────────────────────────────────────┘

# Get latest changes
$ mani run status --all-projects --output table

Name:         status
Description:
Shell:        sh -c
Args:
Command:      git status

┌────────────────────┬─────────────────────────────────────────────────────────────────────────────┐
│ NAME               │ OUTPUT                                                                      │
├────────────────────┼─────────────────────────────────────────────────────────────────────────────┤
│ example            │ On branch main                                                              │
│                    │ Your branch is up to date with 'origin/main'.                               │
│                    │                                                                             │
│                    │ Changes not staged for commit:                                              │
│                    │   (use "git add <file>..." to update what will be committed)                │
│                    │   (use "git checkout -- <file>..." to discard changes in working directory) │
│                    │                                                                             │
│                    │     modified:   README.md                                                   │
│                    │     modified:   mani.yaml                                                   │
│                    │     modified:   ../cmd/list.go                                              │
│                    │                                                                             │
│                    │ no changes added to commit (use "git add" and/or "git commit -a")           │
│                    │                                                                             │
├────────────────────┼─────────────────────────────────────────────────────────────────────────────┤
│ pinto              │ On branch master                                                            │
│                    │ Your branch is up to date with 'origin/master'.                             │
│                    │                                                                             │
│                    │ nothing to commit, working tree clean                                       │
│                    │                                                                             │
├────────────────────┼─────────────────────────────────────────────────────────────────────────────┤
│ dashgrid           │ On branch master                                                            │
│                    │ Your branch is up to date with 'origin/master'.                             │
│                    │                                                                             │
│                    │ nothing to commit, working tree clean                                       │
│                    │                                                                             │
├────────────────────┼─────────────────────────────────────────────────────────────────────────────┤
│ template-generator │ On branch master                                                            │
│                    │ Your branch is up to date with 'origin/master'.                             │
│                    │                                                                             │
│                    │ nothing to commit, working tree clean                                       │
│                    │                                                                             │
└────────────────────┴─────────────────────────────────────────────────────────────────────────────┘

# Run custom ls command for projects with tag bash
$ mani exec 'ls' -t bash

template-generator
CHANGELOG.md
CONTRIBUTING.md
LICENSE
media
README.md
scripts
tests
tp
tp-completion.sh
```
