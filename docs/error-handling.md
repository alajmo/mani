# Error Handling

## Ignoring Task Errors

If you wish to continue task execution even if an error is encountered, use the flag `--ignore-errors` or specify it in the task `spec`. This controls whether the *remaining* commands in a multi-command task run after one of them fails on a given project.

Assuming `task-1` exits non-zero (e.g. `false`) and `task-2` prints `bar`:

- `ignore-errors` set to false (default) — task-2 is skipped after task-1 fails

  ```bash
  $ mani run task-1 task-2 --all --ignore-errors=false
     Project    | Task-1 | Task-2
    ------------+--------+--------
     project-0  |        |
    ------------+--------+--------
     project-1  |        |
    ------------+--------+--------
     project-2  |        |
  ```

- `ignore-errors` set to true — task-2 still runs

  ```bash
  $ mani run task-1 task-2 --all --ignore-errors
     Project    | Task-1 | Task-2
    ------------+--------+--------
     project-0  |        | bar
    ------------+--------+--------
     project-1  |        | bar
    ------------+--------+--------
     project-2  |        | bar
  ```

> mani does not annotate cells with the OS exit code (`exit status N`). The command's own stderr is captured into the cell, so real errors from tools like `git`, `npm`, etc. still surface — only the synthetic exit‑code line is suppressed. Combine `ignore_errors` with `--omit-empty-rows` to hide rows for commands that intentionally exit non‑zero with no output (e.g. `grep -v` with no matches).

## Ignoring Non Existing Project

- `--ignore-non-existing` set to false

  ```bash
  $ mani run task-1 --all
    error: path `/home/test/project-1` does not exist
  ```

- `ignore-unreachable` set to true
  ```bash
  $ mani run task-1 --all --ignore-non-existing
     Project    | Task-1
    ------------+--------
     project-0  | hello
    ------------+--------
     project-1  |
    ------------+--------
     project-2  | hello
  ```
