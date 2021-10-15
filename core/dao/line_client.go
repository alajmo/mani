package dao

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"strings"
	"bufio"

	// "github.com/jedib0t/go-pretty/v6/table"
	color "github.com/logrusorgru/aurora"
	"golang.org/x/term"

	core "github.com/alajmo/mani/core"
)

type LineClient struct {
	cmd     *exec.Cmd
	stdout  io.Reader
	stderr  io.Reader
}

func (t *Task) LineTask(
	entityList EntityList,
	userArgs []string,
	config *Config,
	runFlags *core.RunFlags,
) {
	t.EnvList = GetEnvList(t.Env, userArgs, []string{}, config.GetEnv())

	if runFlags.Parallell {
		t.Parallell = true
	}

	// Set env for sub-commands
	for i := range t.Commands {
		t.Commands[i].EnvList = GetEnvList(t.Commands[i].Env, userArgs, t.EnvList, config.GetEnv())
	}

	var wg sync.WaitGroup

    width, _, err := term.GetSize(0)
	core.CheckIfError(err)
	var header string
	if t.Description != "" {
		header = fmt.Sprintf("%s [%s: %s]", color.Bold("TASK"), t.Name, t.Description)
	} else {
		header = fmt.Sprintf("%s [%s]", color.Bold("TASK"), t.Name)
	}

	fmt.Printf("\n%s %s\n", header, strings.Repeat("*", width - len(header) - 1))

	maxNameLength := entityList.GetLongestNameLength()

	for _, entity := range entityList.Entities {
		wg.Add(1)
		if t.Parallell {
			go t.workList(config, entity, runFlags.DryRun, maxNameLength, &wg)
		} else {
			t.workList(config, entity, runFlags.DryRun, maxNameLength, &wg)
		}
	}

	wg.Wait()
}

func (t Task) workList(
	config *Config,
	entity Entity,
	dryRunFlag bool,
	maxNameLength int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for i, cmd := range t.Commands {
		var header string
		if cmd.Description != "" {
			header = fmt.Sprintf("TASK %d/%d [%s: %s]", i+1, len(t.Commands), cmd.Name, cmd.Description)
		} else {
			header = fmt.Sprintf("TASK %d/%d [%s]", i+1, len(t.Commands), cmd.Name)
		}

		fmt.Println(header)
		runList(cmd.Command, cmd.EnvList, *config, cmd.Shell, entity, dryRunFlag, maxNameLength)
		fmt.Println()
	}

	if t.Command != "" {
		runList(t.Command, t.EnvList, *config, t.Shell, entity, dryRunFlag, maxNameLength)
	}
}

func runList(
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
				fmt.Printf("%s    %s| %s\n", entity.Name, strings.Repeat(" ", maxNameLength - len(entity.Name)), line)
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
