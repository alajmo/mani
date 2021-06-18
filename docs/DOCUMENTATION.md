# Documentation

## Table of Contents

<!-- vim-markdown-toc GFM -->

* [Manifest](#manifest)
  * [Projects](#projects)
    * [Name](#name)
    * [Path](#path)
    * [Url](#url)
    * [Description](#description)
    * [Clone](#clone)
    * [Tags](#tags)
  * [Env](#env)
  * [Shell](#shell)
  * [Theme](#theme)
  * [Tasks](#tasks)
    * [Name](#name-1)
    * [Description](#description-1)
    * [Shell](#shell-1)
    * [Env](#env-1)
    * [Tags](#tags-1)
    * [Projects](#projects-1)
    * [Output](#output)
    * [Command](#command)
    * [Commands](#commands)
* [Environment Variables](#environment-variables)

<!-- vim-markdown-toc -->

## Manifest

The `mani.yaml` config is based on two concepts: __projects__ and __commands__. __Projects__ are simply directories, which may be git repositories, in which case they have an url attribute. __Commands__ are arbitrary shell commands that you write and then run for selected __projects__.

`mani.yaml`
```yaml
# List of Projects
projects:
  - name: pinto # Project name [required]
    path: frontend/pinto # Project path [defaults to NAME]
    url: git@github.com:alajmo/pinto # Project URL [optional]
    description: A vim theme editor # Project description [optional]
    clone: git clone git@github.com:alajmo/pinto # Clone command [defaults to `git clone URL`]
    tags: [frontend] # List of tags [optional]

# List of environment variables that's available to all tasks
env:
  AUTHOR: "alajmo" # Simple string value
  DATE: $(date -u +"%Y-%m-%dT%H:%M:%S%Z") # Shell command substitution

# Shell used for commands [defaults to "sh -c"]
shell: bash -c

theme:
  table: ascii # Available styles: box (default), ascii
  tree: line-bold # Available styles: line (default), line-bold, square, circle, star

# List of tasks
tasks:
  - name: simple # Command name [required]
    command: echo simple # Single line command [required]

  - name: complex
    description: complex task # Task description [optional]
    shell: bash -c # Shell used for this command [defaults to root shell]
    env: # List of environment variables
      branch: master # Simple string value
      num_lines: $(ls -1 | wc -l) # Shell command substitution
    tags: [work] # Target projects with tags [defaults to empty list]
    projects: [awesome] # Target projects [defaults to empty list]
    output: table # Set default output option [defaults to list]

    # Each task can have a single command, multiple commands, or both

    command: | # Multine command
      echo complex
      echo command

    commands: # List of commands
      - name: first # Command name [required]
        description: first command # Single line command [required]
        command: echo first # Single line command [required]

      - name: second # Command name [required]
        description: second command # Single line command [required]
        command: echo second # Single line command [required]
```

### Projects

List of projects that mani will operate on.

#### Name

The name of the project. This is required for each project.

```yaml
projects:
  - name: example
```

#### Path

Path to the project, relative to the directory of the config file. It defaults to the name of the project.

```yaml
projects:
  - name: example
    path: work/example
```

#### Url

The url of the project, which the `mani sync` command will use to download the repository. `mani sync` uses `git clone git@github.com:alajmo/pinto` behind the scenes. So if you want to modify the clone command, checkout the [clone](#clone) property.

```yaml
projects:
  - name: example
    path: git@github.com:alajmo/pinto
```

#### Description

Optional description of the project.

```yaml
projects:
  - name: example
    description: an example repository
```

#### Clone

Clone command that `mani sync` will use to download the repository. It defaults to `git clone URL`.

In-case you want to do modify the clone command, this is the place to do it. For instance, to only clone a single branch:

```yaml
projects:
  - name: example
    clone: git clone git@github.com:alajmo/pinto --branch main
```

#### Tags

A list of tags to associate the project with.

```yaml
projects:
  - name: example
    url: git@github.com:alajmo/pinto
    tags: [work, cli]
```

### Env

A dictionary of key/value pairs that all `tasks` inherit. The value can either be a simple string:

```yaml
env:
  VERSION: v1.0.0
```

or if it is enclosed within `$()`, shell command substitution takes place.

```yaml
env:
  DATE: $(date)
```

### Shell

Shell used for commands, it defaults to "sh -c". Note, you have to provide the flag `-c` for shell programs `bash`, `sh`, etc. if you want a command line string evaluated.

In-case you only want to execute a script file, then the following will do:

```yaml
shell: bash

tasks:
  - name: example
    command: script.sh
```

or

```yaml
shell: bash -c

tasks:
  - name: example
    command: ./script.sh
```

Note, any executable that's in your `PATH` works:

```yaml
shell: node

tasks:
  - name: example
    command: index.js
```

### Theme

The theme property contains key/value pairs that alter the output of table and tree stylings.

```yaml
theme:
  table: ascii # Available styles: box (default), ascii
  tree: line-bold # Available styles: line (default), line-bold, square, circle, star
```

### Tasks

List of predefined tasks that can be run on `projects`.

#### Name

The name of the tasks. This is required for each task.

```yaml
tasks:
  - name: example
    command: echo 123
```

#### Description

An optional string value that describes your `task`.

```yaml
tasks:
  - name: example
    description: print 123
    command: echo 123
```

#### Shell

`Shell` used for this task commands. Defaults to the root `Shell` defined in the global scope (which in turn defaults to `sh -c`).

```yaml
shell: bash

tasks:
  - name: example
    command: script.sh
```

#### Env

A dictionary of key/value pairs, see [env](#env). The value can either be a simple string:

The `env` field is inherited from the global scope and can be overriden in the `task` definition.

For instance:

```yaml
env:
  VERSION: v1.0.0
  BRANCH: main

tasks:
  - name: example
    env:
      VERSION: v2.0.0
    command: |
      echo $VERSION
      echo $BRANCH
```

Will print:

```sh
$ mani run example
v2.0.0
main
```

#### Tags

A list of tags that are used to filter projects when running a `task`.

```yaml
tasks:
  - name: example
    command: echo 123
    tags: [work]
```

This is equivalent to running `mani run example --tags work`

#### Projects

A list of projects that are used to filter projects when running a `task`.

```yaml
tasks:
  - name: example
    command: echo 123
    projects: [pinto]
```

This is equivalent to running `mani run example --projects pinto`

#### Output

Output format when running commands, defaults to `list`. Possible values are: `table` `list`, `markdown` and `HTML`.

```yaml
tasks:
  - name: example
    output: table
    command: echo 123
```

This is equivalent to running `mani run example --output table`

#### Command

Single or multiline command that uses the `shell` program to run in each project its filtered on.

Single-line command:
```yaml
tasks:
  - name: example
    command: echo 123
```

Multi-line command:
```yaml
tasks:
  - name: example
    command: |
      echo 123
      echo 456
```

#### Commands

A `task` also supports running multiple commands. In this case, the `first-command` will be run first, and then the `second-command` will run. Both of its outputs will be displayed.

```yaml
tasks:
  - name: example
    commands:
      - name: first-command
        command: echo first

      - name: second-command
        command: echo first
```

## Environment Variables

`mani` exposes some variables to each command:

Global:

- `MANI_CONFIG_PATH`: Absolute path of the current mani.yaml file

Project specific:

- `MANI_PROJECT_NAME`: The name of project
- `MANI_PROJECT_URL`: The url of the project
- `MANI_PROJECT_PATH` The path to the project in absolute form
