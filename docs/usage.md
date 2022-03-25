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

## Run Some Commands

Now let's run some commands to see everything is working as expected.

```bash
# List all projects
$ mani list projects

# Count number of files in each project in parallel
$ mani exec --all --output table --parallel 'find . -type f | wc -l'
```

Next up:

- [Some more examples](/examples)
- [Familiarize yourself with the mani.yaml config](/config)
- [Checkout mani commands](/commands)
