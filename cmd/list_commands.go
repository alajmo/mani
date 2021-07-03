package cmd

import (
	"github.com/alajmo/mani/core"
	"github.com/spf13/cobra"
)

func listCommandsCmd(configFile *string) *cobra.Command {
	cmd := cobra.Command{
		Aliases: []string { "cmd", "cmds", "command" },
		Use:   "commands [flags]",
		Short: "List commands",
		Long:  "List commands.",
		Example: `  # List commands
  mani list commands`,
		Run: func(cmd *cobra.Command, args []string) {
			listCommands(configFile, args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			_, config, err := core.ReadConfig(*configFile)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			tags := core.GetCommandNames(config.Commands)
			return tags, cobra.ShellCompDirectiveNoFileComp
		},

	}

	return &cmd
}

func listCommands(configFile *string, args []string) {
	_, config, err := core.ReadConfig(*configFile)
	core.CheckIfError(err)

	filteredCommands := core.FilterCommandOnName(config.Commands, args)
	core.PrintCommands(filteredCommands, "list", false)
}
