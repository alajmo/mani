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
	AutoDiscovery    bool
	AutoDiscoverySet bool // whether the user passed --auto-discovery explicitly
	SyncGitignore    bool

	// GitHub-source mode (set when --source github/<name> is used).
	Source      string // raw flag value, e.g. "github/my-org"
	SourcePath  string // common path prefix for the source's projects
	Limit       int
	Archived    bool // include archived repos (default false = exclude)
	Fork        bool // include forks (default true = include)
	NoTags      bool
	NoTopicTags bool
}

type RefreshFlags struct {
	DryRun bool
	Prune  bool
}
