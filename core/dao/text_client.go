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
	entity Entity,
	dryRun bool,
) error {
	entityPath, err := core.GetAbsolutePath(config.Path, entity.Path, entity.Name)
	if err != nil {
		return &core.FailedToParsePath{Name: entityPath}
	}
	if _, err := os.Stat(entityPath); os.IsNotExist(err) {
		return &core.PathDoesNotExist{Path: entityPath}
	}

	defaultArguments := getDefaultArguments(config.Path, config.Dir, entity)
	shellProgram, commandStr := formatShellString(shell, cmdStr)

	// Execute Command
	cmd := exec.Command(shellProgram, commandStr...)
	cmd.Dir = entityPath

	envs := core.MergeEnv(envList, entity.Env, defaultArguments, []string{})

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
