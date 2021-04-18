package cmd

import (
	"fmt"
	"github.com/alajmo/mani/core"
	"github.com/spf13/cobra"
)

func listCmd(configFile *string) *cobra.Command {
	var validArgs = []string{"projects", "tags", "commands"}
	var tags []string
	var projects []string

	cmd := cobra.Command{
		Use:   "list <projects|tags|commands> [flags]",
		Short: "List projects, commands and tags",
		Long:  "List projects, commands and tags.",
		Example: `  # List projects
  mani list projects`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			list(configFile, args, tags, projects)
		},
		ValidArgs: validArgs,
	}

	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "filter projects by their tag")
	cmd.Flags().StringSliceVarP(&projects, "projects", "p", []string{}, "filter tags by their project")

	cmd.MarkFlagCustom("projects", "__mani_parse_projects")
	cmd.MarkFlagCustom("tags", "__mani_parse_tags")

	return &cmd
}

func list(configFile *string, args []string, tags []string, projects []string) {
	_, config, err := core.ReadConfig(*configFile)

	if err != nil {
		fmt.Println(err)
		return
	}

	switch args[0] {
	case "projects":
		filteredProjects := core.FilterProjectOnTag(config.Projects, tags)
		core.PrintProjects(filteredProjects)
	case "tags":
		var filteredTags map[string]struct{}
		if (len(projects) > 0) {
			filteredTags = core.FilterTagOnProject(config.Projects, projects)
		} else {
			filteredTags = core.GetTags(config.Projects)
		}

		core.PrintTags(filteredTags)
	case "commands":
		core.PrintCommands(config.Commands)
	}
}
