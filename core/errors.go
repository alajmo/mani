package core

import (
	"fmt"
)

type FailedToOpenFile struct {
	name string
}

func (f *FailedToOpenFile) Error() string {
	return fmt.Sprintf("error: failed to open %q", f.name)
}

type MissingFile struct {
	name string
}

func (f *MissingFile) Error() string {
	return fmt.Sprintf("error: missing %q", f.name)
}

type FailedToParseFile struct {
	name string
	msg  error
}

type FailedToParsePath struct {
	name string
}

func (f *FailedToParsePath) Error() string {
	return fmt.Sprintf("error: failed to parse path %q", f.name)
}

func (f *FailedToParseFile) Error() string {
	return fmt.Sprintf("error: failed to parse %q \n%s", f.name, f.msg)
}

type PathDoesNotExist struct {
	path string
}

func (p *PathDoesNotExist) Error() string {
	return fmt.Sprintf("fatal: path %q does not exist", p.path)
}

type CommandNotFound struct {
	name string
}

func (c *CommandNotFound) Error() string {
	return fmt.Sprintf("fatal: could not find command %q", c.name)
}

type ConfigNotFound struct {
	names []string
}

func (f *ConfigNotFound) Error() string {
	return fmt.Sprintf("fatal: could not find any configuration file %v in current directory or any of the parent directories", f.names)
}

type FileNotFound struct {
	name string
}

func (f *FileNotFound) Error() string {
	return fmt.Sprintf("fatal: could not find %q (in current directory or any of the parent directories)", f.name)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
