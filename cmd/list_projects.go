package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func listProjectsCmd(config *dao.Config, configErr *error, listFlags *core.ListFlags) *cobra.Command {
	var projectFlags core.ProjectFlags

	cmd := cobra.Command{
		Aliases: []string{"project", "proj", "pr"},
		Use:     "projects [flags]",
		Short:   "List projects",
		Long:    "List projects",
		Example: `  # List projects
  mani list projects`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			listProjects(config, args, listFlags, &projectFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			projectNames := config.GetProjectNames()
			return projectNames, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().BoolVar(&listFlags.Tree, "tree", false, "Tree output")

	cmd.Flags().StringSliceVarP(&projectFlags.Tags, "tags", "t", []string{}, "filter projects by their tag")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetTags()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&projectFlags.Paths, "paths", "d", []string{}, "filter projects by their path")
	err = cmd.RegisterFlagCompletionFunc("paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetProjectPaths()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVar(&projectFlags.Headers, "headers", []string{"project", "tag", "description"}, "Specify headers, defaults to project, tag, description")
	err = cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := []string{"project", "path", "relpath", "description", "url", "tag"}
		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listProjects(config *dao.Config, args []string, listFlags *core.ListFlags, projectFlags *core.ProjectFlags) {
	theme, err := config.GetTheme(listFlags.Theme)
	core.CheckIfError(err)

	if listFlags.Tree {
		tree, err := config.GetProjectsTree(projectFlags.Paths, projectFlags.Tags)
		core.CheckIfError(err)

		print.PrintTree(config, *theme, listFlags, tree)
		return
	}

	allProjects := false
	if len(args) == 0 &&
		len(projectFlags.Paths) == 0 &&
		len(projectFlags.Tags) == 0 {
		allProjects = true
	}

	projects, err := config.FilterProjects(false, allProjects, projectFlags.Paths, args, projectFlags.Tags)
	core.CheckIfError(err)

	options := print.PrintTableOptions{
		Output:               listFlags.Output,
		Theme:                *theme,
		Tree:                 listFlags.Tree,
		OmitEmpty:            false,
		SuppressEmptyColumns: true,
	}

	print.PrintTable(projects, options, projectFlags.Headers, []string{})
}
