package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func listTagsCmd(config *dao.Config, configErr *error, listFlags *core.ListFlags) *cobra.Command {
	var tagFlags core.TagFlags

	cmd := cobra.Command{
		Aliases: []string{"tag"},
		Use:     "tags [tags]",
		Short:   "List tags",
		Long:    "List tags.",
		Example: `  # List all tags
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

	cmd.Flags().StringSliceVar(&tagFlags.Headers, "headers", []string{"tag", "project"}, "specify columns to display [project, tag]")
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

	allTags := config.GetTags()

	var tagsToUse []string
	if len(args) > 0 {
		foundTags := core.Intersection(args, allTags)
		// Could not find one of the provided tags
		if len(foundTags) != len(args) {
			core.CheckIfError(&core.TagNotFound{Tags: args})
		}
		tagsToUse = foundTags
	} else {
		tagsToUse = allTags
	}

	tags, err := config.GetTagAssocations(tagsToUse)
	core.CheckIfError(err)

	if len(tags) == 0 {
		fmt.Println("No tags")
		return
	}

	// Handle JSON/YAML output
	if listFlags.Output == "json" || listFlags.Output == "yaml" {
		outputTags := make([]print.TagOutput, len(tags))
		for i, t := range tags {
			outputTags[i] = print.TagOutput{
				Name:     t.Name,
				Projects: t.Projects,
			}
		}

		if listFlags.Output == "json" {
			err = print.PrintListJSON(outputTags, os.Stdout)
		} else {
			err = print.PrintListYAML(outputTags, os.Stdout)
		}
		core.CheckIfError(err)
		return
	}

	// Table/Markdown/HTML output
	theme.Table.Border.Rows = core.Ptr(false)
	theme.Table.Header.Format = core.Ptr("t")

	options := print.PrintTableOptions{
		Output:           listFlags.Output,
		Theme:            *theme,
		Tree:             listFlags.Tree,
		AutoWrap:         true,
		OmitEmptyRows:    false,
		OmitEmptyColumns: true,
		Color:            *theme.Color,
	}

	fmt.Println()
	print.PrintTable(tags, options, tagFlags.Headers, []string{}, os.Stdout)
	fmt.Println()
}
