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

const (
	bash_completion_func = `
__mani_parse_projects() {
	local mani_output out
	if mani_output=$(mani list projects 2>/dev/null); then
		COMPREPLY=( $( compgen -W "${mani_output[*]}" -- "$cur" ) )
	fi
}

__mani_parse_tags() {
	local mani_output out
	if mani_output=$(mani list tags 2>/dev/null); then
		COMPREPLY=( $( compgen -W "${mani_output[*]}" -- "$cur" ) )
	fi
}

__mani_parse_run()
{
    if [[ "$prev" == "run" ]]; then
        local mani_output out
        if mani_output=$(mani list commands 2>/dev/null); then
            COMPREPLY=( $( compgen -W "${mani_output[*]}" -- "$cur" ) )
        fi
    fi
}

__mani_custom_func() {
	case ${last_command} in
		mani_run)
			__mani_parse_run
			return
			;;
		*)
			;;
	esac
}
`
)

var (
	configFile string
	rootCmd    = &cobra.Command{
		Use:                    appName,
		Short:                  shortAppDesc,
		Long:                   longAppDesc,
		BashCompletionFunction: bash_completion_func,
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
		completionCmd(&configFile),
		execCmd(&configFile),
		runCmd(&configFile),
		listCmd(&configFile),
		syncCmd(&configFile),
		infoCmd(&configFile),
	)

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (by default it checks current and all parent directories for mani.yaml|yml)")
}
