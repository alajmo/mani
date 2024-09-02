# Variables

`mani` supports setting variables for both projects and tasks. Variables can be either strings or commands (encapsulated by $()) that will be evaluated once for each task.

```yaml
projects:
  pinto:
    path: pinto
    url: https://github.com/alajmo/pinto.git
    env:
      foo: bar

tasks:
  ping:
    cmd: |
      echo "$msg"
      echo "$date"
      echo "$foo"
    env:
      msg: text
      date: $(date)
```

## Pass Variables from CLI

To pass variables from the command line, provide them as arguments. For example:

```bash
mani run msg option=123
```

The environment variable option will then be available for use within the task.
