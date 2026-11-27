# Output Format

`mani` supports different output formats for tasks. By default it will use `stream` output, but it's possible to change this via the `--output` flag or specify it in the task `spec`.

The following output formats are available:

- **stream** (default)

  ```
  TASK (1/2) [hello] ***********

  test | world
  test | bar

  TASK (2/2) [foo] ***********

  test | world
  test | bar
  ```

- **table**
  ```
   Project  │ Hello │ Foo
  ──────────┼───────┼───────
   test     │ world │ bar
  ──────────┼───────┼───────
   test-2   │ world │ bar
  ```
- **html**
  ```html
  <table class="">
    <thead>
      <tr>
        <th>project</th>
        <th>hello</th>
        <th>foo</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td>test</td>
        <td>world</td>
        <td>bar</td>
      </tr>
      <tr>
        <td>test-2</td>
        <td>world</td>
        <td>bar</td>
      </tr>
    </tbody>
  </table>
  ```
- **markdown**
  ```markdown
  | project | hello | foo |
  | ------- | ----- | --- |
  | test    | world | bar |
  | test-2  | world | bar |
  ```
- **json**
  ```json
  [
    {
      "project": "test",
      "tasks": ["hello", "foo"],
      "output": ["world", "bar"],
      "exit_code": 0
    },
    {
      "project": "test-2",
      "tasks": ["hello", "foo"],
      "output": ["world", "bar"],
      "exit_code": 0
    }
  ]
  ```
- **jsonl**
  ```json
  {"project":"test","tasks":["hello","foo"],"output":["world","bar"],"exit_code":0}
  {"project":"test-2","tasks":["hello","foo"],"output":["world","bar"],"exit_code":0}
  ```
- **yaml**
  ```yaml
  project: test
  tasks:
    - hello
    - foo
  output:
    - world
    - bar
  exit_code: 0
  ---
  project: test-2
  tasks:
    - hello
    - foo
  output:
    - world
    - bar
  exit_code: 0
  ```

## JSON and YAML

The `json`, `jsonl` and `yaml` formats are meant for scripts and tools like `jq` and `yq`. Nothing but the result is written to stdout, there is no spinner or task header.

| Format  | Result                                    | Written                                              |
| ------- | ----------------------------------------- | ---------------------------------------------------- |
| `json`  | A single array, in the order of projects  | When all projects are done                           |
| `jsonl` | One object per line                       | As each project finishes, in the order they finish   |
| `yaml`  | One document per project, separated by `---` | As each project finishes, in the order they finish |

The format is the same whether the tasks run in parallel or not. Use `jsonl` or `yaml` to get the results of a long run as they come in, or to keep the results of the projects that finished if the run is interrupted.

Each project results in one record:

| Field       | Description                                                                                                                          |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `project`   | Name of the project                                                                                                                  |
| `tasks`     | Names of the tasks and sub-commands that make up the run, the same names as the columns in `table` output                            |
| `output`    | Output lines of all commands in the order they ran, for each command stdout is followed by stderr. The lines are not grouped by task |
| `exit_code` | `0` if all commands succeeded, otherwise the exit code of the first failing command (`-1` if it could not be started or was killed)  |

`mani` itself exits with `0` even if a command fails, as it does for the other output formats, so check `exit_code` to find failures:

```bash
$ mani run test --all -o json | jq -r '.[] | select(.exit_code != 0) | .project'

$ mani run test --all --parallel -o jsonl | jq -r 'select(.exit_code != 0) | .project'
```

A few things to note:

- `tasks` lists all tasks of the run, also the ones that were not run because an earlier task failed
- With `--dry-run`, `output` contains the commands that would have been run
- With `--omit-empty-rows`, projects without output don't result in a record. `--omit-empty-columns` has no effect
- Projects that are skipped with `--ignore-non-existing` don't result in a record
- A task that runs with `tty` enabled can't be used with these formats, since it replaces the `mani` process

### List Commands

`mani list projects`, `mani list tasks` and `mani list tags` support `--output json` and `--output yaml`. The fields are always the ones shown below, the `--headers` and `--tree` flags only apply to the other formats.

`json` is a single array, and `[]` when there is nothing to list. `yaml` is one document per item, and no output when there is nothing to list.

```bash
$ mani list projects -o json
[
  {
    "name": "test",
    "path": "/home/user/projects/test",
    "rel_path": "test",
    "description": "",
    "url": "https://github.com/user/test",
    "tags": ["frontend"],
    "worktrees": [
      {
        "path": "worktrees/feature",
        "branch": "feature"
      }
    ]
  }
]

$ mani list tasks -o json
[
  {
    "name": "hello",
    "description": "Print hello",
    "spec": "default",
    "target": "default"
  }
]

$ mani list tags -o yaml
name: frontend
projects:
  - test
---
name: cli
projects:
  - test-2
```

## Omit Empty Table Rows and Columns

Omit empty outputs using `--omit-empty-rows` and `--omit-empty-columns` flags or task spec. Works for tables, markdown and html formats. For `json`, `jsonl` and `yaml`, `--omit-empty-rows` leaves out the projects without output.

See below for an example:

```bash
$ mani run cmd-1 cmd-2 -s project-1,project-2 -o table

  Project  │ Cmd-1 │ Cmd-2
 ──────────┼───────┼───────
  test     │       │
 ──────────┼───────┼───────
  test-2   │ world │

$ mani run test -s project-1,project-2 -o table --omit-empty-rows --omit-empty-columns

TASKS *******************************

 Project | Cmd-1
---------+--------
 test-2  | world
```
