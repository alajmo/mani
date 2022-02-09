package core

import (
	"fmt"
	"os"
)

type ConfigEnvFailed struct {
	Name string
	Err  error
}

func (c *ConfigEnvFailed) Error() string {
	return fmt.Sprintf("error: failed to evaluate env %q \n %q ", c.Name, c.Err)
}

type Node struct {
	Path     string
	Imports  []string
	Children []*Node
	Visiting bool
	Visited  bool
}

type NodeLink struct {
	A Node
	B Node
}

type FoundCyclicDependency struct {
	Cycles []NodeLink
}

func (c *FoundCyclicDependency) Error() string {
	var msg string
	msg = "Found direct or indirect circular dependency between:\n"
	for i := range c.Cycles {
		msg += fmt.Sprintf(" %s\n %s\n", c.Cycles[i].A.Path, c.Cycles[i].B.Path)
	}

	return msg
}

type FailedToOpenFile struct {
	Name string
}

func (f *FailedToOpenFile) Error() string {
	return fmt.Sprintf("error: failed to open %q", f.Name)
}

type FailedToParsePath struct {
	Name string
}

func (f *FailedToParsePath) Error() string {
	return fmt.Sprintf("error: failed to parse path %q", f.Name)
}

type FailedToParseFile struct {
	Name string
	Msg  error
}

func (f *FailedToParseFile) Error() string {
	return fmt.Sprintf("error: failed to parse %q \n%s", f.Name, f.Msg)
}

type PathDoesNotExist struct {
	Path string
}

func (p *PathDoesNotExist) Error() string {
	return fmt.Sprintf("fatal: path %q does not exist", p.Path)
}

type ProjectNotFound struct {
	Name string
}

func (c *ProjectNotFound) Error() string {
	return fmt.Sprintf("fatal: could not find project %q", c.Name)
}

type TaskNotFound struct {
	Name string
}

func (c *TaskNotFound) Error() string {
	return fmt.Sprintf("fatal: could not find task %q", c.Name)
}

type ThemeNotFound struct {
	Name string
}

func (c *ThemeNotFound) Error() string {
	return fmt.Sprintf("fatal: could not find theme %q", c.Name)
}

type SpecNotFound struct {
	Name string
}

func (c *SpecNotFound) Error() string {
	return fmt.Sprintf("fatal: could not find spec %q", c.Name)
}

type TargetNotFound struct {
	Name string
}

func (c *TargetNotFound) Error() string {
	return fmt.Sprintf("fatal: could not find target %q", c.Name)
}

type ConfigNotFound struct {
	Names []string
}

func (f *ConfigNotFound) Error() string {
	return fmt.Sprintf("fatal: could not find any configuration file %v in current directory or any of the parent directories", f.Names)
}

func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("%s\n", err)
	os.Exit(1)
}
