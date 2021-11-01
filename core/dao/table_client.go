package dao

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

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

	if runFlags.Parallel {
		t.Parallel = true
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

		if t.Parallel {
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
		output, err = runTable(*config, cmd.Command, cmd.EnvList, cmd.Shell, entity, dryRunFlag)
		data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))

		if err != nil && t.Abort {
			return
		}
	}

	if t.Command != "" {
		var output string
		output, _ = runTable(*config, t.Command, t.EnvList, t.Shell, entity, dryRunFlag)
		data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
	}
}

func runTable(
	config Config,
	cmdStr string,
	envList []string,
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

	// fmt.Println("----------------------")
	// fmt.Println(config.Path)
	// fmt.Println(entity.Path)
	// fmt.Println(entity.Name)
	// fmt.Println(entityPath)
	// fmt.Println("----------------------")

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

		output, err := cmd.CombinedOutput()

		return string(output), err
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
