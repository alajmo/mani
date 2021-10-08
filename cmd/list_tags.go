package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func listTagsCmd(config *dao.Config, configErr *error, listFlags *print.ListFlags) *cobra.Command {
	var tagFlags print.ListTagFlags
	var projects []string

	cmd := cobra.Command{
		Aliases: []string{"tag", "tags"},
		Use:     "tags [flags]",
		Short:   "List tags",
		Long:    "List tags.",
		Example: `  # List tags
  mani list tags`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			listTags(config, args, listFlags, &tagFlags, projects)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			tags := config.GetTags()
			return tags, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringSliceVarP(&projects, "projects", "p", []string{}, "filter tags by their project")
	err := cmd.RegisterFlagCompletionFunc("projects", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		projects := config.GetProjectNames()
		return projects, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVar(&tagFlags.Headers, "headers", []string{"name"}, "Specify headers, defaults to name, description")
	err = cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := []string{"name"}
		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listTags(
	config *dao.Config,
	args []string,
	listFlags *print.ListFlags,
	tagFlags *print.ListTagFlags,
	projects []string,
) {
	// Table Style
	// switch config.Theme.Table {
	// case "ascii":
	// 	core.ManiList.Box = core.StyleBoxASCII
	// default:
	// 	core.ManiList.Box = core.StyleBoxDefault
	// }

	allTags := config.GetTags()
	if len(args) == 0 && len(projects) == 0 {
		print.PrintTags(allTags, *listFlags, *tagFlags)
		return
	}

	// TODO: Add dirs and networks here
	if len(args) > 0 && len(projects) == 0 {
		args = core.Intersection(args, allTags)
		print.PrintTags(args, *listFlags, *tagFlags)
	} else if len(args) == 0 && len(projects) > 0 {
		projectTags := config.GetTagsByProject(projects)
		print.PrintTags(projectTags, *listFlags, *tagFlags)
	} else {
		projectTags := config.GetTagsByProject(projects)
		args = core.Intersection(args, projectTags)
		print.PrintTags(args, *listFlags, *tagFlags)
	}
}
