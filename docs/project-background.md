# Project Background

This document contains a little bit of everything:

- Background to `mani` and core design decisions used to develop `mani`
- Comparisons with alternatives
- Roadmap

## Background

`mani` came about because I needed a CLI tool to manage multiple repositories. So, the premise is, you have a bunch of repositories and want the following:

1. a central place for your repositories, containing name, URL, and a small description of the repository
2. ability to clone all repositories in 1 command
3. ability to run ad-hoc and custom commands (perhaps `git status` to see working tree status) on 1, a subset, or all of the repositories
4. ability to get an overview of 1, a subset, or all of the repositories and commands

Now, there's plenty of CLI tools for running cloning multiple repositories, running commands over them, see [similar software](#similar-software), and while I've taken a lot of inspiration from them, there's some core design decision that led me to create `mani`, instead of forking or contributing to an existing solution.

## Design

### Config

A lot of the alternatives to `mani` treat the config file (either using a custom format or JSON) as a state file that is interacted with via their executable.
So the way it works is, you would add a repository to the config file via `sometool add git@github.com/random/xyz`, and then to remove the repository, you'd have to open the config file and remove it manually, taking care to also update the `.gitignore` file.

I think it's a missed opportunity to not let users edit the config file manually for the following reasons:

1. The user can add additional metadata about the repositories
2. The user can order the repositories to their liking to provide a better overview of the repositories, rather than using an alphabetical or random order
3. It's seldom that you add new repositories, so it's not something that should be optimized for

That's why in `mani` you need to edit the config file to add or delete a repository. The exception is when you're setting up `mani` for the first time, then you want it to scan for existing repositories. As a bonus, it also updates your `.gitignore` file with the updated list of repositories.

### Commands

Another missed opportunity is not to have built-in support for commands. For instance, [meta](https://github.com/mateodelnorte/meta), delegates this to 3rd party tools like `make`, which makes you lose out on a few benefits:

1. Fewer tools for developers to learn (albeit `make` is something many are already familiar with)
2. Fewer files to keep track of (1 file instead of 2)
3. Better auto-completion and command discovery

Note, you can still use `make` or regular script files, just call them from the `mani.yaml` config.

So what config format is best suited for this purpose? In my opinion, YAML is a suitable candidate. While it has its issues, I think its purpose as a human-readable config/state file works well. It has all the primitives you'd need in a config language, simple key/value entries, dictionaries, and lists, as well as supporting comments (something which JSON doesn't). We could create a custom format, but then users would have to learn that syntax, so in this case, YAML has a major advantage, almost all software developers are familiar with it.

### Filtering

When we run commands, we need a way to target specific repositories. To make it as flexible as possible, there are three ways to do it in `mani`:

1. **Tag filtering**: target repositories which have a tag, for instance, add a tag `python` to all `python` repositories, then it's as simple as `mani run status -t python`
2. **Directory filtering**: target repositories by which directory they belong to, `mani run status -d frontend`, will target all repositories that are in the `frontend` directory
3. **Project name filtering**: target repositories by their name, `mani run status -p dashgrid`, will target the project `dashgrid`

### General UX

These various features make using `mani` feel more effortless:

- Automatically updating .gitignore when updating the config file
- Rich auto-completion
- Edit the `mani` config file via the `mani edit` command, which opens up the config file in your preferred editor
- Most organizations/people use git, but not everyone uses it or even uses it in the same way, so it's important to provide escape hatches, where people can provide their own VCS and customize commands to clone repositories
- Single binary (most alternatives require Python or Node.js runtime)
- Pretty output when running commands or listing repositories/commands
- Default tags/dirs/name filtering for commands
- Export output as HTML/Markdown from list/run/exec commands

## Similar Software

- [gita](https://github.com/nosarthur/gita)
- [gr](https://github.com/mixu/gr)
- [meta](https://github.com/mateodelnorte/meta)
- [mu-repo](https://github.com/fabioz/mu-repo)
- [myrepos](https://myrepos.branchable.com/)
- [repo](https://source.android.com/setup/develop/repo)
- [vcstool](https://github.com/dirk-thomas/vcstool)

