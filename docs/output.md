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
