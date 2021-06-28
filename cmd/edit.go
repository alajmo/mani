package cmd

import (
	core "github.com/alajmo/mani/core"
	"github.com/spf13/cobra"
)

func editCmd(configFile *string) *cobra.Command {

	cmd := cobra.Command{
		Use:   "edit",
		Short: "Edit mani config",
		Long: `Edit mani config`,

		Example: `  # Edit current context
  mani edit

  # Edit specific mani config
  edit --config path/to/mani/config`,
		Run: func(cmd *cobra.Command, args []string) {
			runEdit(args, configFile)
		},
	}

	return &cmd
}

func runEdit(args []string, configFile *string) {
	configPath, _, err := core.ReadConfig(*configFile)
	core.CheckIfError(err)

	core.EditFile(configPath)
}
