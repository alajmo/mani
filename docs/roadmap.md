# Roadmap

`mani` is under active development. Before **v1.0.0**, I want to finish the following tasks, some miscellaneous fixes and improve code documentation:

- [ ] Add filter tags/projects/paths for `mani sync`
- [ ] Add cd sub-command
- [ ] Add a way to disable spinner on run/exec sub-commands
- [ ] Fix NO_COLOR for `mani init`
- [ ] Add check sub-command from `mani`
- [ ] Add root flag/target to target root mani directory
- [ ] Bring changes from `sake`
  - Add auto-complete description for editing/listing/describing projects
  - Refactor import logic and support recursive nesting of tasks
  - Improve themes (prefix for text output)
  - Fix correct exit code for tasks
  - Remove `commands` and put functionality in `cmd` property which will accepts string or list of tasks/cmds
  - Add flag ignore-errors and any-errors-fatal
- [ ] Add task aliases
- [ ] Add project/task dependency (dependsOn), to wait for certain projects to finish before starting tasks for other projects
- [ ] Update documentation
  - markdown docs
  - manpage
- [ ] Fix windows environment variables

## Future

After **v1.0.0**, focus will be directed to implementing a `tui` for `mani`. The idea is to create something similar to `k9s`, where you have can peruse your projects, tasks, and tags via a `tui`, and execute tasks for selected projects.
