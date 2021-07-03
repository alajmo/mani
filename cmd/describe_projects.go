package cmd

import (
	"github.com/alajmo/mani/core"
	"github.com/spf13/cobra"
)

func describeProjectsCmd(configFile *string) *cobra.Command {
	var tags []string
	var projects []string

	cmd := cobra.Command{
		Aliases: []string { "project", "proj" },
		Use:   "projects [projects] [flags]",
		Short: "Describe projects",
		Long:  "Describe projects.",
		Example: `  # Describe projects
  mani describe projects

  # Describe projects that have tag frontend
  mani describe projects --tags frontend`,
		Run: func(cmd *cobra.Command, args []string) {
			describeProjects(configFile, args, tags, projects)
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

	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "filter projects by their tag")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, config, err := core.ReadConfig(*configFile)

		if err != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		tags := core.GetTags(config.Projects)
		return tags, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func describeProjects(configFile *string, args []string, tags []string, projects []string) {
	configPath, config, err := core.ReadConfig(*configFile)
	core.CheckIfError(err)

	filteredProjects := core.FilterProjectOnTag(config.Projects, tags)
	filteredProjects = core.FilterProjectOnName(filteredProjects, args)
	core.PrintProjects(configPath, filteredProjects, "block", false)
}
