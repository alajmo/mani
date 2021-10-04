package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func editDir(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Use:   "dir",
		Short: "Edit mani dir",
		Long:  `Edit mani dir`,

		Example: `  # Edit a dir called mani
  mani edit dir mani

  # Edit dir in specific mani config
  mani edit --config path/to/mani/config`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			runEditDir(args, *config)
		},
		Args: cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil || len(args) == 1 {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			values := config.GetDirNames()
			return values, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return &cmd
}

func runEditDir(args []string, config dao.Config) {
	if len(args) > 0 {
		config.EditDir(args[0])
	} else {
		config.EditDir("")
	}
}
