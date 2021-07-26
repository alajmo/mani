package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/print"
	"github.com/alajmo/mani/core/dao"
)

func describeTasksCmd(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Aliases: []string { "task", "tasks" },
		Use:   "tasks [tasks] [flags]",
		Short: "Describe tasks",
		Long:  "Describe tasks.",
		Example: `  # Describe tasks
  mani describe tasks`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			describe(config, args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			values := config.GetTaskNames()
			return values, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return &cmd
}

func describe(config *dao.Config, args []string) {
	tasks := config.GetTasksByNames(args)

	for i := range tasks {
		var userEnv []string
		if len(args) > 1 {
			userEnv = args[1:]
		}

		tasks[i].SetEnvList(userEnv, config.GetEnv())
		for j := range tasks[i].Commands {
			tasks[i].Commands[j].SetEnvList(userEnv, config.GetEnv())
		}
	}

	print.PrintTaskBlock(tasks)
}
