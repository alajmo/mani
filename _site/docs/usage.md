# Usage

## Create a New Mani Repository

Run the following command inside a directory containing your `git` repositories, to initialize a mani repo:

```sh
$ mani init
```

This will generate two files:

- `mani.yaml`: contains projects and custom tasks. Any sub-directory that has a `.git` inside it will be included (add the flag `--auto-discovery=false` to turn off this feature)
- `.gitignore`: includes the projects specified in `mani.yaml` file

It can be helpful to initialize the `mani` repository as a git repository so that anyone can easily download the `mani` repository and run `mani sync` to clone all repositories and get the same project setup as you.

## Common Commands

```sh
# Run arbitrary command (list all files for instance)
mani exec --all-projects 'ls -alh'

# List all repositories
mani list projects

# List repositories in a tree-like format
mani tree projects

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
