package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func editTask(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Aliases: []string{"tasks", "tsks", "tsk"},
		Use:     "task [task]",
		Short:   "Edit mani task",
		Long:    `Edit mani task`,

		Example: `  # Edit tasks
  mani edit task

  # Edit task <task>
  mani edit task <task>`,
		Run: func(cmd *cobra.Command, args []string) {
			err := *configErr
			switch e := err.(type) {
			case *core.ConfigNotFound:
				core.CheckIfError(e)
			default:
				runEditTask(args, *config)
			}
		},
		Args: cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil || len(args) == 1 {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			values := config.GetTaskNames()
			return values, cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	return &cmd
}

func runEditTask(args []string, config dao.Config) {
	if len(args) > 0 {
		err := config.EditTask(args[0])
		core.CheckIfError(err)
	} else {
		err := config.EditTask("")
		core.CheckIfError(err)
	}
}
