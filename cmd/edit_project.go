package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func editProject(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Aliases: []string{"projects", "proj", "pr"},
		Use:     "project [project]",
		Short:   "Edit mani project",
		Long:    `Edit mani project`,

		Example: `  # Edit projects
  mani edit project

  # Edit project <project>
  mani edit project <project>`,
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
		DisableAutoGenTag: true,
	}

	return &cmd
}

func runEditProject(args []string, config dao.Config) {
	if len(args) > 0 {
		err := config.EditProject(args[0])
		core.CheckIfError(err)
	} else {
		err := config.EditProject("")
		core.CheckIfError(err)
	}
}
