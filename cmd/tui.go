package cmd

import (
	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/tui"
	"github.com/spf13/cobra"
)

func tuiCmd(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Use:     "tui",
		Aliases: []string{"gui"},
		Short:   "TUI",
		Long: `Clone repositories and add them to gitignore.
In-case you need to enter credentials before cloning, run the command without the parallel flag.`,
		Example: `  # Clone repositories one at a time
  mani sync

  # Clone repositories in parallell
  mani sync --parallel

  # Show cloned projects
  mani sync --status`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			tui.RunTui(config, args)
		},
		DisableAutoGenTag: true,
	}

	return &cmd
}
