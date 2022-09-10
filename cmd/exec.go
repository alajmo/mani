package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/exec"
)

func execCmd(config *dao.Config, configErr *error) *cobra.Command {
	var runFlags core.RunFlags
	var setRunFlags core.SetRunFlags

	cmd := cobra.Command{
		Use:   "exec <command>",
		Short: "Execute arbitrary commands",
		Long: `Execute arbitrary commands.

Single quote your command if you don't want the
file globbing and environments variables expansion to take place
before the command gets executed in each directory.`,

		Example: `  # List files in all projects
  mani exec --all ls

  # List git files that have markdown suffix for all projects
  mani exec --all 'git ls-files | grep -e ".md"'`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)

			// This is necessary since cobra doesn't support pointers for bools
			// (that would allow us to use nil as default value)
			setRunFlags.Parallel = cmd.Flags().Changed("parallel")
			setRunFlags.OmitEmpty = cmd.Flags().Changed("omit-empty")

			execute(args, config, &runFlags, &setRunFlags)
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVar(&runFlags.DryRun, "dry-run", false, "prints the command to see what will be executed")
	cmd.Flags().BoolVarP(&runFlags.Silent, "silent", "s", false, "do not show progress when running tasks")
	cmd.Flags().BoolVar(&runFlags.OmitEmpty, "omit-empty", false, "omit empty results")
	cmd.Flags().BoolVar(&runFlags.Parallel, "parallel", false, "run tasks in parallel for each project")
	cmd.Flags().StringVarP(&runFlags.Output, "output", "o", "", "set output [text|table|markdown|html]")
	err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		valid := []string{"table", "markdown", "html"}
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

	cmd.PersistentFlags().StringVar(&runFlags.Theme, "theme", "default", "set theme")
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

func execute(
	args []string,
	config *dao.Config,
	runFlags *core.RunFlags,
	setRunFlags *core.SetRunFlags,
) {
	projects, err := config.FilterProjects(runFlags.Cwd, runFlags.All, runFlags.Projects, runFlags.Paths, runFlags.Tags)
	core.CheckIfError(err)

	if len(projects) == 0 {
		fmt.Println("No targets")
	} else {
		cmd := strings.Join(args[0:], " ")
		var tasks []dao.Task

		task := dao.Task{Cmd: cmd, Name: "output"}
		taskErrors := make([]dao.ResourceErrors[dao.Task], 1)
		task.ParseTask(*config, &taskErrors[0])

		var configErr = ""
		for _, taskError := range taskErrors {
			if len(taskError.Errors) > 0 {
				configErr = fmt.Sprintf("%s%s", configErr, dao.FormatErrors(taskError.Resource, taskError.Errors))
			}
		}
		if configErr != "" {
			core.CheckIfError(errors.New(configErr))
		}

		for range projects {
			t := dao.Task{}
			err := copier.Copy(&t, &task)
			core.CheckIfError(err)
			tasks = append(tasks, t)
		}

		target := exec.Exec{Projects: projects, Tasks: tasks, Config: *config}
		err := target.Run([]string{}, runFlags, setRunFlags)
		core.CheckIfError(err)
	}
}
