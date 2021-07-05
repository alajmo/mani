package cmd

import (
	"fmt"
	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/print"
	"github.com/spf13/cobra"
)

func runCmd(configFile *string) *cobra.Command {
	var dryRun bool
	var cwd bool
	var allProjects bool
	var tags []string
	var projects []string
	var format string

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
			executeRun(args, configFile, format, dryRun, cwd, allProjects, tags, projects)
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
	cmd.Flags().StringVarP(&format, "format", "f", "list", "Format list|table|markdown|html")

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

	err = cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, _, err := core.ReadConfig(*configFile)

		if err != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validFormats := []string { "table", "markdown", "html" }
		return validFormats, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func executeRun(
	args []string,
	configFile *string,
	format string,
	dryRunFlag bool,
	cwdFlag bool,
	allProjectsFlag bool,
	tagsFlag []string,
	projectsFlag []string,
) {
	configPath, config, err := core.ReadConfig(*configFile)
	core.CheckIfError(err)

	command, err := core.GetCommand(args[0], config.Commands)
	core.CheckIfError(err)

	userArguments := args[1:]
	command.Args = core.ParseUserArguments(command.Args, userArguments)
	userArguments = core.GetUserArguments(command.Args)

	finalProjects := core.FilterProjects(config, cwdFlag, allProjectsFlag, tagsFlag, projectsFlag)
	print.PrintCommandBlocks([]core.Command {*command})

	outputs := make(map[string]string)
	for _, project := range finalProjects {
		output, err := core.RunCommand(configPath, config.Shell, project, command, userArguments, dryRunFlag)
		if err != nil {
			fmt.Println(err)
		}

		outputs[project.Name] = output
	}

	print.PrintRun(format, outputs)
}
