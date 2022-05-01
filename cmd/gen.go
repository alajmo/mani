package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
)

func genCmd() *cobra.Command {
	dir := ""
	cmd := cobra.Command{
		Use:                   "gen",
		Short:                 "Generate man page",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			err := core.GenManPages(dir)
			core.CheckIfError(err)
		},

		DisableAutoGenTag: true,
	}

	cmd.Flags().StringVarP(&dir, "dir", "d", "./", "directory to save manpage to")
	err := cmd.RegisterFlagCompletionFunc("dir", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterDirs
	})
	core.CheckIfError(err)

	return &cmd
}
