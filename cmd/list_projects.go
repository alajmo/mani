package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/print"
	"github.com/alajmo/mani/core/dao"
)

func listProjectsCmd(config *dao.Config, configErr error, listFlags *print.ListFlags) *cobra.Command {
	var projectFlags print.ListProjectFlags

	cmd := cobra.Command{
		Aliases: []string { "project", "proj" },
		Use:   "projects [flags]",
		Short: "List projects",
		Long:  "List projects",
		Example: `  # List projects
  mani list projects`,
		Run: func(cmd *cobra.Command, args []string) {
			listProjects(config, args, listFlags, &projectFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			projectNames := config.GetProjectNames()
			return projectNames, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringSliceVarP(&projectFlags.Tags, "tags", "t", []string{}, "filter projects by their tag")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validTags := config.GetTags()
		return validTags, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVar(&projectFlags.Headers, "headers", []string{ "name", "tags", "description" }, "Specify headers, defaults to name, tags, description")
	err = cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := []string { "name", "path", "relpath", "description", "url", "tags" }
		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listProjects(
	config *dao.Config,
	args []string,
	listFlags *print.ListFlags,
	projectFlags *print.ListProjectFlags,
) {
	tagProjects := config.GetProjectsByTags(projectFlags.Tags)
	nameProjects := config.GetProjectsByName(args)

	filteredProjects := dao.GetUnionProjects(tagProjects, nameProjects, dao.Project{})

	print.PrintProjects(filteredProjects, *listFlags, *projectFlags)
}
