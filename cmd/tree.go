package cmd

import (
	"github.com/spf13/cobra"
	"github.com/jedib0t/go-pretty/v6/list"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/print"
	"github.com/alajmo/mani/core/dao"
)

func treeCmd(config *dao.Config, configErr *error) *cobra.Command {
	var output string
	var dirs []string
	var tags []string

	cmd := cobra.Command {
		Aliases: []string { "t", "tree" },
		Use:   "tree",
		Short: "tree",
		Long:  "list contents of directories in a tree-like format.",
		Example: `  # example
  mani tree`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			runTree(config, &output, &dirs, &tags)
		},
	}

	cmd.PersistentFlags().StringVarP(&output, "output", "o", "tree", "Output tree|markdown|html")
	err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		valid := []string { "tree", "markdown", "html" }
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&dirs, "dirs", "d", []string{}, "filter projects by their directory")
	err = cmd.RegisterFlagCompletionFunc("dirs", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetDirs()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "filter projects by their tag")
	err = cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetTags()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}


func runTree(
	config *dao.Config,
	output *string,
	dirs *[]string,
	tags *[]string,
) {
	switch config.Theme.Tree {
	case "square":
		print.TreeStyle = list.StyleBulletSquare
	case "circle":
		print.TreeStyle = list.StyleBulletCircle
	case "star":
		print.TreeStyle = list.StyleBulletStar
	case "line-bold":
		print.TreeStyle = list.StyleConnectedBold
	default:
		print.TreeStyle = list.StyleConnectedLight
	}

	tree := config.GetProjectsTree(*dirs, *tags)
	print.PrintTree(*output, tree)
}
