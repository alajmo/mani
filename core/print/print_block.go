package print

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/alajmo/mani/core/dao"
)

func PrintProjectBlocks(projects []dao.Project) string {
	if len(projects) == 0 {
		return ""
	}

	output := ""

	output += fmt.Sprintln()

	for i, project := range projects {
		output += fmt.Sprintf("Name: %s\n", project.Name)
		output += fmt.Sprintf("Path: %s\n", project.RelPath)
		output += fmt.Sprintf("Description: %s\n", project.Desc)
		output += fmt.Sprintf("Url: %s\n", project.Url)
		output += fmt.Sprintf("Sync: %t\n", project.IsSync())

		if len(project.Tags) > 0 {
			output += fmt.Sprintf("Tags: %s\n", project.GetValue("Tag", 0))
		}

		if len(project.EnvList) > 0 {
			output += PrintEnv(project.EnvList)
		}

		if i < len(projects)-1 {
			output += fmt.Sprintf("\n--\n\n")
		}
	}

	output += fmt.Sprintln()

	return output
}

func PrintTaskBlock(tasks []dao.Task) string {
	if len(tasks) == 0 {
		return ""
	}

	output := ""
	output += fmt.Sprintln()

	for i, task := range tasks {
		output += fmt.Sprintf("Name: %s\n", task.Name)
		output += fmt.Sprintf("Description: %s\n", task.Desc)
		output += fmt.Sprintf("Theme: %s\n", task.ThemeData.Name)
		output += fmt.Sprintf("Target: \n")
		output += fmt.Sprintf("%4sAll: %t\n", " ", task.TargetData.All)
		output += fmt.Sprintf("%4sCwd: %t\n", " ", task.TargetData.Cwd)
		output += fmt.Sprintf("%4sProjects: %s\n", " ", strings.Join(task.TargetData.Projects, ", "))
		output += fmt.Sprintf("%4sPaths: %s\n", " ", strings.Join(task.TargetData.Paths, ", "))
		output += fmt.Sprintf("%4sTags: %s", " ", strings.Join(task.TargetData.Tags, ", "))

		output += fmt.Sprintln("")

		output += fmt.Sprintf("Spec: \n")
		output += fmt.Sprintf("%4sOutput: %s\n", "", task.SpecData.Output)
		output += fmt.Sprintf("%4sParallel: %t\n", "", task.SpecData.Parallel)
		output += fmt.Sprintf("%4sIgnoreErrors: %t\n", "", task.SpecData.IgnoreErrors)
		output += fmt.Sprintf("%4sOmitEmpty: %t", "", task.SpecData.OmitEmpty)

		output += fmt.Sprintln("")

		if len(task.EnvList) > 0 {
			output += PrintEnv(task.EnvList)
		}

		if task.Cmd != "" {
			output += fmt.Sprintf("Cmd: \n")
			output += PrintCmd(task.Cmd)
		}

		if len(task.Commands) > 0 {
			output += fmt.Sprintf("Commands: \n")
			for _, subCommand := range task.Commands {
				if subCommand.Name != "" {
					if subCommand.Desc != "" {
						output += fmt.Sprintf("%4s - %s: %s\n", " ", subCommand.Name, subCommand.Desc)
					} else {
						output += fmt.Sprintf("%4s - %s\n", " ", subCommand.Name)
					}
				} else {
					output += fmt.Sprintf("%4s - %s\n", " ", "cmd")
				}
			}
		}

		if i < len(tasks)-1 {
			output += fmt.Sprintf("\n--\n\n")
		}
	}
	output += fmt.Sprintln()

	return output
}

func PrintCmd(cmd string) string {
	output := ""
	scanner := bufio.NewScanner(strings.NewReader(cmd))
	for scanner.Scan() {
		output += fmt.Sprintf("%4s%s\n", " ", scanner.Text())
	}

	return output
}

func PrintEnv(env []string) string {
	output := ""
	output += fmt.Sprintf("Env: \n")
	for _, env := range env {
		output += fmt.Sprintf("%4s%s\n", " ", strings.Replace(strings.TrimSuffix(env, "\n"), "=", ": ", 1))
	}

	return output
}
