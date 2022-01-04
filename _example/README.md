# Example

This is an example of how you can use `mani`. Simply save the following content to a file named `mani.yaml` and then run `mani sync` to clone all the repositories.

You can then run some of the [commands](#commands).

```yaml
projects:
  example:
    path: .
    desc: A mani example

  pinto:
    path: frontend/pinto
    url: https://github.com/alajmo/pinto.git
    desc: A vim theme editor
    env:
      default_branch: main
    tags: [frontend, node]

  dashgrid:
    path: frontend/dashgrid
    url: https://github.com/alajmo/dashgrid.git
    desc: A highly customizable drag-and-drop grid
    tags: [lib, node]

  template-generator:
    url: https://github.com/alajmo/template-generator.git
    desc: A simple bash script used to manage boilerplates
    tags: [cli, bash]

tasks:
  git-status:
    desc: show working tree status
    output: table
    parallel: true
    cmd: git status

  git-fetch:
    desc: fetch remote updates
    cmd: git fetch

  git-prune:
    desc: remove local branches which have been deleted on remote
    env:
      remote: origin
    cmd: git remote prune $remote

  git-switch:
    desc: switch branch
    env:
      branch: main
    cmd: git checkout $branch

  git-create:
    desc: create branch
    cmd: git checkout -b $branch

  git-delete:
    desc: delete branch
    cmd: git branch -D $branch

  npm-install:
    desc: run npm install in node repos
    target:
      tags: [node]
    parallel: true
    cmd: npm ci

  npm-setup:
    desc: run npm install and build in node repos
    target:
      tags: [node]
    cmd: |
      npm ci
      npm build

  git-daily:
    desc: show branch, local and remote diffs, last commit and date
    commands:
      - name: branch
        cmd: git rev-parse --abbrev-ref HEAD

      - name: status
        cmd: git status

      - name: local diff
        cmd: git diff --name-only | wc -l

      - name: remote diff
        cmd: |
          current_branch=$(git rev-parse --abbrev-ref HEAD)
          git diff "$current_branch" "origin/$current_branch" --name-only 2> /dev/null | wc -l

      - name: last commit
        cmd: git log -1 --pretty=%B

      - name: commit date
        cmd: git log -1 --format="%cd (%cr)" -n 1 --date=format:"%d  %b %y" | sed 's/ //'
```

## Commands

List all projects as table or tree:
```bash
$ mani list projects
┌────────────────────┬────────────────┬──────────────────────────────────────────────────┐
│ name               │ tags           │ description                                      │
├────────────────────┼────────────────┼──────────────────────────────────────────────────┤
│ example            │                │ A mani example                                   │
│ pinto              │ frontend, node │ A vim theme editor                               │
│ dashgrid           │ lib, node      │ A highly customizable drag-and-drop grid         │
│ template-generator │ cli, bash      │ A simple bash script used to manage boilerplates │
└────────────────────┴────────────────┴──────────────────────────────────────────────────┘

$ mani tree projects
┌─ frontend
│  ├─ dashgrid
│  └─ pinto
└─ template-generator
```

Describe a task:
```bash
$ mani describe task git-daily

Name:            git-daily
Description:     show branch, local and remote diffs, last commit and date
Shell:           sh -c
Env:
Commands:
 - Name:         Branch
   Description:
   Shell:        sh -c
   Env:
   Command:      git rev-parse --abbrev-ref HEAD

 - Name:         L Diff
   Description:
   Shell:        sh -c
   Env:
   Command:      git diff --name-only | wc -l

 - Name:         R Diff
   Description:
   Shell:        sh -c
   Env:
   Command:      current_branch=$(git rev-parse --abbrev-ref HEAD)
                 git diff "$current_branch" "origin/$current_branch" --name-only 2> /dev/null | wc -l


 - Name:         Last commit
   Description:
   Shell:        sh -c
   Env:
   Command:      git log -1 --pretty=%B

 - Name:         Commit date
   Description:
   Shell:        sh -c
   Env:
   Command:      git log -1 --format="%cd (%cr)" -n 1 --date=format:"%d  %b %y" | sed 's/ //'
```

Run a task targeting projects with tag `node` and output results as a table:
```bash
$ mani run git-status -t node

Name:         git-status
Description:  show working tree status
Shell:        sh -c
Env:
Command:      git status

┌──────────┬─────────────────────────────────────────────────┐
│ Project  │ git-status                                      │
├──────────┼─────────────────────────────────────────────────┤
│ pinto    │ On branch master                                │
│          │ Your branch is up to date with 'origin/master'. │
│          │                                                 │
│          │ nothing to commit, working tree clean           │
├──────────┼─────────────────────────────────────────────────┤
│ dashgrid │ On branch master                                │
│          │ Your branch is up to date with 'origin/master'. │
│          │                                                 │
│          │ nothing to commit, working tree clean           │
└──────────┴─────────────────────────────────────────────────┘
```

Run custom `ls` command for projects with tag bash:
```bash
$ mani exec 'ls' --paths frontend

[pinto] ****************************************************************************************

bin           LICENSE       package-lock.json  README.md    src
CHANGELOG.md  package.json  postcss.config.js  screenshots  vite.config.js

[dashgrid] *************************************************************************************

CHANGELOG  demo  dist  LICENSE  package.json  package-lock.json  README.md  specs  src
```
