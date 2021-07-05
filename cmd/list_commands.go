package cmd

import (
	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/print"
	"github.com/spf13/cobra"
)

func listCommandsCmd(configFile *string, listFlags *core.ListFlags) *cobra.Command {
	var commandFlags core.ListCommandFlags

	cmd := cobra.Command{
		Aliases: []string { "cmd", "cmds", "command" },
		Use:   "commands [flags]",
		Short: "List commands",
		Long:  "List commands.",
		Example: `  # List commands
  mani list commands`,
		Run: func(cmd *cobra.Command, args []string) {
			listCommands(configFile, args, listFlags, &commandFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			_, config, err := core.ReadConfig(*configFile)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			commands := core.GetCommandNames(config.Commands)
			return commands, cobra.ShellCompDirectiveNoFileComp
		},

	}

	cmd.Flags().StringSliceVar(&commandFlags.Headers, "headers", []string{ "name", "description" }, "Specify headers, defaults to name, description")
	err := cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, _, err := core.ReadConfig(*configFile)

		if err != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := []string { "name", "description" }

		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listCommands(
	configFile *string,
	args []string,
	listFlags *core.ListFlags,
	commandFlags *core.ListCommandFlags,
) {
	_, config, err := core.ReadConfig(*configFile)
	core.CheckIfError(err)

	filteredCommands := core.FilterCommandOnName(config.Commands, args)
	print.PrintCommands(filteredCommands, *listFlags, *commandFlags)
}
