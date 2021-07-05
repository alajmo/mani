package cmd

import (
	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/print"
	"github.com/spf13/cobra"
)

func listProjectsCmd(configFile *string, listFlags *core.ListFlags) *cobra.Command {
	var projectFlags core.ListProjectFlags

	cmd := cobra.Command{
		Aliases: []string { "project", "proj" },
		Use:   "projects [flags]",
		Short: "List projects",
		Long:  "List projects",
		Example: `  # List projects
  mani list projects`,
		Run: func(cmd *cobra.Command, args []string) {
			listProjects(configFile, args, listFlags, &projectFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			_, config, err := core.ReadConfig(*configFile)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			projectNames := core.GetProjectNames(config.Projects)
			return projectNames, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringSliceVarP(&projectFlags.Tags, "tags", "t", []string{}, "filter projects by their tag")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, config, err := core.ReadConfig(*configFile)

		if err != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validTags := core.GetTags(config.Projects)
		return validTags, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVar(&projectFlags.Headers, "headers", []string{ "name", "tags", "description" }, "Specify headers, defaults to name, tags, description")
	err = cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, _, err := core.ReadConfig(*configFile)

		if err != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := []string { "name", "path", "description", "url", "tags" }

		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listProjects(configFile *string, args []string, listFlags *core.ListFlags, projectFlags *core.ListProjectFlags) {
	configPath, config, err := core.ReadConfig(*configFile)
	core.CheckIfError(err)

	filteredProjects := core.FilterProjectOnTag(config.Projects, projectFlags.Tags)
	filteredProjects = core.FilterProjectOnName(filteredProjects, args)

	print.PrintProjects(configPath, filteredProjects, *listFlags, *projectFlags)
}
