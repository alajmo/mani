package dao

import (
	"fmt"
	"io"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	// "golang.org/x/crypto/ssh"
	// "golang.org/x/crypto/ssh/agent"

	"gopkg.in/yaml.v3"
	"github.com/theckman/yacspin"
	"github.com/melbahja/goph"

	core "github.com/alajmo/mani/core"
)

var (
	build_mode = "dev"
)

type CommandInterface interface {
	RunRemoteCmd() (string, error)
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
	Ref			string			  `yaml:"ref"`
}

type Command struct {
	CommandBase `yaml:",inline"`
}

type Task struct {
	Output			string

	Projects		[]string
	ProjectPaths	[]string

	Dirs			[]string
	DirPaths		[]string

	Tags			[]string

	Abort			bool
	Commands		[]Command
	CommandBase		`yaml:",inline"`
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

func (c *CommandBase) SetEnvList(userEnv []string, parentEnv []string, configEnv []string) {
	pEnv, err := core.EvaluateEnv(parentEnv)
	core.CheckIfError(err)

	cmdEnv, err := core.EvaluateEnv(c.GetEnv())
	core.CheckIfError(err)

	globalEnv, err := core.EvaluateEnv(configEnv)
	core.CheckIfError(err)

	envList := core.MergeEnv(userEnv, cmdEnv, pEnv, globalEnv)

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

func getDefaultArguments(configPath string, entity Entity) []string {
	// Default arguments
	maniConfigPath := fmt.Sprintf("MANI_CONFIG_PATH=%s", configPath)
	maniConfigDir := fmt.Sprintf("MANI_CONFIG_DIR=%s", filepath.Dir(configPath))
	projectNameEnv := fmt.Sprintf("MANI_PROJECT_NAME=%s", entity.Name)
	projectPathEnv := fmt.Sprintf("MANI_PROJECT_PATH=%s", entity.Path)

	defaultArguments := []string {maniConfigPath, maniConfigDir, projectNameEnv, projectPathEnv}

	return defaultArguments
}

func formatShellString(shell string, command string) (string, []string) {
	shellProgram := strings.SplitN(shell, " ", 2)
	return shellProgram[0], append(shellProgram[1:], command)
}

func TaskSpinner() (yacspin.Spinner, error) {
	var cfg yacspin.Config

	// NOTE: Don't print the spinner in tests since it causes
	// golden files to produce different results.
	if build_mode == "TEST" {
		cfg = yacspin.Config {
			Frequency:       100 * time.Millisecond,
			CharSet:         yacspin.CharSets[9],
			SuffixAutoColon: false,
			Writer: io.Discard,
		}
	} else {
		cfg = yacspin.Config {
			Frequency:       100 * time.Millisecond,
			CharSet:         yacspin.CharSets[9],
			SuffixAutoColon: false,
		}
	}

	spinner, err := yacspin.New(cfg)

	return *spinner, err
}


func (c CommandBase) RunCmd(
	config Config,
	shell string,
	entity Entity,
	dryRun bool,
) (string, error) {
	entityPath, err := core.GetAbsolutePath(config.Path, entity.Path, entity.Name)
	if err != nil {
		return "", &core.FailedToParsePath{ Name: entityPath }
	}
	if _, err := os.Stat(entityPath); os.IsNotExist(err) {
		return "", &core.PathDoesNotExist{ Path: entityPath }
	}

	defaultArguments := getDefaultArguments(config.Path, entity)

	var shellProgram string
	var commandStr []string

	if c.Ref != "" {
		refTask, err := config.GetTask(c.Ref)
		if err != nil {
			return "", err
		}

		shellProgram, commandStr = formatShellString(refTask.Shell, refTask.Command)
	} else {
		shellProgram, commandStr = formatShellString(shell, c.Command)
	}

	// Execute Command
	cmd := exec.Command(shellProgram, commandStr...)
	cmd.Dir = entityPath

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
		return "", &core.FailedToParsePath{ Name: projectPath }
	}
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return "", &core.PathDoesNotExist{ Path: projectPath }
	}
	// TODO: FIX THIS
	// defaultArguments := getDefaultArguments(configPath, project)

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

func (c CommandBase) RunRemoteCmd(
	config Config,
	shell string,
	entity Entity,
	dryRun bool,
) (string, error) {
	auth, err := goph.UseAgent()
	core.CheckIfError(err)

	client, err := goph.New("samir", "192.168.0.107", auth)
	core.CheckIfError(err)

	defer client.Close()

	entityPath, err := core.GetAbsolutePath(config.Path, entity.Path, entity.Name)
	if err != nil {
		return "", &core.FailedToParsePath{ Name: entityPath }
	}
	if _, err := os.Stat(entityPath); os.IsNotExist(err) {
		return "", &core.PathDoesNotExist{ Path: entityPath }
	}

	defaultArguments := getDefaultArguments(config.Path, entity)

	var shellProgram string
	var commandStr []string

	if c.Ref != "" {
		refTask, err := config.GetTask(c.Ref)
		if err != nil {
			return "", err
		}

		shellProgram, commandStr = formatShellString(refTask.Shell, refTask.Command)
	} else {
		shellProgram, commandStr = formatShellString(shell, c.Command)
	}

	// Execute Command
	cmd, err := client.Command(shellProgram, commandStr...)
	core.CheckIfError(err)

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
