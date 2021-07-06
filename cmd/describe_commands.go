package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core/print"
	"github.com/alajmo/mani/core/dao"
)

func describeCommandsCmd(config *dao.Config, configErr error) *cobra.Command {
	cmd := cobra.Command{
		Aliases: []string { "cmd", "cmds", "command" },
		Use:   "commands [commands] [flags]",
		Short: "Describe commands",
		Long:  "Describe commands.",
		Example: `  # Describe commands
  mani describe commands`,
		Run: func(cmd *cobra.Command, args []string) {
			describe(config, args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			commandNames := config.GetCommandNames()
			return commandNames, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return &cmd
}

func describe(config *dao.Config, args []string) {
	filteredCommands := config.GetCommandsByNames(args)

	print.PrintCommandBlocks(filteredCommands)
}
