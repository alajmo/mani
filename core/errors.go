package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/gookit/color"
)

type ConfigEnvFailed struct {
	Name string
	Err  string
}

func (c *ConfigEnvFailed) Error() string {
	return fmt.Sprintf("failed to evaluate env `%s` \n  %s", c.Name, c.Err)
}

type AlreadyManiDirectory struct {
	Dir string
}

func (c *AlreadyManiDirectory) Error() string {
	return fmt.Sprintf("`%s` is already a mani directory\n", c.Dir)
}

type ZeroNotAllowed struct {
	Name string
}

func (c *ZeroNotAllowed) Error() string {
	return fmt.Sprintf("invalid value for %s, cannot be 0", c.Name)
}

type FailedToOpenFile struct {
	Name string
}

func (f *FailedToOpenFile) Error() string {
	return fmt.Sprintf("failed to open `%s`", f.Name)
}

type FailedToParsePath struct {
	Name string
}

func (f *FailedToParsePath) Error() string {
	return fmt.Sprintf("failed to parse path `%s`", f.Name)
}

type PathDoesNotExist struct {
	Path string
}

func (p *PathDoesNotExist) Error() string {
	return fmt.Sprintf("path `%s` does not exist", p.Path)
}

type TagNotFound struct {
	Tags []string
}

func (c *TagNotFound) Error() string {
	tags := "`" + strings.Join(c.Tags, "`, `") + "`"
	return fmt.Sprintf("cannot find tags %s", tags)
}

type DirNotFound struct {
	Dirs []string
}

func (c *DirNotFound) Error() string {
	dirs := "`" + strings.Join(c.Dirs, "`, `") + "`"
	return fmt.Sprintf("cannot find paths %s", dirs)
}

type NoTargets struct{}

func (c *NoTargets) Error() string {
	return "no matching projects found"
}

type ProjectNotFound struct {
	Name []string
}

func (c *ProjectNotFound) Error() string {
	projects := "`" + strings.Join(c.Name, "`, `") + "`"
	return fmt.Sprintf("cannot find projects %s", projects)
}

type TaskNotFound struct {
	Name []string
}

func (c *TaskNotFound) Error() string {
	tasks := "`" + strings.Join(c.Name, "`, `") + "`"
	return fmt.Sprintf("cannot find tasks %s", tasks)
}

type ThemeNotFound struct {
	Name string
}

func (c *ThemeNotFound) Error() string {
	return fmt.Sprintf("cannot find theme `%s`", c.Name)
}

type SpecNotFound struct {
	Name string
}

func (c *SpecNotFound) Error() string {
	return fmt.Sprintf("cannot find spec `%s`", c.Name)
}

type SpecOutputError struct {
	Name   string
	Output string
}

func (c *SpecOutputError) Error() string {
	return fmt.Sprintf("invalid output for spec `%s`, found `%s`, expected one of: stream, table, html, markdown, json, yaml", c.Name, c.Output)
}

type TargetNotFound struct {
	Name string
}

func (c *TargetNotFound) Error() string {
	return fmt.Sprintf("cannot find target `%s`", c.Name)
}

type TargetTagsExprError struct {
	Name string
	Err  error
}

func (c *TargetTagsExprError) Error() string {
	return fmt.Sprintf("invalid tags_expr for target `%s`, %s", c.Name, c.Err.Error())
}

type TagExprInvalid struct {
	Expression string
}

func (c *TagExprInvalid) Error() string {
	return fmt.Sprintf("invalid tags expression: %s", c.Expression)
}

type ConfigNotFound struct {
	Names []string
}

func (f *ConfigNotFound) Error() string {
	return fmt.Sprintf("cannot find any configuration file %v in current directory or any of the parent directories", f.Names)
}

type WorktreePathRequired struct{}

func (c *WorktreePathRequired) Error() string {
	return "worktree path is required"
}

type FailedToCreateWorktree struct {
	Path   string
	Output string
	Err    error
}

func (c *FailedToCreateWorktree) Error() string {
	return fmt.Sprintf("failed to create worktree `%s`: %s - %s", c.Path, c.Err, c.Output)
}

type FailedToRemoveWorktree struct {
	Path   string
	Output string
	Err    error
}

func (c *FailedToRemoveWorktree) Error() string {
	return fmt.Sprintf("failed to remove worktree `%s`: %s - %s", c.Path, c.Err, c.Output)
}

type ConfigErr struct {
	Msg string
}

func (f *ConfigErr) Error() string {
	return f.Msg
}

func CheckIfError(err error) {
	if err != nil {
		switch err.(type) {
		case *ConfigErr:
			// Errors are already mapped with `error:` prefix
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		default:
			fmt.Fprintf(os.Stderr, "%s: %v\n", color.FgRed.Sprintf("error"), err)
			os.Exit(1)
		}
	}
}

func Exit(err error) {
	switch err := err.(type) {
	case *ConfigErr:
		// Errors are already mapped with `error:` prefix
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "%s: %v\n", color.FgRed.Sprintf("error"), err)
		os.Exit(1)
	}
}
