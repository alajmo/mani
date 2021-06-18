# Motivation

There's plenty of CLI tools for handling multiple repositories, see [similar software](#similar-software), and while I've taken a lot of inspiration from them, there's some core design decision that led me to create `mani`, instead of forking or contributing to an existing solution. This document will contain some of those design decisions.

## User Story

You have a bunch of repositories and want the following:

1. a centralized place for your repositories, containing name, URL and small description of the project
2. ability to clone all repositories in 1 command
3. ability to run ad-hoc and custom commands (perhaps `git status` to see working tree status) on 1, a subset, or all of the repositories
4. ability to get an overview of 1, a subset, or all of the projects/commands

## Solution

Given our user stories/problem formulation, let's address all of the problems:

1. Create a config format that contains all of the repositories along with some meta data
2. Create a CLI tool that is able to read this config and clone the repositories
3. Add functionality to CLI tool to run shell commands and add the possibility for custom commands inside the config file
4. Add filtering options for projects/commands to our CLI tool

### Config

A lot of many-repo CLI tools treat the config file (using either a custom format or json) as a state file that is only interacted with via the executable. So you add a project to the config file via `sometool add git@github.com/random/xyz` and then to remove it, you have to open the config file and remove it manually, taking care to also update the `.gitignore` file.

In my opinion, I think it's a missed opportunity to not let users interact manually with the config file, tend to it as their garden, perhaps align projects in a order that makes sense (instead of alphabetic order, or the random order in which they added the repositories), or add a comment about the repository. It's also seldom you add new repositories, so it's not something that should be optimizied for. Another reason is consistency, users expect that if you add a repository via the CLI tool, then you should be able to remove it via the CLI tool.

That's why in `mani` you need to edit the config file to add or delete a repository (except for when you initialize mani in a directory, then it will scan the directories for git repositories). As a bonus, it also updates your `.gitignore` file with the updated list of repositories.

### Commands

Another big part which a lot of similar softwares miss and delegate to other tools (be it `make` or bash scripts), is custom commands. There are some benefits to include custom commands directly into to the tool:

1. Less tools for developers to learn
2. Less files to keep track of
3. Uniform way to target commands for certain projects

And you can still use make/script files if you have long/complex commands, just call the scripts from `mani`.

So what config format is best suited for this purpose? In my opinion, YAML is a suitable candidate. While it has its issues, I think its purpose as a human-readable config/state file works really well. It has all the primitives you'd need in a config language, simple key/value entries, dictionaries and lists, as well as supporting comments (something which JSON doesn't). We could create a custom format, but then users would have to learn that syntax, so in this case, YAML has a major advantage, almost all software developers are familiar with it.

### Filtering

When we run commands, we need a way to target specific repositories. Now, there are multiple ways to solve this, but the solution should be as flexible as possible. So, to support both dynamic and static filtering. Dynamic means that we're interested in different subset of the repositories, for instance, run a command for all backend services, or all backend services that have a certain tag. Static means that this command always targets certain projects.

To support dynamic filtering, we provide 3 mechanism for filtering:

1. Tag filtering: target projects which have a tag, for instance, add a tag called `c++`, that all c++ projects have
2. Path filtering: target projects by which directory they belong to
3. Project name filtering: target projects by their name

To support static filtering, we simply add a property to `tasks` that contains the predefined tag/path or project names this command should run with.

### General UX

Then there's various small features which makes using mani feel more effortless:

- automatically updating .gitignore when updating the config file
- auto-completion provided by [cobra](https://github.com/spf13/cobra)
- edit the mani config file via the `mani edit` command, which opens up the config file in your preferred editor
- escape hatches: most organizations/people use git, but not everyone uses it, or uses it in the same way, so it's important to provide escape hatches (to support the 1% users), where people can provide their own VCS and customize commands to clone repositories

## Similar Software

- [gita](https://github.com/nosarthur/gita)
- [gr](https://github.com/mixu/gr)
- [meta](https://github.com/mateodelnorte/meta)
- [mu-repo](https://github.com/fabioz/mu-repo)
- [myrepos](https://myrepos.branchable.com/)
- [repo](https://source.android.com/setup/develop/repo)
- [vcstool](https://github.com/dirk-thomas/vcstool)
