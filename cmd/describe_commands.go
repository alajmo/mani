package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/print"
	"github.com/alajmo/mani/core/dao"
)

func describeCommandsCmd(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Aliases: []string { "cmd", "cmds", "command" },
		Use:   "commands [commands] [flags]",
		Short: "Describe commands",
		Long:  "Describe commands.",
		Example: `  # Describe commands
  mani describe commands`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			describe(config, args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			commandNames := config.GetCommandNames()
			return commandNames, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return &cmd
}

func describe(config *dao.Config, args []string) {
	commands := config.GetCommandsByNames(args)

	for i := range commands {
		var userEnv []string
		if len(args) > 1 {
			userEnv = args[1:]
		}

		commands[i].SetEnvList(userEnv, config.GetEnv())
	}

	print.PrintCommandBlocks(commands)
}
