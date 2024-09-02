package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/exec"
)

func initCmd() *cobra.Command {
	var initFlags core.InitFlags

	cmd := cobra.Command{
		Use:   "init",
		Short: "Initialize a mani repository",
		Long: `Initialize a mani repository.

Creates a new mani repository by generating a mani.yaml configuration file 
and a .gitignore file in the current directory.`,

		Example: `  # Initialize with default settings
  mani init

  # Initialize without auto-discovering projects
  mani init --auto-discovery=false

  # Initialize without updating .gitignore
  mani init --sync-gitignore=false`,

		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			foundProjects, err := dao.InitMani(args, initFlags)
			core.CheckIfError(err)

			if initFlags.AutoDiscovery {
				exec.PrintProjectInit(foundProjects)
			}
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVar(&initFlags.AutoDiscovery, "auto-discovery", true, "automatically discover and add Git repositories to mani.yaml")
	cmd.Flags().BoolVarP(&initFlags.SyncGitignore, "sync-gitignore", "g", true, "synchronize .gitignore file")

	return &cmd
}
