package cmd

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func runCmd(config *dao.Config, configErr *error) *cobra.Command {
	var edit bool
	var serial bool
	var dryRun bool
	var cwd bool
	var describe bool
	var allProjects bool
	var dirs []string
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
			run(args, config, output, describe, dryRun, edit, serial, cwd, allProjects, dirs, tags, projects)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			return config.GetTasks(), cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().BoolVar(&describe, "describe", true, "Print task information")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "don't execute any task, just print the output of the task to see what will be executed")
	cmd.Flags().BoolVarP(&edit, "edit", "e", false, "Edit task")
	cmd.Flags().BoolVarP(&serial, "serial", "s", false, "Run tasks in serial")
	cmd.Flags().BoolVarP(&cwd, "cwd", "k", false, "current working directory")
	cmd.Flags().BoolVarP(&allProjects, "all-projects", "a", false, "target all projects")
	cmd.Flags().StringSliceVarP(&dirs, "dirs", "d", []string{}, "target projects by their path")
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

	err = cmd.RegisterFlagCompletionFunc("dirs", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetDirs()
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
	editFlag bool,
	serialFlag bool,
	cwdFlag bool,
	allProjectsFlag bool,
	dirsFlag []string,
	tagsFlag []string,
	projectsFlag []string,
) {
	var taskNames []string
	var userArgs []string
	// Seperate user arguments from task names
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			userArgs = append(userArgs, arg)
		} else {
			taskNames = append(taskNames, arg)
		}
	}

	if (editFlag) {
		if len(args) > 0 {
			config.EditTask(taskNames[0])
			return
		} else {
			config.EditTask("")
			return
		}
	}

	for i, cmd := range taskNames {
		task, err := config.GetTask(cmd)
		core.CheckIfError(err)

		if task.Output != "" && outputFlag == "" {
			outputFlag = task.Output
		}

		if len(dirsFlag) == 0 {
			dirsFlag = task.Dirs
		}

		if len(tagsFlag) == 0 {
			tagsFlag = task.Tags
		}

		if len(projectsFlag) == 0 {
			projectsFlag = task.Projects
		}

		projects := config.FilterProjects(cwdFlag, allProjectsFlag, dirsFlag, tagsFlag, projectsFlag)
		if len(projects) == 0 {
			fmt.Println("No projects targeted")
			continue
		}

		runTask(task, projects, userArgs, config, outputFlag, serialFlag, describeFlag, dryRunFlag)

		// Newline seperator between tasks
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
	serialFlag bool,
	describeFlag bool,
	dryRunFlag bool,
) {
	task.SetEnvList(userArgs, []string{}, config.GetEnv())

	// Set env for sub-commands
	for i := range task.Commands {
		task.Commands[i].SetEnvList(userArgs, task.EnvList, config.GetEnv())
	}

	if describeFlag {
		print.PrintTaskBlock([]dao.Task {*task})
	}

	spinner, err := dao.TaskSpinner()
	core.CheckIfError(err)

	err = spinner.Start()
	core.CheckIfError(err)

	var data print.TableOutput

	// Table Style
	switch config.Theme.Table {
		case "ascii":
			print.ManiList.Box = print.StyleBoxASCII
		default:
			print.ManiList.Box = print.StyleBoxDefault
	}

	// Headers
	data.Headers = append(data.Headers, "Project")

	if task.Command != "" {
		data.Headers = append(data.Headers, task.Name)
	}

	for _, cmd := range task.Commands {
		if cmd.Ref != "" {
			refTask, err := config.GetTask(cmd.Ref)
			core.CheckIfError(err)

			if cmd.Name != "" {
				data.Headers = append(data.Headers, cmd.Name)
			} else {
				data.Headers = append(data.Headers, refTask.Name)
			}
		} else {
			data.Headers = append(data.Headers, cmd.Name)
		}
	}

	for _, project := range projects {
		data.Rows = append(data.Rows, table.Row { project.Name })
	}

	// Data
	var wg sync.WaitGroup

	for i, project := range projects {
		wg.Add(1)

		if (serialFlag) {
			spinner.Message(fmt.Sprintf(" %v", project.Name))
			worker(&data, *task, project, dryRunFlag, serialFlag, i, &wg)
		} else {
			spinner.Message(" Running")
			go worker(&data, *task, project, dryRunFlag, serialFlag, i, &wg)
		}
	}

	wg.Wait()

	err = spinner.Stop()
	core.CheckIfError(err)

	print.PrintRun(outputFlag, data)
}

func worker(
	data *print.TableOutput,
	task dao.Task,
	project dao.Project,
	dryRunFlag bool,
	serialFlag bool,
	i int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	if task.Command != "" {
		output, err := task.RunCmd(config, task.Shell, project, dryRunFlag)
		if err != nil {
			data.Rows[i] = append(data.Rows[i], err)
		} else {
			data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
		}
	}

	for _, cmd := range task.Commands {
		output, err := cmd.RunCmd(config, cmd.Shell, project, dryRunFlag)
		if err != nil {
			data.Rows[i] = append(data.Rows[i], output)
			return
		} else {
			data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
		}
	}
}
