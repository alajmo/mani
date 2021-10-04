package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func describeNetworksCmd(config *dao.Config, configErr *error) *cobra.Command {
	var networkFlags print.ListNetworkFlags

	cmd := cobra.Command{
		Aliases: []string{"network", "net", "n"},
		Use:     "networks [networks] [flags]",
		Short:   "Describe networks",
		Long:    "Describe networks.",
		Example: `  # Describe networks
  mani describe networks

  # Describe networks that have tag frontend
  mani describe networks --tags frontend`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			describeNetworks(config, args, networkFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			projectNames := config.GetProjectNames()
			return projectNames, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringSliceVarP(&networkFlags.Tags, "tags", "t", []string{}, "filter networks by their tag")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetTags()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().BoolVarP(&networkFlags.Edit, "edit", "e", false, "Edit project")

	return &cmd
}

func describeNetworks(
	config *dao.Config,
	args []string,
	networkFlags print.ListNetworkFlags,
) {
	if networkFlags.Edit {
		if len(args) > 0 {
			config.EditNetworks(args[0])
		} else {
			config.EditNetworks("")
		}
	} else {
		networksName := config.GetNetworksByName(args)
		networksTag := config.GetNetworksByTag(networkFlags.Tags)

		filteredNetworks := dao.GetIntersectNetworks(networksName, networksTag)

		print.PrintNetworkBlocks(filteredNetworks)
	}
}
