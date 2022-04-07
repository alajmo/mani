package dao

import (
	"os"
	"os/exec"
	"strings"

	core "github.com/alajmo/mani/core"
)

func RunTable(
	config Config,
	cmdStr string,
	envList []string,
	shell string,
	project Project,
	dryRun bool,
) (string, error) {
	projectPath, err := core.GetAbsolutePath(config.Path, project.Path, project.Name)
	if err != nil {
		return "", &core.FailedToParsePath{Name: projectPath}
	}
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return "", &core.PathDoesNotExist{Path: projectPath}
	}

	defaultArguments := getDefaultArguments(config.Path, config.Dir, project)
	shellProgram, commandStr := formatShellString(shell, cmdStr)

	// Execute Command
	cmd := exec.Command(shellProgram, commandStr...)
	cmd.Dir = projectPath

	envs := core.MergeEnv(envList, project.EnvList, defaultArguments, []string{})

	var output string
	if dryRun {
		for _, arg := range envs {
			env := strings.SplitN(arg, "=", 2)
			os.Setenv(env[0], env[1])
		}

		output = os.ExpandEnv(cmdStr)
	} else {
		cmd.Env = append(os.Environ(), envs...)
		output, err := cmd.CombinedOutput()

		return string(output), err
	}

	return output, nil
}
