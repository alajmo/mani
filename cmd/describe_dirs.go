package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func describeDirsCmd(config *dao.Config, configErr *error) *cobra.Command {
	var dirFlags core.DirFlags

	cmd := cobra.Command{
		Aliases: []string{"dir", "drs", "d"},
		Use:     "dirs [dirs] [flags]",
		Short:   "Describe dirs",
		Long:    "Describe dirs.",
		Example: `  # Describe dirs
  mani describe dirs

  # Describe dirs that have tag frontend
  mani describe dirs --tags frontend`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			describeDirs(config, args, dirFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			names := config.GetDirNames()
			return names, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringSliceVarP(&dirFlags.Tags, "tags", "t", []string{}, "filter dirs by their tag")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetTags()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVar(&dirFlags.DirPaths, "dir-paths", []string{}, "filter dirs by their path")
	err = cmd.RegisterFlagCompletionFunc("dir-paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetDirPaths()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().BoolVarP(&dirFlags.Edit, "edit", "e", false, "Edit dir")

	return &cmd
}

func describeDirs(
	config *dao.Config,
	args []string,
	dirFlags core.DirFlags,
) {
	if dirFlags.Edit {
		if len(args) > 0 {
			config.EditDir(args[0])
		} else {
			config.EditDir("")
		}
	} else {
		allDirs := false
		if (len(args) == 0 &&
			len(dirFlags.DirPaths) == 0 &&
			len(dirFlags.Tags) == 0) {
			allDirs = true
		}

		dirs := config.FilterDirs(false, allDirs, dirFlags.DirPaths, args, dirFlags.Tags)
		print.PrintDirBlocks(dirs)
	}
}
