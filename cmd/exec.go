package cmd

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func execCmd(config *dao.Config, configErr *error) *cobra.Command {
	var dryRun bool
	var cwd bool
	var allProjects bool
	var projectPaths []string
	var tags []string
	var projects []string
	var output string

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
			core.CheckIfError(*configErr)
			execute(args, config, output, dryRun, cwd, allProjects, projectPaths, tags, projects)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "don't execute any command, just print the output of the command to see what will be executed")
	cmd.Flags().BoolVarP(&cwd, "cwd", "k", false, "current working directory")
	cmd.Flags().BoolVarP(&allProjects, "all-projects", "a", false, "target all projects")
	cmd.Flags().StringSliceVarP(&projectPaths, "project-paths", "d", []string{}, "target projects by their path")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "target projects by their tag")
	cmd.Flags().StringSliceVarP(&projects, "projects", "p", []string{}, "target projects by their name")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output list|table|markdown|html")

	err := cmd.RegisterFlagCompletionFunc("projects", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		projects := config.GetProjectNames()
		return projects, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	err = cmd.RegisterFlagCompletionFunc("project-paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetProjectDirs()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	err = cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		tags := config.GetTags()
		return tags, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	err = cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		valid := []string{"table", "markdown", "html"}
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func execute(
	args []string,
	config *dao.Config,
	outputFlag string,
	dryRunFlag bool,
	cwdFlag bool,
	allProjectsFlag bool,
	projectPathsFlag []string,
	tagsFlag []string,
	projectsFlag []string,
) {
	// Table Style
	// switch config.Theme.Table {
	// case "ascii":
	// 	core.ManiList.Box = core.StyleBoxASCII
	// default:
	// 	core.ManiList.Box = core.StyleBoxDefault
	// }

	projects := config.FilterProjects(cwdFlag, allProjectsFlag, projectPathsFlag, projectsFlag, tagsFlag)

	if len(projects) == 0 {
		fmt.Println("No projects targeted")
		return
	}

	spinner, err := dao.TaskSpinner()
	core.CheckIfError(err)

	err = spinner.Start()
	core.CheckIfError(err)

	cmd := strings.Join(args[0:], " ")
	var data core.TableOutput

	data.Headers = table.Row{"Project", "Output"}

	for i, project := range projects {
		data.Rows = append(data.Rows, table.Row{project.Name})

		spinner.Message(fmt.Sprintf(" %v", project.Name))

		output, err := dao.ExecCmd(config.Path, project, cmd, dryRunFlag)
		if err != nil {
			data.Rows[i] = append(data.Rows[i], err)
		} else {
			data.Rows[i] = append(data.Rows[i], output)
		}
	}

	err = spinner.Stop()
	core.CheckIfError(err)

	// render.Render(outputFlag, data)
}
