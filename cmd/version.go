package cmd

import (
	"fmt"
	color "github.com/logrusorgru/aurora"
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
	fmt.Println(color.Blue(fmt.Sprintf(secFmt, "Version:")), version)
	fmt.Println(color.Blue(fmt.Sprintf(secFmt, "Commit:")), commit)
	fmt.Println(color.Blue(fmt.Sprintf(secFmt, "Date:")), date)
}
