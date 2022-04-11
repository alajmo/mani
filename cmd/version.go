package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	version, commit, date = "dev", "none", "n/a"
)

func versionCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "version",
		Short: "Print version/build info",
		Long:  "Print version/build info.",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
		},
	}

	return &cmd
}

func printVersion() {
	const secFmt = "%-10s "

	fmt.Printf("Version: %-10s\n", version)
	fmt.Printf("Commit: %-10s\n", commit)
	fmt.Printf("Date: %-10s\n", date)
}
