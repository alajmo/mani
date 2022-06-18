# Examples

This is an example of how you can use `mani`. Simply save the following content to a file named `mani.yaml` and then run `mani sync` to clone all the repositories. If you already have your own repositories, just omit the `projects` section.

You can then run some of the [commands](#commands) or checkout [git-quick-stats.sh](https://git-quick-stats.sh/) for additional git statistics and run them via `mani` for multiple projects.

### Config

```yaml
projects:
  example:
    path: .
    desc: A mani example

  pinto:
    path: frontend/pinto
    url: https://github.com/alajmo/pinto.git
    desc: A vim theme editor
    tags: [frontend, node]

  dashgrid:
    path: frontend/dashgrid
    url: https://github.com/alajmo/dashgrid.git
    desc: A highly customizable drag-and-drop grid
    tags: [frontend, lib, node]

  template-generator:
    url: https://github.com/alajmo/template-generator.git
    desc: A simple bash script used to manage boilerplates
    tags: [cli, bash]
    env:
      branch: master

specs:
  custom:
    output: table
    parallel: true

targets:
  all:
    all: true

themes:
  custom:
    table:
      options:
        draw_border: true
        separate_columns: true
        separate_header: true
        separate_rows: true

tasks:
  git-status:
    desc: show working tree status
    spec: custom
    target: all
    cmd: git status -s

  git-last-commit-msg:
    desc: show last commit
    cmd: git log -1 --pretty=%B

  git-last-commit-date:
    desc: show last commit date
    cmd: |
      git log -1 --format="%cd (%cr)" -n 1 --date=format:"%d  %b %y" \
      | sed 's/ //'

  git-branch:
    desc: show current git branch
    cmd: git rev-parse --abbrev-ref HEAD

  npm-install:
    desc: run npm install in node repos
    target:
      tags: [node]
    cmd: npm install

  git-overview:
    desc: show branch, local and remote diffs, last commit and date
    spec: custom
    target: all
    theme: custom
    commands:
      - task: git-branch
      - task: git-last-commit-msg
      - task: git-last-commit-date
```

## Commands

### List all Projects as Table or Tree:

```bash
$ mani list projects

 Project            | Tag                 | Description
--------------------+---------------------+--------------------------------------------------
 example            |                     | A mani example
 pinto              | frontend, node      | A vim theme editor
 dashgrid           | frontend, lib, node | A highly customizable drag-and-drop grid
 template-generator | cli, bash           | A simple bash script used to manage boilerplates

$ mani list projects --tree
┌─ frontend
│  ├─ pinto
│  └─ dashgrid
└─ template-generator
```

### Describe Task

```bash
$ mani describe task git-overview

Name: git-overview
Description: show branch, local and remote diffs, last commit and date
Theme: custom
Target:
    All: true
    Cwd: false
    Projects:
    Paths:
    Tags:
Spec:
    Output: table
    Parallel: true
    IgnoreError: false
    OmitEmpty: false
Commands:
     - git-branch: show current git branch
     - git-last-commit-msg: show last commit
     - git-last-commit-date: show last commit date
```

### Run a Task Targeting Projects with Tag `node` and Output Table

```bash
$ mani run npm-install --tags node

TASK [npm-install: run npm install in node repos] *********************************

pinto |
pinto | up to date, audited 911 packages in 928ms
pinto |
pinto | 71 packages are looking for funding
pinto |   run `npm fund` for details
pinto |
pinto | 15 vulnerabilities (9 moderate, 6 high)
pinto |
pinto | To address issues that do not require attention, run:
pinto |   npm audit fix
pinto |
pinto | To address all issues (including breaking changes), run:
pinto |   npm audit fix --force
pinto |
pinto | Run `npm audit` for details.

TASK [npm-install: run npm install in node repos] *********************************

dashgrid |
dashgrid | up to date, audited 960 packages in 1s
dashgrid |
dashgrid | 87 packages are looking for funding
dashgrid |   run `npm fund` for details
dashgrid |
dashgrid | 14 vulnerabilities (2 low, 2 moderate, 10 high)
dashgrid |
dashgrid | To address all issues possible (including breaking changes), run:
dashgrid |   npm audit fix --force
dashgrid |
dashgrid | Some issues need review, and may require choosing
dashgrid | a different dependency.
dashgrid |
dashgrid | Run `npm audit` for details.
```

### Run Custom Command for All Projects

```bash
$ mani exec --all --output table --parallel 'find . -type f | wc -l'

 Project            | Output
--------------------+--------
 example            | 31016
 pinto              | 14444
 dashgrid           | 16527
 template-generator | 42
```
