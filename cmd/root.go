package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core/dao"
)

const (
	appName      = "mani"
	shortAppDesc = "manage multiple repositories and run commands across them"
	longAppDesc  = `Manage multiple repositories and run commands across them.

Documentation: https://manicli.com
Report issues:  https://github.com/alajmo/mani/issues`
)

var (
	config         dao.Config
	configErr      error
	configFilepath string
	userConfigPath string
	color          bool
	buildMode      = ""
	version        = "dev"
	commit         = "none"
	date           = "n/a"
	rootCmd        = &cobra.Command{
		Use:     appName,
		Short:   shortAppDesc,
		Long:    longAppDesc,
		Version: version,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// When user input's wrong command or flag
		os.Exit(1)
	}
}

func init() {
	// Modify default shell in-case we're on windows
	if runtime.GOOS == "windows" {
		dao.DEFAULT_SHELL = "powershell -NoProfile"
		dao.DEFAULT_SHELL_PROGRAM = "powershell"
	}

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&configFilepath, "config", "c", "", "specify config")
	rootCmd.PersistentFlags().StringVarP(&userConfigPath, "user-config", "u", "", "specify user config")
	rootCmd.PersistentFlags().BoolVar(&color, "color", true, "enable color")

	rootCmd.AddCommand(
		completionCmd(),
		genCmd(),
		initCmd(),
		execCmd(&config, &configErr),
		runCmd(&config, &configErr),
		listCmd(&config, &configErr),
		describeCmd(&config, &configErr),
		syncCmd(&config, &configErr),
		editCmd(&config, &configErr),
		checkCmd(&configErr),
		tuiCmd(&config, &configErr),
	)

	rootCmd.SetVersionTemplate(fmt.Sprintf("Version: %-10s\nCommit: %-10s\nDate: %-10s\n", version, commit, date))

	if buildMode == "man" {
		rootCmd.AddCommand(genDocsCmd(longAppDesc))
	}

	rootCmd.DisableAutoGenTag = true
}

func initConfig() {
	config, configErr = dao.ReadConfig(configFilepath, userConfigPath, color)
}
