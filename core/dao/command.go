package dao

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	core "github.com/alajmo/mani/core"
)

type Command struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Args        map[string]string `yaml:"args"`
	Shell		string            `yaml:"shell"`
	Command     string            `yaml:"command"`
}

func (c Command) GetValue(key string) string {
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

type ProjectOutput struct {
	ProjectName string
	Output string
}

func (c Command) ParseUserArguments(userArguments []string) map[string]string {
	// Runtime arguments
	args := make(map[string]string)
	for _, arg := range userArguments {
		kv := strings.SplitN(arg, "=", 2)
		args[kv[0]] = kv[1]
	}

	// Default arguments
	for k, v := range c.Args {
		if (args[k] == "") {
			args[k] = v
		}
	}

	return args
}

func GetUserArguments(commandArgs map[string]string) []string {
	var args []string
	for k, v := range commandArgs {
		args = append(args, fmt.Sprintf("%v=%v", k, v))
	}

	return args
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

func RunCommand(
	configPath string,
	shell string,
	project Project,
	command *Command,
	userArguments []string,
	dryRun bool,
) (string, error){
	projectPath, err := GetAbsolutePath(configPath, project.Path, project.Name)
	if err != nil {
		return "", &core.FailedToParsePath{projectPath}
	}
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return "", &core.PathDoesNotExist{projectPath}
	}

	defaultArguments := getDefaultArguments(configPath, project)

	// Execute Command
	shellProgram, commandStr := formatShellString(shell, command.Command)
	cmd := exec.Command(shellProgram, commandStr...)
	cmd.Dir = projectPath

	var output string = ""
	if dryRun {
		for _, arg := range defaultArguments {
			env := strings.SplitN(arg, "=", 2)
			os.Setenv(env[0], env[1])
		}

		for _, arg := range userArguments {
			env := strings.SplitN(arg, "=", 2)
			os.Setenv(env[0], env[1])
		}

		output = os.ExpandEnv(command.Command)
	} else {
		cmd.Env = append(os.Environ(), defaultArguments...)
		cmd.Env = append(cmd.Env, userArguments...)
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
		return "", &core.FailedToParsePath{ projectPath }
	}
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return "", &core.PathDoesNotExist{projectPath}
	}
	defaultArguments := getDefaultArguments(configPath, project)

	// Execute Command
	shellProgram, commandStr := formatShellString(shell, cmdString)
	cmd := exec.Command(shellProgram, commandStr...)
	cmd.Dir = projectPath

	var output string = ""
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
