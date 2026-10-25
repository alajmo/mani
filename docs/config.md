# Config

The mani.yaml config is based on the following concepts:

- **projects** are directories, which may be git repositories, in which case they have an URL attribute
- **tasks** are shell commands that you write and then run for selected **projects**
- **specs** are configs that alter **task** execution and output
- **targets** are configs that provide shorthand filtering of **projects** when executing **tasks**
- **themes** are used to modify the output of `mani` commands
- **env** are environment variables that can be defined globally, per project and per task

**Specs**, **targets** and **themes** use a default object by default that the user can override to modify execution of mani commands.

Check the [files](#files) and [environment](#environment) section to see how the config file is loaded.

Below is a config file detailing all of the available options and their defaults.

```yaml
# Import projects/tasks/env/specs/themes/targets from other configs
import:
  - ./some-dir/mani.yaml

# Base directory for this file's relative project paths
# Relative to this config file; does not affect imported configs
project_dir: .

# Shell used for commands
# If you use any other program than bash, zsh, sh, node, and python
# then you have to provide the command flag if you want the command-line string evaluted
# For instance: bash -c
shell: bash

# If set to true, mani will override the URL of any existing remote
# and remove remotes not found in the config
sync_remotes: false

# If set to true, mani will remove worktrees that exist on disk
# but are not defined in the config
remove_orphaned_worktrees: false

# Determines whether the .gitignore should be updated when syncing projects
sync_gitignore: true

# When running the TUI, specifies whether it should reload when the mani config is changed
reload_tui_on_change: false

# List of Projects
projects:
  # Project name [required]
  pinto:
    # Determines if the project should be synchronized during 'mani sync'
    sync: true

    # Project path relative to project_dir
    # Defaults to project name if not specified
    path: frontend/pinto

    # Repository URL
    url: git@github.com:alajmo/pinto

    # Project description
    desc: A vim theme editor

    # Custom clone command
    # Defaults to "git clone URL PATH"
    clone: git clone git@github.com:alajmo/pinto --branch main

    # Branch to use as primary HEAD when cloning
    # Defaults to repository's primary HEAD
    branch:

    # When true, clones only the specified branch or primary HEAD
    single_branch: false

    # Project tags
    tags: [dev]

    # Remote repositories
    # Key is the remote name, value is the URL
    remotes:
      foo: https://github.com/bar

    # Git worktrees
    # path: Required, relative to project directory (or absolute)
    # branch: Optional, defaults to path basename
    # Auto-discovered by 'mani init', created by 'mani sync'
    worktrees:
      - path: hotfix                    # branch defaults to "hotfix"
      - path: feature-branch
        branch: feature/awesome
      - path: ../project-staging        # worktree outside project dir
        branch: staging

    # Project-specific environment variables
    env:
      # Simple string value
      branch: main

      # Shell command substitution
      date: $(date -u +"%Y-%m-%dT%H:%M:%S%Z")

# List of Specs
specs:
  default:
    # Output format for task results
    # Options: stream, table, html, markdown
    output: stream

    # Enable parallel task execution
    parallel: false

    # Maximum number of concurrent tasks when running in parallel
    forks: 4

    # When true, continues execution if a command fails in a multi-command task
    ignore_errors: false

    # When true, skips project entries in the config that don't exist
    # on the filesystem without throwing an error
    ignore_non_existing: false

    # Hide projects with no command output
    omit_empty_rows: false

    # Hide columns with no data
    omit_empty_columns: false

    # Clear screen before task execution (TUI only)
    clear_output: true

# List of targets
targets:
  default:
    # Select all projects
    all: false

    # Select project in current working directory
    cwd: false

    # Select projects by name
    projects: []

    # Select projects by path
    paths: []

    # Select projects by tag
    tags: []

    # Select projects by tag expression
    tags_expr: ''

# Environment variables available to all tasks
env:
  # Simple string value
  AUTHOR: 'alajmo'

# Shell command substitution
DATE: $(date -u +"%Y-%m-%dT%H:%M:%S%Z")

# List of tasks
tasks:
  # Command name [required]
  simple-2: echo "hello world"

  # Command name [required]
  simple-1:
    cmd: |
      echo "hello world"
    desc: simple command 1

  # Command name [required]
  advanced-command:
    # Task description
    desc: complex task

    # Task theme
    theme: default

    # Shell interpreter
    shell: bash

    # Task-specific environment variables
    env:
      # Static value
      branch: main

      # Dynamic shell command output
      num_lines: $(ls -1 | wc -l)

    # Can reference a predefined spec:
    # spec: custom_spec
    # or define one inline:
    spec:
      output: table
      parallel: true
      forks: 4
      ignore_errors: false
      ignore_non_existing: true
      omit_empty_rows: true
      omit_empty_columns: true

    # Can reference a predefined target:
    # target: custom_target
    # or define one inline:
    target:
      all: true
      cwd: false
      projects: [pinto]
      paths: [frontend]
      tags: [dev]
      tags_expr: (prod || dev) && !test

    # Single multi-line command
    cmd: |
      echo complex
      echo command

    # Multiple commands. Use either `cmd` or `commands`, not both.
    # Each entry is either an inline command or a reference to another
    # task via `task:`. When referencing a task, only its `cmd`/`shell`
    # are reused — `target`, `spec` and `env` come from the wrapping
    # task (or its CLI flags), so referenced tasks compose cleanly
    # without recursing through `mani run`.
    commands:
      # Inline command with a custom shell
      - name: node-example
        shell: node
        cmd: console.log("hello world from node.js");

      # Reference to another task defined above
      - task: simple-1

# List of themes
# Styling Options:
#   Fg (foreground color): Empty string (""), hex color, or named color from W3C standard
#   Bg (background color): Empty string (""), hex color, or named color from W3C standard
#   Format: Empty string (""), "lower", "title", "upper"
#   Attribute: Empty string (""), "bold", "italic", "underline"
#   Alignment: Empty string (""), "left", "center", "right"
themes:
  # Theme name [required]
  default:
    # Stream Output Configuration
    stream:
      # Include project name prefix for each line
      prefix: true

      # Colors to alternate between for each project prefix
      prefix_colors:
        ['#d787ff', '#00af5f', '#d75f5f', '#5f87d7', '#00af87', '#5f00ff']

      # Add a header before each project
      header: true

      # String value that appears before the project name in the header
      header_prefix: 'TASK'

      # Fill remaining spaces with a character after the prefix
      header_char: '*'

    # Table Output Configuration
    table:
      # Table style
      # Available options: ascii, light, bold, double, rounded
      style: ascii

      # Border options for table output
      border:
        around: false # Border around the table
        columns: true # Vertical border between columns
        header: true # Horizontal border between headers and rows
        rows: false # Horizontal border between rows

      header:
        fg: '#d787ff'
        attr: bold
        format: ''

      title_column:
        fg: '#5f87d7'
        attr: bold
        format: ''

    # Tree View Configuration
    tree:
      # Tree style
      # Available options: ascii, light, bold, double, rounded, bullet-square, bullet-circle, bullet-star
      style: light

    # Block Display Configuration
    block:
      key:
        fg: '#5f87d7'
      separator:
        fg: '#5f87d7'
      value:
        fg:
      value_true:
        fg: '#00af5f'
      value_false:
        fg: '#d75f5f'

      # TUI Configuration
      tui:
        default:
          fg:
          bg:
          attr:

        border:
          fg:
        border_focus:
          fg: '#d787ff'

        title:
          fg:
          bg:
          attr:
          align: center
        title_active:
          fg: '#000000'
          bg: '#d787ff'
          attr:
          align: center

        button:
          fg:
          bg:
          attr:
          format:
        button_active:
          fg: '#080808'
          bg: '#d787ff'
          attr:
          format:

        table_header:
          fg: '#d787ff'
          bg:
          attr: bold
          align: left
          format:

        item:
          fg:
          bg:
          attr:
        item_focused:
          fg: '#ffffff'
          bg: '#262626'
          attr:
        item_selected:
          fg: '#5f87d7'
          bg:
          attr:
        item_dir:
          fg: '#d787ff'
          bg:
          attr:
        item_ref:
          fg: '#d787ff'
          bg:
          attr:

        search_label:
          fg: '#d7d75f'
          bg:
          attr: bold
        search_text:
          fg:
          bg:
          attr:

        filter_label:
          fg: '#d7d75f'
          bg:
          attr: bold
        filter_text:
          fg:
          bg:
          attr:

        shortcut_label:
          fg: '#00af5f'
          bg:
          attr:
        shortcut_text:
          fg:
          bg:
          attr:

```

## Combining Projects

Projects with the same name in different config files are combined. This lets a
shared config define repositories while a local config adds tags, environment
variables or overrides. Each project name may appear only once per file.

- A config overrides values set by files it imports, directly or indirectly.
- Independent configs may supply the same value. Different values are an error;
  set the value in a config that imports both to resolve it.
- The user config and files imported only by it provide fallback values. Workspace
  configs take precedence over them.
- `tags` are combined without duplicates. `env`, `remotes` and `worktrees` are
  merged by environment name, remote name and cleaned worktree path respectively.
- Other fields are merged individually. Explicit `false` and empty strings
  override inherited values. An omitted field inherits its value. Empty
  collections add nothing; they do not remove inherited entries.

An explicit `path` is relative to `project_dir` in the file setting it. Absolute
paths remain absolute. Both settings support environment variables and `~/`.
If no definition sets `path`, it defaults to the project name beside the config
providing its `url`, using that file's `project_dir`. If no URL is set, import
precedence chooses the defining config. A user config supplies this default only
when the project has no workspace definition. If independent definitions imply
different default paths, set `path` explicitly to remove the ambiguity.

This keeps a tag-only overlay from moving a shared repository. Displayed relative
paths and `--paths` filters are relative to the root config. Project env commands
run only after merging and config validation; overridden commands do not run.
Worktree branch defaults are applied after the winning worktree entry is selected.

```yaml
# shared/mani.yaml
project_dir: ..
projects:
  app:
    url: git@github.com:team/app.git
    tags: [shared]
    env:
      MODE: production
```

```yaml
# mani.yaml
import: [shared/mani.yaml]
projects:
  app:
    tags: [local]
    sync: false
    env:
      MODE: development
```

The result is one `app` project beside the root config, with the shared URL,
tags `[shared, local]`, `sync: false`, and `MODE=development`.
These merging rules apply only to projects. Tasks, global env, specs, targets and
themes retain their existing import behavior.

## Files

When running a command, `mani` will check the current directory and all parent directories for the following files: `mani.yaml`, `mani.yml`, `.mani.yaml`, `.mani.yml` .

Additionally, it will import (if found) a config file from:

- Linux: `$XDG_CONFIG_HOME/mani/config.yaml` or `$HOME/.config/mani/config.yaml` if `$XDG_CONFIG_HOME` is not set.
- Darwin: `$HOME/Library/Application Support/mani/config.yaml`
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
