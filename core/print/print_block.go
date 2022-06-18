package print

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/alajmo/mani/core/dao"
)

func PrintProjectBlocks(projects []dao.Project) {
	if len(projects) == 0 {
		return
	}

	fmt.Println()

	for i, project := range projects {
		fmt.Printf("Name: %s\n", project.Name)
		fmt.Printf("Path: %s\n", project.RelPath)
		fmt.Printf("Description: %s\n", project.Desc)
		fmt.Printf("Url: %s\n", project.Url)
		fmt.Printf("Sync: %t\n", project.IsSync())

		if len(project.Tags) > 0 {
			fmt.Printf("Tags: %s\n", project.GetValue("Tag", 0))
		}

		if len(project.EnvList) > 0 {
			printEnv(project.EnvList)
		}

		if i < len(projects)-1 {
			fmt.Printf("\n--\n\n")
		}
	}

	fmt.Println()
}

func PrintTaskBlock(tasks []dao.Task) {
	if len(tasks) == 0 {
		return
	}

	fmt.Println()

	for i, task := range tasks {
		fmt.Printf("Name: %s\n", task.Name)
		fmt.Printf("Description: %s\n", task.Desc)
		fmt.Printf("Theme: %s\n", task.ThemeData.Name)
		fmt.Printf("Target: \n")
		fmt.Printf("%4sAll: %t\n", " ", task.TargetData.All)
		fmt.Printf("%4sCwd: %t\n", " ", task.TargetData.Cwd)
		fmt.Printf("%4sProjects: %s\n", " ", strings.Join(task.TargetData.Projects, ", "))
		fmt.Printf("%4sPaths: %s\n", " ", strings.Join(task.TargetData.Paths, ", "))
		fmt.Printf("%4sTags: %s", " ", strings.Join(task.TargetData.Tags, ", "))

		fmt.Println("")

		fmt.Printf("Spec: \n")
		fmt.Printf("%4sOutput: %s\n", "", task.SpecData.Output)
		fmt.Printf("%4sParallel: %t\n", "", task.SpecData.Parallel)
		fmt.Printf("%4sIgnoreErrors: %t\n", "", task.SpecData.IgnoreErrors)
		fmt.Printf("%4sOmitEmpty: %t", "", task.SpecData.OmitEmpty)

		fmt.Println("")

		if len(task.EnvList) > 0 {
			printEnv(task.EnvList)
		}

		if task.Cmd != "" {
			fmt.Printf("Cmd: \n")
			printCmd(task.Cmd)
		}

		if len(task.Commands) > 0 {
			fmt.Printf("Commands: \n")
			for _, subCommand := range task.Commands {
				if subCommand.Name != "" {
					if subCommand.Desc != "" {
						fmt.Printf("%4s - %s: %s\n", " ", subCommand.Name, subCommand.Desc)
					} else {
						fmt.Printf("%4s - %s\n", " ", subCommand.Name)
					}
				} else {
					fmt.Printf("%4s - %s\n", " ", "cmd")
				}
			}
		}

		if i < len(tasks)-1 {
			fmt.Printf("\n--\n\n")
		}
	}
	fmt.Println()
}

func printCmd(cmd string) {
	scanner := bufio.NewScanner(strings.NewReader(cmd))
	for scanner.Scan() {
		fmt.Printf("%4s%s\n", " ", scanner.Text())
	}
}

func printEnv(env []string) {
	fmt.Printf("Env: \n")
	for _, env := range env {
		fmt.Printf("%4s%s\n", " ", strings.Replace(strings.TrimSuffix(env, "\n"), "=", ": ", 1))
	}
}
