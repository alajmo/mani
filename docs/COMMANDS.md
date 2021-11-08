# Commands

A collection of commands.

## Git

```yaml
tasks:
  # Work

  sync:
    desc: update all of your branches set to track remote ones
    cmd: |
      branch=$(git rev-parse --abbrev-ref HEAD)

      git remote update
      git rebase origin/$branch

  git-status:
    desc: show status
    cmd: git status

  git-checkout:
    desc: switch branch
    env:
      branch: main
    cmd: git checkout $branch

  git-create-branch:
    desc: create branch
    env:
      branch: main
    cmd: git checkout -b $branch

  git-stash:
    desc: store uncommited changes
    cmd: git stash

  git-merge-long-lived-branch:
    desc: merges long-lived branch
    cmd: |
      git checkout $new_branch
      git merge -s ours $old_branch
      git checkout $old_branch
      git merge $new_branch

  git-replace-branch:
    desc: force replace one branch with another
    cmd: |
      git push -f origin $new_branch:$old_branch

  # Update

  git-fetch:
    desc: fetch remote update
    cmd: git fetch

  git-pull:
    desc: pull remote updates and rebase
    cmd: git pull --rebase

  git-pull-rebase:
    desc: pull remote updates
    cmd: git pull

  git-set-url:
    desc: Set remote url
    env:
      base: git@github.com:alajmo
    cmd: |
      repo=$(basename "$PWD")
      git remote set-url origin "$base/$repo.git"

  git-set-upstream-url:
    desc: set upstream url
    cmd: |
      current_branch=$(git rev-parse --abbrev-ref HEAD)
      git branch --set-upstream-to="origin/$current_branch" "$current_branch"

  # Clean

  git-reset:
    desc: reset repo
    env:
      args: ''
    cmd: git reset $args

  git-clean:
    desc: remove all untracked files/folders
    cmd: git clean -dfx

  git-prune-local-branches:
    desc: remove local branches which have been deleted on remote
    env:
      remote: origin
    cmd: git remote prune $remote

  git-delete-branch:
    desc: deletes local and remote branch
    cmd: |
      git branch -D $branch
      git push origin --delete $branch

  git-maintenance:
    desc:  Clean up unnecessary files and optimize the local repository
    cmd: git maintenance run --auto

  # Branch Info

  git-current-branch:
    desc: print current branch
    cmd: git rev-parse --abbrev-ref HEAD

  git-branch-all:
    desc: show git branches, remote and local
    commands:
      - name: all
        cmd: git branch -a -vv

      - name: local
        cmd: git branch

      - name: remote
        cmd: git branch -r

  git-branch-merge-status:
    desc: show merge status of branches
    commands:
      - name: merged
        env:
          branch: ""
        cmd: git branch -a --merged $branch

      - name: unmerged
        env:
          branch: ""
        cmd: git branch -a --no-merged $branch

  git-branch-activity:
    desc: list branches ordered by most recent commit
    commands:
      - name: branch
        cmd: git for-each-ref --sort=committerdate refs/heads/ --format='%(HEAD) %(refname:short)'

      - name: commit
        cmd: git for-each-ref --sort=committerdate refs/heads/ --format='%(objectname:short)'

      - name: message
        cmd: git for-each-ref --sort=committerdate refs/heads/ --format='%(contents:subject)'

      - name: author
        cmd: git for-each-ref --sort=committerdate refs/heads/ --format='%(authorname)'

      - name: date
        cmd: git for-each-ref --sort=committerdate refs/heads/ --format='(%(color:green)%(committerdate:relative)%(color:reset))'

  # Commit Info

  git-head:
    desc: show log information of HEAD
    cmd: git log -1 HEAD

  git-log:
    desc: show 3 latest logs
    env:
      n: 3
    cmd: git --no-pager log --decorate --graph --oneline -n $n

  git-log-full:
    desc: show detailed logs
    cmd: git --no-pager log --color --graph --pretty=format:'%Cred%h%Creset -%C(yellow)%d%Creset %s %Cgreen(%cr) %C(bold blue)<%an>%Creset' --abbrev-commit

  git-show-commit:
    desc: show detailed commit information
    env:
      commit: ''
    cmd: git show $commit

  # Remote Info

  git-remote:
    desc: show remote settings
    cmd: git remote -v

  # Tags

  git-tags:
    desc: show tags
    cmd: git tag -n

  git-tags-newest:
    desc: get the newest tag
    cmd: git describe --tags

  # Author

  git-show-author:
    desc: show number commits per author
    cmd: git shortlog -s -n --all --no-merges

  # Diff

  git-diff-stats:
    desc: git display differences
    cmd: git diff

  git-diff-stat:
    desc: show edit statistics
    cmd: git diff --stat

  git-difftool:
    desc: show differences using a tool
    cmd: git difftool

  # Misc

  git-overview:
    desc: "show # commits, # branches, # authors, last commit date"
    commands:
      - name: "# commits"
        cmd: git rev-list --all --count

      - name: "# branches"
        cmd: git branch | wc -l

      - name: "# authors"
        cmd: git shortlog -s -n --all --no-merges | wc -l

      - name: last commit
        cmd: git log -1 --pretty=%B

      - name: commit date
        cmd: git log -1 --format="%cd (%cr)" -n 1 --date=format:"%d  %b %y" | sed 's/ //'

  git-daily:
    desc: show branch, local and remote diffs, last commit and date
    commands:
      - name: branch
        cmd: git rev-parse --abbrev-ref HEAD

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
