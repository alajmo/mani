package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func syncCmd(config *dao.Config, configErr *error) *cobra.Command {
	var syncFlags core.SyncFlags

	cmd := cobra.Command{
		Use:   "sync",
		Short: "Clone repositories and add them to gitignore",
		Long: `Clone repositories and add them to gitignore.
In-case you need to enter credentials before cloning, run the command with the parallell flag.`,
		Example: `  # Clone repositories one at a time
  mani sync

  # Clone repositories one by one
  mani sync --parallell=false`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			runSync(config, syncFlags)
		},
	}

	cmd.Flags().BoolVarP(&syncFlags.Parallell, "parallell", "p", true, "Clone projects in parallell")

	return &cmd
}

func runSync(config *dao.Config, syncFlags core.SyncFlags) {
	config.SyncDirs(config.Dir, syncFlags.Parallell)
	config.SyncProjects(config.Dir, syncFlags.Parallell)
}
