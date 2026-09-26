package print

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/alajmo/mani/core/dao"
)

var FORMATTER Formatter
var COLORIZE bool
var BLOCK dao.Block

func PrintProjectBlocks(projects []dao.Project, colorize bool, block dao.Block, f Formatter) string {
	if len(projects) == 0 {
		return ""
	}

	FORMATTER = f
	COLORIZE = colorize
	BLOCK = block

	var output strings.Builder
	output.WriteString("\n")

	for i, project := range projects {
		output.WriteString(printKeyValue(false, "", "name", ":", project.Name, *block.Key, *block.Value))
		output.WriteString(printKeyValue(false, "", "sync", ":", strconv.FormatBool(project.IsSync()), *block.Key, trueOrFalse(project.IsSync())))
		if project.Desc != "" {
			output.WriteString(printKeyValue(false, "", "description", ":", project.Desc, *block.Key, *block.Value))
		}
		if project.RelPath != project.Name {
			output.WriteString(printKeyValue(false, "", "path", ":", project.RelPath, *block.Key, *block.Value))
		}

		output.WriteString(printKeyValue(false, "", "url", ":", project.URL, *block.Key, *block.Value))

		if len(project.RemoteList) > 0 {
			output.WriteString(printKeyValue(false, "", "remotes", ":", "", *block.Key, *block.Value))
			for _, remote := range project.RemoteList {
				output.WriteString(printKeyValue(true, "", remote.Name, ":", remote.URL, *block.Key, *block.Value))
			}
		}

		if len(project.WorktreeList) > 0 {
			output.WriteString(printKeyValue(false, "", "worktrees", ":", "", *block.Key, *block.Value))
			for _, wt := range project.WorktreeList {
				output.WriteString(printKeyValue(true, "- ", "path", ":", wt.Path, *block.Key, *block.Value))
				output.WriteString(printKeyValue(true, "  ", "branch", ":", wt.Branch, *block.Key, *block.Value))
			}
		}

		if project.Branch != "" {
			output.WriteString(printKeyValue(false, "", "branch", ":", project.Branch, *block.Key, *block.Value))
		}

		output.WriteString(printKeyValue(false, "", "single_branch", ":", strconv.FormatBool(project.IsSingleBranch()), *block.Key, trueOrFalse(project.IsSingleBranch())))

		if len(project.Tags) > 0 {
			output.WriteString(printKeyValue(false, "", "tags", ":", project.GetValue("tag", 0), *block.Key, *block.Value))
		}

		if len(project.EnvList) > 0 {
			output.WriteString(printEnv(project.EnvList, block))
		}

		if i < len(projects)-1 {
			output.WriteString("\n--\n\n")
		}
	}

	output.WriteString("\n")

	return output.String()
}

func PrintTaskBlock(tasks []dao.Task, colorize bool, block dao.Block, f Formatter) string {
	if len(tasks) == 0 {
		return ""
	}
	FORMATTER = f
	COLORIZE = colorize
	BLOCK = block

	var output strings.Builder
	output.WriteString("\n")

	for i, task := range tasks {
		output.WriteString(printKeyValue(false, "", "name", ":", task.Name, *block.Key, *block.Value))
		output.WriteString(printKeyValue(false, "", "description", ":", task.Desc, *block.Key, *block.Value))
		output.WriteString(printKeyValue(false, "", "theme", ":", task.ThemeData.Name, *block.Key, *block.Value))
		output.WriteString(printKeyValue(false, "", "target", ":", "", *block.Key, *block.Value))
		output.WriteString(printKeyValue(true, "", "all", ":", strconv.FormatBool(task.TargetData.All), *block.Key, trueOrFalse(task.TargetData.All)))
		output.WriteString(printKeyValue(true, "", "cwd", ":", strconv.FormatBool(task.TargetData.Cwd), *block.Key, trueOrFalse(task.TargetData.Cwd)))
		output.WriteString(printKeyValue(true, "", "projects", ":", strings.Join(task.TargetData.Projects, ", "), *block.Key, *block.Value))
		output.WriteString(printKeyValue(true, "", "paths", ":", strings.Join(task.TargetData.Paths, ", "), *block.Key, *block.Value))
		output.WriteString(printKeyValue(true, "", "tags", ":", strings.Join(task.TargetData.Tags, ", "), *block.Key, *block.Value))
		output.WriteString(printKeyValue(true, "", "tags_expr", ":", task.TargetData.TagsExpr, *block.Key, *block.Value))

		output.WriteString(printKeyValue(false, "", "spec", ":", "", *block.Key, *block.Value))
		output.WriteString(printKeyValue(true, "", "output", ":", task.SpecData.Output, *block.Key, *block.Value))
		output.WriteString(printKeyValue(true, "", "parallel", ":", strconv.FormatBool(task.SpecData.Parallel), *block.Key, trueOrFalse(task.SpecData.Parallel)))
		output.WriteString(printKeyValue(true, "", "ignore_errors", ":", strconv.FormatBool(task.SpecData.IgnoreErrors), *block.Key, trueOrFalse(task.SpecData.IgnoreErrors)))
		output.WriteString(printKeyValue(true, "", "omit_empty_rows", ":", strconv.FormatBool(task.SpecData.OmitEmptyRows), *block.Key, trueOrFalse(task.SpecData.OmitEmptyRows)))
		output.WriteString(printKeyValue(true, "", "omit_empty_columns", ":", strconv.FormatBool(task.SpecData.OmitEmptyColumns), *block.Key, trueOrFalse(task.SpecData.OmitEmptyColumns)))

		if len(task.EnvList) > 0 {
			output.WriteString(printEnv(task.EnvList, block))
		}

		if task.Cmd != "" {
			output.WriteString(printKeyValue(false, "", "cmd", ":", "", *block.Key, *block.Value))
			output.WriteString(printCmd(task.Cmd))
		}

		if len(task.Commands) > 0 {
			output.WriteString(printKeyValue(false, "", "commands", ":", "", *block.Key, *block.Value))
			for _, subCommand := range task.Commands {
				if subCommand.Name != "" {
					if subCommand.Desc != "" {
						output.WriteString(printKeyValue(true, "- ", subCommand.Name, ":", subCommand.Desc, *block.Key, *block.Value))
					} else {
						output.WriteString(printKeyValue(true, "- ", subCommand.Name, "", "", *block.Key, *block.Value))
					}
				} else {
					output.WriteString(printKeyValue(true, "- ", "cmd", "", "", *block.Value, *block.Value))
				}
			}
		}

		if i < len(tasks)-1 {
			output.WriteString("\n--\n\n")
		}
	}
	output.WriteString("\n")

	return output.String()
}

type Formatter interface {
	Format(prefix string, key string, value string, separator string, keyColor *dao.ColorOptions, valueColor *dao.ColorOptions) string
}

func printKeyValue(
	padding bool,
	prefix string,
	key string,
	separator string,
	value string,
	keyStyle dao.ColorOptions,
	valueStyle dao.ColorOptions,
) string {
	if !COLORIZE {
		str := fmt.Sprintf("%s%s%s %s\n", prefix, key, separator, value)
		if padding {
			return fmt.Sprintf("%4s%s", " ", str)
		}
		return str
	}

	str := FORMATTER.Format(prefix, key, value, separator, &keyStyle, &valueStyle)
	str += "\n"

	if padding {
		str = fmt.Sprintf("%4s%s", " ", str)
	}

	return str
}

func printCmd(cmd string) string {
	var output strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(cmd))
	for scanner.Scan() {
		fmt.Fprintf(&output, "%4s%s\n", " ", scanner.Text())
	}

	return output.String()
}

func printEnv(env []string, block dao.Block) string {
	var output strings.Builder

	output.WriteString(printKeyValue(false, "", "env", ":", "", *block.Key, *block.Value))

	for _, env := range env {
		parts := strings.SplitN(strings.TrimSuffix(env, "\n"), "=", 2)
		output.WriteString(printKeyValue(true, "", parts[0], ":", parts[1], *block.Key, *block.Value))
	}

	return output.String()
}

func trueOrFalse(value bool) dao.ColorOptions {
	if value {
		return *BLOCK.ValueTrue
	}
	return *BLOCK.ValueFalse
}

type TviewFormatter struct{}
type GookitFormatter struct{}

func (t TviewFormatter) Format(
	prefix string,
	key string,
	value string,
	separator string,
	keyColor *dao.ColorOptions,
	valueColor *dao.ColorOptions,
) string {
	sepStr := fmt.Sprintf("[%s:-:%s]%s", *BLOCK.Separator.Fg, *BLOCK.Separator.Attr, separator)
	return fmt.Sprintf(
		"[%s:-:%s]%s%s[-::-]%s[-:-:-] [%s:-:%s]%s",
		*keyColor.Fg, *keyColor.Attr, prefix, key, sepStr, *valueColor.Fg, *valueColor.Attr, value,
	)
}

func (g GookitFormatter) Format(
	prefix string,
	key string,
	value string,
	separator string,
	keyColor *dao.ColorOptions,
	valueColor *dao.ColorOptions,
) string {
	prefixStr := dao.StyleString(prefix, *keyColor, true)
	keyStr := dao.StyleString(key, *keyColor, true)
	sepStr := dao.StyleString(separator, *BLOCK.Separator, true)
	valueStr := dao.StyleString(value, *valueColor, true)

	return fmt.Sprintf("%s%s%s %s", prefixStr, keyStr, sepStr, valueStr)
}
