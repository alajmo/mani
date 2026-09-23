package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func refreshCmd() *cobra.Command {
	var dryRun, prune, autoDiscovery bool

	cmd := cobra.Command{
		Use:   "refresh",
		Short: "Update mani.yaml with new projects from remote sources and local clones",
		Long: `Update the 'projects:' mapping in mani.yaml from two places:

  1. Remote sources — re-query each entry in the 'sources:' block and append
     any new upstream repositories.
  2. Local clones — unless --auto-discovery=false, walk the config directory
     for git repositories that aren't listed yet and append them (the same
     discovery 'mani init' performs). This is how a repo you 'git clone' into
     the workspace by hand gets picked up.

By default, refresh only adds projects. Existing entries are left untouched —
any local edits (renames, custom tags, env, etc.) are preserved. Projects
whose upstream repository is gone are reported as stale; use --prune to also
remove them. Pruning is scoped to each source's remote, so projects from a
different remote (or discovered locally) are never removed.

The 'sources:' block is optional: a config from a plain 'mani init' has none,
and refresh still picks up local clones. Sources do require the provider CLI
to be installed and authenticated:
  - github sources need 'gh'
  - gitea/forgejo/codeberg sources need 'tea'`,

		Example: `  # Append new upstream repos and locally-cloned repos to projects:
  mani refresh

  # Show what would change without writing
  mani refresh --dry-run

  # Only re-query remote sources, skip local discovery
  mani refresh --auto-discovery=false

  # Also delete projects whose upstream repo is gone
  mani refresh --prune`,

		Run: func(cmd *cobra.Command, args []string) {
			// Only override the config's auto_discovery field when the user
			// actually passed --auto-discovery.
			var autoDiscoveryOverride *bool
			if cmd.Flags().Changed("auto-discovery") {
				autoDiscoveryOverride = &autoDiscovery
			}

			res, err := dao.RefreshMani("", prune, dryRun, autoDiscoveryOverride)
			core.CheckIfError(err)

			printRefreshSummary(res, prune)
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would change but do not modify mani.yaml")
	cmd.Flags().BoolVar(&prune, "prune", false, "remove projects whose upstream repository no longer exists")
	cmd.Flags().BoolVar(&autoDiscovery, "auto-discovery", true, "discover locally-cloned git repositories and add them to mani.yaml")

	return &cmd
}

func printRefreshSummary(res *dao.RefreshResult, prune bool) {
	if res == nil {
		return
	}

	prefix := ""
	if res.DryRun {
		prefix = "[dry-run] "
	}

	switch {
	case len(res.Added) == 0 && len(res.Removed) == 0 && len(res.Stale) == 0:
		fmt.Printf("%smani.yaml is up to date.\n", prefix)
	default:
		fmt.Printf("%sRefreshed %s\n", prefix, res.ConfigPath)
	}

	if len(res.Added) > 0 {
		fmt.Printf("\n%sAdded (%d):\n", prefix, len(res.Added))
		for _, p := range res.Added {
			fmt.Printf("  + %s\n", p.Name)
		}
	}
	if len(res.Removed) > 0 {
		fmt.Printf("\n%sRemoved (%d):\n", prefix, len(res.Removed))
		for _, p := range res.Removed {
			fmt.Printf("  - %s\n", p.Name)
		}
	}
	if len(res.Stale) > 0 {
		fmt.Printf("\nStale (%d) — upstream gone but not pruned (pass --prune to remove):\n", len(res.Stale))
		for _, p := range res.Stale {
			fmt.Printf("  ? %s\n", p.Name)
		}
	}
}
