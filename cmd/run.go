package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func runCmd(config *dao.Config, configErr *error) *cobra.Command {
	var dryRun bool
	var cwd bool
	var describe bool
	var allProjects bool
	var tags []string
	var projects []string
	var output string

	cmd := cobra.Command{
		Use:   "run <task> [flags]",
		Short: "Run tasks",
		Long: `Run tasks.

The tasks are specified in a mani.yaml file along with the projects you can target.`,

		Example: `  # Run task 'pwd' for all projects
  mani run pwd --all-projects

  # Checkout branch 'development' for all projects that have tag 'backend'
  mani run checkout -t backend branch=development`,

		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			run(args, config, output, describe, dryRun, cwd, allProjects, tags, projects)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			return config.GetTasks(), cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().BoolVar(&describe, "describe", true, "Print task information")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "don't execute any task, just print the output of the task to see what will be executed")
	cmd.Flags().BoolVarP(&cwd, "cwd", "k", false, "current working directory")
	cmd.Flags().BoolVarP(&allProjects, "all-projects", "a", false, "target all projects")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "target projects by their tag")
	cmd.Flags().StringSliceVarP(&projects, "projects", "p", []string{}, "target projects by their name")
	cmd.Flags().StringVarP(&output, "output", "o", "list", "Output list|table|markdown|html")

	err := cmd.RegisterFlagCompletionFunc("projects", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		projects := config.GetProjectNames()
		return projects, cobra.ShellCompDirectiveDefault
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

		valid := []string { "table", "markdown", "html" }
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func run(
	args []string,
	config *dao.Config,
	outputFlag string,
	describeFlag bool,
	dryRunFlag bool,
	cwdFlag bool,
	allProjectsFlag bool,
	tagsFlag []string,
	projectsFlag []string,
) {
	projects := config.FilterProjects(cwdFlag, allProjectsFlag, tagsFlag, projectsFlag)

	var taskNames []string
	var userArgs []string
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			userArgs = append(userArgs, arg)
		} else {
			taskNames = append(taskNames, arg)
		}
	}

	for i, cmd := range taskNames {
		task, err := config.GetTask(cmd)
		core.CheckIfError(err)

		runTask(task, projects, userArgs, config, outputFlag, describeFlag, dryRunFlag)

		if i < len(taskNames) {
			fmt.Println()
		}
	}
}

func runTask(
	task *dao.Task,
	projects []dao.Project,
	userArgs []string,
	config *dao.Config,
	outputFlag string,
	describeFlag bool,
	dryRunFlag bool,
) {
	task.SetEnvList(userArgs, config.GetEnv())

	// Set env for sub-commands
	for i := range task.Commands {
		task.Commands[i].SetEnvList(userArgs, config.GetEnv())
	}

	if describeFlag {
		print.PrintTaskBlock([]dao.Task {*task})
	}

	spinner, err := dao.TaskSpinner()
	core.CheckIfError(err)

	err = spinner.Start()
	var data print.TableOutput

	// Headers
	data.Headers = append(data.Headers, "PROJECT")

	if task.Command != "" {
		data.Headers = append(data.Headers, task.Name)
	}

	for _, cmd := range task.Commands {
		data.Headers = append(data.Headers, cmd.Name)
	}

	for i, project := range projects {
		data.Rows = append(data.Rows, table.Row { project.Name })

		spinner.Message(fmt.Sprintf(" %v", project.Name))

		if task.Command != "" {
			output, err := task.RunCmd(config.Path, config.Shell, project, dryRunFlag)
			if err != nil {
				fmt.Println(err)
			}
			data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
		}

		for _, cmd := range task.Commands {
			output, err := cmd.RunCmd(config.Path, config.Shell, project, dryRunFlag)
			if err != nil {
				fmt.Println(err)
			}
			data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
		}
	}

	err = spinner.Stop()
	core.CheckIfError(err)

	print.PrintRun(outputFlag, data)
}
