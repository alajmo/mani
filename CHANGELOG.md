# Changelog

## v0.5.0 - unreleased

### Added

- Add MANI environment variable that is cwd of the current context mani.yaml file
- Add mani edit command which opens mani.yaml in preferred editor
- Add describe cmd, display commands and projects in detail
- Append default shell to commands
- Update info cmd, print git version and number of projects, commands and tags
- Sync creates gitignore file if not found
- Use CLI spinner when syncing projects
- Add output formats table, markdown and html
- Add no-borders, no-headers flags to print
- Allow users to specify headers to be printed in list command
- Give user feedback on all projects synced

### Fixed

- Refactor list and describe cmd to use sub-commands
- Output args at top for run commands instead of for each run

### Misc

- With no projects to sync, output helpful message: "No projects to sync"
- Refactor and make code more DRY

## v0.4.0

### Added

- Allow users to set global and command level shell commands

## v0.3.0

### Misc

- Update golang version and dependencies
- Add integration tests

### Added

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

### Fixed

- Fix crashing on not found config file
- Check possible, non-handled nil/err values
- Don't add project to gitignore if doesn't have a url
- Remove path if path is same as name
- Fix gitignore sync, removing old entries
- Fix broken init command
- Fix so path accepts environment variables
- Fix auto-complete when not in mani directory
