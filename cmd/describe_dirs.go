package cmd

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func describeDirsCmd(config *dao.Config, configErr *error) *cobra.Command {
	var tags []string
	var dirPaths []string
	var edit bool
	var dirs []string

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
			describeDirs(config, args, tags, dirPaths, dirs, edit)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			names := config.GetDirNames()
			return names, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "filter dirs by their tag")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetTags()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVar(&dirPaths, "dir-paths", []string{}, "filter dirs by their path")
	err = cmd.RegisterFlagCompletionFunc("dir-paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetDirPaths()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().BoolVarP(&edit, "edit", "e", false, "Edit dir")

	return &cmd
}

func describeDirs(
	config *dao.Config,
	args []string,
	tags []string,
	dirPaths []string,
	dirs []string,
	edit bool,
) {
	if edit {
		if len(args) > 0 {
			config.EditDir(args[0])
		} else {
			config.EditDir("")
		}
	} else {
		dirNames := config.GetDirsByName(args)
		fmt.Println("=========================")
		fmt.Println(dirNames)
		fmt.Println("=========================")

		dirPaths := config.GetDirsByPath(dirPaths)
		// fmt.Println(dirPaths)
		dirTags := config.GetDirsByTags(tags)
		// fmt.Println(dirTags)
		// fmt.Println(dirPaths)
		// fmt.Println("=========================")

		filtered := dao.GetIntersectDirs(dirNames, dirTags)
		filtered = dao.GetIntersectDirs(filtered, dirPaths)

		print.PrintDirBlocks(filtered)
	}
}
