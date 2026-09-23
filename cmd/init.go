package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/exec"
)

func initCmd() *cobra.Command {
	var initFlags core.InitFlags

	cmd := cobra.Command{
		Use:   "init",
		Short: "Initialize a mani repository",
		Long: `Initialize a mani repository.

Creates a new mani repository by generating a mani.yaml configuration file
and a .gitignore file in the current directory.

With --source <provider>/<name>, mani fetches repositories from a remote
source and scaffolds mani.yaml from them. The query is persisted in a
'sources:' block so 'mani refresh' can pick up new repos later.

Supported providers:
  - github/<name>                 auto-detect org or user           (requires 'gh')
  - github-org/<name>             explicit organization             (requires 'gh')
  - github-user/<name>            explicit user                     (requires 'gh')
  - gitea/<name>, forgejo/<name>,
    codeberg/<name>               auto-detect (all Gitea-compatible) (requires 'tea')
  - gitea-org/<name>, forgejo-org/<name>, codeberg-org/<name>        (requires 'tea')
  - gitea-user/<name>, forgejo-user/<name>, codeberg-user/<name>    (requires 'tea')

The provider's CLI must be installed and authenticated before use.
The auto-detect form (github/, gitea/, …) probes the org endpoint first
and falls back to the user endpoint; explicit -org/-user forms skip the
probe and go directly to the right endpoint.`,

		Example: `  # Initialize with default settings (local auto-discovery)
  mani init

  # Initialize from a GitHub organization
  mani init --source github/my-org

  # Initialize from a Codeberg user
  mani init --source codeberg/some-user

  # Initialize from a GitHub org, placing all repos under a common path
  mani init --source github/my-org --source-path my-org

  # Initialize from a Gitea org, including archived repos and forks
  mani init --source gitea/my-org --archived

  # Initialize from a GitHub org with a smaller limit and no topic tags
  mani init --source github/my-org --limit 50 --no-topic-tags

  # Initialize without auto-discovering projects
  mani init --auto-discovery=false

  # Initialize without updating .gitignore
  mani init --sync-gitignore=false`,

		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if initFlags.SourcePath != "" && initFlags.Source == "" {
				core.CheckIfError(fmt.Errorf("--source-path requires --source"))
			}

			initFlags.AutoDiscoverySet = cmd.Flags().Changed("auto-discovery")

			foundProjects, err := dao.InitMani(args, initFlags)
			core.CheckIfError(err)

			// Only print the discovered-from-disk list when we did a
			// local discovery walk. In --source mode the projects came
			// from GitHub, not the filesystem.
			if initFlags.AutoDiscovery && initFlags.Source == "" {
				exec.PrintProjectInit(foundProjects)
			}
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVar(&initFlags.AutoDiscovery, "auto-discovery", true, "automatically discover and add Git repositories to mani.yaml")
	cmd.Flags().BoolVarP(&initFlags.SyncGitignore, "sync-gitignore", "g", true, "synchronize .gitignore file")

	cmd.Flags().StringVar(&initFlags.Source, "source", "", "remote source to scaffold projects from (e.g. github/my-org, gitea/foo, codeberg/joe)")
	cmd.Flags().StringVar(&initFlags.SourcePath, "source-path", "", "common path prefix under which the source's projects are placed")
	cmd.Flags().IntVar(&initFlags.Limit, "limit", 1000, "maximum number of repositories to fetch from the remote source")
	cmd.Flags().BoolVar(&initFlags.Archived, "archived", false, "include archived repositories (default: exclude)")
	cmd.Flags().BoolVar(&initFlags.Fork, "fork", true, "include forked repositories")
	cmd.Flags().BoolVar(&initFlags.NoTags, "no-tags", false, "omit synthesized tags from generated projects")
	cmd.Flags().BoolVar(&initFlags.NoTopicTags, "no-topic-tags", false, "omit topic=<name> tags")

	return &cmd
}
