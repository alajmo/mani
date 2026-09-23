package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func projectRemoveCmd(config *dao.Config, configErr *error) *cobra.Command {
	var force bool

	cmd := cobra.Command{
		Aliases: []string{"rm"},
		Use:     "remove <name>",
		Short:   "Remove a project from mani.yaml",
		Long: `Remove a project from mani.yaml.

The project entry, together with the comment and blank lines directly above
it, is removed; all other formatting is preserved. The project's files on
disk are NOT deleted. You are prompted for confirmation unless --force is
given.`,

		Example: `  # Remove a project (prompts for confirmation)
  mani project rm my-repo

  # Remove without prompting
  mani project rm my-repo --force`,

		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)

			name := args[0]
			if _, err := config.GetProject(name); err != nil {
				core.CheckIfError(&core.ProjectNotFound{Name: []string{name}})
			}

			if !force && !confirm(fmt.Sprintf("Remove project `%s` from %s?", name, config.Path)) {
				fmt.Println("Aborted.")
				return
			}

			core.CheckIfError(dao.RemoveProject(config.Path, name))

			fmt.Printf("Removed project `%s` from %s\n", name, config.Path)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil || len(args) == 1 {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}
			return config.GetProjectNames(), cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip the confirmation prompt")

	return &cmd
}

// confirm asks a yes/no question on stdin, defaulting to no.
func confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}
