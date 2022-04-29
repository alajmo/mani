package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func editCmd(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Aliases: []string{"e", "ed"},
		Use:     "edit [flags]",
		Short:   "Edit mani config",
		Long:    `Edit mani config`,

		Example: `  # Edit current context
  mani edit

  # Edit specific mani config
  edit edit --config path/to/mani/config`,
		Run: func(cmd *cobra.Command, args []string) {
			err := *configErr
			switch e := err.(type) {
			case *core.ConfigNotFound:
				core.CheckIfError(e)
			default:
				runEdit(args, *config)
			}
		},
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		editTask(config, configErr),
		editProject(config, configErr),
	)

	return &cmd
}

func runEdit(args []string, config dao.Config) {
	err := config.EditConfig()
	core.CheckIfError(err)
}
