package cmd

import (
	"github.com/alajmo/mani/core"
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

	var filteredTags []string
	if len(projects) > 0 {
		filteredTags = core.FilterTagOnProject(config.Projects, projects)
	} else {
		filteredTags = core.GetTags(config.Projects)
	}

	core.PrintTags(filteredTags)
}
