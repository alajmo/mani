package dao

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	core "github.com/alajmo/mani/core"
)

func RunText(
	cmdStr string,
	envList []string,
	config Config,
	shell string,
	project Project,
	dryRun bool,
) error {
	projectPath, err := core.GetAbsolutePath(config.Path, project.Path, project.Name)
	if err != nil {
		return &core.FailedToParsePath{Name: projectPath}
	}
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return &core.PathDoesNotExist{Path: projectPath}
	}

	defaultArguments := getDefaultArguments(config.Path, config.Dir, project)
	shellProgram, commandStr := formatShellString(shell, cmdStr)

	// Execute Command
	cmd := exec.Command(shellProgram, commandStr...)
	cmd.Dir = projectPath

	envs := core.MergeEnv(envList, project.EnvList, defaultArguments, []string{})

	if dryRun {
		for _, arg := range envs {
			env := strings.SplitN(arg, "=", 2)
			os.Setenv(env[0], env[1])
		}

		fmt.Println(os.ExpandEnv(cmdStr))
	} else {
		cmd.Env = append(os.Environ(), envs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()

		return err
	}

	return nil
}
