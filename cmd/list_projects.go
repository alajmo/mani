package cmd

import (
	"github.com/alajmo/mani/core"
	"github.com/spf13/cobra"
)

func listProjectsCmd(configFile *string) *cobra.Command {
	var listRaw bool
	var tags []string

	cmd := cobra.Command{
		Use:   "projects [flags]",
		Short: "List projects",
		Long:  "List projects",
		Example: `  # List projects
  mani list projects`,
		Run: func(cmd *cobra.Command, args []string) {
			listProjects(configFile, args, listRaw, tags)
		},
	}

	cmd.Flags().BoolVar(&listRaw, "list-raw", false, "When listing objects, ignore description")
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

func listProjects(configFile *string, args []string, listRaw bool, tags []string) {
	configPath, config, err := core.ReadConfig(*configFile)
	core.CheckIfError(err)

	filteredProjects := core.FilterProjectOnTag(config.Projects, tags)
	core.PrintProjects(configPath, filteredProjects, "list", listRaw)
}
