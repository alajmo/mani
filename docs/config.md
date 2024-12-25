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

# Shell used for commands
# If you use any other program than bash, zsh, sh, node, and python
# then you have to provide the command flag if you want the command-line string evaluted
# For instance: bash -c
shell: bash

# If set to true, mani will override the URL of any existing remote
# and remove remotes not found in the config
sync_remotes: false

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

 # Project path relative to the config file
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
   tags_expr: ""

# Environment variables available to all tasks
env:
  # Simple string value
  AUTHOR: "alajmo"

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

 # Can reference predefined spec:
 # spec: custom_spec
 # or define inline:
 spec:
   output: table
   parallel: true
   forks: 4
   ignore_errors: false
   ignore_non_existing: true
   omit_empty_rows: true
   omit_empty_columns: true

 # Can reference predefined target:
 # target: custom_target
 # or define inline:
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

 # Multiple commands
 commands:
   # Node.js command example
   - name: node-example
       shell: node
     cmd: console.log("hello world from node.js");

   # Reference to another task
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
   prefix_colors: ["#d787ff", "#00af5f", "#d75f5f", "#5f87d7", "#00af87", "#5f00ff"]

   # Add a header before each project
   header: true

   # String value that appears before the project name in the header
   header_prefix: "TASK"

   # Fill remaining spaces with a character after the prefix
   header_char: "*"

 # Table Output Configuration
 table:
   # Table style
   # Available options: ascii, light, bold, double, rounded
   style: ascii

   # Border options for table output
   border:
     around: false  # Border around the table
     columns: true  # Vertical border between columns
     header: true   # Horizontal border between headers and rows
     rows: false    # Horizontal border between rows

   header:
     fg: "#d787ff"
     attr: bold
     format: ""

   title_column:
     fg: "#5f87d7"
     attr: bold
     format: ""

 # Tree View Configuration
 tree:
   # Tree style
   # Available options: ascii, light, bold, double, rounded, bullet-square, bullet-circle, bullet-star
   style: light

 # Block Display Configuration
 block:
   key:
     fg: "#5f87d7"
   separator:
     fg: "#5f87d7"
   value:
     fg:
   value_true:
     fg: "#00af5f"
   value_false:
     fg: "#d75f5f"

  # TUI Configuration
  tui:
    default:
      fg:
      bg:
      attr:

    border:
      fg:
    border_focus:
      fg: "#d787ff"

    title:
      fg:
      bg:
      attr:
      align: center
    title_active:
      fg: "#000000"
      bg: "#d787ff"
      attr:
      align: center

    button:
      fg:
      bg:
      attr:
      format:
    button_active:
      fg: "#080808"
      bg: "#d787ff"
      attr:
      format:

    table_header:
      fg: "#d787ff"
      bg:
      attr: bold
      align: left
      format:

    item:
      fg:
      bg:
      attr:
    item_focused:
      fg: "#ffffff"
      bg: "#262626"
      attr:
    item_selected:
      fg: "#5f87d7"
      bg:
      attr:
    item_dir:
      fg: "#d787ff"
      bg:
      attr:
    item_ref:
      fg: "#d787ff"
      bg:
      attr:

    search_label:
      fg: "#d7d75f"
      bg:
      attr: bold
    search_text:
      fg:
      bg:
      attr:

    filter_label:
      fg: "#d7d75f"
      bg:
      attr: bold
    filter_text:
      fg:
      bg:
      attr:

    shortcut_label:
      fg: "#00af5f"
      bg:
      attr:
    shortcut_text:
      fg:
      bg:
      attr:
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
