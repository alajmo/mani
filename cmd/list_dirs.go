package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func listDirsCmd(config *dao.Config, configErr *error, listFlags *core.ListFlags) *cobra.Command {
	var dirFlags core.DirFlags

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

	cmd.Flags().StringSliceVarP(&dirFlags.Paths, "paths", "p", []string{}, "filter dirs by their path")
	err = cmd.RegisterFlagCompletionFunc("paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
	listFlags *core.ListFlags,
	dirFlags *core.DirFlags,
) {
	allDirs := false
	if len(args) == 0 &&
		len(dirFlags.Paths) == 0 &&
		len(dirFlags.Tags) == 0 {
		allDirs = true
	}

	dirs := config.FilterDirs(false, allDirs, dirFlags.Paths, args, dirFlags.Tags)
	dao.PrintDirs(config, dirs, *listFlags, *dirFlags)
}
