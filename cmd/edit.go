package cmd

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func editCmd(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Aliases: []string{"e", "ed"},
		Use:     "edit",
		Short:   "Edit mani config",
		Long:    `Edit mani config`,

		Example: `  # Edit current context
  mani edit

  # Edit specific mani config
  edit edit --config path/to/mani/config`,
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Should handle all cases correctly, now it just handles cases
			// when it can't find the config file, but what about permissions errors,
			// Golang errors from GetWD, etc.
			// Perhaps solution is to panic on those errors since something
			// must have gone horribly wrong.
			err := *configErr
			fmt.Println(err)
			switch e := err.(type) {
			case *core.ConfigNotFound:
				core.CheckIfError(e)
			default:
				runEdit(args, *config)
			}
		},
	}

	cmd.AddCommand(
		editDir(config, configErr),
		editTask(config, configErr),
		editProject(config, configErr),
	)

	return &cmd
}

func runEdit(args []string, config dao.Config) {
	// config.EditConfig()
}
