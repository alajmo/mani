package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func editProject(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Use:   "project",
		Short: "Edit mani project",
		Long:  `Edit mani project`,

		Example: `  # Edit a project called mani
  mani edit project mani

  # Edit project in specific mani config
  mani edit --config path/to/mani/config`,
		Run: func(cmd *cobra.Command, args []string) {
			err := *configErr
			switch e := err.(type) {
			case *core.ConfigNotFound:
				core.CheckIfError(e)
			default:
				runEditProject(args, *config)
			}
		},
		Args: cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil || len(args) == 1 {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			values := config.GetProjectNames()
			return values, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return &cmd
}

func runEditProject(args []string, config dao.Config) {
	if len(args) > 0 {
		config.EditProject(args[0])
	} else {
		config.EditProject("")
	}
}
