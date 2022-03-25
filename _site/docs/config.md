# Config Reference

The `mani.yaml` config is based on two concepts: __projects__ and __tasks__. __Projects__ are simply directories, which may be git repositories, in which case they have an URL attribute. __Tasks__ are arbitrary shell commands that you write and then run for selected __projects__.

`mani.yaml`
```yaml
# Import projects/tasks/env/themes from other configs
import:
  - some-dir/mani.yaml

# List of Projects
projects:
  # Project name [required]
  pinto:

    # Project path [defaults to project name]
    path: frontend/pinto

    # Project URL [optional]
    url: git@github.com:alajmo/pinto

    # Project description [optional]
    desc: A vim theme editor

    # Clone command [defaults to `git clone URL`]
    clone: git clone git@github.com:alajmo/pinto

    # List of tags [optional]
    tags: [frontend]

    # List of project specific environment variables
    env:
      # Simple string value
      branch: main

    # If project should be synced automatically [default to true]
    sync: true

# List of environment variables that are available to all tasks
env:
  # Simple string value
  AUTHOR: "alajmo"

  # Shell command substitution
  DATE: $(date -u +"%Y-%m-%dT%H:%M:%S%Z")

# Shell used for commands [defaults to "bash"]
shell: bash

# List of themes
themes:
  # Theme name
  default:
    # Available styles: light (default), ascii
    table: ascii

    # Available styles: connected-light (default), connected-bold, bullet-square, bullet-circle, bullet-star
    tree: connected-bold

# List of Specs
specs:
  default:
    # Available options: text/table/html/markdown
    output: text

    # Whether to run the tasks in parallel per project
	parallel: false

    # If ignore_error is set to true and multiple commands are set for a task, then the exit code is not 0
	ignore_error: false

    # If a command returns an empty string, then the project won't be shown
	omit_empty: false

# List of targets
targets:
  all:
    all: true

# List of tasks
tasks:
  # Command name [required]
  simple-1:
    cmd: echo simple

  # Short-form for a command
  simple-2: echo simple

  # Command name [required]
  complex:

    # Task description [optional]
    desc: complex task

    # Specify theme [optional, defaults to default]
    theme: default

    # Shell used for this command [defaults to root shell]
    shell: bash -c

    # List of environment variables
    env:
      # Simple string value
      branch: master

      # Shell command substitution
      num_lines: $(ls -1 | wc -l)

    # Target projects inline
    target:
      # Target all projects [defaults to false]
      all: false

      # Target current working directory [defaults to false]
      cwd: false

      # Target specific projects [defaults to empty list]
      projects: [awesome]

      # Target projects under a directory [defaults to empty list]
      paths: [frontend]

      # Target projects with tags [defaults to empty list]
      tags: [work]

    # Or specify a target defined globally
    # target: all

    # Specify specs inline
    spec:
      # Set default output option, available styles: text(default)/table/html/markdown
      output: table

      # Run tasks per project in parallel [defaults to false]
      parallel: true

      # Run all task commands even if there is an error [defaults to false]
      ignore_error: true

      # If a command returns an empty string, then the project won't be shown in the table
      omit_empty: false

    # Or specify a spec defined globally
    # spec: default

    # Each task can have a single command, multiple commands, OR both

    # Multine command
    cmd: |
      echo complex
      echo command

    # List of commands
    commands:
      # Basic command
      - name: first
        desc: first command
        cmd: echo first

      # Reference another task
      - task: simple
```

## Import

A list of configs to import.

```yaml
import:
  - dir-a/mani.yaml
  - dir-b/tasks.yaml
```

## Env

Global A dictionary of key/value pairs that all `tasks` inherit. The value can either be a simple string:

```yaml
env:
  VERSION: v1.0.0
```

or if it is enclosed within `$()`, shell command substitution takes place.

```yaml
env:
  DATE: $(date)
```

### Default Env

`mani` exposes some variables to each command:

Global:

- `MANI_CONFIG_PATH`: Absolute path of the current mani.yaml file
- `MANI_CONFIG_DIR`: Directory of the current mani.yaml file

Project specific:

- `MANI_PROJECT_NAME`: The name of the project
- `MANI_PROJECT_PATH` The path to the project in absolute form

## Shell

Shell used for commands, it defaults to `sh -c`. Note, you have to provide the flag `-c` for shell programs `bash`, `sh`, etc. if you want a command-line string evaluated.

In case you only want to execute a script file, then the following will do:

```yaml
shell: bash

tasks:
  example:
    cmd: script.sh
```

or

```yaml
shell: bash -c

tasks:
  example:
    cmd: ./script.sh
```

Note, any executable that's in your `PATH` works:

```yaml
shell: node

tasks:
  example:
    cmd: index.js
```

## Themes

List of themes that alter styling of some `mani` commands. By default, all tasks use a `default` theme. To override the default theme, simply create a theme with name `default`.

```yaml
themes:
  default:
    table: ascii # Available styles: light (default), ascii
    tree: bullet-star # Available styles: connected-light (default), connected-bold, bullet-square, bullet-circle, bullet-star
```

## Specs

List of specs. By default, all tasks use a `default` spec. To override the default spec, simply create a spec with name `default`.

```yaml
specs:
  default:
    # Available options: text(default)/table/html/markdown
    output: text

    # Whether to run the tasks in parallel per project (defaults to false)
	parallel: false

    # If ignore_error is set to true and multiple commands are set for a task, then the exit code is not 0. (defaults to false)
	ignore_error: false

    # If a command returns an empty string, then the project won't be shown (defaults to false)
	omit_empty: false
```

## Targets

List of targets. By default, all tasks use a `default` target. To override the default target, simply create a spec with name `default`.

```yaml
targets:
  default:
    projects: []
    paths: []
    tags: []
    all: false,
    cwd: false

  custom:
    all: true
```

## Projects

List of projects that `mani` will operate on.

### Name

The name of the project. This is required for each project.

```yaml
projects:
  example:
    path: work/example
```

### Path

Path to the project, relative to the directory of the config file. It defaults to the name of the project.

```yaml
projects:
  example:
    path: work/example
```

### Url

The URL of the project, which the `mani sync` command will use to download the repository. `mani sync` uses `git clone git@github.com:alajmo/pinto` behind the scenes. So if you want to modify the clone command, check out the [clone](#clone) property.

```yaml
projects:
  example:
    path: git@github.com:alajmo/pinto
```

### Desc

Optional description of the project.

```yaml
projects:
  example:
    desc: an example repository
```

### Clone

Clone command that `mani sync` will use to clone the repository. It defaults to `git clone URL`.

In case you want to do modify the clone command, this is the place to do it. For instance, to only clone a single branch:

```yaml
projects:
  example:
    clone: git clone git@github.com:alajmo/pinto --branch main
```

### Tags

A list of tags to associate the project with.

```yaml
projects:
  example:
    url: git@github.com:alajmo/pinto
    tags: [work, cli]
```

### Env

A dictionary of key/value pairs per project. The value can either be a simple string:

```yaml
env:
  branch: main
```

or if it is enclosed within `$()`, shell command substitution takes place.

```yaml
env:
  DATE: $(date)
```

### Sync

Whether a project should be cloned when running `mani sync`.

```yaml
projects:
  example:
    url: git@github.com:alajmo/pinto
    sync: true
```

## Tasks

List of predefined tasks that can be run on `projects`.

### Name

The `name` of the task. This is required for each task.

```yaml
tasks:
  example:
    cmd: echo 123
```

### Desc

An optional string value that describes the `task`.

```yaml
tasks:
  example:
    desc: print 123
    cmd: echo 123
```

### Shell

The `shell` used for this task commands. Defaults to the root `shell` defined in the global scope (which in turn defaults to `bash`).

```yaml
shell: bash

tasks:
  example:
    cmd: ./script.sh
```

### Env

A dictionary of key/value pairs, see [env](#env).

The `env` field is inherited from the global scope and can be overridden in the `task` definition.

For instance:

```yaml
env:
  VERSION: v1.0.0
  BRANCH: main

tasks:
  example:
    env:
      VERSION: v2.0.0
    cmd: |
      echo $VERSION
      echo $BRANCH
```

Will print:

```sh
$ mani run example
v2.0.0
main
```

### Theme

Specify which `theme` to use. You can either specify an existing theme or declare it inline.

```yaml
themes:
  ascii:
    table: ascii

tasks:
  example:
    # Inline
    theme:
      table: ascii

    # Reference
    # theme: ascii

    cmd: echo 123
```

If no `theme` is specified, then it will use the default `theme`.

### Spec

`Spec` contains miscellaneous properties for each task:

```yaml
specs:
  custom:
    # Available options: text/table/html/markdown
    output: text

    # Whether to run the tasks in parallel per project
	parallel: false

    # If ignore_error is true and task has multiple commands,
    # then it will stop executing if error is encountered
	ignore_error: false

    # If a command returns an empty string, then the project won't be shown
	omit_empty: false

tasks:
  example:
    # Inline
    spec:
      output: table
      parallel: true
      ignore_error: true
      omit_empty: true

    # Reference
    # spec: custom

    cmd: echo 123
```

The spec properties are overriden when running `mani` with a spec flag manually. If no `spec` is specified, then it will use the default `spec`.

### Target

Specify projects by default when running a task. You can either specify an existing target or declare it inline.

```yaml
targets:
  custom:
    projects: [awesome]
    paths: [frontend]
    tags: [work]

tasks:
  example:
    # Inline
    target:
      all: false
      cwd: false
      projects: [awesome]
      paths: [frontend]
      tags: [work]

    # Reference
    # target: custom

    cmd: echo 123
```

This is equivalent to running `mani run example --all=false --cwd=false --projects awesome --paths frontend --tags work`.
The target property is overriden when running `mani` with a target flag manually. If no `target` is specified, then it will use the default `target`.

### Cmd

A single or multiline command that uses the `shell` program to run in each project it's filtered on.

Single-line cmd:
```yaml
tasks:
  example:
    cmd: echo 123
```

Multi-line cmd:
```yaml
tasks:
  example:
    cmd: |
      echo 123
      echo 456
```

Short-form:
```yaml
tasks:
  example: echo 123
```

### Commands

A `task` also supports running multiple commands. In this example, the `first-command` will be run first, and then the `echo` command will run. Both of its outputs will be displayed.

```yaml
tasks:
  echo: echo 123

  example:
    ignore_error: true
    commands:
      - name: first-command
        cmd: echo first

      - task: echo
```
