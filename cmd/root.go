package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"github.com/spf13/viper"
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

func initConfig() {
	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (by default it checks current and all parent directories for mani.yaml|yml)")
	// rootCmd.PersistentFlags().StringVar(&cfgFile   , "config", "", "config file (default is $HOME/.cobra.yaml)")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	rootCmd.AddCommand(
		versionCmd(),
		initCmd(),
		completionCmd(),
		execCmd(&configFile),
		runCmd(&configFile),
		listCmd(&configFile),
		describeCmd(&configFile),
		syncCmd(&configFile),
		infoCmd(&configFile),
		editCmd(&configFile),
	)
}
