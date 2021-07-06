package cmd

import (
	"fmt"
	core "github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/print"
	"github.com/spf13/cobra"
	"strings"
)

func execCmd(configFile *string) *cobra.Command {
	var dryRun bool
	var cwd bool
	var allProjects bool
	var tags []string
	var projects []string
	var format string

	cmd := cobra.Command{
		Use:   "exec <command>",
		Short: "Execute arbitrary commands",
		Long: `Execute arbitrary commands.

Single quote your command if you don't want the file globbing and environments variables expansion to take place
before the command gets executed in each directory.`,

		Example: `  # List files in all projects
  mani exec ls --all-projects

  # List all git files that have markdown suffix
  mani exec 'git ls-files | grep -e ".md"' --all-projects`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			executeCmd(args, configFile, format, dryRun, cwd, allProjects, tags, projects)
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

func executeCmd(
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

	finalProjects := core.FilterProjects(config, cwdFlag, allProjectsFlag, tagsFlag, projectsFlag)

	cmd := strings.Join(args[0:], " ")
	var outputs []core.ProjectOutput
	for _, project := range finalProjects {
		output, err := core.ExecCmd(configPath, config.Shell, project, cmd, dryRunFlag)
		if err != nil {
			fmt.Println(err)
		}

		outputs = append(outputs, core.ProjectOutput { 
			ProjectName: project.Name, 
			Output: output,
		})
	}

	print.PrintRun(format, outputs)
}
