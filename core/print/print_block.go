package print

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/alajmo/mani/core/dao"
)

var PURPLE = "#9068BD"
var GREEN = "#b5bd68"
var BLUE = "#94B8D8"
var RED = "#cc6666"
var WHITE = "#E0E0E0"

var KEY_COLOR = BLUE
var VALUE_COLOR = WHITE
var TRUE_COLOR = GREEN
var FALSE_COLOR = RED
var CMD_COLOR = PURPLE
var ENV_COLOR = PURPLE

func PrintProjectBlocks(projects []dao.Project, colorize bool) string {
	if len(projects) == 0 {
		return ""
	}

	output := ""

	output += fmt.Sprintln()

	for i, project := range projects {
		output += printKeyValue("", "Name", project.Name, ":", colorize, KEY_COLOR, false)
		if project.Desc != "" {
			output += printKeyValue("", "Description", project.Desc, ":", colorize, KEY_COLOR, false)
		}
		if project.RelPath != project.Name {
			output += printKeyValue("", "Path", project.RelPath, ":", colorize, KEY_COLOR, false)
		}
		output += printKeyValue("", "Url", project.Url, ":", colorize, KEY_COLOR, false)
		output += printKeyValue("", "Sync", strconv.FormatBool(project.IsSync()), ":", colorize, KEY_COLOR, false)

		if len(project.Tags) > 0 {
			output += printKeyValue("", "Tags", project.GetValue("Tag", 0), ":", colorize, KEY_COLOR, false)
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
		output += printKeyValue("", "Name", task.Name, ":", colorize, KEY_COLOR, false)
		output += printKeyValue("", "Description", task.Desc, ":", colorize, KEY_COLOR, false)
		output += printKeyValue("", "Theme", task.ThemeData.Name, ":", colorize, KEY_COLOR, false)
		output += printKeyValue("", "Target", "", ":", colorize, KEY_COLOR, false)
		output += printKeyValue("", "All", strconv.FormatBool(task.TargetData.All), ":", colorize, KEY_COLOR, true)
		output += printKeyValue("", "Cwd", strconv.FormatBool(task.TargetData.Cwd), ":", colorize, KEY_COLOR, true)
		output += printKeyValue("", "Projects", strings.Join(task.TargetData.Projects, ", "), ":", colorize, KEY_COLOR, true)
		output += printKeyValue("", "Paths", strings.Join(task.TargetData.Paths, ", "), ":", colorize, KEY_COLOR, true)
		output += printKeyValue("", "Tags", strings.Join(task.TargetData.Tags, ", "), ":", colorize, KEY_COLOR, true)

		output += printKeyValue("", "Spec", "", ":", colorize, KEY_COLOR, false)
		output += printKeyValue("", "Output", task.SpecData.Output, ":", colorize, KEY_COLOR, true)
		output += printKeyValue("", "Parallel", strconv.FormatBool(task.SpecData.Parallel), ":", colorize, KEY_COLOR, true)
		output += printKeyValue("", "IgnoreErrors", strconv.FormatBool(task.SpecData.IgnoreErrors), ":", colorize, KEY_COLOR, true)
		output += printKeyValue("", "OmitEmpty", strconv.FormatBool(task.SpecData.OmitEmpty), ":", colorize, KEY_COLOR, true)

		if len(task.EnvList) > 0 {
			output += printEnv(task.EnvList, colorize)
		}

		if task.Cmd != "" {
			output += printKeyValue("", "Cmd", "", ":", colorize, KEY_COLOR, false)
			output += printCmd(task.Cmd)
		}

		if len(task.Commands) > 0 {
			output += printKeyValue("", "Commands", "", ":", colorize, KEY_COLOR, false)
			for _, subCommand := range task.Commands {
				if subCommand.Name != "" {
					if subCommand.Desc != "" {
						output += printKeyValue("- ", subCommand.Name, subCommand.Desc, ":", colorize, CMD_COLOR, true)
					} else {
						output += printKeyValue("- ", subCommand.Name, "", "", colorize, CMD_COLOR, true)
					}
				} else {
					output += printKeyValue("- ", "cmd", "", "", colorize, CMD_COLOR, true)
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
	valueColor := VALUE_COLOR
	if value == "true" {
		valueColor = TRUE_COLOR
	} else if value == "false" {
		valueColor = FALSE_COLOR
	}

	if colorize {
		str = fmt.Sprintf("%s[%s:b]%s[-:-]%s [%s::-]%s\n", prefix, color, key, seperator, valueColor, value)
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

	output += printKeyValue("", "Env", "", ":", colorize, KEY_COLOR, false)

	for _, env := range env {
		parts := strings.SplitN(strings.TrimSuffix(env, "\n"), "=", 2)
		output += printKeyValue("", parts[0], parts[1], ":", colorize, ENV_COLOR, true)
	}

	return output
}
