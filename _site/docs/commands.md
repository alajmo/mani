# Commands

## mani

repositories manager and task runner

### Options

```
  -c, --config string        config file (default is current and all parent directories)
  -h, --help                 help for mani
      --no-color             disable color
  -u, --user-config string   user config
```

## run

Run tasks

### Synopsis

Run tasks.

The tasks are specified in a mani.yaml file along with the projects you can target.

```
run <task> [flags]
```

### Examples

```
  # Run task 'pwd' for all projects
  mani run pwd --all

  # Checkout branch 'development' for all projects that have tag 'backend'
  mani run checkout -t backend branch=development
```

### Options

```
  -a, --all                target all projects
  -k, --cwd                current working directory
      --describe           print task information
      --dry-run            prints the task to see what will be executed
  -e, --edit               edit task
  -h, --help               help for run
      --omit-empty         omit empty results when running a task
  -o, --output string      set output [text|table|html|markdown]
      --parallel           run tasks in parallel for each project
  -d, --paths strings      target projects by their path
  -p, --projects strings   target projects by their name
  -t, --tags strings       target projects by their tag
      --theme string       set theme
```

## exec

Execute arbitrary commands

### Synopsis

Execute arbitrary commands.

Single quote your command if you don't want the file globbing and environments variables expansion to take place
before the command gets executed in each directory.

```
exec <command> [flags]
```

### Examples

```
  # List files in all projects
  mani exec ls --all

  # List all git files that have markdown suffix
  mani exec 'git ls-files | grep -e ".md"' --all
```

### Options

```
  -a, --all                target all projects
  -k, --cwd                current working directory
      --dry-run            prints the command to see what will be executed
  -h, --help               help for exec
      --omit-empty         omit empty results when running a command
  -o, --output string      set output [text|table|markdown|html]
      --parallel           run tasks in parallel
  -d, --paths strings      target projects by their path
  -p, --projects strings   target projects by their name
  -t, --tags strings       target projects by their tag
      --theme string       set theme (default "default")
```

## init

Initialize a mani repository

### Synopsis

Initialize a mani repository.

Creates a mani repository - a directory with configuration file mani.yaml and a .gitignore file.

```
init [flags]
```

### Examples

```
  # Basic example
  mani init

  # Skip auto-discovery of projects
  mani init --auto-discovery=false
```

### Options

```
      --auto-discovery   walk current directory and find git repositories to add to mani.yaml (default true)
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
```

### Options

```
  -h, --help       help for sync
  -p, --parallel   clone projects in parallel
  -s, --status     print sync status only
```

## edit

Edit mani config

### Synopsis

Edit mani config

```
edit [flags]
```

### Examples

```
  # Edit current context
  mani edit

  # Edit specific mani config
  edit edit --config path/to/mani/config
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
  # Edit a project called mani
  mani edit project mani

  # Edit project in specific mani config
  mani edit --config path/to/mani/config
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
  # Edit a task called status
  mani edit task status

  # Edit task in specific mani config
  mani edit task status --config path/to/mani/config
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
  # List projects
  mani list projects
```

### Options

```
      --headers strings   set headers. Available headers: project, path, relpath, description, url, tag (default [project,tag,description])
  -h, --help              help for projects
  -d, --paths strings     filter projects by their path
  -t, --tags strings      filter projects by their tag
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
  # List tags
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
  # List tasks
  mani list tasks
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
  # Describe projects
  mani describe projects

  # Describe projects that have tag frontend
  mani describe projects --tags frontend
```

### Options

```
  -e, --edit            Edit project
  -h, --help            help for projects
  -d, --paths strings   filter projects by their path
  -t, --tags strings    filter projects by their tag
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
  # Describe tasks
  mani describe tasks
```

### Options

```
  -e, --edit   edit task
  -h, --help   help for tasks
```

## gen

Generate man page

```
gen [flags]
```

### Options

```
      --dir string   directory to save manpages to (default "./")
  -h, --help         help for gen
```

## version

Print version/build info

### Synopsis

Print version/build info.

```
version [flags]
```

### Options

```
  -h, --help   help for version
```

