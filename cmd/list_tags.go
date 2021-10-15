package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func listTagsCmd(config *dao.Config, configErr *error, listFlags *core.ListFlags) *cobra.Command {
	var tagFlags core.TagFlags

	cmd := cobra.Command{
		Aliases: []string{"tag", "tags"},
		Use:     "tags [flags]",
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
	}

	cmd.Flags().StringSliceVar(&tagFlags.Headers, "headers", []string{"name", "projects", "directories"}, "Specify headers, defaults to name, description")
	err := cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := []string{"tag", "projects", "directories"}
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
	allTags := config.GetTags()
	if len(args) > 0 {
		args = core.Intersection(args, allTags)
		m := config.GetTagAssocations(args)
		dao.PrintTags(config, m, *listFlags, *tagFlags)
	} else {
		m := config.GetTagAssocations(allTags)
		dao.PrintTags(config, m, *listFlags, *tagFlags)
	}
}
