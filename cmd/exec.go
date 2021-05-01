package cmd

import (
	"fmt"
	core "github.com/alajmo/mani/core"
	"github.com/spf13/cobra"
	"strings"
)

func execCmd(configFile *string) *cobra.Command {
	var dryRun bool
	var cwd bool
	var allProjects bool
	var tags []string
	var projects []string

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
			executeCmd(args, configFile, dryRun, cwd, allProjects, tags, projects)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "don't execute any command, just print the output of the command to see what will be executed")
	cmd.Flags().BoolVarP(&cwd, "cwd", "k", false, "current working directory")
	cmd.Flags().BoolVarP(&allProjects, "all-projects", "a", false, "target all projects")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "target projects by their tag")
	cmd.Flags().StringSliceVarP(&projects, "projects", "p", []string{}, "target projects by their name")

	cmd.MarkFlagCustom("projects", "__mani_parse_projects")
	cmd.MarkFlagCustom("tags", "__mani_parse_tags")

	return &cmd
}

func executeCmd(args []string, configFile *string, dryRunFlag bool, cwdFlag bool, allProjectsFlag bool, tagsFlag []string, projectsFlag []string) {
	configPath, config, err := core.ReadConfig(*configFile)

	if err != nil {
		fmt.Println(err)
		return
	}

	var finalProjects []core.Project
	if allProjectsFlag {
		finalProjects = config.Projects
	} else {
		var tagProjects []core.Project
		if len(tagsFlag) > 0 {
			tagProjects = core.GetProjectsByTag(tagsFlag, config.Projects)
		}

		var projects []core.Project
		if len(projectsFlag) > 0 {
			projects = core.GetProjects(projectsFlag, config.Projects)
		}

		var cwdProject core.Project
		if cwdFlag {
			cwdProject = core.GetCwdProject(config.Projects)
		}

		finalProjects = core.GetUnionProjects(tagProjects, projects, cwdProject)
	}

	cmd := strings.Join(args[0:], " ")
	for _, project := range finalProjects {
		err := core.ExecCmd(configPath, project, cmd, dryRunFlag)

		if err != nil {
			fmt.Println(err)
		}
	}

}
