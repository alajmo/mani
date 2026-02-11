package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func listTasksCmd(config *dao.Config, configErr *error, listFlags *core.ListFlags) *cobra.Command {
	var taskFlags core.TaskFlags

	cmd := cobra.Command{
		Aliases: []string{"task", "tsk", "tsks"},
		Use:     "tasks [tasks]",
		Short:   "List tasks",
		Long:    "List tasks.",
		Example: `  # List all tasks
  mani list tasks

  # List tasks by name
  mani list task <task>`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			listTasks(config, args, listFlags, &taskFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			values := config.GetTaskNames()
			return values, cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().StringSliceVar(&taskFlags.Headers, "headers", []string{"task", "description"}, "specify columns to display [task, description, target, spec]")
	err := cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := []string{"task", "description", "target", "spec"}
		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listTasks(
	config *dao.Config,
	args []string,
	listFlags *core.ListFlags,
	taskFlags *core.TaskFlags,
) {
	tasks, err := config.GetTasksByNames(args)
	core.CheckIfError(err)

	theme, err := config.GetTheme(listFlags.Theme)
	core.CheckIfError(err)

	if len(tasks) == 0 {
		fmt.Println("No tasks")
		return
	}

	// Handle JSON/YAML output
	if listFlags.Output == "json" || listFlags.Output == "yaml" {
		outputTasks := make([]print.TaskOutput, len(tasks))
		for i, t := range tasks {
			outputTasks[i] = print.TaskOutput{
				Name:        t.Name,
				Description: t.Desc,
				Spec:        t.SpecData.Name,
				Target:      t.TargetData.Name,
			}
		}

		if listFlags.Output == "json" {
			err = print.PrintListJSON(outputTasks, os.Stdout)
		} else {
			err = print.PrintListYAML(outputTasks, os.Stdout)
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
	print.PrintTable(tasks, options, taskFlags.Headers, []string{}, os.Stdout)
	fmt.Println()
}
