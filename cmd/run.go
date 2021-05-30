package cmd

import (
	"fmt"
	"github.com/alajmo/mani/core"
	"github.com/spf13/cobra"
)

func runCmd(configFile *string) *cobra.Command {
	var dryRun bool
	var cwd bool
	var allProjects bool
	var tags []string
	var projects []string

	cmd := cobra.Command{
		Use:   "run <command> [flags]",
		Short: "Run commands",
		Long: `Run commands.

The commands are specified in a mani.yaml file along with the projects you can target.`,

		Example: `  # Run task 'pwd' for all projects
  mani run pwd --all-projects

  # Checkout branch 'development' for all projects that have tag 'backend'
  mani run checkout -t backend branch=development`,

		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			executeRun(args, configFile, dryRun, cwd, allProjects, tags, projects)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "don't execute any command, just print the output of the command to see what will be executed")
	cmd.Flags().BoolVarP(&cwd, "cwd", "k", false, "current working directory")
	cmd.Flags().BoolVarP(&allProjects, "all-projects", "a", false, "target all projects")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "target projects by their tag")
	cmd.Flags().StringSliceVarP(&projects, "projects", "p", []string{}, "target projects by their name")

	cmd.MarkFlagCustom("projects", "__mani_parse_projects")
	cmd.MarkFlagCustom("tags", "__mani_parse_tags")

	return &cmd
}

func executeRun(args []string, configFile *string, dryRunFlag bool, cwdFlag bool, allProjectsFlag bool, tagsFlag []string, projectsFlag []string) {
	configPath, config, err := core.ReadConfig(*configFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	command, err := core.GetCommand(args[0], config.Commands)
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
			tagProjects = core.GetProjectsByTag(tagsFlag, config.Projects)
		}

		var projects []core.Project
		if len(projectsFlag) > 0 {
			projects = core.GetProjects(projectsFlag, config.Projects)
		}

		var cwdProject core.Project
		if cwdFlag {
			cwdProject = core.GetCwdProject(config.Projects)
		}

		finalProjects = core.GetUnionProjects(tagProjects, projects, cwdProject)
	}

	userArguments := args[1:]
	core.PrintCommand(command)
	for _, project := range finalProjects {
		err := core.RunCommand(configPath, project, command, userArguments, dryRunFlag)

		if err != nil {
			fmt.Println(err)
		}
	}

}
