# Commands

A collection of various commands.

## Git

```yaml
tasks:
  # Work

  - name: sync
    description: update all of your branches set to track remote ones
    command: |
      branch=$(git rev-parse --abbrev-ref HEAD)

      git remote update
      git rebase origin/$branch

  - name: git-status
    description: show status
    command: git status

  - name: git-checkout
    description: switch branch
    env:
      branch: main
    command: git checkout $branch

  - name: git-create-branch
    description: create branch
    env:
      branch: main
    command: git checkout -b $branch

  - name: git-stash
    description: store uncommited changes
    command: git stash

  - name: git-merge-long-lived-branch
    description: merges long-lived branch
    command: |
      git checkout $new_branch
      git merge -s ours $old_branch
      git checkout $old_branch
      git merge $new_branch

  - name: git-replace-branch
    description: force replace one branch with another
    command: |
      git push -f origin $new_branch:$old_branch

  # Update

  - name: git-fetch
    description: fetch remote update
    command: git fetch

  - name: git-pull
    description: pull remote updates and rebase
    command: git pull --rebase

  - name: git-pull-rebase
    description: pull remote updates
    command: git pull

  - name: git-set-url
    description: Set remote url
    env:
      base: git@github.com:alajmo
    command: |
      repo=$(basename "$PWD")
      git remote set-url origin "$base/$repo.git"

  - name: git-set-upstream-url
    description: set upstream url
    command: |
      current_branch=$(git rev-parse --abbrev-ref HEAD)
      git branch --set-upstream-to="origin/$current_branch" "$current_branch"

  # Clean

  - name: git-reset
    description: reset repo
    env:
      args: ''
    command: git reset $args

  - name: git-clean
    description: remove all untracked files/folders
    command: git clean -dfx

  - name: git-prune-local-branches
    description: remove local branches which have been deleted on remote
    env:
      remote: origin
    command: git remote prune $remote

  - name: git-delete-branch
    description: deletes local and remote branch
    command: |
      git branch -D $branch
      git push origin --delete $branch

  - name: git-maintenance
    description:  Clean up unnecessary files and optimize the local repository
    command: git maintenance run --auto

  # Branch Info

  - name: git-current-branch
    description: print current branch
    command: git rev-parse --abbrev-ref HEAD

  - name: git-branch-all
    description: show git branches, remote and local
    commands:
      - name: all
        command: git branch -a -vv

      - name: local
        command: git branch

      - name: remote
        command: git branch -r

  - name: git-branch-merge-status
    description: show merge status of branches
    commands:
      - name: merged
        env:
          branch: ""
        command: git branch -a --merged $branch

      - name: unmerged
        env:
          branch: ""
        command: git branch -a --no-merged $branch

  - name: git-branch-activity
    description: list branches ordered by most recent commit
    commands:
      - name: branch
        command: git for-each-ref --sort=committerdate refs/heads/ --format='%(HEAD) %(refname:short)'

      - name: commit
        command: git for-each-ref --sort=committerdate refs/heads/ --format='%(objectname:short)'

      - name: message
        command: git for-each-ref --sort=committerdate refs/heads/ --format='%(contents:subject)'

      - name: author
        command: git for-each-ref --sort=committerdate refs/heads/ --format='%(authorname)'

      - name: date
        command: git for-each-ref --sort=committerdate refs/heads/ --format='(%(color:green)%(committerdate:relative)%(color:reset))'

  # Commit Info

  - name: git-head
    description: show log information of HEAD
    command: git log -1 HEAD

  - name: git-log
    description: show 3 latest logs
    env:
      n: 3
    command: git --no-pager log --decorate --graph --oneline -n $n

  - name: git-log-full
    description: show detailed logs
    command: git --no-pager log --color --graph --pretty=format:'%Cred%h%Creset -%C(yellow)%d%Creset %s %Cgreen(%cr) %C(bold blue)<%an>%Creset' --abbrev-commit

  - name: git-show-commit
    description: show detailed commit information
    env:
      commit: ''
    command: git show $commit

  # Remote Info

  - name: git-remote
    description: show remote settings
    command: git remote -v

  # Tags

  - name: git-tags
    description: show tags
    command: git tag -n

  - name: git-tags-newest
    description: get the newest tag
    command: git describe --tags

  # Author

  - name: git-show-author
    description: show number commits per author
    command: git shortlog -s -n --all --no-merges

  # Diff

  - name: git-diff-stats
    description: git display differences
    command: git diff

  - name: git-diff-stat
    description: show edit statistics
    command: git diff --stat

  - name: git-difftool
    description: show differences using a tool
    command: git difftool

  # Misc

  - name: git-overview
    description: "show # commits, # branches, # authors, last commit date"
    commands:
      - name: "# commits"
        command: git rev-list --all --count

      - name: "# branches"
        command: git branch | wc -l

      - name: "# authors"
        command: git shortlog -s -n --all --no-merges | wc -l

      - name: last commit
        command: git log -1 --pretty=%B

      - name: commit date
        command: git log -1 --format="%cd (%cr)" -n 1 --date=format:"%d  %b %y" | sed 's/ //'

  - name: git-daily
    description: show branch, local and remote diffs, last commit and date
    commands:
      - name: branch
        command: git rev-parse --abbrev-ref HEAD

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
