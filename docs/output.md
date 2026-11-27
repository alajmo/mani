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

The `json` and `yaml` formats are meant for scripts and tools like `jq` and `yq`. Nothing but the result is written to stdout, there is no spinner or task header.

Each project results in one record:

| Field       | Description                                                                                                                         |
| ----------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `project`   | Name of the project                                                                                                                 |
| `tasks`     | Names of the tasks and sub-commands that make up the run, the same names as the columns in `table` output                           |
| `output`    | Output lines of all commands in the order they ran, for each command stdout is followed by stderr                                   |
| `exit_code` | `0` if all commands succeeded, otherwise the exit code of the first failing command (`-1` if it could not be started or was killed) |

`mani` itself exits with `0` even if a command fails, as it does for the other output formats, so check `exit_code` to find failures:

```bash
$ mani run test --all -o json | jq -r '.[] | select(.exit_code != 0) | .project'
```

With `--dry-run`, `output` contains the commands that would have been run.

### Parallel Execution

When tasks run in parallel (`--parallel` or `parallel: true` in the task spec), results are written as soon as each project finishes, in the order they finish:

- `json` is written as one object per line instead of an array

  ```bash
  $ mani run hello foo --all --parallel -o json
  {"project":"test-2","tasks":["hello","foo"],"output":["world","bar"],"exit_code":0}
  {"project":"test","tasks":["hello","foo"],"output":["world","bar"],"exit_code":0}

  $ mani run hello foo --all --parallel -o json | jq -r '.project'
  ```

- `yaml` is always written as one document per project, separated by `---`

### List Commands

`mani list projects`, `mani list tasks` and `mani list tags` also support `--output json` and `--output yaml`. All fields are included, the `--headers` and `--tree` flags only apply to the other formats. An empty result is `[]` in `json` and no output in `yaml`.

```bash
$ mani list projects -o json
[
  {
    "name": "test",
    "path": "/home/user/projects/test",
    "rel_path": "test",
    "description": "",
    "url": "https://github.com/user/test",
    "tags": ["frontend"]
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
```

## Omit Empty Table Rows and Columns

Omit empty outputs using `--omit-empty-rows` and `--omit-empty-columns` flags or task spec. Works for tables, markdown and html formats.

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
