# Error Handling

## Ignoring Task Errors

If you wish to continue task execution even if an error is encountered, use the flag `--ignore-errors` or specify it in the task `spec`.

- `ignore-errors` set to false

  ```bash
  $ mani run task-1 task-2 --all --ignore-errors=false
     Project    | Task-1        | Task-2
    ------------+---------------+--------
     project-0  |               |
                | exit status 1 |
    ------------+---------------+--------
     project-1  |               |
                | exit status 1 |
    ------------+---------------+--------
     project-2  |               |
                | exit status 1 |
  ```

- `ignore-errors` set to true
  ```bash
     Project    | Task-1        | Task-2
    ------------+---------------+--------
     project-0  |               | bar
                | exit status 1 |
    ------------+---------------+--------
     project-1  |               | bar
                | exit status 1 |
    ------------+---------------+--------
     project-2  |               | bar
                | exit status 1 |
  ```

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
