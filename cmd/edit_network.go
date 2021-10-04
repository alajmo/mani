package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func editNetwork(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Use:   "network",
		Short: "Edit mani network",
		Long:  `Edit mani network`,

		Example: `  # Edit a network called server
  mani edit network server

  # Edit network in specific mani config
  mani edit --config path/to/mani/config`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			runEditNetwork(args, *config)
		},
		Args: cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil || len(args) == 1 {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			values := config.GetNetworkNames()
			return values, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return &cmd
}

func runEditNetwork(args []string, config dao.Config) {
	if len(args) > 0 {
		config.EditNetworks(args[0])
	} else {
		config.EditNetworks("")
	}
}
