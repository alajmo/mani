package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func describeTasksCmd(config *dao.Config, configErr *error) *cobra.Command {
	var taskFlags core.TaskFlags

	cmd := cobra.Command{
		Aliases: []string{"task", "tsk"},
		Use:     "tasks [tasks]",
		Short:   "Describe tasks",
		Long:    "Describe tasks.",
		Example: `  # Describe all tasks
  mani describe tasks

  # Describe task <task>
  mani describe task <task>`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			describe(config, args, taskFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			values := config.GetTaskNames()
			return values, cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVarP(&taskFlags.Edit, "edit", "e", false, "edit task")

	return &cmd
}

func describe(config *dao.Config, args []string, taskFlags core.TaskFlags) {
	if taskFlags.Edit {
		if len(args) > 0 {
			err := config.EditTask(args[0])
			core.CheckIfError(err)
		} else {
			err := config.EditTask("")
			core.CheckIfError(err)
		}
	} else {
		tasks, err := config.GetTasksByNames(args)
		core.CheckIfError(err)

		if len(tasks) == 0 {
			fmt.Println("No tasks")
		} else {
			for i := range tasks {
				envs, err := dao.ParseTaskEnv(tasks[i].Env, []string{}, []string{}, []string{})
				core.CheckIfError(err)

				tasks[i].EnvList = envs

				for j := range tasks[i].Commands {
					envs, err = dao.ParseTaskEnv(tasks[i].Commands[j].Env, []string{}, []string{}, []string{})
					core.CheckIfError(err)

					tasks[i].Commands[j].EnvList = envs
				}
			}

			print.PrintTaskBlock(tasks)
		}
	}
}
