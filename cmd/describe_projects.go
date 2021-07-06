package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/print"
	"github.com/alajmo/mani/core/dao"
)

func describeProjectsCmd(config *dao.Config, configErr error) *cobra.Command {
	var tags []string
	var projects []string

	cmd := cobra.Command{
		Aliases: []string { "project", "proj" },
		Use:   "projects [projects] [flags]",
		Short: "Describe projects",
		Long:  "Describe projects.",
		Example: `  # Describe projects
  mani describe projects

  # Describe projects that have tag frontend
  mani describe projects --tags frontend`,
		Run: func(cmd *cobra.Command, args []string) {
			describeProjects(config, args, tags, projects)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			projectNames := config.GetProjectNames()
			return projectNames, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "filter projects by their tag")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		tags := config.GetTags()
		return tags, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func describeProjects(config *dao.Config, args []string, tags []string, projects []string) {
	tagProjects  := config.GetProjectsByTags(tags)
	nameProjects := config.GetProjectsByName(args)

	filteredProjects := dao.GetUnionProjects(tagProjects, nameProjects, dao.Project{})

	print.PrintProjectBlocks(filteredProjects)
}
