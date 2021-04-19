package cmd

import (
	"github.com/alajmo/mani/core"
	"github.com/spf13/cobra"
)

func listCmd(configFile *string) *cobra.Command {
	var validArgs = []string{"projects", "tags", "commands"}
	var listRaw bool
	var tags []string
	var projects []string

	cmd := cobra.Command{
		Use:   "list <projects|tags|commands> [flags]",
		Short: "List projects, commands and tags",
		Long:  "List projects, commands and tags.",
		Example: `  # List projects
  mani list projects`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			list(configFile, args, listRaw, tags, projects)
		},
		ValidArgs: validArgs,
	}

	cmd.Flags().BoolVar(&listRaw, "list-raw", false, "When listing objects, ignore description")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "filter projects by their tag")
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

	err = cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, config, err := core.ReadConfig(*configFile)

		if err != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		tags := core.GetTags(config.Projects)
		return tags, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func list(configFile *string, args []string, listRaw bool, tags []string, projects []string) {
	_, config, err := core.ReadConfig(*configFile)
	core.CheckIfError(err)

	switch args[0] {
	case "projects":
		filteredProjects := core.FilterProjectOnTag(config.Projects, tags)
		core.PrintProjects(filteredProjects, listRaw)
	case "tags":
		var filteredTags []string
		if len(projects) > 0 {
			filteredTags = core.FilterTagOnProject(config.Projects, projects)
		} else {
			filteredTags = core.GetTags(config.Projects)
		}

		core.PrintTags(filteredTags)
	case "commands":
		core.PrintCommands(config.Commands, listRaw)
	}
}
