package core

// CMD Flags

type ListFlags struct {
	Output string
	Theme  string
	Tree   bool
}

type ProjectFlags struct {
	Tags    []string
	Paths   []string
	Headers []string
	Edit    bool
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
	Theme    string

	All      bool
	Projects []string
	Paths    []string
	Tags     []string

	OmitEmpty bool
	Output    string
}

type SetRunFlags struct {
	Parallel  bool
	OmitEmpty bool
}

type SyncFlags struct {
	Parallel bool
	Status   bool
}

type InitFlags struct {
	AutoDiscovery bool
	Vcs           string
}
