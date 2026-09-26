package dao

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestConfig_DuplicateProjectName(t *testing.T) {
	originalProjects := []Project{
		{Name: "project-a", Path: "sub-1/project-a"},
		{Name: "project-a", Path: "sub-2/project-a"},
		{Name: "project-b", Path: "sub-3/project-b"},
	}

	var projects []Project
	projects = append(projects, originalProjects...)
	RenameDuplicates(projects)

	if projects[0].Name != originalProjects[0].Path {
		t.Fatalf(`Wanted: %q, Found: %q`, projects[0].Path, originalProjects[0].Name)
	}

	if projects[1].Name != originalProjects[1].Path {
		t.Fatalf(`Wanted: %q, Found: %q`, projects[1].Path, originalProjects[1].Name)
	}

	if originalProjects[2].Name != projects[2].Name {
		t.Fatalf(`Wanted: %q, Found: %q`, projects[2].Name, originalProjects[2].Name)
	}
}

// readTestConfig writes the config files to a temporary directory, with mani.yaml as the
// root config, and reads it. Returns the config and the temporary directory.
func readTestConfig(t *testing.T, files map[string]string) (Config, string, error) {
	t.Helper()

	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Use an empty user config so the user config of whoever runs the tests isn't imported
	userConfig := filepath.Join(t.TempDir(), "config.yaml")
	if _, found := files["user-config.yaml"]; found {
		userConfig = filepath.Join(dir, "user-config.yaml")
	} else if err := os.WriteFile(userConfig, nil, 0644); err != nil {
		t.Fatal(err)
	}

	config, err := ReadConfig(filepath.Join(dir, "mani.yaml"), userConfig, false)
	return config, dir, err
}

func TestConfig_ProjectDir(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected map[string]string // Project paths, relative to the root config unless absolute
	}{
		{
			name: "unset, paths resolve against the config directory",
			files: map[string]string{
				"mani.yaml": "import: [base/mani.yaml]",
				"base/mani.yaml": `
projects:
  project1:
  tools:
    path: libs/tools
`,
			},
			expected: map[string]string{
				"project1": "base/project1",
				"tools":    "base/libs/tools",
			},
		},
		{
			name: "parent directory of an imported config",
			files: map[string]string{
				"mani.yaml": "import: [base/mani.yaml]",
				"base/mani.yaml": `
project_dir: ..
projects:
  base:
  project1:
  tools:
    path: libs/tools
  sibling:
    path: ../sibling
  absolute:
    path: /opt/absolute
`,
			},
			expected: map[string]string{
				"base":     "base",
				"project1": "project1",
				"tools":    "libs/tools",
				"sibling":  "../sibling",
				"absolute": "/opt/absolute",
			},
		},
		{
			name: "subdirectory of an imported config",
			files: map[string]string{
				"mani.yaml": "import: [base/mani.yaml]",
				"base/mani.yaml": `
project_dir: hello
projects:
  project1:
  tools:
    path: libs/tools
`,
			},
			expected: map[string]string{
				"project1": "base/hello/project1",
				"tools":    "base/hello/libs/tools",
			},
		},
		{
			name: "only applies to the config file it's set in",
			files: map[string]string{
				"mani.yaml": `
import: [base/mani.yaml]
project_dir: repos
projects:
  project1:
`,
				"base/mani.yaml": `
projects:
  project2:
`,
			},
			expected: map[string]string{
				"project1": "repos/project1",
				"project2": "base/project2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, dir, err := readTestConfig(t, tt.files)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(config.ProjectList) != len(tt.expected) {
				t.Fatalf("expected %d projects, got %d", len(tt.expected), len(config.ProjectList))
			}

			for _, project := range config.ProjectList {
				expected, found := tt.expected[project.Name]
				if !found {
					t.Fatalf("unexpected project %q", project.Name)
				}

				if filepath.IsAbs(expected) {
					if project.Path != expected {
						t.Errorf("project %q: expected path %q, got %q", project.Name, expected, project.Path)
					}
					continue
				}

				// The relative path is relative to the root config
				if project.Path != filepath.Join(dir, expected) {
					t.Errorf("project %q: expected path %q, got %q", project.Name, filepath.Join(dir, expected), project.Path)
				}
				if project.RelPath != expected {
					t.Errorf("project %q: expected relpath %q, got %q", project.Name, expected, project.RelPath)
				}
			}
		})
	}
}

func TestConfig_MergeProjects(t *testing.T) {
	config, dir, err := readTestConfig(t, map[string]string{
		"mani.yaml": `
import: [base/mani.yaml]
projects:
  myproject:
  project1:
    desc: shared project
    tags: [local, react]
    env:
      B: mine
      A: shared
    remotes:
      fork: git@github.com:me/project1.git
    worktrees:
      - path: ../project1-wt
  project2:
    sync: false
`,
		"base/mani.yaml": `
project_dir: ..
projects:
  project1:
    url: git@github.com:owner/project1.git
    desc: shared project
    tags: [react]
    env:
      A: shared
    remotes:
      upstream: git@github.com:owner/project1.git
  project2:
    url: git@github.com:owner/project2.git
`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Merged projects keep the position of the config that owns them, not the one adding to them
	names := []string{}
	for _, project := range config.ProjectList {
		names = append(names, project.Name)
	}
	if expected := []string{"myproject", "project1", "project2"}; !slices.Equal(names, expected) {
		t.Fatalf("expected projects %v, got %v", expected, names)
	}

	project1, err := config.GetProject("project1")
	if err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		field    string
		got      any
		expected any
	}{
		{"path", project1.Path, filepath.Join(dir, "project1")},
		{"relpath", project1.RelPath, "project1"},
		{"context", project1.GetContext(), filepath.Join(dir, "base/mani.yaml")},
		{"url", project1.URL, "git@github.com:owner/project1.git"},
		{"desc", project1.Desc, "shared project"},
		{"tags", project1.Tags, []string{"react", "local"}},
		{"env", project1.EnvList, []string{"A=shared", "B=mine"}},
		{"remotes", project1.RemoteList, []Remote{
			{Name: "upstream", URL: "git@github.com:owner/project1.git"},
			{Name: "fork", URL: "git@github.com:me/project1.git"},
		}},
		{"worktrees", project1.WorktreeList, []Worktree{{Path: "../project1-wt", Branch: "project1-wt"}}},
	}
	for _, check := range checks {
		if !equalValues(check.got, check.expected) {
			t.Errorf("project1 %s: expected %v, got %v", check.field, check.expected, check.got)
		}
	}

	project2, err := config.GetProject("project2")
	if err != nil {
		t.Fatal(err)
	}
	if project2.IsSync() {
		t.Errorf("project2: expected sync to be false")
	}
}

func equalValues(a, b any) bool {
	switch a := a.(type) {
	case []string:
		return slices.Equal(a, b.([]string))
	case []Remote:
		return slices.Equal(a, b.([]Remote))
	case []Worktree:
		return slices.Equal(a, b.([]Worktree))
	default:
		return a == b
	}
}

func TestConfig_MergeProjects_DefaultPath(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected string
	}{
		{
			name: "adding tags doesn't move a project",
			files: map[string]string{
				"mani.yaml": `
import: [base/mani.yaml]
projects:
  project1:
    tags: [local]
`,
				"base/mani.yaml": `
projects:
  project1:
    url: git@github.com:owner/project1.git
`,
			},
			expected: "base/project1",
		},
		{
			name: "path set in the config without url",
			files: map[string]string{
				"mani.yaml": `
import: [base/mani.yaml]
projects:
  project1:
    path: code/project1
`,
				"base/mani.yaml": `
projects:
  project1:
    url: git@github.com:owner/project1.git
`,
			},
			expected: "code/project1",
		},
		{
			name: "identical path from different configs",
			files: map[string]string{
				"mani.yaml": `
import: [base/mani.yaml]
projects:
  project1:
    path: project1
    tags: [local]
`,
				"base/mani.yaml": `
project_dir: ..
projects:
  project1:
    path: project1
`,
			},
			expected: "project1",
		},
		{
			name: "user config adds tags",
			files: map[string]string{
				"mani.yaml": `
projects:
  project1:
    url: git@github.com:owner/project1.git
`,
				"user-config.yaml": `
projects:
  project1:
    tags: [local]
`,
			},
			expected: "project1",
		},
		{
			name: "importing config overrides url",
			files: map[string]string{
				"mani.yaml": `
import: [base/mani.yaml]
projects:
  project1:
    url: git@github.com:me/project1.git
`,
				"base/mani.yaml": `
projects:
  project1:
    url: git@github.com:owner/project1.git
`,
			},
			expected: "project1",
		},
		{
			name: "no url, the importing config wins",
			files: map[string]string{
				"mani.yaml": `
import: [base/mani.yaml]
projects:
  project1:
    tags: [local]
`,
				"base/mani.yaml": `
projects:
  project1:
    tags: [shared]
`,
			},
			expected: "project1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, dir, err := readTestConfig(t, tt.files)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(config.ProjectList) != 1 {
				t.Fatalf("expected 1 project, got %d", len(config.ProjectList))
			}

			expected := filepath.Join(dir, tt.expected)
			if config.ProjectList[0].Path != expected {
				t.Errorf("expected path %q, got %q", expected, config.ProjectList[0].Path)
			}
		})
	}
}

func TestConfig_MergeProjects_Errors(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected []string // Substrings of the error, {dir} is replaced with the temporary directory
	}{
		{
			name: "same project twice in one file",
			files: map[string]string{
				"mani.yaml": `
projects:
  project1:
    path: project1
  project1:
    path: project1
`,
			},
			expected: []string{
				"project `project1` is defined more than once in the same file, first definition is at {dir}/mani.yaml:3",
				"--> {dir}/mani.yaml:5",
			},
		},
		{
			name: "conflicting fields in configs that don't import each other",
			files: map[string]string{
				"mani.yaml": "import: [a.yaml, b.yaml]",
				"a.yaml": `
projects:
  project1:
    url: git@github.com:a/project1.git
    desc: a
    sync: true
`,
				"b.yaml": `
projects:
  project1:
    url: git@github.com:b/project1.git
    desc: b
    sync: false
`,
			},
			expected: []string{
				"project `project1` has conflicting `url`: `git@github.com:b/project1.git` here and `git@github.com:a/project1.git` in {dir}/a.yaml:3, override it in a config that imports both",
				"project `project1` has conflicting `desc`: `b` here and `a` in {dir}/a.yaml:3, override it in a config that imports both",
				"project `project1` has conflicting `sync`: `false` here and `true` in {dir}/a.yaml:3, override it in a config that imports both",
				"--> {dir}/b.yaml:3",
			},
		},
		{
			name: "conflicting path in configs that don't import each other",
			files: map[string]string{
				"mani.yaml": "import: [a.yaml, b.yaml]",
				"a.yaml":    "projects:\n  project1:\n    path: a\n",
				"b.yaml":    "projects:\n  project1:\n    path: b\n",
			},
			expected: []string{"project `project1` has conflicting `path`: `{dir}/b` here and `{dir}/a` in {dir}/a.yaml:2, override it in a config that imports both"},
		},
		{
			name: "ambiguous default path in configs that don't import each other",
			files: map[string]string{
				"mani.yaml":   "import: [a/mani.yaml, b/mani.yaml]",
				"a/mani.yaml": "projects:\n  project1:\n    tags: [a]\n",
				"b/mani.yaml": "projects:\n  project1:\n    tags: [b]\n",
			},
			expected: []string{
				"project `project1` resolves to `{dir}/b/project1` here and `{dir}/a/project1` in {dir}/a/mani.yaml:2, set `path` in one of the definitions",
			},
		},
		{
			name: "ambiguous default path with the same url in configs that don't import each other",
			files: map[string]string{
				"mani.yaml":   "import: [a/mani.yaml, b/mani.yaml]",
				"a/mani.yaml": "projects:\n  project1:\n    url: git@github.com:owner/project1.git\n",
				"b/mani.yaml": "projects:\n  project1:\n    url: git@github.com:owner/project1.git\n",
			},
			expected: []string{
				"project `project1` resolves to `{dir}/b/project1` here and `{dir}/a/project1` in {dir}/a/mani.yaml:2, set `path` in one of the definitions",
			},
		},
		{
			name: "conflicting env, remote and worktree in configs that don't import each other",
			files: map[string]string{
				"mani.yaml": "import: [a.yaml, b.yaml]",
				"a.yaml": `
projects:
  project1:
    url: git@github.com:owner/project1.git
    env:
      A: a
    remotes:
      origin: git@github.com:a/project1.git
    worktrees:
      - path: wt
        branch: a
`,
				"b.yaml": `
projects:
  project1:
    env:
      A: b
    remotes:
      origin: git@github.com:b/project1.git
    worktrees:
      - path: wt
        branch: b
`,
			},
			expected: []string{
				"project `project1` has conflicting env `A`: `b` here and `a` in {dir}/a.yaml:3",
				"project `project1` has conflicting remote `origin`: `git@github.com:b/project1.git` here and `git@github.com:a/project1.git` in {dir}/a.yaml:3",
				"project `project1` has conflicting worktree `wt`: `b` here and `a` in {dir}/a.yaml:3",
				"--> {dir}/b.yaml:3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, dir, err := readTestConfig(t, tt.files)
			if err == nil {
				t.Fatal("expected error but got none")
			}

			for _, expected := range tt.expected {
				expected = strings.ReplaceAll(expected, "{dir}", dir)
				if !strings.Contains(err.Error(), expected) {
					t.Errorf("expected error to contain:\n%s\ngot:\n%s", expected, err)
				}
			}
		})
	}
}

func TestConfig_MergeProjects_Overrides(t *testing.T) {
	tests := []struct {
		name              string
		files             map[string]string
		expectedEnv       []string
		expectedRemotes   []Remote
		expectedWorktrees []Worktree
	}{
		{
			name: "importing config overrides env, remotes and worktrees",
			files: map[string]string{
				"mani.yaml": `
import: [base/mani.yaml]
projects:
  project1:
    env:
      A: mine
      C: mine
    remotes:
      origin: git@github.com:me/project1.git
    worktrees:
      - path: wt
        branch: mine
`,
				"base/mani.yaml": `
project_dir: ..
projects:
  project1:
    url: git@github.com:owner/project1.git
    env:
      A: shared
      B: shared
    remotes:
      origin: git@github.com:owner/project1.git
      upstream: git@github.com:owner/project1.git
    worktrees:
      - path: wt
        branch: shared
`,
			},
			expectedEnv: []string{"A=mine", "B=shared", "C=mine"},
			expectedRemotes: []Remote{
				{Name: "origin", URL: "git@github.com:me/project1.git"},
				{Name: "upstream", URL: "git@github.com:owner/project1.git"},
			},
			expectedWorktrees: []Worktree{{Path: "wt", Branch: "mine"}},
		},
		{
			name: "indirect import",
			files: map[string]string{
				"mani.yaml": `
import: [base.yaml]
projects:
  project1:
    env:
      A: root
`,
				"base.yaml": `
import: [common.yaml]
projects:
  project1:
    env:
      A: base
      B: base
`,
				"common.yaml": `
projects:
  project1:
    url: git@github.com:owner/project1.git
    env:
      A: common
      B: common
`,
			},
			expectedEnv: []string{"A=root", "B=base"},
		},
		{
			name: "importing config settles a conflict between configs that don't import each other",
			files: map[string]string{
				"mani.yaml": `
import: [a.yaml, b.yaml]
projects:
  project1:
    env:
      A: root
`,
				"a.yaml": `
projects:
  project1:
    url: git@github.com:owner/project1.git
    env:
      A: a
`,
				"b.yaml": `
projects:
  project1:
    env:
      A: b
`,
			},
			expectedEnv: []string{"A=root"},
		},
		{
			name: "config imported twice is overridden by the second importer",
			files: map[string]string{
				"mani.yaml": "import: [a.yaml, b.yaml]",
				"a.yaml":    "import: [common.yaml]",
				"b.yaml": `
import: [common.yaml]
projects:
  project1:
    env:
      A: b
`,
				"common.yaml": `
projects:
  project1:
    url: git@github.com:owner/project1.git
    env:
      A: common
`,
			},
			expectedEnv: []string{"A=b"},
		},
		{
			name: "root config overrides user config",
			files: map[string]string{
				"mani.yaml": `
projects:
  project1:
    url: git@github.com:owner/project1.git
    env:
      A: root
`,
				"user-config.yaml": `
projects:
  project1:
    env:
      A: user
      B: user
`,
			},
			expectedEnv: []string{"A=root", "B=user"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, _, err := readTestConfig(t, tt.files)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			project, err := config.GetProject("project1")
			if err != nil {
				t.Fatal(err)
			}

			if !slices.Equal(project.EnvList, tt.expectedEnv) {
				t.Errorf("expected env %v, got %v", tt.expectedEnv, project.EnvList)
			}
			if tt.expectedRemotes != nil && !slices.Equal(project.RemoteList, tt.expectedRemotes) {
				t.Errorf("expected remotes %v, got %v", tt.expectedRemotes, project.RemoteList)
			}
			if tt.expectedWorktrees != nil && !slices.Equal(project.WorktreeList, tt.expectedWorktrees) {
				t.Errorf("expected worktrees %v, got %v", tt.expectedWorktrees, project.WorktreeList)
			}
		})
	}
}

func TestConfig_MergeProjects_FieldOverrides(t *testing.T) {
	config, dir, err := readTestConfig(t, map[string]string{
		"mani.yaml": `
import: [base/mani.yaml]
projects:
  project1:
    path: code/project1
    url: git@github.com:me/project1.git
    desc: mine
    clone: git clone git@github.com:me/project1.git code/project1
    branch: dev
    single_branch: true
    sync: false
`,
		"base/mani.yaml": `
projects:
  project1:
    path: project1
    url: git@github.com:owner/project1.git
    desc: shared
    clone: git clone git@github.com:owner/project1.git project1
    branch: main
    single_branch: false
    sync: true
    tags: [shared]
`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	project, err := config.GetProject("project1")
	if err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		field    string
		got      any
		expected any
	}{
		{"path", project.Path, filepath.Join(dir, "code/project1")},
		{"context", project.GetContext(), filepath.Join(dir, "mani.yaml")},
		{"url", project.URL, "git@github.com:me/project1.git"},
		{"desc", project.Desc, "mine"},
		{"clone", project.Clone, "git clone git@github.com:me/project1.git code/project1"},
		{"branch", project.Branch, "dev"},
		{"single_branch", *project.SingleBranch, true},
		{"sync", project.IsSync(), false},
		{"tags", project.Tags, []string{"shared"}},
	}
	for _, check := range checks {
		if !equalValues(check.got, check.expected) {
			t.Errorf("project1 %s: expected %v, got %v", check.field, check.expected, check.got)
		}
	}
}

func TestConfig_MergeProjects_WorktreePaths(t *testing.T) {
	config, _, err := readTestConfig(t, map[string]string{
		"mani.yaml": `
import: [base/mani.yaml]
projects:
  project1:
    worktrees:
      - path: ./staging
        branch: main
`,
		"base/mani.yaml": `
project_dir: ..
projects:
  project1:
    url: git@github.com:owner/project1.git
    worktrees:
      - path: staging
      - path: hotfix/
`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// staging and ./staging are the same worktree, and the importing config wins
	expected := []Worktree{{Path: "./staging", Branch: "main"}, {Path: "hotfix/", Branch: "hotfix"}}
	if !slices.Equal(config.ProjectList[0].WorktreeList, expected) {
		t.Errorf("expected worktrees %v, got %v", expected, config.ProjectList[0].WorktreeList)
	}
}

func TestConfig_RelPath(t *testing.T) {
	config, _, err := readTestConfig(t, map[string]string{
		"mani.yaml": `
import: [sub/mani.yaml, base/mani.yaml]
project_dir: repos
projects:
  project1:
`,
		"sub/mani.yaml": `
projects:
  project2:
`,
		"base/mani.yaml": `
project_dir: ..
projects:
  project3:
`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Relative paths are relative to the root config, regardless of which config defines the project
	expected := map[string]string{"project1": "repos/project1", "project2": "sub/project2", "project3": "project3"}
	for _, project := range config.ProjectList {
		if project.RelPath != expected[project.Name] {
			t.Errorf("project %q: expected relpath %q, got %q", project.Name, expected[project.Name], project.RelPath)
		}
	}

	projects, err := config.GetProjectsByPath([]string{"repos"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if names := getProjectNames(projects); !slices.Equal(names, []string{"project1"}) {
		t.Errorf("expected --paths repos to match [project1], got %v", names)
	}

	// Paths offered by the TUI and shell completion, which are then passed to --paths
	if paths := config.GetProjectPaths(); !slices.Equal(paths, []string{"repos", "sub"}) {
		t.Errorf("expected project paths [repos sub], got %v", paths)
	}
}

func TestConfig_MergeTasks(t *testing.T) {
	config, dir, err := readTestConfig(t, map[string]string{
		"mani.yaml": `
import: [shared.yaml]
tasks:
  build:
    cmd: echo personal
  test:
    desc: personal tests
    env:
      MODE: personal
  deploy:
    tty: false
`,
		"shared.yaml": `
tasks:
  build:
    desc: build it
    commands:
      - cmd: echo one
    env:
      A: shared
  test:
    cmd: echo test
    spec:
      output: table
    env:
      MODE: shared
      B: shared
  deploy:
    cmd: echo deploy
    tty: true
  lint: echo lint
`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// A merged task takes the position of the config its cmd comes from
	names := []string{}
	for _, task := range config.TaskList {
		names = append(names, task.Name)
	}
	if expected := []string{"build", "test", "deploy", "lint"}; !slices.Equal(names, expected) {
		t.Fatalf("expected tasks %v, got %v", expected, names)
	}

	build, _ := config.GetTask("build")
	test, _ := config.GetTask("test")
	deploy, _ := config.GetTask("deploy")

	checks := []struct {
		field    string
		got      any
		expected any
	}{
		// cmd and commands are one field, so the importing config's cmd replaces the shared commands
		{"build cmd", build.Cmd, "echo personal"},
		{"build commands", len(build.Commands), 0},
		{"build desc", build.Desc, "build it"},
		{"build env", ParseNodeEnv(build.Env), []string{"A=shared"}},
		{"build context", build.GetContext(), filepath.Join(dir, "mani.yaml")},
		{"test cmd", test.Cmd, "echo test"},
		{"test desc", test.Desc, "personal tests"},
		{"test spec", test.SpecData.Output, "table"},
		{"test env", ParseNodeEnv(test.Env), []string{"MODE=personal", "B=shared"}},
		{"test context", test.GetContext(), filepath.Join(dir, "shared.yaml")},
		{"deploy cmd", deploy.Cmd, "echo deploy"},
		{"deploy tty", deploy.TTY, false},
	}
	for _, check := range checks {
		if !equalValues(check.got, check.expected) {
			t.Errorf("%s: expected %v, got %v", check.field, check.expected, check.got)
		}
	}
}

func TestConfig_MergeEnvList(t *testing.T) {
	config, _, err := readTestConfig(t, map[string]string{
		"mani.yaml": `
import: [shared.yaml]
env:
  LEVEL: personal
`,
		"shared.yaml": `
env:
  REGION: eu
  LEVEL: shared
`,
		"user-config.yaml": `
env:
  LEVEL: user
  USER_ONLY: "true"
`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"LEVEL=personal", "REGION=eu", "USER_ONLY=true"}
	if !slices.Equal(config.EnvList, expected) {
		t.Errorf("expected env %v, got %v", expected, config.EnvList)
	}
}

func TestConfig_UserConfig(t *testing.T) {
	// The user config, and the configs it imports, are overridden by every other config,
	// not only the root config
	config, dir, err := readTestConfig(t, map[string]string{
		"mani.yaml": "import: [shared.yaml]",
		"shared.yaml": `
env:
  LEVEL: shared
projects:
  project1:
    url: git@github.com:owner/project1.git
    desc: shared
    env:
      A: shared
tasks:
  build:
    cmd: echo shared
`,
		"user-config.yaml": `
import: [user-extra.yaml]
env:
  LEVEL: user
  USER_ONLY: "true"
projects:
  project1:
    desc: user
    tags: [user]
    env:
      A: user
      B: user
tasks:
  build:
    desc: user build
    cmd: echo user
`,
		"user-extra.yaml": `
env:
  LEVEL: extra
  EXTRA: "true"
`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	project, err := config.GetProject("project1")
	if err != nil {
		t.Fatal(err)
	}
	build, err := config.GetTask("build")
	if err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		field    string
		got      any
		expected any
	}{
		{"env", config.EnvList, []string{"LEVEL=shared", "USER_ONLY=true", "EXTRA=true"}},
		{"project desc", project.Desc, "shared"},
		{"project tags", project.Tags, []string{"user"}},
		{"project env", project.EnvList, []string{"A=shared", "B=user"}},
		{"task cmd", build.Cmd, "echo shared"},
		{"task desc", build.Desc, "user build"},
		{"task context", build.GetContext(), filepath.Join(dir, "shared.yaml")},
	}
	for _, check := range checks {
		if !equalValues(check.got, check.expected) {
			t.Errorf("%s: expected %v, got %v", check.field, check.expected, check.got)
		}
	}
}

func TestConfig_MergeTasksAndEnv_Errors(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected []string // Substrings of the error, {dir} is replaced with the temporary directory
	}{
		{
			name: "same task twice in one file",
			files: map[string]string{
				"mani.yaml": "tasks:\n  build: echo a\n  build: echo a\n",
			},
			expected: []string{
				"task `build` is defined more than once in the same file, first definition is at {dir}/mani.yaml:2",
				"--> {dir}/mani.yaml:3",
			},
		},
		{
			name: "conflicting task fields in configs that don't import each other",
			files: map[string]string{
				"mani.yaml": "import: [a.yaml, b.yaml]",
				"a.yaml":    "tasks:\n  build:\n    cmd: echo a\n    spec:\n      output: table\n",
				"b.yaml":    "tasks:\n  build:\n    cmd: echo b\n    spec:\n      output: stream\n",
			},
			expected: []string{
				"task `build` has conflicting `cmd`: `echo b` here and `echo a` in {dir}/a.yaml:2, override it in a config that imports both",
				"task `build` has conflicting `spec`: `{output: stream}` here and `{output: table}` in {dir}/a.yaml:2",
				"--> {dir}/b.yaml:2",
			},
		},
		{
			name: "conflicting task env in configs that don't import each other",
			files: map[string]string{
				"mani.yaml": "import: [a.yaml, b.yaml]",
				"a.yaml":    "tasks:\n  build:\n    cmd: echo\n    env:\n      A: a\n",
				"b.yaml":    "tasks:\n  build:\n    env:\n      A: b\n",
			},
			expected: []string{"task `build` has conflicting env `A`: `b` here and `a` in {dir}/a.yaml:2"},
		},
		{
			name: "conflicting env in configs that don't import each other",
			files: map[string]string{
				"mani.yaml": "import: [a.yaml, b.yaml]",
				"a.yaml":    "env:\n  A: a\n",
				"b.yaml":    "env:\n  A: b\n",
			},
			expected: []string{
				"conflicting env `A`: `b` here and `a` in {dir}/a.yaml:2, override it in a config that imports both",
				"--> {dir}/b.yaml:2",
			},
		},
		{
			name: "same env twice in one file",
			files: map[string]string{
				"mani.yaml": "env:\n  A: 1\n  A: 2\n",
			},
			expected: []string{
				// No hint to override it in another config, since it's the same file
				"conflicting env `A`: `2` here and `1` in {dir}/mani.yaml:2\n",
				"--> {dir}/mani.yaml:3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, dir, err := readTestConfig(t, tt.files)
			if err == nil {
				t.Fatal("expected error but got none")
			}

			for _, expected := range tt.expected {
				expected = strings.ReplaceAll(expected, "{dir}", dir)
				if !strings.Contains(err.Error(), expected) {
					t.Errorf("expected error to contain:\n%s\ngot:\n%s", expected, err)
				}
			}
		})
	}
}

func TestConfig_EvaluateProjectEnv(t *testing.T) {
	config, _, err := readTestConfig(t, map[string]string{
		"mani.yaml": `
import: [shared.yaml]
projects:
  app:
    env:
      TOKEN: mine
`,
		"shared.yaml": `
projects:
  app:
    path: app
    env:
      TOKEN: $(exit 1)
      GREETING: $(echo hi)
`,
	})
	// The overridden TOKEN would fail if it was evaluated
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"TOKEN=mine", "GREETING=hi\n"}
	if !slices.Equal(config.ProjectList[0].EnvList, expected) {
		t.Errorf("expected env %q, got %q", expected, config.ProjectList[0].EnvList)
	}
}

func TestConfig_EvaluateProjectEnv_Errors(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected []string // Substrings of the error, {dir} is replaced with the temporary directory
	}{
		{
			name: "failing env of a project in one config",
			files: map[string]string{
				"mani.yaml": "projects:\n  app:\n    path: app\n    env:\n      TOKEN: $(echo boom; exit 1)\n",
			},
			expected: []string{"failed to evaluate env `TOKEN`", "boom", "--> {dir}/mani.yaml:2"},
		},
		{
			name: "failing env points to the config that sets it",
			files: map[string]string{
				"mani.yaml":   "import: [shared.yaml]\nprojects:\n  app:\n    env:\n      TOKEN: $(exit 1)\n",
				"shared.yaml": "projects:\n  app:\n    path: app\n",
			},
			expected: []string{"failed to evaluate env `TOKEN`", "--> {dir}/mani.yaml:3"},
		},
		{
			name: "command output containing a line number",
			files: map[string]string{
				"mani.yaml": "projects:\n  app:\n    path: app\n    env:\n      TOKEN: \"$(echo 'sh: line 1: foo: command not found'; exit 1)\"\n",
			},
			expected: []string{"failed to evaluate env `TOKEN`", "--> {dir}/mani.yaml:2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, dir, err := readTestConfig(t, tt.files)
			if err == nil {
				t.Fatal("expected error but got none")
			}

			for _, expected := range tt.expected {
				expected = strings.ReplaceAll(expected, "{dir}", dir)
				if !strings.Contains(err.Error(), expected) {
					t.Errorf("expected error to contain:\n%s\ngot:\n%s", expected, err)
				}
			}
		})
	}
}

func TestConfig_EvaluateProjectEnv_InvalidConfig(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected string
	}{
		{
			name: "conflict between configs",
			files: map[string]string{
				"mani.yaml": "import: [a.yaml, b.yaml]\nprojects:\n  app:\n    path: app\n    env:\n      TOKEN: $(exit 1)\n",
				"a.yaml":    "tasks:\n  build: echo a\n",
				"b.yaml":    "tasks:\n  build: echo b\n",
			},
			expected: "task `build` has conflicting `cmd`",
		},
		{
			name: "invalid task",
			files: map[string]string{
				"mani.yaml": "projects:\n  app:\n    path: app\n    env:\n      TOKEN: $(exit 1)\ntasks:\n  build:\n    cmd: echo\n    theme: missing\n",
			},
			expected: "cannot find theme `missing`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := readTestConfig(t, tt.files)
			if err == nil {
				t.Fatal("expected error but got none")
			}

			if !strings.Contains(err.Error(), tt.expected) {
				t.Errorf("expected error to contain %q, got:\n%s", tt.expected, err)
			}

			// Env commands shouldn't run for a config that's invalid anyway
			if strings.Contains(err.Error(), "failed to evaluate env") {
				t.Errorf("expected project env not to be evaluated, got:\n%s", err)
			}
		})
	}
}

func TestConfig_MergeTasks_FieldErrors(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected []string // Substrings of the error, {dir} is replaced with the temporary directory
	}{
		{
			name: "invalid theme reference from the importing config",
			files: map[string]string{
				"mani.yaml":   "import: [shared.yaml]\ntasks:\n  build:\n    theme: missing\n",
				"shared.yaml": "tasks:\n  build:\n    cmd: make\n",
			},
			expected: []string{"cannot find theme `missing`", "--> {dir}/mani.yaml:3"},
		},
		{
			name: "invalid inline spec from the importing config",
			files: map[string]string{
				"mani.yaml":   "import: [shared.yaml]\ntasks:\n  build:\n    spec:\n      parallel: maybe\n",
				"shared.yaml": "tasks:\n  build:\n    cmd: make\n",
			},
			expected: []string{"cannot unmarshal", "--> {dir}/mani.yaml:5"},
		},
		{
			name: "invalid target reference from the config that owns the task",
			files: map[string]string{
				"mani.yaml":   "import: [shared.yaml]\ntasks:\n  build:\n    desc: build it\n",
				"shared.yaml": "tasks:\n  build:\n    cmd: make\n    target: missing\n",
			},
			expected: []string{"cannot find target `missing`", "--> {dir}/shared.yaml:2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, dir, err := readTestConfig(t, tt.files)
			if err == nil {
				t.Fatal("expected error but got none")
			}

			for _, expected := range tt.expected {
				expected = strings.ReplaceAll(expected, "{dir}", dir)
				if !strings.Contains(err.Error(), expected) {
					t.Errorf("expected error to contain:\n%s\ngot:\n%s", expected, err)
				}
			}
		})
	}
}
