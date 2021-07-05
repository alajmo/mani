package cmd

import (
	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/print"
	"github.com/spf13/cobra"
)

func listTagsCmd(configFile *string) *cobra.Command {
	var projects []string

	cmd := cobra.Command {
		Aliases: []string { "tag" },
		Use:   "tags [flags]",
		Short: "List tags",
		Long:  "List tags.",
		Example: `  # List tags
  mani list tags`,
		Run: func(cmd *cobra.Command, args []string) {
			listTags(configFile, args, projects)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			_, config, err := core.ReadConfig(*configFile)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			tags := core.GetTags(config.Projects)
			return tags, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringSliceVarP(&projects, "projects", "p", []string{}, "filter tags by their project")
	err := cmd.RegisterFlagCompletionFunc("projects", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, config, err := core.ReadConfig(*configFile)

		if err != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		projects := core.GetProjectNames(config.Projects)
		return projects, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listTags(configFile *string, args []string, projects []string) {
	_, config, err := core.ReadConfig(*configFile)
	core.CheckIfError(err)

	allTags := core.GetTags(config.Projects)
	if (len(args) == 0 && len(projects) == 0) {
		print.PrintTags(allTags)
		return
	}

	if (len(args) > 0 && len(projects) == 0) {
		args = core.Intersection(args, allTags)
		print.PrintTags(args)
	} else if (len(args) == 0 && len(projects) > 0) {
		projectTags := core.FilterTagOnProject(config.Projects, projects)
		print.PrintTags(projectTags)
	} else {
		projectTags := core.FilterTagOnProject(config.Projects, projects)
		args = core.Intersection(args, projectTags)
		print.PrintTags(args)
	}
}
