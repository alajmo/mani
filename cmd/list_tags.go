package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func listTagsCmd(config *dao.Config, configErr *error, listFlags *core.ListFlags) *cobra.Command {
	var tagFlags core.TagFlags

	cmd := cobra.Command{
		Aliases: []string{"tag", "tags"},
		Use:     "tags [tags] [flags]",
		Short:   "List tags",
		Long:    "List tags.",
		Example: `  # List tags
  mani list tags`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			listTags(config, args, listFlags, &tagFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			tags := config.GetTags()
			return tags, cobra.ShellCompDirectiveNoFileComp
		},
      DisableAutoGenTag: true,
	}

    cmd.Flags().StringSliceVar(&tagFlags.Headers, "headers", []string{"tag", "project"}, "set headers. Available headers: tag, project")
	err := cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := []string{"tag", "project"}
		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listTags(
	config *dao.Config,
	args []string,
	listFlags *core.ListFlags,
	tagFlags *core.TagFlags,
) {
	theme, err := config.GetTheme(listFlags.Theme)
	core.CheckIfError(err)

	options := print.PrintTableOptions{
		Output:               listFlags.Output,
		Theme:                *theme,
		Tree:                 listFlags.Tree,
		OmitEmpty:            false,
		SuppressEmptyColumns: true,
	}

	allTags := config.GetTags()
	if len(args) > 0 {
		args = core.Intersection(args, allTags)
		m, err := config.GetTagAssocations(args)
		core.CheckIfError(err)
		print.PrintTable(m, options, tagFlags.Headers, []string{})
	} else {
		m, err := config.GetTagAssocations(allTags)
		core.CheckIfError(err)
		print.PrintTable(m, options, tagFlags.Headers, []string{})
	}
}
