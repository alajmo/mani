package core

import "slices"

// Output formats of the run and exec commands, and of the spec output option
var RunOutputs = []string{"stream", "table", "html", "markdown", "json", "yaml"}

// Output formats of the list commands
var ListOutputs = []string{"table", "markdown", "html", "json", "yaml"}

// CheckOutput returns an error if output is not one of the valid output formats.
func CheckOutput(output string, valid []string) error {
	if slices.Contains(valid, output) {
		return nil
	}

	return &OutputError{Output: output, Valid: valid}
}

// CMD Flags

type TUIFlags struct {
	Theme  string
	Reload bool
}

type ListFlags struct {
	Output string
	Theme  string
	Tree   bool
}

type DescribeFlags struct {
	Theme string
}

type SetProjectFlags struct {
	All    bool
	Cwd    bool
	Target bool
}

type ProjectFlags struct {
	All      bool
	Cwd      bool
	Tags     []string
	TagsExpr string
	Paths    []string
	Projects []string
	Target   string
	Headers  []string
	Edit     bool
}

type TagFlags struct {
	Headers []string
}

type TaskFlags struct {
	Headers []string
	Edit    bool
}

type RunFlags struct {
	Edit     bool
	Parallel bool
	DryRun   bool
	Silent   bool
	Describe bool
	Cwd      bool
	TTY      bool
	Theme    string
	Target   string
	Spec     string

	All      bool
	Projects []string
	Paths    []string
	Tags     []string
	TagsExpr string

	IgnoreErrors      bool
	IgnoreNonExisting bool
	OmitEmptyRows     bool
	OmitEmptyColumns  bool
	Output            string
	Forks             uint32
}

type SetRunFlags struct {
	TTY bool

	All bool
	Cwd bool

	Parallel          bool
	OmitEmptyColumns  bool
	OmitEmptyRows     bool
	IgnoreErrors      bool
	IgnoreNonExisting bool
	Forks             bool
}

type SyncFlags struct {
	IgnoreSyncState         bool
	Parallel                bool
	SyncGitignore           bool
	Status                  bool
	SyncRemotes             bool
	RemoveOrphanedWorktrees bool
	Forks                   uint32
}

type SetSyncFlags struct {
	Parallel                bool
	SyncGitignore           bool
	SyncRemotes             bool
	RemoveOrphanedWorktrees bool
	Forks                   bool
}

type InitFlags struct {
	AutoDiscovery bool
	SyncGitignore bool
}
