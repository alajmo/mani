package cmd

import (
	"fmt"
	color "github.com/logrusorgru/aurora"
	"github.com/samiralajmovic/loop/core"
	"github.com/spf13/cobra"
	"strings"
)

func listCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "List projects, commands and tags",
		Long:  "List projects, commands and tags",
		Run: func(cmd *cobra.Command, args []string) {
			list(args)
		},
	}

	return &cmd
}

func list(args []string) {
	config := core.ReadConfig()
	if len(args) > 0 {
		for _, arg := range args {
			switch arg {
			case "projects":
				printProjects(config.Projects)
			case "commands":
				printCommands(config.Commands)
			case "tags":
				tags := core.GetAllTags(config.Projects)
				printTags(tags)
			}
		}
	} else {
		printProjects(config.Projects)
		printCommands(config.Commands)
		tags := core.GetAllTags(config.Projects)
		printTags(tags)
	}
}

func printProjects(projects []core.Project) {
	fmt.Println(color.Blue("Projects").Underline())
	fmt.Println()
	for _, project := range projects {
		fmt.Println(color.Green(project.Name))
		if project.Description != "" {
			fmt.Println(project.Description)
		}

		if len(project.Tags) > 0 {
			fmt.Println(color.Red(strings.Join(project.Tags, ",")))
		}

		fmt.Println()
	}
}

func printCommands(commands []core.Command) {
	fmt.Println(color.Blue("Commands").Underline())
	fmt.Println()
	for _, command := range commands {
		fmt.Println(color.Green(command.Name))
		if command.Description != "" {
			fmt.Println(command.Description)
		}
		fmt.Println()
	}
}

func printTags(tags map[string]struct{}) {
	fmt.Println(color.Blue("Tags").Underline())
	fmt.Println()
	for tag := range tags {
		fmt.Println(color.Red(tag))
	}
	fmt.Println()
}
