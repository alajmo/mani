package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func describeProjectsCmd(config *dao.Config, configErr *error) *cobra.Command {
	var projectFlags core.ProjectFlags

	cmd := cobra.Command{
		Aliases: []string{"project", "proj"},
		Use:     "projects [projects] [flags]",
		Short:   "Describe projects",
		Long:    "Describe projects.",
		Example: `  # Describe projects
  mani describe projects

  # Describe projects that have tag frontend
  mani describe projects --tags frontend`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			describeProjects(config, args, projectFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			projectNames := config.GetProjectNames()
			return projectNames, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringSliceVarP(&projectFlags.Tags, "tags", "t", []string{}, "filter projects by their tag")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetTags()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&projectFlags.ProjectPaths, "project-paths", "d", []string{}, "filter projects by their path")
	err = cmd.RegisterFlagCompletionFunc("project-paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetProjectDirs()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().BoolVarP(&projectFlags.Edit, "edit", "e", false, "Edit project")

	return &cmd
}

func describeProjects(
	config *dao.Config,
	args []string,
	projectFlags core.ProjectFlags,
) {
	if projectFlags.Edit {
		if len(args) > 0 {
			config.EditProject(args[0])
		} else {
			config.EditProject("")
		}
	} else {
		allProjects := false
		if (len(args) == 0 &&
			len(projectFlags.ProjectPaths) == 0 &&
			len(projectFlags.Tags) == 0) {
			allProjects = true
		}

		projects := config.FilterProjects(false, allProjects, projectFlags.ProjectPaths, args, projectFlags.Tags)
		print.PrintProjectBlocks(projects)
	}
}
