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

Creates a mani repository - a directory with configuration file mani.yaml and a .gitignore file.`,
		Example: `  # Basic example
  mani init

  # Skip auto-discovery of projects
  mani init --auto-discovery=false`,

		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			configDir, foundProjects, err := dao.InitMani(args, initFlags)
			core.CheckIfError(err)

			if initFlags.AutoDiscovery {
				exec.PrintProjectInit(configDir, foundProjects)
			}
		},
	}

	cmd.Flags().BoolVar(&initFlags.AutoDiscovery, "auto-discovery", true, "walk current directory and find git repositories to add to mani.yaml")
	cmd.Flags().StringVar(&initFlags.Vcs, "vcs", "git", "Initialize directory using version control system. Acceptable values: git, none")
	err := cmd.RegisterFlagCompletionFunc("vcs", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		valid := []string{"git", "none"}
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}
