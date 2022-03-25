# Usage

## Create a New Mani Repository

Run the following command inside a directory containing your `git` repositories:

```bash
$ mani init
```

This will generate two files:

- `mani.yaml`: contains projects and custom tasks. Any sub-directory that has a `.git` inside it will be included (add the flag `--auto-discovery=false` to turn off this feature)
- `.gitignore`: includes the projects specified in `mani.yaml` file. To opt out, use `mani init --vcs=none`

It can be helpful to initialize the `mani` repository as a git repository so that anyone can easily download the `mani` repository and run `mani sync` to clone all repositories and get the same project setup as you.

## Default Config

`mani` will attempt to read in an additional file found in different places depending on OS:

- Linux: `$XDG_CONFIG_HOME/mani/config.yaml` or `$HOME/.config/mani/config.yaml` if `$XDG_CONFIG_HOME` is not set.
- Darwin: `$HOME/Library/Application/mani`
- Windows: `%AppData%/mani`

To change the directory, run `mani` with the flag `--user-config-dir <path>`.

## Commands to Get You Started

```bash
# Run arbitrary command (list all files for instance)
mani exec --all 'ls -alh'

# List all repositories
mani list projects

# List repositories in a tree-like format
mani list projects --tree

# Describe available tasks
mani describe tasks

# Run task for projects that have the frontend tag
mani run list-files --tags frontend

# Run task for projects under a specific directory
mani run list-files --paths work/

# Run task for specific project
mani run list-files --project project-a

# Open up mani.yaml in your preferred editor
mani edit
```
