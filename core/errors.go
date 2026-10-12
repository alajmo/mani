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
	return fmt.Sprintf("invalid output for spec `%s`, found `%s`, expected one of: stream, table, html, markdown", c.Name, c.Output)
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

type DefinedTwice struct {
	Kind  string // project or task
	Name  string
	Other string // Location of the first definition
}

func (c *DefinedTwice) Error() string {
	return fmt.Sprintf("%s `%s` is defined more than once in the same file, first definition is at %s", c.Kind, c.Name, c.Other)
}

type SetTwice struct {
	Field      string // env, remote or worktree
	Name       string
	Value      string
	OtherValue string
	OtherLine  int // Line of the first value
	Line       int // Line of the second value
}

func (c *SetTwice) ConfigLine() int {
	return c.Line
}

func (c *SetTwice) Error() string {
	return fmt.Sprintf("%s `%s` is set more than once with different values: `%s` here and `%s` at line %d", c.Field, c.Name, c.Value, c.OtherValue, c.OtherLine)
}

type AnchorContainsItself struct {
	Anchor string
	Line   int // Line of the merge key (<<) that merges the anchor into itself
}

func (c *AnchorContainsItself) ConfigLine() int {
	return c.Line
}

func (c *AnchorContainsItself) Error() string {
	return fmt.Sprintf("anchor `%s` value contains itself", c.Anchor)
}

type Conflict struct {
	Kind       string // project or task, empty for the env section of a config
	Name       string
	Field      string
	Value      string
	OtherValue string
	Other      string // Location of the other definition
}

func (c *Conflict) Error() string {
	subject := "conflicting"
	if c.Kind != "" {
		subject = fmt.Sprintf("%s `%s` has conflicting", c.Kind, c.Name)
	}
	return fmt.Sprintf("%s %s: `%s` here and `%s` in %s, override it in a config that imports both", subject, c.Field, c.Value, c.OtherValue, c.Other)
}

type ProjectPathAmbiguous struct {
	Name      string
	Path      string
	OtherPath string
	Other     string // Location of the other definition
}

func (c *ProjectPathAmbiguous) Error() string {
	return fmt.Sprintf("project `%s` resolves to `%s` here and `%s` in %s, set `path` in one of the definitions", c.Name, c.Path, c.OtherPath, c.Other)
}

type InvalidEnv struct {
	Name string // Env whose value isn't a single value, empty if env itself isn't a mapping
	Line int    // Line of the invalid value in the config
}

func (c *InvalidEnv) ConfigLine() int {
	return c.Line
}

func (c *InvalidEnv) Error() string {
	if c.Name == "" {
		return "`env` must be a mapping of names to values, for instance `env: {NAME: value}`"
	}
	return fmt.Sprintf("env `%s` must be a single value", c.Name)
}

type InvalidRemotes struct {
	Name string // Remote whose url isn't a single value, empty if remotes itself isn't a mapping
	Line int    // Line of the invalid value in the config
}

func (c *InvalidRemotes) ConfigLine() int {
	return c.Line
}

func (c *InvalidRemotes) Error() string {
	if c.Name == "" {
		return "`remotes` must be a mapping of names to urls, for instance `remotes: {upstream: git@github.com:owner/repo.git}`"
	}
	return fmt.Sprintf("remote `%s` must be a single url", c.Name)
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
