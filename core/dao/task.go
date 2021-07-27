package dao

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"github.com/theckman/yacspin"

	core "github.com/alajmo/mani/core"
)

type CommandInterface interface {
	RunCmd() (string, error)
	GetEnv() ([]string)
	SetEnvList() ([]string)
	GetValue(string) (string)
}

type CommandBase struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Env         yaml.Node		  `yaml:"env"`
	EnvList     []string
	Shell		string            `yaml:"shell"`
	Command     string            `yaml:"command"`
}

type Task struct {
	Commands []Command	`yaml:"commands"`
	CommandBase			`yaml:",inline"`
}

type Command struct {
	CommandBase `yaml:",inline"`
}

func (c CommandBase) GetEnv() []string {
	var envs []string
	count := len(c.Env.Content)

	for i := 0; i < count; i += 2 {
		env := fmt.Sprintf("%v=%v", c.Env.Content[i].Value, c.Env.Content[i + 1].Value)
		envs = append(envs, env)
	}

	return envs
}

func (c *CommandBase) SetEnvList(userEnv []string, configEnv []string) {
	cmdEnv, err := core.EvaluateEnv(c.GetEnv())
	core.CheckIfError(err)

	globalEnv, err := core.EvaluateEnv(configEnv)
	core.CheckIfError(err)

	envList := core.MergeEnv(userEnv, cmdEnv, globalEnv)

	c.EnvList = envList
}

func (c CommandBase) GetValue(key string) string {
	switch key {
	case "Name", "name":
		return c.Name
	case "Description", "description":
		return c.Description
	case "Shell", "shell":
		return c.Shell
	case "Command", "command":
		return c.Command
	}

	return ""
}

func getDefaultArguments(configPath string, project Project) []string {
	// Default arguments
	maniConfigPath := fmt.Sprintf("MANI_CONFIG_PATH=%s", configPath)
	maniConfigDir := fmt.Sprintf("MANI_CONFIG_DIR=%s", filepath.Dir(configPath))
	projectNameEnv := fmt.Sprintf("MANI_PROJECT_NAME=%s", project.Name)
	projectUrlEnv := fmt.Sprintf("MANI_PROJECT_URL=%s", project.Url)
	projectPathEnv := fmt.Sprintf("MANI_PROJECT_PATH=%s", project.Path)

	defaultArguments := []string {maniConfigPath, maniConfigDir, projectNameEnv, projectUrlEnv, projectPathEnv}

	return defaultArguments
}

func (c CommandBase) RunCmd(
	configPath string,
	shell string,
	project Project,
	dryRun bool,
) (string, error) {
	projectPath, err := GetAbsolutePath(configPath, project.Path, project.Name)
	if err != nil {
		return "", &core.FailedToParsePath{ Name: projectPath }
	}
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return "", &core.PathDoesNotExist{ Path: projectPath }
	}

	defaultArguments := getDefaultArguments(configPath, project)

	// Execute Command
	shellProgram, commandStr := formatShellString(shell, c.Command)
	cmd := exec.Command(shellProgram, commandStr...)
	cmd.Dir = projectPath

	var output string
	if dryRun {
		for _, arg := range defaultArguments {
			env := strings.SplitN(arg, "=", 2)
			os.Setenv(env[0], env[1])
		}

		for _, arg := range c.EnvList {
			env := strings.SplitN(arg, "=", 2)
			os.Setenv(env[0], env[1])
		}

		output = os.ExpandEnv(c.Command)
	} else {
		cmd.Env = append(os.Environ(), defaultArguments...)
		cmd.Env = append(cmd.Env, c.EnvList...)
		out, _ := cmd.CombinedOutput()
		output = string(out)
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
	projectPath, err := GetAbsolutePath(configPath, project.Path, project.Name)
	if err != nil {
		return "", &core.FailedToParsePath{ Name: projectPath }
	}
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return "", &core.PathDoesNotExist{ Path: projectPath }
	}
	defaultArguments := getDefaultArguments(configPath, project)

	// Execute Command
	shellProgram, commandStr := formatShellString(shell, cmdString)
	cmd := exec.Command(shellProgram, commandStr...)
	cmd.Dir = projectPath

	var output string
	if dryRun {
		for _, arg := range defaultArguments {
			env := strings.SplitN(arg, "=", 2)
			os.Setenv(env[0], env[1])
		}

		output = os.ExpandEnv(cmdString)
	} else {
		cmd.Env = append(os.Environ(), defaultArguments...)
		out, _ := cmd.CombinedOutput()
		output = string(out)
	}

	return output, nil
}

func formatShellString(shell string, command string) (string, []string) {
	shellProgram := strings.SplitN(shell, " ", 2)
	return shellProgram[0], append(shellProgram[1:], command)
}

func TaskSpinner() (yacspin.Spinner, error) {
	cfg := yacspin.Config{
		Frequency:       100 * time.Millisecond,
		CharSet:         yacspin.CharSets[9],
		SuffixAutoColon: false,
	}

	spinner, err := yacspin.New(cfg)

	return *spinner, err
}
