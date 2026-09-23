package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core/dao"
)

func projectCmd(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Aliases: []string{"projects", "proj"},
		Use:     "project",
		Short:   "Add or remove projects in mani.yaml",
		Long: `Add or remove projects in mani.yaml.

Edits are lossless: comments, blank lines, indentation, and any
hand-customized fields outside the touched entry are preserved
byte-for-byte. The edited file is validated as YAML before it is written,
so a failed edit never corrupts the config.`,

		Example: `  # Add a remote project (cloned on next sync)
  mani project add my-repo --url git@github.com:org/my-repo.git --tags backend

  # Remove a project
  mani project rm my-repo`,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		projectAddCmd(config, configErr),
		projectRemoveCmd(config, configErr),
	)

	return &cmd
}
