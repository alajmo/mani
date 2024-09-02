package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func listProjectsCmd(
	config *dao.Config,
	configErr *error,
	listFlags *core.ListFlags,
) *cobra.Command {
	var projectFlags core.ProjectFlags
	var setProjectFlags core.SetProjectFlags

	cmd := cobra.Command{
		Aliases: []string{"project", "proj", "pr"},
		Use:     "projects [projects]",
		Short:   "List projects",
		Long:    "List projects.",
		Example: `  # List all projects
  mani list projects

  # List projects by name
  mani list projects <project>

  # List projects by tags
  mani list projects --tags <tag>

  # List projects by paths
  mani list projects --paths <path>

	# List projects matching a tag expression
	mani run <task> --tags-expr '<tag-1> || <tag-2>'`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)

			setProjectFlags.All = cmd.Flags().Changed("all")
			setProjectFlags.Cwd = cmd.Flags().Changed("cwd")
			setProjectFlags.Target = cmd.Flags().Changed("target")

			listProjects(config, args, listFlags, &projectFlags, &setProjectFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			projectNames := config.GetProjectNames()
			return projectNames, cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVar(&listFlags.Tree, "tree", false, "display output in tree format")

	cmd.Flags().BoolVarP(&projectFlags.All, "all", "a", true, "select all projects")
	cmd.Flags().BoolVarP(&projectFlags.Cwd, "cwd", "k", false, "select current working directory")

	cmd.Flags().StringSliceVarP(&projectFlags.Tags, "tags", "t", []string{}, "select projects by tags")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		options := config.GetTags()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringVarP(&projectFlags.TagsExpr, "tags-expr", "E", "", "select projects by tags expression")
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&projectFlags.Paths, "paths", "d", []string{}, "select projects by paths")
	err = cmd.RegisterFlagCompletionFunc("paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		options := config.GetProjectPaths()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringVarP(&projectFlags.Target, "target", "T", "", "select projects by target name")
	err = cmd.RegisterFlagCompletionFunc("target", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		values := config.GetTargetNames()
		return values, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVar(&projectFlags.Headers, "headers", []string{"project", "tag", "description"}, "specify columns to display [project, path, relpath, description, url, tag]")
	err = cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		validHeaders := []string{"project", "path", "relpath", "description", "url", "tag"}
		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listProjects(
	config *dao.Config,
	args []string,
	listFlags *core.ListFlags,
	projectFlags *core.ProjectFlags,
	setProjectFlags *core.SetProjectFlags,
) {
	theme, err := config.GetTheme(listFlags.Theme)
	core.CheckIfError(err)

	if listFlags.Tree {
		tree, err := config.GetProjectsTree(projectFlags.Paths, projectFlags.Tags)
		core.CheckIfError(err)
		print.PrintTree(config, *theme, listFlags, tree)
		return
	}

	projectFlags.Projects = args
	if !setProjectFlags.All {
		isNoFiltersSet := len(projectFlags.Projects) == 0 &&
			len(projectFlags.Paths) == 0 &&
			len(projectFlags.Tags) == 0 &&
			projectFlags.TagsExpr == "" &&
			!setProjectFlags.Cwd &&
			!setProjectFlags.Target
		projectFlags.All = isNoFiltersSet
	}
	projects, err := config.GetFilteredProjects(projectFlags)
	core.CheckIfError(err)

	if len(projects) == 0 {
		fmt.Println("No matching projects found")
	} else {
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
		print.PrintTable(projects, options, projectFlags.Headers, []string{}, os.Stdout)
		fmt.Println()
	}
}
