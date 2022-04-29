package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/exec"
)

func syncCmd(config *dao.Config, configErr *error) *cobra.Command {
	var syncFlags core.SyncFlags

	cmd := cobra.Command{
		Use:     "sync",
		Aliases: []string{"clone"},
		Short:   "Clone repositories and add them to gitignore",
		Long: `Clone repositories and add them to gitignore.
In-case you need to enter credentials before cloning, run the command without the parallel flag.`,
		Example: `  # Clone repositories one at a time
  mani sync

  # Clone repositories in parallell
  mani sync --parallel`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)

			// This is necessary since cobra doesn't support pointers for bools
			// (that would allow us to use nil as default value)
			syncFlags.Parallel = cmd.Flags().Changed("parallel")
			runSync(config, syncFlags)
		},
      DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVarP(&syncFlags.Parallel, "parallel", "p", false, "clone projects in parallel")
	cmd.Flags().BoolVarP(&syncFlags.Status, "status", "s", false, "print sync status only")

	return &cmd
}

func runSync(config *dao.Config, syncFlags core.SyncFlags) {
	if !syncFlags.Status {
		err := exec.UpdateGitignoreIfExists(config)
		core.CheckIfError(err)

		err = exec.CloneRepos(config, syncFlags.Parallel)
		core.CheckIfError(err)
	}

	err := exec.PrintProjectStatus(config)
	core.CheckIfError(err)
}
