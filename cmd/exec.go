package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func execCmd(config *dao.Config, configErr *error) *cobra.Command {
	var runFlags core.RunFlags

	cmd := cobra.Command{
		Use:   "exec <command>",
		Short: "Execute arbitrary commands",
		Long: `Execute arbitrary commands.

Single quote your command if you don't want the file globbing and environments variables expansion to take place
before the command gets executed in each directory.`,

		Example: `  # List files in all projects
  mani exec ls --all-projects

  # List all git files that have markdown suffix
  mani exec 'git ls-files | grep -e ".md"' --all-projects`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			execute(args, config, runFlags)
		},
	}

	cmd.Flags().BoolVar(&runFlags.DryRun, "dry-run", false, "don't execute any command, just print the output of the command to see what will be executed")
	cmd.Flags().BoolVar(&runFlags.Parallel, "parallel", false, "Run tasks in parallel")
	cmd.Flags().StringVarP(&runFlags.Output, "output", "o", "list", "Output list|table|markdown|html")
	err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		valid := []string{"table", "markdown", "html"}
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().BoolVarP(&runFlags.Cwd, "cwd", "k", false, "current working directory")

	cmd.Flags().BoolVarP(&runFlags.AllProjects, "all", "a", false, "target all projects")

	cmd.Flags().StringSliceVarP(&runFlags.Projects, "projects", "p", []string{}, "target projects by their name")
	err = cmd.RegisterFlagCompletionFunc("projects", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		projects := config.GetProjectNames()
		return projects, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&runFlags.Paths, "paths", "g", []string{}, "target directories by their path")
	err = cmd.RegisterFlagCompletionFunc("paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		projectPaths := config.GetProjectPaths()
		dirPaths := config.GetDirPaths()

		options := append(projectPaths, dirPaths...)

		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&runFlags.Tags, "tags", "t", []string{}, "target entities by their tag")
	err = cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		tags := config.GetTags()
		return tags, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func execute(
	args []string,
	config *dao.Config,
	runFlags core.RunFlags,
) {
	projectEntities, dirEntities := config.GetEntities(runFlags)

	if len(projectEntities) == 0 && len(dirEntities) == 0 {
		fmt.Println("No targets")
	} else {
		cmd := strings.Join(args[0:], " ")
		if len(projectEntities) > 0 {
			entityList := dao.EntityList{
				Type:     "Project",
				Entities: projectEntities,
			}

			dao.RunExec(cmd, entityList, config, &runFlags)
		}

		if len(dirEntities) > 0 {
			entityList := dao.EntityList{
				Type:     "Directory",
				Entities: dirEntities,
			}
			dao.RunExec(cmd, entityList, config, &runFlags)
		}
	}
}
