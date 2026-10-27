# Usage

## Initialize Mani

Run the following command inside a directory containing your `git` repositories:

```bash
mani init
```

This will generate:

- `mani.yaml`: Contains projects and custom tasks. Any subdirectory that has a `.git` directory will be included (add the flag `--auto-discovery=false` to turn off this feature)
- `.gitignore`: (only when inside a git repo) Includes the projects specified in `mani.yaml` file. To opt out, use `mani init --sync-gitignore=false`.

It can be helpful to initialize the `mani` repository as a git repository so that anyone can easily download the `mani` repository and run `mani sync` to clone all repositories and get the same project setup as you.

## Example Commands

```bash
# List all projects
mani list projects

# Run git status across all projects
mani exec --all git status

# Run git status across all projects in parallel with output in table format
mani exec --all --parallel --output table git status
```

Next up:

- [Some more examples](/examples)
- [Familiarize yourself with the mani.yaml config](/config)
- [Checkout mani commands](/commands)
