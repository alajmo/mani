package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func describeProjectsCmd(config *dao.Config, configErr *error) *cobra.Command {
	var projectFlags core.ProjectFlags

	cmd := cobra.Command{
		Aliases: []string{"project", "prj"},
		Use:     "projects [projects]",
		Short:   "Describe projects",
		Long:    "Describe projects.",
		Example: `  # Describe all projects
  mani describe projects

  # Describe project <project>
  mani describe projects <project>

  # Describe projects that have tag <tag>
  mani describe projects --tags <tag>

  # Describe projects matching paths <path>
  mani describe projects --paths <path>`,
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
		DisableAutoGenTag: true,
	}

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

	cmd.Flags().BoolVarP(&projectFlags.Edit, "edit", "e", false, "edit project")

	return &cmd
}

func describeProjects(
	config *dao.Config,
	args []string,
	projectFlags core.ProjectFlags,
) {
	if projectFlags.Edit {
		if len(args) > 0 {
			err := config.EditProject(args[0])
			core.CheckIfError(err)
		} else {
			err := config.EditProject("")
			core.CheckIfError(err)
		}
	} else {
		allProjects := false
		if len(args) == 0 &&
			len(projectFlags.Paths) == 0 &&
			len(projectFlags.Tags) == 0 {
			allProjects = true
		}

		projects, err := config.FilterProjects(false, allProjects, args, projectFlags.Paths, projectFlags.Tags)
		core.CheckIfError(err)
		if len(projects) == 0 {
			fmt.Println("No projects")
		} else {
			print.PrintProjectBlocks(projects)
		}
	}
}
