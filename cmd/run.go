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


type RunFlags struct {
	Edit bool
	Serial bool
	DryRun bool
	Describe bool
	Cwd bool
	SshConfig string

	AllProjects bool
	Projects []string
	ProjectPaths []string

	AllDirs bool
	Dirs []string
	DirPaths []string

	AllNetworks bool
	Networks []string
	NetworkHosts []string

	Tags []string
	Output string
}

func runCmd(config *dao.Config, configErr *error) *cobra.Command {
	var runFlags RunFlags

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
			run(args, config, &runFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			return config.GetTasks(), cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().BoolVar(&runFlags.Describe, "describe", true, "Print task information")
	cmd.Flags().BoolVar(&runFlags.DryRun, "dry-run", false, "don't execute any task, just print the output of the task to see what will be executed")
	cmd.Flags().BoolVarP(&runFlags.Edit, "edit", "e", false, "Edit task")
	cmd.Flags().BoolVarP(&runFlags.Serial, "serial", "s", false, "Run tasks in serial")
	cmd.Flags().StringVarP(&runFlags.Output, "output", "o", "", "Output list|table|markdown|html")

	cmd.Flags().BoolVarP(&runFlags.Cwd, "cwd", "k", false, "current working directory")

	cmd.Flags().BoolVarP(&runFlags.AllProjects, "all-projects", "a", false, "target all projects")
	cmd.Flags().StringSliceVarP(&runFlags.Projects, "projects", "p", []string{}, "target projects by their name")
	cmd.Flags().StringSliceVar(&runFlags.ProjectPaths, "project-paths", []string{}, "target projects by their path")

	cmd.Flags().BoolVar(&runFlags.AllDirs, "all-dirs", false, "target all dirs")
	cmd.Flags().StringSliceVarP(&runFlags.Dirs, "dirs", "d", []string{}, "target directories by their name")
	cmd.Flags().StringSliceVar(&runFlags.DirPaths, "dir-paths", []string{}, "target directories by their path")

	cmd.Flags().BoolVar(&runFlags.AllNetworks, "all-networks", false, "target all networks")
	cmd.Flags().StringSliceVarP(&runFlags.Networks, "networks", "n", []string{}, "target networks by their name")
	cmd.Flags().StringSliceVar(&runFlags.NetworkHosts, "network-hosts", []string{}, "target networks by their host")

	cmd.Flags().StringSliceVarP(&runFlags.Tags, "tags", "t", []string{}, "target entities by their tag")

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

	err = cmd.RegisterFlagCompletionFunc("dirs", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		dirs := config.GetDirNames()
		return dirs, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	err = cmd.RegisterFlagCompletionFunc("dir-paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetDirPaths()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	err = cmd.RegisterFlagCompletionFunc("networks", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		dirs := config.GetNetworkNames()
		return dirs, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	err = cmd.RegisterFlagCompletionFunc("network-hosts", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetAllHosts()
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
	runFlags *RunFlags,
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

	if (runFlags.Edit) {
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

		if task.Output != "" && runFlags.Output == "" {
			runFlags.Output = task.Output
		}

		if len(runFlags.Projects) == 0 {
			runFlags.Projects = task.Projects
		}

		if len(runFlags.ProjectPaths) == 0 {
			runFlags.ProjectPaths = task.ProjectPaths
		}

		if len(runFlags.Dirs) == 0 {
			runFlags.Dirs = task.Dirs
		}

		if len(runFlags.DirPaths) == 0 {
			runFlags.DirPaths = task.DirPaths
		}

		if len(runFlags.Tags) == 0 {
			runFlags.Tags = task.Tags
		}

		projects := config.FilterProjects(runFlags.Cwd, runFlags.AllProjects, runFlags.ProjectPaths, runFlags.Projects, runFlags.Tags)

		dirs := config.FilterDirs(runFlags.Cwd, runFlags.AllDirs, runFlags.DirPaths, runFlags.Dirs, runFlags.Tags)

		networks := config.FilterNetworks(runFlags.AllNetworks, runFlags.Networks, runFlags.NetworkHosts, runFlags.Tags)

		if len(projects) > 0 {
			var entities []dao.Entity
			for i := range projects {
				var entity dao.Entity
				entity.Name = projects[i].Name
				entity.Path = projects[i].Path

				entities = append(entities, entity)
			}

			runTask(task, "Project", entities, userArgs, config, runFlags)
		}

		if len(dirs) > 0 {
			var entities []dao.Entity
			for i := range dirs {
				var entity dao.Entity
				entity.Name = dirs[i].Name
				entity.Path = dirs[i].Path

				entities = append(entities, entity)
			}

			runTask(task, "Directory", entities, userArgs, config, runFlags)
		}

		if len(networks) > 0 {
			// TODO: Add ssh credentials
			// Load SSH config

			var entities []dao.Entity
			for i := range networks {
				var entity dao.Entity
				entity.Name = networks[i].Name

				entities = append(entities, entity)
			}

			runTask(task, "Networks", entities, userArgs, config, runFlags)
		}

		if len(projects) == 0 && len(dirs) == 0 && len(networks) == 0 {
			fmt.Println("No targets")
			continue
		}

		// Newline seperator between tasks
		if i < len(taskNames) {
			fmt.Println()
		}
	}
}

func runTask(
	task *dao.Task,
	entityType string,
	entities []dao.Entity,
	userArgs []string,
	config *dao.Config,
	runFlags *RunFlags,
) {
	task.SetEnvList(userArgs, []string{}, config.GetEnv())

	// Set env for sub-commands
	for i := range task.Commands {
		task.Commands[i].SetEnvList(userArgs, task.EnvList, config.GetEnv())
	}

	if runFlags.Describe {
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
	data.Headers = append(data.Headers, entityType)

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

	for _, entity := range entities {
		data.Rows = append(data.Rows, table.Row { entity.Name })
	}

	// Data
	var wg sync.WaitGroup

	for i, entity := range entities {
		wg.Add(1)

		if (runFlags.Serial) {
			spinner.Message(fmt.Sprintf(" %v", entity.Name ))
			worker(&data, *task, entity, runFlags.DryRun, runFlags.Serial, i, &wg)
		} else {
			spinner.Message(" Running")
			go worker(&data, *task, entity, runFlags.DryRun, runFlags.Serial, i, &wg)
		}
	}

	wg.Wait()

	err = spinner.Stop()
	core.CheckIfError(err)

	// print.PrintRun(runFlags.Output, data)
}

func worker(
	data *print.TableOutput,
	task dao.Task,
	entity dao.Entity,
	dryRunFlag bool,
	serialFlag bool,
	i int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	if task.Command != "" {
		output, err := task.RunCmd(config, task.Shell, entity, dryRunFlag)
		if err != nil {
			data.Rows[i] = append(data.Rows[i], err)
		} else {
			data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
		}
	}

	for _, cmd := range task.Commands {
		output, err := cmd.RunCmd(config, cmd.Shell, entity, dryRunFlag)
		if err != nil {
			data.Rows[i] = append(data.Rows[i], output)
			return
		} else {
			data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
		}
	}
}
