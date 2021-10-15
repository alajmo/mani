package dao

import (
	"fmt"
	"os"
	"strings"
	"os/exec"
	"sync"
	"bytes"

	"github.com/jedib0t/go-pretty/v6/table"

	core "github.com/alajmo/mani/core"
	// render "github.com/alajmo/mani/core/render"
)

func (t *Task) TableTask(
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

	spinner, err := TaskSpinner()
	core.CheckIfError(err)

	err = spinner.Start()
	core.CheckIfError(err)

	var data core.TableOutput

	/**
	** Headers
	**/
	data.Headers = append(data.Headers, entityList.Type)

	// Append Command name if set
	if t.Command != "" {
		data.Headers = append(data.Headers, t.Name)
	}

	// Append Command names if set
	for _, cmd := range t.Commands {
		if cmd.Task != "" {
			task, err := config.GetTask(cmd.Task)
			core.CheckIfError(err)

			if cmd.Name != "" {
				data.Headers = append(data.Headers, cmd.Name)
			} else {
				data.Headers = append(data.Headers, task.Name)
			}
		} else {
			data.Headers = append(data.Headers, cmd.Name)
		}
	}

	for _, entity := range entityList.Entities {
		data.Rows = append(data.Rows, table.Row{entity.Name})
	}

	/**
	** Values
	**/
	var wg sync.WaitGroup

	for i, entity := range entityList.Entities {
		wg.Add(1)

		if t.Parallell {
			spinner.Message(" Running")
			go t.work(config, &data, entity, runFlags.DryRun, i, &wg)
		} else {
			spinner.Message(fmt.Sprintf(" %v", entity.Name))
			t.work(config, &data, entity, runFlags.DryRun, i, &wg)
		}
	}

	wg.Wait()

	err = spinner.Stop()
	core.CheckIfError(err)

	t.printTable(data)
}

func (t Task) work(
	config *Config,
	data *core.TableOutput,
	entity Entity,
	dryRunFlag bool,
	i int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for _, cmd := range t.Commands {
		var output string
		var err error
		output, err = runTable(cmd.Command, cmd.EnvList, *config, cmd.Shell, entity, dryRunFlag)

		if err != nil {
			data.Rows[i] = append(data.Rows[i], output)
			return
		} else {
			data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
		}
	}

	if t.Command != "" {
		var output string
		var err error
		output, err = runTable(t.Command, t.EnvList, *config, t.Shell, entity, dryRunFlag)

		if err != nil {
			data.Rows[i] = append(data.Rows[i], err)
		} else {
			data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
		}
	}
}

func runTable(
	cmdStr string,
	envList []string,
	config Config,
	shell string,
	entity Entity,
	dryRun bool,
) (string, error) {
	entityPath, err := core.GetAbsolutePath(config.Path, entity.Path, entity.Name)
	if err != nil {
		return "", &core.FailedToParsePath{Name: entityPath}
	}
	if _, err := os.Stat(entityPath); os.IsNotExist(err) {
		return "", &core.PathDoesNotExist{Path: entityPath}
	}

	defaultArguments := getDefaultArguments(config.Path, config.Dir, entity)
	shellProgram, commandStr := formatShellString(shell, cmdStr)

	// Execute Command
	cmd := exec.Command(shellProgram, commandStr...)
	cmd.Dir = entityPath

	var output string
	if dryRun {
		for _, arg := range defaultArguments {
			env := strings.SplitN(arg, "=", 2)
			os.Setenv(env[0], env[1])
		}

		for _, arg := range envList {
			env := strings.SplitN(arg, "=", 2)
			os.Setenv(env[0], env[1])
		}

		output = os.ExpandEnv(cmdStr)
	} else {
		cmd.Env = append(os.Environ(), defaultArguments...)
		cmd.Env = append(cmd.Env, envList...)

		var outb bytes.Buffer
		var errb bytes.Buffer

		cmd.Stdout = &outb
		cmd.Stderr = &errb

		err := cmd.Run()
		if err != nil {
			output = errb.String()
		} else {
			output = outb.String()
		}

		return output, err
	}

	return output, nil
}

func ExecCmd(
	configPath string,
	shell string,
	project Project,
	cmdString string,
	dryRun bool,
) (string, error) {
	projectPath, err := core.GetAbsolutePath(configPath, project.Path, project.Name)
	if err != nil {
		return "", &core.FailedToParsePath{Name: projectPath}
	}
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return "", &core.PathDoesNotExist{Path: projectPath}
	}
	// TODO: FIX THIS
	// defaultArguments := getDefaultArguments(config.Path, config.Dir, entity)

	// Execute Command
	shellProgram, commandStr := formatShellString(shell, cmdString)

	cmd := exec.Command(shellProgram, commandStr...)
	cmd.Dir = projectPath

	var output string
	if dryRun {
		// for _, arg := range defaultArguments {
		// 	env := strings.SplitN(arg, "=", 2)
		// 	os.Setenv(env[0], env[1])
		// }

		output = os.ExpandEnv(cmdString)
	} else {
		// cmd.Env = append(os.Environ(), defaultArguments...)
		out, _ := cmd.CombinedOutput()
		output = string(out)
	}

	return output, nil
}

func (task Task) printTable(data core.TableOutput) {
	switch task.ThemeData.Table {
	case "ascii":
		core.ManiList.Box = core.StyleBoxASCII
	default:
		core.ManiList.Box = core.StyleBoxDefault
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	t.AppendHeader(data.Headers)

	for _, row := range data.Rows {
		t.AppendRow(row)
		t.AppendSeparator()
	}

	switch task.Output {
	case "markdown":
		t.RenderMarkdown()
	case "html":
		t.RenderHTML()
	default:
		t.Render()
	}
}