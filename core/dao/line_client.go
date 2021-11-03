package dao

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	core "github.com/alajmo/mani/core"
)

func RunList(
	cmdStr string,
	envList []string,
	config Config,
	shell string,
	entity Entity,
	dryRun bool,
	maxNameLength int,
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

	if dryRun {
		for _, arg := range defaultArguments {
			env := strings.SplitN(arg, "=", 2)
			os.Setenv(env[0], env[1])
		}

		for _, arg := range envList {
			env := strings.SplitN(arg, "=", 2)
			os.Setenv(env[0], env[1])
		}

		fmt.Println(os.ExpandEnv(cmdStr))
	} else {
		cmd.Env = append(os.Environ(), defaultArguments...)
		cmd.Env = append(cmd.Env, envList...)
		r, err := cmd.StdoutPipe()
		core.CheckIfError(err)
		cmd.Stderr = cmd.Stdout

		done := make(chan struct{})
		scanner := bufio.NewScanner(r)

		go func() {
			for scanner.Scan() {
				line := scanner.Text()
				fmt.Printf("%s    %s| %s\n", entity.Name, strings.Repeat(" ", maxNameLength-len(entity.Name)), line)
			}
			done <- struct{}{}
		}()
		err = cmd.Start()
		core.CheckIfError(err)
		<-done

		err = cmd.Wait()

		return err
	}

	return nil
}
