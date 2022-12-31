# Roadmap

`mani` is under active development. Before **v1.0.0**, I want to finish the following tasks, some miscellaneous fixes and improve code documentation:

- [x] Add filter tags/projects/paths for `mani sync`
- [x] Add check sub-command
- [x] Add a way to disable spinner on run/exec sub-commands

- [ ] Bring changes from `sake`
  - Add auto-complete description for editing/listing/describing projects
  - Refactor import logic and support recursive nesting of tasks
  - Improve themes (prefix for text output)
  - Fix correct exit code for tasks
  - Remove `commands` and put functionality in `cmd` property which will accepts string or list of tasks/cmds
  - Add flag ignore-errors and any-errors-fatal

- [ ] Update documentation
  - markdown docs
  - manpage

- [ ] Improve output
  - Add new table format output (tasks in 1st column, output in 2nd, one table per server)
  - Add new format output (tasks in column, project output in row)

- [ ] Improve tasks
  - [ ] Return correct error exit codes when running tasks
    - [ ] serial playbook
    - [ ] parallel playbook
  - [ ] Omit certain tasks from auto-completion and/or being called directly (mainly tasks which are called by other tasks)
  - [ ] Repress certain task output if exit code is 0, otherwise displayed
  - [ ] Summary of task execution at the end
  - [ ] Pass environment variables between tasks
  - [ ] Access exit code of the previous task
  - [ ] Conditional task execution
  - [ ] Tags/servers filtering launching different comands on different projects #6
  - [ ] Ensure command (check existence of file, software, etc.)
  - [ ] on-error/on-success task execution
  - [ ] Cleanup task that is always ran
  - [ ] Log task execution to a file
  - [ ] Abort if certain env variables are not present (required envs)
  - [ ] Add --step mode flag or config setting to prompt before executing a task
  - [ ] Add yaml to command mapper
  - [ ] Add task aliases
  - [ ] Add project/task dependency (dependsOn), to wait for certain projects to finish before starting tasks for other projects

## Future

After **v1.0.0**, focus will be directed to implementing a `tui` for `mani`. The idea is to create something similar to `k9s`, where you have can peruse your projects, tasks, and tags via a `tui`, and execute tasks for selected projects.
