package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func completionCmd(configFile *string) *cobra.Command {
	cmd := cobra.Command{
		Use:   "completion",
		Short: "Output shell completion code for bash",
		Long:  `Output shell completion code for bash.

Auto-complete requires bash-completion. There's two ways to add mani auto-completion:
- Source the completion script in your ~/.bashrc file:
  echo 'source <(mani completion)' >>~/.bashrc
or
- Add the completion script to the /etc/bash_completion.d directory:
  mani completion >/etc/bash_completion.d/mani`,
		Run: func(cmd *cobra.Command, args []string) {
			generateCompletion(configFile)
		},
	}

	return &cmd
}

func generateCompletion(configFile *string) {
	fmt.Println(*configFile)
	rootCmd.GenBashCompletion(os.Stdout)
}
