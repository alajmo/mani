package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/exec"
)

func execCmd(config *dao.Config, configErr *error) *cobra.Command {
	var runFlags core.RunFlags
	var setRunFlags core.SetRunFlags

	cmd := cobra.Command{
		Use:   "exec <command>",
		Short: "Execute arbitrary commands",
		Long: `Execute arbitrary commands.
Use single quotes around your command to prevent file globbing and 
environment variable expansion from occurring before the command is 
executed in each directory.`,

		Example: `  # List files in all projects
  mani exec --all ls

  # List git files with markdown suffix in all projects
  mani exec --all 'git ls-files | grep -e ".md"'`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)

			// This is necessary since cobra doesn't support pointers for bools
			// (that would allow us to use nil as default value)
			setRunFlags.TTY = cmd.Flags().Changed("tty")
			setRunFlags.Parallel = cmd.Flags().Changed("parallel")
			setRunFlags.OmitEmptyRows = cmd.Flags().Changed("omit-empty-rows")
			setRunFlags.OmitEmptyColumns = cmd.Flags().Changed("omit-empty-columns")
			setRunFlags.IgnoreErrors = cmd.Flags().Changed("ignore-errors")
			setRunFlags.IgnoreNonExisting = cmd.Flags().Changed("ignore-non-existing")
			setRunFlags.Forks = cmd.Flags().Changed("forks")
			setRunFlags.Cwd = cmd.Flags().Changed("cwd")
			setRunFlags.All = cmd.Flags().Changed("all")

			if setRunFlags.Forks {
				forks, err := cmd.Flags().GetUint32("forks")
				core.CheckIfError(err)
				if forks == 0 {
					core.Exit(&core.ZeroNotAllowed{Name: "forks"})
				}
				runFlags.Forks = forks
			}

			execute(args, config, &runFlags, &setRunFlags)
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVar(&runFlags.TTY, "tty", false, "replace current process")
	cmd.Flags().BoolVar(&runFlags.DryRun, "dry-run", false, "print commands without executing them")
	cmd.Flags().BoolVarP(&runFlags.Silent, "silent", "s", false, "hide progress when running tasks")
	cmd.Flags().BoolVar(&runFlags.IgnoreNonExisting, "ignore-non-existing", false, "ignore non-existing projects")
	cmd.Flags().BoolVar(&runFlags.IgnoreErrors, "ignore-errors", false, "ignore errors")
	cmd.Flags().BoolVar(&runFlags.OmitEmptyRows, "omit-empty-rows", false, "omit empty rows in table output")
	cmd.Flags().BoolVar(&runFlags.OmitEmptyColumns, "omit-empty-columns", false, "omit empty columns in table output")
	cmd.Flags().BoolVar(&runFlags.Parallel, "parallel", false, "run tasks in parallel across projects")
	cmd.Flags().Uint32P("forks", "f", 4, "maximum number of concurrent processes")
	cmd.Flags().BoolVarP(&runFlags.Cwd, "cwd", "k", false, "use current working directory")
	cmd.Flags().BoolVarP(&runFlags.All, "all", "a", false, "target all projects")

	cmd.Flags().StringVarP(&runFlags.Output, "output", "o", "", "set output format [stream|table|markdown|html|json|yaml]")
	err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		valid := []string{"stream", "table", "markdown", "html", "json", "yaml"}
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringVarP(&runFlags.Spec, "spec", "J", "", "set spec")
	err = cmd.RegisterFlagCompletionFunc("spec", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		values := config.GetSpecNames()
		return values, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&runFlags.Projects, "projects", "p", []string{}, "select projects by name")
	err = cmd.RegisterFlagCompletionFunc("projects", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		projects := config.GetProjectNames()
		return projects, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&runFlags.Paths, "paths", "d", []string{}, "select projects by path")
	err = cmd.RegisterFlagCompletionFunc("paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetProjectPaths()

		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&runFlags.Tags, "tags", "t", []string{}, "select projects by tag")
	err = cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		tags := config.GetTags()
		return tags, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringVarP(&runFlags.TagsExpr, "tags-expr", "E", "", "select projects by tags expression")
	core.CheckIfError(err)

	cmd.Flags().StringVarP(&runFlags.Target, "target", "T", "", "target projects by target name")
	err = cmd.RegisterFlagCompletionFunc("target", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		values := config.GetTargetNames()
		return values, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.PersistentFlags().StringVar(&runFlags.Theme, "theme", "", "set theme")
	err = cmd.RegisterFlagCompletionFunc("theme", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		names := config.GetThemeNames()

		return names, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func execute(
	args []string,
	config *dao.Config,
	runFlags *core.RunFlags,
	setRunFlags *core.SetRunFlags,
) {
	cmd := strings.Join(args[0:], " ")
	var tasks []dao.Task

	tasks, projects, err := dao.ParseCmd(cmd, runFlags, setRunFlags, config)
	core.CheckIfError(err)

	target := exec.Exec{Projects: projects, Tasks: tasks, Config: *config}
	err = target.Run([]string{}, runFlags, setRunFlags)
	core.CheckIfError(err)
}
