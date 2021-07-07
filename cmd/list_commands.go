package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/print"
	"github.com/alajmo/mani/core/dao"
)

func listCommandsCmd(config *dao.Config, configErr *error, listFlags *print.ListFlags) *cobra.Command {
	var commandFlags print.ListCommandFlags

	cmd := cobra.Command{
		Aliases: []string { "cmd", "cmds", "command" },
		Use:   "commands [flags]",
		Short: "List commands",
		Long:  "List commands.",
		Example: `  # List commands
  mani list commands`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			listCommands(config, args, listFlags, &commandFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			commands := config.GetCommandNames()
			return commands, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringSliceVar(&commandFlags.Headers, "headers", []string{ "name", "description" }, "Specify headers, defaults to name, description")
	err := cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := []string { "name", "description" }
		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listCommands(
	config *dao.Config,
	args []string,
	listFlags *print.ListFlags,
	commandFlags *print.ListCommandFlags,
) {
	filteredCommands := config.GetCommandsByNames(args)
	print.PrintCommands(filteredCommands, *listFlags, *commandFlags)
}
