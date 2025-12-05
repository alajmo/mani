package cmd

import (
	"strings"

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

		Example: `  # Execute task for all projects
  mani run <task> --all

  # Execute a task in parallel with a maximum of 8 concurrent processes
  mani run <task> --projects <project> --parallel --forks 8

  # Execute task for a specific projects
  mani run <task> --projects <project>

  # Execute a task for projects with specific tags
  mani run <task> --tags <tag>

  # Execute a task for projects matching specific paths
  mani run <task> --paths <path>

  # Execute a task for all projects matching a tag expression
  mani run <task> --tags-expr 'active || git' <tag>

  # Execute a task with environment variables from shell
  mani run <task> key=value`,

		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)

			// This is necessary since cobra doesn't support pointers for bools
			// (that would allow us to use nil as default value)
			setRunFlags.TTY = cmd.Flags().Changed("tty")
			setRunFlags.Cwd = cmd.Flags().Changed("cwd")
			setRunFlags.All = cmd.Flags().Changed("all")

			setRunFlags.Parallel = cmd.Flags().Changed("parallel")
			setRunFlags.OmitEmptyRows = cmd.Flags().Changed("omit-empty-rows")
			setRunFlags.OmitEmptyColumns = cmd.Flags().Changed("omit-empty-columns")
			setRunFlags.IgnoreErrors = cmd.Flags().Changed("ignore-errors")
			setRunFlags.IgnoreNonExisting = cmd.Flags().Changed("ignore-non-existing")
			setRunFlags.Forks = cmd.Flags().Changed("forks")

			if setRunFlags.Forks {
				forks, err := cmd.Flags().GetUint32("forks")
				core.CheckIfError(err)
				if forks == 0 {
					core.Exit(&core.ZeroNotAllowed{Name: "forks"})
				}
				runFlags.Forks = forks
			}

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

	cmd.Flags().BoolVar(&runFlags.TTY, "tty", false, "replace current process")
	cmd.Flags().BoolVar(&runFlags.Describe, "describe", false, "display task information")
	cmd.Flags().BoolVar(&runFlags.DryRun, "dry-run", false, "display the task without execution")
	cmd.Flags().BoolVarP(&runFlags.Silent, "silent", "s", false, "hide progress output during task execution")
	cmd.Flags().BoolVar(&runFlags.IgnoreNonExisting, "ignore-non-existing", false, "skip non-existing projects")
	cmd.Flags().BoolVar(&runFlags.IgnoreErrors, "ignore-errors", false, "continue execution despite errors")
	cmd.Flags().BoolVar(&runFlags.OmitEmptyRows, "omit-empty-rows", false, "hide empty rows in table output")
	cmd.Flags().BoolVar(&runFlags.OmitEmptyColumns, "omit-empty-columns", false, "hide empty columns in table output")
	cmd.Flags().BoolVar(&runFlags.Parallel, "parallel", false, "execute tasks in parallel across projects")
	cmd.Flags().BoolVarP(&runFlags.Edit, "edit", "e", false, "edit task")
	cmd.Flags().Uint32P("forks", "f", 4, "maximum number of concurrent processes")

	cmd.Flags().StringVarP(&runFlags.Output, "output", "o", "", "set output format [stream|table|markdown|html|json|yaml]")
	err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		valid := []string{"stream", "table", "html", "markdown", "json", "yaml"}
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringVarP(&runFlags.Spec, "spec", "J", "", "set spec")
	err = cmd.RegisterFlagCompletionFunc("spec", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		values := config.GetSpecNames()
		return values, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().BoolVarP(&runFlags.Cwd, "cwd", "k", false, "select current working directory")

	cmd.Flags().BoolVarP(&runFlags.All, "all", "a", false, "select all projects")

	cmd.Flags().StringSliceVarP(&runFlags.Projects, "projects", "p", []string{}, "select projects by name")
	err = cmd.RegisterFlagCompletionFunc("projects", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		projects := config.GetProjectNames()
		return projects, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&runFlags.Paths, "paths", "d", []string{}, "select projects by path")
	err = cmd.RegisterFlagCompletionFunc("paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		options := config.GetProjectPaths()

		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&runFlags.Tags, "tags", "t", []string{}, "select projects by tag")
	err = cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		tags := config.GetTags()
		return tags, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringVarP(&runFlags.TagsExpr, "tags-expr", "E", "", "select projects by tags expression")
	core.CheckIfError(err)

	cmd.Flags().StringVarP(&runFlags.Target, "target", "T", "", "select projects by target name")
	err = cmd.RegisterFlagCompletionFunc("target", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		values := config.GetTargetNames()
		return values, cobra.ShellCompDirectiveDefault
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

	var tasks []dao.Task
	var projects []dao.Project
	var err error
	if len(taskNames) == 1 {
		tasks, projects, err = dao.ParseSingleTask(taskNames[0], runFlags, setRunFlags, config)
	} else {
		tasks, projects, err = dao.ParseManyTasks(taskNames, runFlags, setRunFlags, config)
	}
	core.CheckIfError(err)

	target := exec.Exec{Projects: projects, Tasks: tasks, Config: *config}
	err = target.Run(userArgs, runFlags, setRunFlags)
	core.CheckIfError(err)
}
