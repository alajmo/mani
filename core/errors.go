package core

import (
	"fmt"
	"log"
)

type ConfigEnvFailed struct {
	Name string
	Err  error
}

func (c *ConfigEnvFailed) Error() string {
	return fmt.Sprintf("error: failed to evaluate env `%s` \n `%s` ", c.Name, c.Err)
}

type FailedToOpenFile struct {
	Name string
}

func (f *FailedToOpenFile) Error() string {
	return fmt.Sprintf("error: failed to open `%s`", f.Name)
}

type FailedToParsePath struct {
	Name string
}

func (f *FailedToParsePath) Error() string {
	return fmt.Sprintf("error: failed to parse path `%s`", f.Name)
}

type FailedToParseFile struct {
	Name string
	Msg  error
}

func (f *FailedToParseFile) Error() string {
	return fmt.Sprintf("failed to parse `%s` \n%s", f.Name, f.Msg)
}

type PathDoesNotExist struct {
	Path string
}

func (p *PathDoesNotExist) Error() string {
	return fmt.Sprintf("error: path `%s` does not exist", p.Path)
}

type ProjectNotFound struct {
	Name string
}

func (c *ProjectNotFound) Error() string {
	return fmt.Sprintf("error: cannot find project `%s`", c.Name)
}

type TaskNotFound struct {
	Name string
}

func (c *TaskNotFound) Error() string {
	return fmt.Sprintf("cannot find task `%s`", c.Name)
}

type ThemeNotFound struct {
	Name string
}

func (c *ThemeNotFound) Error() string {
	return fmt.Sprintf("error: cannot find theme `%s`", c.Name)
}

type SpecNotFound struct {
	Name string
}

func (c *SpecNotFound) Error() string {
	return fmt.Sprintf("error: cannot find spec `%s`", c.Name)
}

type TargetNotFound struct {
	Name string
}

func (c *TargetNotFound) Error() string {
	return fmt.Sprintf("error: cannot find target `%s`", c.Name)
}

type ConfigNotFound struct {
	Names []string
}

func (f *ConfigNotFound) Error() string {
	return fmt.Sprintf("error: cannot find any configuration file %v in current directory or any of the parent directories", f.Names)
}

func CheckIfError(err error) {
	if err == nil {
		return
	}

	log.SetFlags(0)
	log.Fatalf("%s\n", err)
}
