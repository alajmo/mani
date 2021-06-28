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
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			_, config, err := core.ReadConfig(*configFile)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			return core.GetCommands(config.Commands), cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "don't execute any command, just print the output of the command to see what will be executed")
	cmd.Flags().BoolVarP(&cwd, "cwd", "k", false, "current working directory")
	cmd.Flags().BoolVarP(&allProjects, "all-projects", "a", false, "target all projects")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "target projects by their tag")
	cmd.Flags().StringSliceVarP(&projects, "projects", "p", []string{}, "target projects by their name")

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

func executeRun(args []string, configFile *string, dryRunFlag bool, cwdFlag bool, allProjectsFlag bool, tagsFlag []string, projectsFlag []string) {
	configPath, config, err := core.ReadConfig(*configFile)
	core.CheckIfError(err)

	command, err := core.GetCommand(args[0], config.Commands)
	core.CheckIfError(err)

	if command.Shell != "" {
		config.Shell = command.Shell
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
		err := core.RunCommand(configPath, config.Shell, project, command, userArguments, dryRunFlag)

		if err != nil {
			fmt.Println(err)
		}
	}

}
