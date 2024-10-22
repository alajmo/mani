package print

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/alajmo/mani/core/dao"
)

var GREEN = "#b5bd68"
var BLUE = "#94B8D8"
var RED = "#cc6666"

func PrintProjectBlocks(projects []dao.Project, colorize bool) string {
	if len(projects) == 0 {
		return ""
	}

	output := ""

	output += fmt.Sprintln()

	for i, project := range projects {
		output += printKeyValue("", "Name", project.Name, ":", colorize, GREEN, false)
		if project.Desc != "" {
			output += printKeyValue("", "Description", project.Desc, ":", colorize, GREEN, false)
		}
		if project.RelPath != project.Name {
			output += printKeyValue("", "Path", project.RelPath, ":", colorize, GREEN, false)
		}
		output += printKeyValue("", "Url", project.Url, ":", colorize, GREEN, false)
		output += printKeyValue("", "Sync", strconv.FormatBool(project.IsSync()), ":", colorize, GREEN, false)

		if len(project.Tags) > 0 {
			output += printKeyValue("", "Tags", project.GetValue("Tag", 0), ":", colorize, GREEN, false)
		}

		if len(project.EnvList) > 0 {
			output += printEnv(project.EnvList, colorize)
		}

		if i < len(projects)-1 {
			output += fmt.Sprintf("\n--\n\n")
		}
	}

	output += fmt.Sprintln()

	return output
}

func PrintTaskBlock(tasks []dao.Task, colorize bool) string {
	if len(tasks) == 0 {
		return ""
	}

	output := ""
	output += fmt.Sprintln()

	for i, task := range tasks {
		output += printKeyValue("", "Name", task.Name, ":", colorize, GREEN, false)
		output += printKeyValue("", "Description", task.Desc, ":", colorize, GREEN, false)
		output += printKeyValue("", "Theme", task.ThemeData.Name, ":", colorize, GREEN, false)
		output += printKeyValue("", "Target", "", ":", colorize, GREEN, false)
		output += printKeyValue("", "All", strconv.FormatBool(task.TargetData.All), ":", colorize, GREEN, true)
		output += printKeyValue("", "Cwd", strconv.FormatBool(task.TargetData.Cwd), ":", colorize, GREEN, true)
		output += printKeyValue("", "Projects", strings.Join(task.TargetData.Projects, ", "), ":", colorize, GREEN, true)
		output += printKeyValue("", "Paths", strings.Join(task.TargetData.Paths, ", "), ":", colorize, GREEN, true)
		output += printKeyValue("", "Tags", strings.Join(task.TargetData.Tags, ", "), ":", colorize, GREEN, true)

		output += printKeyValue("", "Spec", "", ":", colorize, GREEN, false)
		output += printKeyValue("", "Output", task.SpecData.Output, ":", colorize, GREEN, true)
		output += printKeyValue("", "Parallel", strconv.FormatBool(task.SpecData.Parallel), ":", colorize, GREEN, true)
		output += printKeyValue("", "IgnoreErrors", strconv.FormatBool(task.SpecData.IgnoreErrors), ":", colorize, GREEN, true)
		output += printKeyValue("", "OmitEmpty", strconv.FormatBool(task.SpecData.OmitEmpty), ":", colorize, GREEN, true)

		if len(task.EnvList) > 0 {
			output += printEnv(task.EnvList, colorize)
		}

		if task.Cmd != "" {
			output += printKeyValue("", "Cmd", "", ":", colorize, GREEN, false)
			output += printCmd(task.Cmd)
		}

		if len(task.Commands) > 0 {
			output += printKeyValue("", "Commands", "", ":", colorize, GREEN, false)
			for _, subCommand := range task.Commands {
				if subCommand.Name != "" {
					if subCommand.Desc != "" {
						output += printKeyValue("- ", subCommand.Name, subCommand.Desc, ":", colorize, BLUE, true)
					} else {
						output += printKeyValue("- ", subCommand.Name, "", "", colorize, BLUE, true)
					}
				} else {
					output += printKeyValue("- ", "cmd", "", "", colorize, BLUE, true)
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

func printKeyValue(prefix string, key string, value string, seperator string, colorize bool, color string, padding bool) string {
	str := ""
	valueColor := "-"
	if value == "true" {
		valueColor = GREEN
	} else if value == "false" {
		valueColor = RED
	}

	if colorize {
		str = fmt.Sprintf("%s[%s:b]%s[-::-]%s [%s::-]%s\n", prefix, "white", key, seperator, valueColor, value)
	} else {
		str = fmt.Sprintf("%s%s: %s\n", prefix, key, seperator, value)
	}

	if padding {
		str = fmt.Sprintf("%4s%s", " ", str)
	}

	return str
}

func printCmd(cmd string) string {
	output := ""
	scanner := bufio.NewScanner(strings.NewReader(cmd))
	for scanner.Scan() {
		output += fmt.Sprintf("%4s%s\n", " ", scanner.Text())
	}

	return output
}

func printEnv(env []string, colorize bool) string {
	output := ""

	output += printKeyValue("", "Env", "", ":", colorize, GREEN, false)

	for _, env := range env {
		parts := strings.SplitN(strings.TrimSuffix(env, "\n"), "=", 2)
		output += printKeyValue("", parts[0], parts[1], ":", colorize, BLUE, true)
	}

	return output
}
