# Changelog

## v0.12.0

### Features

- Add option to omit empty results
- Add --vcs flag to mani init to choose vcs
- Add default import from user config directory
- [BREAKING CHANGE]: Add spec property to allow reusing common properties
- Add target property to allow reusing common properties

### Fixes

- Fix header bug in run print when task has both commands and cmd
- Fix `mani edit` to run even if config file is malformed (wrong YAML syntax)

### Misc

- [BREAKING CHANGE]: Move tree feature to list projects as a flag instead of it being a special sub-command
- [BREAKING CHANGE]: Rename flag --all-projects to all
- Remove legacy code related to Dirs entity
- Change default value of --parallel flag to false when syncing
- Allow omitting the -c when specifying shell for bash, zsh, sh, node and python

## v0.11.1

### Features

- Add `env` property to projects to enable project specific variables

### Fixes

- Use syncmap to allow safe concurrent writes when running `mani sync` in parallel, previously there was a race condition that occurred when cloning many repos

## v0.10.0

### Features

- Add ability to import projects, tasks and themes
- Possible to run tasks in parallel now per each project
- Add sub-commands project/task to edit command to open editor at line corresponding to project/task
- Add edit flag to describe/run sub-commands to open up editor
- Sync projects in parallel by default and add flag serial to opt out
- Add support for referencing commands in Commands property
- Run commands in serial, if one fails, dont run other tasks
- Add directory entity, similar to project, just without a url/clone property

### Misc

- Add new acceptable filenames Manifile, Manifile.yaml, Manifile.yml
- Don't create .gitignore if no projects with url exists on mani init/sync
- List tags now shows associated dirs/projects
- If user uses a cwd/tag/project/dir flag, then disable task targets
- [BREAKING CHANGE:] A lot of syntax changes, use object notation instead of array list for projects, themes and tasks

## v0.6.1

### Features

- Add dirs filtering property to commands struct

### Fixes

- Correct project path in gitignore file when running mani init

### Misc

- Update help text for dirs flag

## v0.6.0

### Features

- New tree command that list contents of projects in a tree-like format
- Add filtering on directory for tree/list/describe/run/exec cmd
- Add global environment variables
- Add describe flag to run cmd to suppress command information
- Add sub-commands
- Add possibility to run multiple commands from cli
- Add default tags/projects/output to tasks
- Add new table style that can be configured only from mani config
- Add progress spinner for run/exec cmd

### Misc

- [BREAKING CHANGE]: Renamed args in command block to env
- [BREAKING CHANGE]: Renamed commands in root block to tasks
- Environment variables now support shell execution
- Rename flag format to output when listing

## v0.5.1

### Fixes

- Fix auto-complete for flag format in list command

## v0.5.0

### Features

- Add MANI environment variable that is cwd of the current context mani.yaml file
- Add mani edit command which opens mani.yaml in preferred editor
- Add describe cmd, display commands and projects in detail
- Append default shell to commands
- Add output formats table, markdown and html
- Add no-borders, no-headers flags to print
- Allow users to specify headers to be printed in list command
- Sync creates gitignore file if not found
- Use CLI spinner when syncing projects
- Update info cmd to print git version

### Fixes

- Output args at top for run commands instead of for each run
- Output error message when running commands in non-mani directory that require mani config

### Misc

- Refactor and make code more DRY
- Refactor list and describe cmd to use sub-commands
- With no projects to sync, output helpful message: "No projects to sync"
- With all projects synced, output helpful message: "All projects synced"

## v0.4.0

### Features

- Allow users to set global and command level shell commands

## v0.3.0

### Features

- Add support for running from nested sub-directories
- Add info sub-command that shows which configuration file is being used
- Add flag to point to config file
- Accept different config names (.mani, .mani.yaml, .mani.yml, mani.yaml, mani.yml)
- Add new command exec to run arbitrary command
- Add config flag
- Add first argument to init should be path, if empty, current dir
- Add completion for all commands bash
- Update auto-discovery to equal true by default
- Add option to filter list command on tags and projects
- Add Nicer output on failed git sync
- Add cwd flag to target current directory
- Add comment section in .gitignore so users can modify the gitignore without mani overwriting all parts
- Improved listing for projects/tags

### Fixes

- Fix crashing on not found config file
- Check possible, non-handled nil/err values
- Don't add project to gitignore if doesn't have a url
- Remove path if path is same as name
- Fix gitignore sync, removing old entries
- Fix broken init command
- Fix so path accepts environment variables
- Fix auto-complete when not in mani directory

### Misc

- Update golang version and dependencies
- Add integration tests
