# Config

The mani.yaml config is based on the following concepts:

- **projects** are directories, which may be git repositories, in which case they have an URL attribute
- **tasks** are shell commands that you write and then run for selected **projects**
- **specs** are configs that alter **task** execution and output
- **targets** are configs that provide shorthand filtering of **projects** when executing **tasks**
- **themes** are used to modify the output of `mani` commands

**Specs**, **targets** and **themes** use a default object by default that the user can override to modify execution of mani commands.

Check the [files](#files) and [environment](#environment) section to see how the config file is loaded.

Below is a config file detailing all of the available options and their defaults.

```yaml
# Import projects/tasks/env/specs/themes/targets from other configs [optional]
import:
  - ./some-dir/mani.yaml

# If this is set to true mani will override the url of any existing remote
sync_remotes: false

# List of Projects
projects:
  # Project name [required]
  pinto:
    # Project path relative to the config file. Defaults to project name [optional]
    path: frontend/pinto

    # Project URL [optional]
    url: git@github.com:alajmo/pinto

    # Project description [optional]
    desc: A vim theme editor

    # Override clone command [defaults to "git clone URL PATH"]
    clone: git clone git@github.com:alajmo/pinto --branch main

    # List of tags [optional]
    tags: [dev]

    # If project should be synced when running mani sync [optional]
    sync: true

    # List of remotes, the key is the name and value is the url [optional]
    remotes:
      somekey: https://github.com/some-other-remote

    # List of project specific environment variables [optional]
    env:
      # Simple string value
      branch: main

      # Shell command substitution
      date: $(date -u +"%Y-%m-%dT%H:%M:%S%Z")

# List of environment variables that are available to all tasks
env:
  # Simple string value
  AUTHOR: "alajmo"

  # Shell command substitution
  DATE: $(date -u +"%Y-%m-%dT%H:%M:%S%Z")

# Shell used for commands [optional]
# If you use any other program than bash, zsh, sh, node, and python
# then you have to provide the command flag if you want the command-line string evaluted
# For instance: bash -c
shell: bash

# List of themes
themes:
  # Theme name
  default:
    # Tree options [optional]
    tree:
      # Tree style [optional]
      # Available options: light, bold, ascii, bullet-square, bullet-circle, bullet-star
      style: light

    # Text options [optional]
    text:
      # Include project name prefix for each line [optional]
      prefix: true

      # Colors to alternate between for each project prefix [optional]
      # Available options: green, blue, red, yellow, magenta, cyan
      prefix_colors: ["green", "blue", "red", "yellow", "magenta", "cyan"]

      # Add a header before each project [optional]
      header: true

      # String value that appears before the project name in the header [optional]
      header_prefix: "TASK"

      # Fill remaining spaces with a character after the prefix [optional]
      header_char: "*"

    # Table options [optional]
    table:
      # Table style [optional]
      # Available options: light, bold, double, rounded, ascii
      style: light

      # Text format options for headers and rows in table output [optional]
      # Available options: default, lower, title, upper
      format:
        header: default
        row: default

      # Border options for table output [optional]
      options:
        draw_border: false
        separate_columns: true
        separate_header: true
        separate_rows: false
        separate_footer: false

      # Color, attr and align options [optional]
      # Available options for fg/bg: green, blue, red, yellow, magenta, cyan
      # Available options for align: left, center, justify, right
      # Available options for attr: normal, bold, faint, italic, underline, crossed_out
      color:
        header:
          project:
            fg:
            bg:
            align: left
            attr: normal

          tag:
            fg:
            bg:
            align: left
            attr: normal

          desc:
            fg:
            bg:
            align: left
            attr: normal

          task:
            fg:
            bg:
            align: left
            attr: normal

          rel_path:
            fg:
            bg:
            align: left
            attr: normal

          path:
            fg:
            bg:
            align: left
            attr: normal

          url:
            fg:
            bg:
            align: left
            attr: normal

          output:
            fg:
            bg:
            align: left
            attr: normal

        row:
          project:
            fg:
            bg:
            align: left
            attr: normal

          tag:
            fg:
            bg:
            align: left
            attr: normal

          desc:
            fg:
            # bg:
            align: left
            attr: normal

          task:
            fg:
            # bg:
            align: left
            attr: normal

          rel_path:
            fg:
            bg:
            align: left
            attr: normal

          path:
            fg:
            bg:
            align: left
            attr: normal

          url:
            fg:
            bg:
            align: left
            attr: normal

          output:
            fg:
            bg:
            align: left
            attr: normal

        border:
          header:
            fg:
            bg:

          row:
            fg:
            bg:

          row_alt:
            fg:
            bg:

          footer:
            fg:
            bg:


# List of Specs [optional]
specs:
  default:
    # The preferred output format for a task
    # Available options: text, table, html, markdown
    output: text

    # Option to run tasks in parallel
    parallel: false

    # If ignore_errors is set to true and multiple commands are set for a task, then the exit code is not 0
    ignore_errors: true

    # If command(s) in result in an empty output, the project row will be hidden
    omit_empty: false

# List of targets [optional]
targets:
  default:
    # Target all projects
    all: false

    # Target current working directory project
    cwd: false

    # Specify projects via project name
    projects: []

    # Specify projects via project path
    paths: []

    # Specify projects via project tags
    tags: []

# List of tasks
tasks:
  # Command name [required]
  simple-1:
    cmd: |
      echo "hello world"
    desc: simple command 1

  # Short-form for a command
  simple-2: echo "hello world"

  # Command name [required]
  advanced-command:
    # Task description [optional]
    desc: complex task

    # Specify theme [optional]
    theme: default

    # Shell used for this command [optional]
    shell: bash

    # List of environment variables [optional]
    env:
      # Simple string value
      branch: master

      # Shell command substitution
      num_lines: $(ls -1 | wc -l)

    # Spec reference [optional]
    # spec: default

    # Or specify specs inline
    spec:
      output: table
      parallel: true
      ignore_errors: false
      omit_empty: true

    # Target reference [optional]
    # target: default

    # Or specify targets inline
    target:
      all: true
      cwd: false
      projects: [pinto]
      paths: [frontend]
      tags: [dev]

    # Each task can have a single command, multiple commands, OR both

    # Multine command
    cmd: |
      echo complex
      echo command

    # List of commands
    commands:
      # Basic command
      - name: inline-command
        cmd: echo "Hello World"

      # Reference another task
      - task: simple-1
```

## Files

When running a command, `mani` will check the current directory and all parent directories for the following files: `mani.yaml`, `mani.yml`, `.mani.yaml`, `.mani.yml` .

Additionally, it will import (if found) a config file from:

- Linux: `$XDG_CONFIG_HOME/mani/config.yaml` or `$HOME/.config/mani/config.yaml` if `$XDG_CONFIG_HOME` is not set.
- Darwin: `$HOME/Library/Application/mani`
- Windows: `%AppData%\mani`

Both the config and user config can be specified via flags or environments variables.

## Environment

```txt
MANI_CONFIG
    Override config file path

MANI_USER_CONFIG
    Override user config file path

NO_COLOR
    If this env variable is set (regardless of value) then all colors will be disabled
```
