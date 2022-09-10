// This source will generate
//   - core/mani.1
//   - docs/commands.md
//
// and is not included in the final build.

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
)

func genDocsCmd(longAppDesc string) *cobra.Command {
	cmd := cobra.Command{
		Use:                   "gen-docs",
		Short:                 "Generate man and markdown pages",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			err := core.CreateManPage(
				longAppDesc,
				version,
				date,
				rootCmd,
				runCmd(&config, &configErr),
				execCmd(&config, &configErr),
				initCmd(),
				syncCmd(&config, &configErr),
				editCmd(&config, &configErr),
				listCmd(&config, &configErr),
				describeCmd(&config, &configErr),
				checkCmd(&config, &configErr),
				genCmd(),
			)
			core.CheckIfError(err)
		},

		DisableAutoGenTag: true,
	}

	return &cmd
}
