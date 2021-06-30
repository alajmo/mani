package cmd

import (
	"github.com/alajmo/mani/core"
	"github.com/spf13/cobra"
)

func describeCommandsCmd(configFile *string) *cobra.Command {
	cmd := cobra.Command{
		Use:   "commands [commands] [flags]",
		Short: "Describe commands",
		Long:  "Describe commands.",
		Example: `  # Describe commands
  mani describe commands`,
		Run: func(cmd *cobra.Command, args []string) {
			describe(configFile, args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			_, config, err := core.ReadConfig(*configFile)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			commandNames := core.GetCommandNames(config.Commands)
			return commandNames, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return &cmd
}

func describe(configFile *string, args []string) {
	_, config, err := core.ReadConfig(*configFile)
	core.CheckIfError(err)

	filteredCommands := core.FilterCommandOnName(config.Commands, args)
	core.PrintCommands(filteredCommands, "block", false)
}
