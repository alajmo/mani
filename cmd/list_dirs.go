package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func listDirsCmd(config *dao.Config, configErr *error, listFlags *print.ListFlags) *cobra.Command {
	var dirFlags print.ListDirFlags

	cmd := cobra.Command{
		Aliases: []string{"dir", "dr", "d"},
		Use:     "dirs [flags]",
		Short:   "List dirs",
		Long:    "List dirs",
		Example: `  # List dirs
  mani list dirs`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			listDirs(config, args, listFlags, &dirFlags)
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

	cmd.Flags().StringSliceVar(&dirFlags.Headers, "headers", []string{"name", "tags", "description"}, "Specify headers, defaults to name, tags, description")
	err = cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := []string{"name", "path", "relpath", "description", "url", "tags"}
		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listDirs(
	config *dao.Config,
	args []string,
	listFlags *print.ListFlags,
	dirFlags *print.ListDirFlags,
) {
	// Table Style
	switch config.Theme.Table {
	case "ascii":
		print.ManiList.Box = print.StyleBoxASCII
	default:
		print.ManiList.Box = print.StyleBoxDefault
	}

	dirName := config.GetDirsByName(args)
	dirPaths := config.GetDirsByPath(dirFlags.DirPaths)
	dirTags := config.GetDirsByTags(dirFlags.Tags)

	filteredDirs := dao.GetIntersectDirs(dirName, dirTags)
	filteredDirs = dao.GetIntersectDirs(filteredDirs, dirPaths)

	print.PrintDirs(filteredDirs, *listFlags, *dirFlags)
}
