package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func describeProjectsCmd(
	config *dao.Config,
	configErr *error,
	describeFlags *core.DescribeFlags,
) *cobra.Command {
	var projectFlags core.ProjectFlags
	var setProjectFlags core.SetProjectFlags

	cmd := cobra.Command{
		Aliases: []string{"project", "prj"},
		Use:     "projects [projects]",
		Short:   "Describe projects",
		Long:    "Describe projects.",
		Example: `  # Describe all projects
  mani describe projects

  # Describe projects by name
  mani describe projects <project>

  # Describe projects by tags
  mani describe projects --tags <tag>

  # Describe projects by paths
  mani describe projects --paths <path>

	# Describe projects matching a tag expression
	mani run <task> --tags-expr '<tag-1> || <tag-2>'`,

		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)

			setProjectFlags.All = cmd.Flags().Changed("all")
			setProjectFlags.Cwd = cmd.Flags().Changed("cwd")
			setProjectFlags.Target = cmd.Flags().Changed("target")

			describeProjects(config, args, &projectFlags, &setProjectFlags, describeFlags)
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

	cmd.Flags().BoolVarP(&projectFlags.All, "all", "a", true, "select all projects")
	cmd.Flags().BoolVarP(&projectFlags.Cwd, "cwd", "k", false, "select current working directory")

	cmd.Flags().StringSliceVarP(&projectFlags.Tags, "tags", "t", []string{}, "filter projects by tags")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetTags()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringVarP(&projectFlags.TagsExpr, "tags-expr", "E", "", "target projects by tags expression")
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&projectFlags.Paths, "paths", "d", []string{}, "filter projects by paths")
	err = cmd.RegisterFlagCompletionFunc("paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetProjectPaths()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringVarP(&projectFlags.Target, "target", "T", "", "target projects by target name")
	err = cmd.RegisterFlagCompletionFunc("target", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		values := config.GetTargetNames()
		return values, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().BoolVarP(&projectFlags.Edit, "edit", "e", false, "edit project")

	return &cmd
}

func describeProjects(
	config *dao.Config,
	args []string,
	projectFlags *core.ProjectFlags,
	setProjectFlags *core.SetProjectFlags,
	describeFlags *core.DescribeFlags,
) {
	if projectFlags.Edit {
		if len(args) > 0 {
			err := config.EditProject(args[0])
			core.CheckIfError(err)
		} else {
			err := config.EditProject("")
			core.CheckIfError(err)
		}
	} else {

		projectFlags.Projects = args
		if !setProjectFlags.All {
			// If no flags are set, use all and empty default target (but not the modified one by user)
			// If target is set, use the defaults from that target and respect other flags
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
			theme, err := config.GetTheme(describeFlags.Theme)
			core.CheckIfError(err)

			output := print.PrintProjectBlocks(projects, true, theme.Block, print.GookitFormatter{})
			fmt.Print(output)
		}
	}
}
