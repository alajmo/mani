package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/exec"
)

func syncCmd(config *dao.Config, configErr *error) *cobra.Command {
	var projectFlags core.ProjectFlags
	var syncFlags = core.SyncFlags{Forks: 4}
	var setSyncFlags core.SetSyncFlags

	cmd := cobra.Command{
		Use:     "sync",
		Aliases: []string{"clone"},
		Short:   "Clone repositories and update .gitignore",
		Long: `Clone repositories and update .gitignore file.
For repositories requiring authentication, disable parallel cloning to enter
credentials for each repository individually.`,
		Example: `  # Clone repositories one at a time
  mani sync

  # Clone repositories in parallel
  mani sync --parallel

  # Disable updating .gitignore file
  mani sync --sync-gitignore=false

  # Sync project remotes. This will modify the projects .git state
  mani sync --sync-remotes

  # Clone repositories even if project sync field is set to false
  mani sync --ignore-sync-state

  # Display sync status
  mani sync --status`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)

			// This is necessary since cobra doesn't support pointers for bools
			// (that would allow us to use nil as default value)
			setSyncFlags.Parallel = cmd.Flags().Changed("parallel")
			setSyncFlags.SyncGitignore = cmd.Flags().Changed("sync-gitignore")
			setSyncFlags.SyncRemotes = cmd.Flags().Changed("sync-remotes")
			setSyncFlags.RemoveOrphanedWorktrees = cmd.Flags().Changed("remove-orphaned-worktrees")
			setSyncFlags.Forks = cmd.Flags().Changed("forks")

			if setSyncFlags.Forks {
				forks, err := cmd.Flags().GetUint32("forks")
				core.CheckIfError(err)
				if forks == 0 {
					core.Exit(&core.ZeroNotAllowed{Name: "forks"})
				}
				syncFlags.Forks = forks
			}

			runSync(config, args, projectFlags, syncFlags, setSyncFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			projectNames := config.GetProjectNames()
			return projectNames, cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVarP(&syncFlags.SyncRemotes, "sync-remotes", "r", false, "update git remote state")
	cmd.Flags().BoolVarP(&syncFlags.RemoveOrphanedWorktrees, "remove-orphaned-worktrees", "w", false, "remove git worktrees not in config")
	cmd.Flags().BoolVarP(&syncFlags.SyncGitignore, "sync-gitignore", "g", true, "sync gitignore")
	cmd.Flags().BoolVar(&syncFlags.IgnoreSyncState, "ignore-sync-state", false, "sync project even if the project's sync field is set to false")
	cmd.Flags().BoolVarP(&syncFlags.Parallel, "parallel", "p", false, "clone projects in parallel")
	cmd.Flags().BoolVarP(&syncFlags.Status, "status", "s", false, "display status only")
	cmd.Flags().Uint32P("forks", "f", 4, "maximum number of concurrent processes")

	// Targets
	cmd.Flags().StringSliceVarP(&projectFlags.Tags, "tags", "t", []string{}, "clone projects by tags")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetTags()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringVarP(&projectFlags.TagsExpr, "tags-expr", "E", "", "clone projects by tag expression")
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&projectFlags.Paths, "paths", "d", []string{}, "clone projects by path")
	err = cmd.RegisterFlagCompletionFunc("paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetProjectPaths()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func runSync(
	config *dao.Config,
	args []string,
	projectFlags core.ProjectFlags,
	syncFlags core.SyncFlags,
	setSyncFlags core.SetSyncFlags,
) {

	// If no flag is set for targetting projects, then assume all projects
	var allProjects bool
	if len(args) == 0 &&
		projectFlags.TagsExpr == "" &&
		len(projectFlags.Paths) == 0 &&
		len(projectFlags.Tags) == 0 {
		allProjects = true
	}

	projects, err := config.FilterProjects(false, allProjects, args, projectFlags.Paths, projectFlags.Tags, projectFlags.TagsExpr)
	core.CheckIfError(err)

	if !syncFlags.Status {
		if setSyncFlags.SyncRemotes {
			config.SyncRemotes = &syncFlags.SyncRemotes
		}

		if setSyncFlags.RemoveOrphanedWorktrees {
			config.RemoveOrphanedWorktrees = &syncFlags.RemoveOrphanedWorktrees
		}

		if setSyncFlags.SyncGitignore {
			config.SyncGitignore = &syncFlags.SyncGitignore
		}

		if *config.SyncGitignore {
			err := exec.UpdateGitignoreIfExists(config)
			core.CheckIfError(err)
		}

		err = exec.CloneRepos(config, projects, syncFlags)
		core.CheckIfError(err)
	}

	err = exec.PrintProjectStatus(config, projects)
	core.CheckIfError(err)
}
