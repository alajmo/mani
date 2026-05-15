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
      "tasks": ["hello"],
      "output": ["world"],
      "exit_code": 0
    },
    {
      "project": "test-2",
      "tasks": ["hello"],
      "output": ["world"],
      "exit_code": 0
    }
  ]
  ```

- **yaml**
  ```yaml
  project: test
  tasks:
    - hello
  output:
    - world
  exit_code: 0
  ---
  project: test-2
  tasks:
    - hello
  output:
    - world
  exit_code: 0
  ```

  YAML format outputs each result as a separate YAML document, making it suitable for streaming output and processing with tools like `yq`.

## Structured Output Benefits

The `json` and `yaml` output formats provide structured data that includes:

- **project**: The name of the project the command was run on
- **tasks**: An array of task names that were executed
- **output**: An array of output lines from the command
- **exit_code**: The exit code of the command (0 for success, non-zero for failure)

This is especially useful for:

- Piping output to `jq` or `yq` for further processing
- Storing results in document databases or structured logs
- Capturing exit codes for each project when running commands across multiple repositories
- Building automation scripts that need to process the results programmatically

## Streaming Output with Parallel Execution

When running with `--parallel`, the `json` and `yaml` output formats use a streaming format where each result is output immediately as it completes:

- **JSON streaming**: Each result is a single-line JSON object followed by a newline
  ```bash
  $ mani run hello --all --output json --parallel
  {"project":"test","tasks":["hello"],"output":["world"],"exit_code":0}
  {"project":"test-2","tasks":["hello"],"output":["world"],"exit_code":0}
  ```

- **YAML streaming**: Each result is a separate YAML document (separated by `---`)
  ```bash
  $ mani run hello --all --output yaml --parallel
  ---
  project: test
  tasks:
    - hello
  output:
    - world
  exit_code: 0
  ---
  project: test-2
  tasks:
    - hello
  output:
    - world
  exit_code: 0
  ```

This streaming format is ideal for processing results as they complete without waiting for all tasks to finish.

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
