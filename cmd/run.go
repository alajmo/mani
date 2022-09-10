package cmd

import (
	"fmt"
	"strings"

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/exec"
)

func runCmd(config *dao.Config, configErr *error) *cobra.Command {
	var runFlags core.RunFlags
	var setRunFlags core.SetRunFlags

	cmd := cobra.Command{
		Use:   "run <task>",
		Short: "Run tasks",
		Long: `Run tasks.

The tasks are specified in a mani.yaml file along with the projects you can target.`,

		Example: `  # Run task <task> for all projects
  mani run <task> --all

  # Run task <task> for all projects <project>
  mani run <task> --projects <project>

  # Run task <task> for all projects that have tags <tag>
  mani run <task> --tags <tag>

  # Run task <task> for all projects matching paths <path>
  mani run <task> --paths <path>

  # Run task <task> and pass in env value from shell
  mani run <task> key=value`,

		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)

			// This is necessary since cobra doesn't support pointers for bools
			// (that would allow us to use nil as default value)
			setRunFlags.Parallel = cmd.Flags().Changed("parallel")
			setRunFlags.OmitEmpty = cmd.Flags().Changed("omit-empty")

			run(args, config, &runFlags, &setRunFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			return config.GetTaskNameAndDesc(), cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVar(&runFlags.Describe, "describe", false, "print task information")
	cmd.Flags().BoolVar(&runFlags.DryRun, "dry-run", false, "prints the task to see what will be executed")
	cmd.Flags().BoolVarP(&runFlags.Silent, "silent", "s", false, "do not show progress when running tasks")
	cmd.Flags().BoolVar(&runFlags.OmitEmpty, "omit-empty", false, "omit empty results")
	cmd.Flags().BoolVar(&runFlags.Parallel, "parallel", false, "run tasks in parallel for each project")
	cmd.Flags().BoolVarP(&runFlags.Edit, "edit", "e", false, "edit task")

	cmd.Flags().StringVarP(&runFlags.Output, "output", "o", "", "set output [text|table|html|markdown]")
	err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		valid := []string{"text", "table", "html", "markdown"}
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().BoolVarP(&runFlags.Cwd, "cwd", "k", false, "current working directory")

	cmd.Flags().BoolVarP(&runFlags.All, "all", "a", false, "target all projects")

	cmd.Flags().StringSliceVarP(&runFlags.Projects, "projects", "p", []string{}, "target projects by names")
	err = cmd.RegisterFlagCompletionFunc("projects", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		projects := config.GetProjectNames()
		return projects, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&runFlags.Paths, "paths", "d", []string{}, "target projects by paths")
	err = cmd.RegisterFlagCompletionFunc("paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetProjectPaths()

		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&runFlags.Tags, "tags", "t", []string{}, "target projects by tags")
	err = cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		tags := config.GetTags()
		return tags, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.PersistentFlags().StringVar(&runFlags.Theme, "theme", "", "set theme")
	err = cmd.RegisterFlagCompletionFunc("theme", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		names := config.GetThemeNames()

		return names, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func run(
	args []string,
	config *dao.Config,
	runFlags *core.RunFlags,
	setRunFlags *core.SetRunFlags,
) {
	var taskNames []string
	var userArgs []string
	// Separate user arguments from task names
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			userArgs = append(userArgs, arg)
		} else {
			taskNames = append(taskNames, arg)
		}
	}

	if runFlags.Edit {
		if len(args) > 0 {
			_ = config.EditTask(taskNames[0])
			return
		} else {
			_ = config.EditTask("")
			return
		}
	}

	for _, taskName := range taskNames {
		task, err := config.GetTask(taskName)
		core.CheckIfError(err)

		projects, err := config.GetTaskProjects(task, runFlags)
		core.CheckIfError(err)

		var tasks []dao.Task
		for range projects {
			t := dao.Task{}
			err := copier.Copy(&t, &task)
			core.CheckIfError(err)

			tasks = append(tasks, t)
		}

		if len(projects) == 0 {
			fmt.Println("No targets")
		} else {
			target := exec.Exec{Projects: projects, Tasks: tasks, Config: *config}

			err := target.Run(userArgs, runFlags, setRunFlags)
			core.CheckIfError(err)
		}
	}
}
