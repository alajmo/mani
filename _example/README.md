# Example

A demo of mani.

![example output of mani](../res/output.gif)

The demo is based on the following mani config.

`mani.yaml`
```yaml
projects:
  - name: example
    path: .
    description: A mani example

  - name: pinto
    path: frontend/pinto
    url: https://github.com/alajmo/pinto.git
    description: A vim theme editor
    tags: [frontend, node]

  - name: dashgrid
    path: frontend/dashgrid
    url: https://github.com/alajmo/dashgrid.git
    description: A highly customizable drag-and-drop grid
    tags: [lib, node]

  - name: template-generator
    url: https://github.com/alajmo/template-generator.git
    description: A simple bash script used to manage boilerplates
    tags: [cli, bash]

tasks:
  - name: git-status
    description: show working tree status
    command: git status

  - name: git-fetch
    description: fetch remote updates
    command: git fetch

  - name: git-prune
    description: remove local branches which have been deleted on remote
    env:
      remote: origin
    command: git remote prune $remote

  - name: git-switch
    description: switch branch
    env:
      branch: main
    command: git checkout $branch

  - name: git-create
    description: create branch
    command: git checkout -b $branch

  - name: git-delete
    description: delete branch
    command: git branch -D $branch

  - name: npm-install
    description: run npm install in node repos
    tags: [node]
    command: npm ci

  - name: npm-setup
    description: run npm install and build in node repos
    tags: [node]
    command: |
      npm ci
      npm build

  - name: git-daily
    description: show branch, local and remote diffs, last commit and date
    commands:
      - name: branch
        command: git rev-parse --abbrev-ref HEAD

      - name: status
        command: git status

      - name: local diff
        command: git diff --name-only | wc -l

      - name: remote diff
        command: |
          current_branch=$(git rev-parse --abbrev-ref HEAD)
          git diff "$current_branch" "origin/$current_branch" --name-only 2> /dev/null | wc -l

      - name: last commit
        command: git log -1 --pretty=%B

      - name: commit date
        command: git log -1 --format="%cd (%cr)" -n 1 --date=format:"%d  %b %y" | sed 's/ //'
```

Given the above `mani.yaml` we can run commands like:

Initialize mani, any sub-directory that has a `.git` inside it will be included:
```bash
$ mani init
✓ Initialized mani repository in /home/samir/tmp
```

Sync repositories (will clone any repository that is not cloned yet):
```bash
$ mani sync
pinto

Cloning into '/home/samir/tmp/frontend/pinto'...
remote: Enumerating objects: 1003, done.
remote: Counting objects: 100% (236/236), done.
remote: Compressing objects: 100% (175/175), done.
remote: Total 1003 (delta 94), reused 135 (delta 53), pack-reused 767
Receiving objects: 100% (1003/1003), 4.56 MiB | 10.55 MiB/s, done.
Resolving deltas: 100% (389/389), done.

dashgrid

Cloning into '/home/samir/tmp/frontend/dashgrid'...
remote: Enumerating objects: 790, done.
remote: Counting objects: 100% (34/34), done.
remote: Compressing objects: 100% (27/27), done.
remote: Total 790 (delta 22), reused 10 (delta 7), pack-reused 756
Receiving objects: 100% (790/790), 756.73 KiB | 6.58 MiB/s, done.
Resolving deltas: 100% (409/409), done.

template-generator

Cloning into '/home/samir/tmp/template-generator'...
remote: Enumerating objects: 188, done.
remote: Counting objects: 100% (121/121), done.
remote: Compressing objects: 100% (75/75), done.
remote: Total 188 (delta 72), reused 91 (delta 43), pack-reused 67
Receiving objects: 100% (188/188), 133.67 KiB | 1.59 MiB/s, done.
Resolving deltas: 100% (94/94), done.
All projects synced
```

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

$ mani tree
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
$ mani run git-status -t node --output table

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
$ mani exec 'ls' --dirs frontend

pinto
bin
CHANGELOG.md
LICENSE
package.json
package-lock.json
postcss.config.js
README.md
screenshots
src
vite.config.js


dashgrid
CHANGELOG
demo
dist
LICENSE
package.json
package-lock.json
README.md
specs
src
```
