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
		Short: "Print version",
		Long:  "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
		},
	}

	return &cmd
}

func printVersion() {
	const secFmt = "%-10s "
	fmt.Println(fmt.Sprintf(secFmt, "Version"), version)
	fmt.Println(fmt.Sprintf(secFmt, "Commit"), commit)
	fmt.Println(fmt.Sprintf(secFmt, "Date"), date)
}
