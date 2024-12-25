# Commands

## mani

repositories manager and task runner

### Synopsis

mani is a CLI tool that helps you manage multiple repositories.

It's useful when you are working with microservices, multi-project systems, multiple libraries, or just a collection 
of repositories and want a central place for pulling all repositories and running commands across them.

### Options

```
      --color                enable color (default true)
  -c, --config string        specify config
  -h, --help                 help for mani
  -u, --user-config string   specify user config
```

## run

Run tasks

### Synopsis

Run tasks.

The tasks are specified in a mani.yaml file along with the projects you can target.

```
run <task>
```

### Examples

```
  # Execute task for all projects
  mani run <task> --all

  # Execute a task in parallel with a maximum of 8 concurrent processes
  mani run <task> --projects <project> --parallel --forks 8

  # Execute task for a specific projects
  mani run <task> --projects <project>

  # Execute a task for projects with specific tags
  mani run <task> --tags <tag>

  # Execute a task for projects matching specific paths
  mani run <task> --paths <path>

  # Execute a task for all projects matching a tag expression
  mani run <task> --tags-expr 'active || git' <tag>

  # Execute a task with environment variables from shell
  mani run <task> key=value
```

### Options

```
  -a, --all                   select all projects
  -k, --cwd                   select current working directory
      --describe              display task information
      --dry-run               display the task without execution
  -e, --edit                  edit task
  -f, --forks uint32          maximum number of concurrent processes (default 4)
  -h, --help                  help for run
      --ignore-errors         continue execution despite errors
      --ignore-non-existing   skip non-existing projects
      --omit-empty-columns    hide empty columns in table output
      --omit-empty-rows       hide empty rows in table output
  -o, --output string         set output format [stream|table|markdown|html]
      --parallel              execute tasks in parallel across projects
  -d, --paths strings         select projects by path
  -p, --projects strings      select projects by name
  -s, --silent                hide progress output during task execution
  -J, --spec string           set spec
  -t, --tags strings          select projects by tag
  -E, --tags-expr string      select projects by tags expression
  -T, --target string         select projects by target name
      --theme string          set theme
      --tty                   replace current process
```

## exec

Execute arbitrary commands

### Synopsis

Execute arbitrary commands.
Use single quotes around your command to prevent file globbing and 
environment variable expansion from occurring before the command is 
executed in each directory.

```
exec <command> [flags]
```

### Examples

```
  # List files in all projects
  mani exec --all ls

  # List git files with markdown suffix in all projects
  mani exec --all 'git ls-files | grep -e ".md"'
```

### Options

```
  -a, --all                   target all projects
  -k, --cwd                   use current working directory
      --dry-run               print commands without executing them
  -f, --forks uint32          maximum number of concurrent processes (default 4)
  -h, --help                  help for exec
      --ignore-errors         ignore errors
      --ignore-non-existing   ignore non-existing projects
      --omit-empty-columns    omit empty columns in table output
      --omit-empty-rows       omit empty rows in table output
  -o, --output string         set output format [stream|table|markdown|html]
      --parallel              run tasks in parallel across projects
  -d, --paths strings         select projects by path
  -p, --projects strings      select projects by name
  -s, --silent                hide progress when running tasks
  -J, --spec string           set spec
  -t, --tags strings          select projects by tag
  -E, --tags-expr string      select projects by tags expression
  -T, --target string         target projects by target name
      --theme string          set theme
      --tty                   replace current process
```

## init

Initialize a mani repository

### Synopsis

Initialize a mani repository.

Creates a new mani repository by generating a mani.yaml configuration file 
and a .gitignore file in the current directory.

```
init [flags]
```

### Examples

```
  # Initialize with default settings
  mani init

  # Initialize without auto-discovering projects
  mani init --auto-discovery=false

  # Initialize without updating .gitignore
  mani init --sync-gitignore=false
```

### Options

```
      --auto-discovery   automatically discover and add Git repositories to mani.yaml (default true)
  -h, --help             help for init
  -g, --sync-gitignore   synchronize .gitignore file (default true)
```

## sync

Clone repositories and update .gitignore

### Synopsis

Clone repositories and update .gitignore file.
For repositories requiring authentication, disable parallel cloning to enter
credentials for each repository individually.

```
sync [flags]
```

### Examples

```
  # Clone repositories one at a time
  mani sync

  # Clone repositories in parallell
  mani sync --parallel

  # Disable updating .gitignore file
  mani sync --sync-gitingore=false

  # Sync project remotes. This will modify the projects .git state
  mani sync --sync-remotes

	# Clone repositories even if project sync field is set to false
  mani sync --ignore-sync-state

  # Display sync status
  mani sync --status
```

### Options

```
  -f, --forks uint32        maximum number of concurrent processes (default 4)
  -h, --help                help for sync
      --ignore-sync-state   sync project even if the project's sync field is set to false
  -p, --parallel            clone projects in parallel
  -d, --paths strings       clone projects by path
  -s, --status              display status only
  -g, --sync-gitignore      sync gitignore (default true)
  -r, --sync-remotes        update git remote state
  -t, --tags strings        clone projects by tags
  -E, --tags-expr string    clone projects by tag expression
```

## edit

Open up mani config file

### Synopsis

Open up mani config file in $EDITOR.

```
edit [flags]
```

### Examples

```
  # Edit current context
  mani edit
```

### Options

```
  -h, --help   help for edit
```

## edit project

Edit mani project

### Synopsis

Edit mani project in $EDITOR.

```
edit project [project] [flags]
```

### Examples

```
  # Edit projects
  mani edit project

  # Edit project <project>
  mani edit project <project>
```

### Options

```
  -h, --help   help for project
```

## edit task

Edit mani task

### Synopsis

Edit mani task in $EDITOR.

```
edit task [task] [flags]
```

### Examples

```
  # Edit tasks
  mani edit task

  # Edit task <task>
  mani edit task <task>
```

### Options

```
  -h, --help   help for task
```

## list projects

List projects

### Synopsis

List projects.

```
list projects [projects] [flags]
```

### Examples

```
  # List all projects
  mani list projects

  # List projects by name
  mani list projects <project>

  # List projects by tags
  mani list projects --tags <tag>

  # List projects by paths
  mani list projects --paths <path>

	# List projects matching a tag expression
	mani run <task> --tags-expr '<tag-1> || <tag-2>'
```

### Options

```
  -a, --all                select all projects (default true)
  -k, --cwd                select current working directory
      --headers strings    specify columns to display [project, path, relpath, description, url, tag] (default [project,tag,description])
  -h, --help               help for projects
  -d, --paths strings      select projects by paths
  -t, --tags strings       select projects by tags
  -E, --tags-expr string   select projects by tags expression
  -T, --target string      select projects by target name
      --tree               display output in tree format
```

### Options inherited from parent commands

```
  -o, --output string   set output format [table|markdown|html] (default "table")
      --theme string    set theme (default "default")
```

## list tags

List tags

### Synopsis

List tags.

```
list tags [tags] [flags]
```

### Examples

```
  # List all tags
  mani list tags
```

### Options

```
      --headers strings   specify columns to display [project, tag] (default [tag,project])
  -h, --help              help for tags
```

### Options inherited from parent commands

```
  -o, --output string   set output format [table|markdown|html] (default "table")
      --theme string    set theme (default "default")
```

## list tasks

List tasks

### Synopsis

List tasks.

```
list tasks [tasks] [flags]
```

### Examples

```
  # List all tasks
  mani list tasks

  # List tasks by name
  mani list task <task>
```

### Options

```
      --headers strings   specify columns to display [task, description, target, spec] (default [task,description])
  -h, --help              help for tasks
```

### Options inherited from parent commands

```
  -o, --output string   set output format [table|markdown|html] (default "table")
      --theme string    set theme (default "default")
```

## describe projects

Describe projects

### Synopsis

Describe projects.

```
describe projects [projects] [flags]
```

### Examples

```
  # Describe all projects
  mani describe projects

  # Describe projects by name
  mani describe projects <project>

  # Describe projects by tags
  mani describe projects --tags <tag>

  # Describe projects by paths
  mani describe projects --paths <path>

	# Describe projects matching a tag expression
	mani run <task> --tags-expr '<tag-1> || <tag-2>'
```

### Options

```
  -a, --all                select all projects (default true)
  -k, --cwd                select current working directory
  -e, --edit               edit project
  -h, --help               help for projects
  -d, --paths strings      filter projects by paths
  -t, --tags strings       filter projects by tags
  -E, --tags-expr string   target projects by tags expression
  -T, --target string      target projects by target name
```

### Options inherited from parent commands

```
      --theme string   set theme (default "default")
```

## describe tasks

Describe tasks

### Synopsis

Describe tasks.

```
describe tasks [tasks] [flags]
```

### Examples

```
  # Describe all tasks
  mani describe tasks

  # Describe task <task>
  mani describe task <task>
```

### Options

```
  -e, --edit   edit task
  -h, --help   help for tasks
```

### Options inherited from parent commands

```
      --theme string   set theme (default "default")
```

## tui

TUI

### Synopsis

Run TUI

```
tui [flags]
```

### Examples

```
  # Open tui
  mani tui
```

### Options

```
  -h, --help               help for tui
  -r, --reload-on-change   reload mani on config change
      --theme string       set theme (default "default")
```

## check

Validate config

### Synopsis

Validate config.

```
check [flags]
```

### Examples

```
  # Validate config
  mani check
```

### Options

```
  -h, --help   help for check
```

## gen

Generate man page

```
gen
```

### Options

```
  -d, --dir string   directory to save manpage to (default "./")
  -h, --help         help for gen
```

