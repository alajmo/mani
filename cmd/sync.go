package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/exec"
)

func syncCmd(config *dao.Config, configErr *error) *cobra.Command {
	var projectFlags core.ProjectFlags
	var syncFlags core.SyncFlags

	cmd := cobra.Command{
		Use:     "sync",
		Aliases: []string{"clone"},
		Short:   "Clone repositories and add them to gitignore",
		Long: `Clone repositories and add them to gitignore.
In-case you need to enter credentials before cloning, run the command without the parallel flag.`,
		Example: `  # Clone repositories one at a time
  mani sync

  # Clone repositories in parallell
  mani sync --parallel

  # Show cloned projects
  mani sync --status`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)

			// This is necessary since cobra doesn't support pointers for bools
			// (that would allow us to use nil as default value)
			syncFlags.Parallel = cmd.Flags().Changed("parallel")
			runSync(config, args, projectFlags, syncFlags)
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

	cmd.Flags().BoolVarP(&syncFlags.Parallel, "parallel", "p", false, "clone projects in parallel")
	cmd.Flags().BoolVarP(&syncFlags.Status, "status", "s", false, "print sync status only")

	cmd.Flags().StringSliceVarP(&projectFlags.Tags, "tags", "t", []string{}, "filter projects by tags")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetTags()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&projectFlags.Paths, "paths", "d", []string{}, "filter projects by paths")
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

func runSync(config *dao.Config, args []string, projectFlags core.ProjectFlags, syncFlags core.SyncFlags) {
	allProjects := false
	if len(args) == 0 &&
		len(projectFlags.Paths) == 0 &&
		len(projectFlags.Tags) == 0 {
		allProjects = true
	}

	projects, err := config.FilterProjects(false, allProjects, args, projectFlags.Paths, projectFlags.Tags)
	core.CheckIfError(err)

	if !syncFlags.Status {
		err := exec.UpdateGitignoreIfExists(config)
		core.CheckIfError(err)

		err = exec.CloneRepos(config, projects, syncFlags.Parallel)
		core.CheckIfError(err)
	}

	err = exec.PrintProjectStatus(config, projects)
	core.CheckIfError(err)
}
