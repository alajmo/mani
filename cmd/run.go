package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func runCmd(config *dao.Config, configErr *error) *cobra.Command {
	var runFlags core.RunFlags

	cmd := cobra.Command{
		Use:   "run <task> [flags]",
		Short: "Run tasks",
		Long: `Run tasks.

The tasks are specified in a mani.yaml file along with the projects you can target.`,

		Example: `  # Run task 'pwd' for all projects
  mani run pwd --project-all

  # Checkout branch 'development' for all projects that have tag 'backend'
  mani run checkout -t backend branch=development`,

		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			// core.DebugPrint(config.GetTaskNames())
			run(args, config, &runFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			return config.GetTaskNames(), cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().BoolVar(&runFlags.Describe, "describe", false, "Print task information")
	cmd.Flags().BoolVar(&runFlags.DryRun, "dry-run", false, "don't execute any task, just print the output of the task to see what will be executed")
	cmd.Flags().BoolVarP(&runFlags.Edit, "edit", "e", false, "Edit task")
	cmd.Flags().BoolVarP(&runFlags.Serial, "serial", "s", false, "Run tasks in serial")
	cmd.Flags().StringVarP(&runFlags.Output, "output", "o", "", "Output list|table|markdown|html")

	cmd.Flags().BoolVarP(&runFlags.Cwd, "cwd", "k", false, "current working directory")

	cmd.Flags().BoolVar(&runFlags.AllProjects, "project-all", false, "target all projects")
	cmd.Flags().StringSliceVarP(&runFlags.Projects, "projects", "p", []string{}, "target projects by their name")
	cmd.Flags().StringSliceVar(&runFlags.ProjectPaths, "project-paths", []string{}, "target projects by their path")

	cmd.Flags().BoolVar(&runFlags.AllDirs, "dir-all", false, "target all dirs")
	cmd.Flags().StringSliceVarP(&runFlags.Dirs, "dirs", "d", []string{}, "target directories by their name")
	cmd.Flags().StringSliceVar(&runFlags.DirPaths, "dir-paths", []string{}, "target directories by their path")

	cmd.Flags().BoolVar(&runFlags.AllNetworks, "network-all", false, "target all networks")
	cmd.Flags().StringSliceVarP(&runFlags.Networks, "networks", "n", []string{}, "target networks by their name")
	cmd.Flags().StringSliceVar(&runFlags.Hosts, "hosts", []string{}, "target networks by their host")

	cmd.Flags().StringSliceVarP(&runFlags.Tags, "tags", "t", []string{}, "target entities by their tag")

	err := cmd.RegisterFlagCompletionFunc("projects", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		projects := config.GetProjectNames()
		return projects, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	err = cmd.RegisterFlagCompletionFunc("project-paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetProjectDirs()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	err = cmd.RegisterFlagCompletionFunc("dirs", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		dirs := config.GetDirNames()
		return dirs, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	err = cmd.RegisterFlagCompletionFunc("dir-paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetDirPaths()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	err = cmd.RegisterFlagCompletionFunc("networks", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		dirs := config.GetNetworkNames()
		return dirs, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	err = cmd.RegisterFlagCompletionFunc("hosts", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetAllHosts()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	err = cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		tags := config.GetTags()
		return tags, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	err = cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		valid := []string{"table", "markdown", "html"}
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func run(
	args []string,
	config *dao.Config,
	runFlags *core.RunFlags,
) {
	var taskNames []string
	var userArgs []string
	// Seperate user arguments from task names
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			userArgs = append(userArgs, arg)
		} else {
			taskNames = append(taskNames, arg)
		}
	}

	if runFlags.Edit {
		if len(args) > 0 {
			config.EditTask(taskNames[0])
			return
		} else {
			config.EditTask("")
			return
		}
	}

	for _, name := range taskNames {
		task, err := config.GetTask(name)
		core.CheckIfError(err)

		projectEntities, dirEntities, networkEntities := config.ParseTask(task, *runFlags)

		if len(projectEntities) == 0 &&  len(dirEntities) == 0 && len(networkEntities) == 0{
			fmt.Println("No targets")
		} else {
			if len(projectEntities) > 0 {
				entityList := dao.EntityList {
					Type: "Project",
					Entities: projectEntities,
				}

				task.RunTask(entityList, userArgs, config, runFlags)
			}

			if len(dirEntities) > 0 {
				entityList := dao.EntityList {
					Type: "Directory",
					Entities: dirEntities,
				}
				task.RunTask(entityList, userArgs, config, runFlags)
			}

			if len(networkEntities) > 0 {
				entityList := dao.EntityList {
					Type: "Host",
					Entities: networkEntities,
				}
				task.RunTask(entityList, userArgs, config, runFlags)
			}
		}
	}
}
