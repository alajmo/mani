package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

const (
	appName      = "mani"
	shortAppDesc = "mani is a tool used to manage multiple repositories"
	longAppDesc  = `mani is a tool used to manage multiple repositories`
)

var (
	configFile string
	rootCmd    = &cobra.Command{
		Use:   appName,
		Short: shortAppDesc,
		Long:  longAppDesc,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()
	rootCmd.AddCommand(
		versionCmd(),
		initCmd(),
		completionCmd(),
		execCmd(&configFile),
		runCmd(&configFile),
		listCmd(&configFile),
		syncCmd(&configFile),
		infoCmd(&configFile),
		editCmd(&configFile),
	)

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (by default it checks current and all parent directories for mani.yaml|yml)")
}
