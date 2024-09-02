package core

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
	IgnoreSyncState bool
	Parallel        bool
	SyncGitignore   bool
	Status          bool
	SyncRemotes     bool
	Forks           uint32
}

type SetSyncFlags struct {
	Parallel      bool
	SyncGitignore bool
	SyncRemotes   bool
	Forks         bool
}

type InitFlags struct {
	AutoDiscovery bool
	SyncGitignore bool
}
