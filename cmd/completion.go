package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

func completionCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:
Bash:

  $ source <(mani completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ mani completion bash > /etc/bash_completion.d/mani
  # macOS:
  $ mani completion bash > /usr/local/etc/bash_completion.d/mani

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ mani completion zsh > "${fpath[1]}/_mani"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ mani completion fish | source

  # To load completions for each session, execute once:
  $ mani completion fish > ~/.config/fish/completions/mani.fish

PowerShell:

  PS> mani completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> mani completion powershell > mani.ps1
  # and source this file from your PowerShell profile.
		`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run:                   generateCompletion,
	}

	return &cmd
}

func generateCompletion(cmd *cobra.Command, args []string) {
	switch args[0] {
	case "bash":
		cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		cmd.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
	}
}
