package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "version",
		Short: "Print version/build info",
		Long:  "Print version/build info.",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
		},
		DisableAutoGenTag: true,
	}

	return &cmd
}

func printVersion() {
	fmt.Printf("Version: %-10s\n", version)
	fmt.Printf("Commit: %-10s\n", commit)
	fmt.Printf("Date: %-10s\n", date)
}
