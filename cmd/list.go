package cmd

import (
	"fmt"
	"github.com/samiralajmovic/mani/core"
	"github.com/spf13/cobra"
)

func listCmd(configFile *string) *cobra.Command {
	var validArgs = []string{"projects", "tags", "commands"}

	cmd := cobra.Command{
		Use:   "list <projects|tags|commands>",
		Short: "List projects, commands and tags",
		Long:  "List projects, commands and tags.",
		Example: `  # List projects
  mani list projects`,
		Args:  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			list(configFile, args)
		},
		ValidArgs: validArgs,
	}

	return &cmd
}

func list(configFile *string, args []string) {
	_, config, err := core.ReadConfig(*configFile)

	if err != nil {
		fmt.Println(err)
		return
	}

	switch args[0] {
	case "projects":
		core.PrintProjects(config.Projects)
	case "commands":
		core.PrintCommands(config.Commands)
	case "tags":
		tags := core.GetAllTags(config.Projects)
		core.PrintTags(tags)
	}
}
