package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func editCmd(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Use:   "edit",
		Short: "Edit mani config",
		Long: `Edit mani config`,

		Example: `  # Edit current context
  mani edit

  # Edit specific mani config
  edit --config path/to/mani/config`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			runEdit(args, *config)
		},
	}

	return &cmd
}

func runEdit(args []string, config dao.Config) {
	config.EditFile()
}
