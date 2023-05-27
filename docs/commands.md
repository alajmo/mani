# Commands

## mani

repositories manager and task runner

### Synopsis

mani is a CLI tool that helps you manage multiple repositories.

It's useful when you want a central place for pulling all repositories and running commands over them.

You specify repository and tasks in a config file and then run the commands over all or a subset of the repositories.


### Options

```
  -c, --config string        specify config
  -h, --help                 help for mani
      --no-color             disable color
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
  # Run task <task> for all projects
  mani run <task> --all

  # Run task <task> for all projects <project>
  mani run <task> --projects <project>

  # Run task <task> for all projects that have tags <tag>
  mani run <task> --tags <tag>

  # Run task <task> for all projects matching paths <path>
  mani run <task> --paths <path>

  # Run task <task> and pass in env value from shell
  mani run <task> key=value
```

### Options

```
  -a, --all                   target all projects
  -k, --cwd                   current working directory
      --describe              print task information
      --dry-run               prints the task to see what will be executed
  -e, --edit                  edit task
  -h, --help                  help for run
      --ignore-errors         ignore errors
      --ignore-non-existing   ignore non-existing projects
      --omit-empty            omit empty results
  -o, --output string         set output [text|table|html|markdown]
      --parallel              run tasks in parallel for each project
  -d, --paths strings         target projects by paths
  -p, --projects strings      target projects by names
  -s, --silent                do not show progress when running tasks
  -t, --tags strings          target projects by tags
      --theme string          set theme
```

## exec

Execute arbitrary commands

### Synopsis

Execute arbitrary commands.

Single quote your command if you don't want the
file globbing and environments variables expansion to take place
before the command gets executed in each directory.

```
exec <command> [flags]
```

### Examples

```
  # List files in all projects
  mani exec --all ls

  # List git files that have markdown suffix for all projects
  mani exec --all 'git ls-files | grep -e ".md"'
```

### Options

```
  -a, --all                   target all projects
  -k, --cwd                   current working directory
      --dry-run               prints the command to see what will be executed
  -h, --help                  help for exec
      --ignore-errors         ignore errors
      --ignore-non-existing   ignore non-existing projects
      --omit-empty            omit empty results
  -o, --output string         set output [text|table|markdown|html]
      --parallel              run tasks in parallel for each project
  -d, --paths strings         target projects by paths
  -p, --projects strings      target projects by names
  -s, --silent                do not show progress when running tasks
  -t, --tags strings          target projects by tags
      --theme string          set theme (default "default")
```

## init

Initialize a mani repository

### Synopsis

Initialize a mani repository.

Creates a mani repository - a directory with config file mani.yaml and a .gitignore file.

```
init [flags]
```

### Examples

```
  # Basic example
  mani init

  # Skip auto-discovery of projects
  mani init --auto-discovery=false

  # Skip creation of .gitignore file
  mani init --vcs=none
```

### Options

```
      --auto-discovery   walk current directory and add git repositories to mani.yaml (default true)
  -h, --help             help for init
      --vcs string       initialize directory using version control system. Acceptable values: <git|none> (default "git")
```

## sync

Clone repositories and add them to gitignore

### Synopsis

Clone repositories and add them to gitignore.
In-case you need to enter credentials before cloning, run the command without the parallel flag.

```
sync [flags]
```

### Examples

```
  # Clone repositories one at a time
  mani sync

  # Clone repositories in parallell
  mani sync --parallel

  # Show cloned projects
  mani sync --status
```

### Options

```
  -h, --help            help for sync
  -p, --parallel        clone projects in parallel
  -d, --paths strings   filter projects by paths
  -s, --status          print sync status only
  -r, --sync-remotes    update existing remotes
  -t, --tags strings    filter projects by tags
```

## edit

Open up mani config file in $EDITOR

### Synopsis

Open up mani config file in $EDITOR

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

Edit mani project

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

Edit mani task

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

List projects

```
list projects [projects] [flags]
```

### Examples

```
  # List all projects
  mani list projects

  # List projects <project>
  mani list projects <project>

  # List projects that have tag <tag>
  mani list projects --tags <tag>

  # List projects matching paths <path>
  mani list projects --paths <path>
```

### Options

```
      --headers strings   set headers. Available headers: project, path, relpath, description, url, tag (default [project,tag,description])
  -h, --help              help for projects
  -d, --paths strings     filter projects by paths
  -t, --tags strings      filter projects by tags
      --tree              tree output
```

### Options inherited from parent commands

```
  -o, --output string   set output [table|markdown|html] (default "table")
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
      --headers strings   set headers. Available headers: tag, project (default [tag,project])
  -h, --help              help for tags
```

### Options inherited from parent commands

```
  -o, --output string   set output [table|markdown|html] (default "table")
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

  # List task <task>
  mani list task <task>
```

### Options

```
      --headers strings   set headers. Available headers: task, description (default [task,description])
  -h, --help              help for tasks
```

### Options inherited from parent commands

```
  -o, --output string   set output [table|markdown|html] (default "table")
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

  # Describe project <project>
  mani describe projects <project>

  # Describe projects that have tag <tag>
  mani describe projects --tags <tag>

  # Describe projects matching paths <path>
  mani describe projects --paths <path>
```

### Options

```
  -e, --edit            edit project
  -h, --help            help for projects
  -d, --paths strings   filter projects by paths
  -t, --tags strings    filter projects by tags
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

