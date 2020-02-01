package cmd

import (
	"fmt"
	"github.com/samiralajmovic/loop/core"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	var dryRun bool
	var allProjects bool
	var tags []string
	var projects []string

	cmd := cobra.Command{
		Use:   "run",
		Short: "Run",
		Long: `Run
		Run`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			executeRun(args, dryRun, allProjects, tags, projects)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print what will be ran for each project")
	cmd.Flags().BoolVarP(&allProjects, "all", "a", false, "Specify all projects")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Specify tags")
	cmd.Flags().StringSliceVarP(&projects, "projects", "p", []string{}, "Specify project")

	return &cmd
}

func executeRun(args []string, dryRunFlag bool, allProjectsFlag bool, tagsFlag []string, projectsFlag []string) {
	config := core.ReadConfig()

	command, err := core.GetCommand(args[0], config.Commands)
	userArguments := args[1:]

	if err != nil {
		fmt.Println(err)
		return
	}

	var finalProjects []core.Project
	if allProjectsFlag {
		finalProjects = config.Projects
	} else {
		var tagProjects []core.Project
		if len(tagsFlag) > 0 {
			tagProjects, err = core.GetProjectsByTag(tagsFlag, config.Projects)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		var projects []core.Project
		if len(projectsFlag) > 0 {
			projects, err = core.GetProjects(projectsFlag, config.Projects)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		finalProjects = core.GetUnionProjects(tagProjects, projects)
	}

	core.PrintCommand(command)
	for _, project := range finalProjects {
		err := core.RunCommand(project, command, userArguments, dryRunFlag)

		if err != nil {
			fmt.Println(err)
		}
	}

}
