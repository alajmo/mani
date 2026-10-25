package dao

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// readTestConfig writes the config files to a temporary directory, with mani.yaml as the
// root config, and reads it. {dir} in the files is replaced with the temporary directory.
// Returns the config and the temporary directory.
func readTestConfig(t *testing.T, files map[string]string) (Config, string, error) {
	t.Helper()

	dir := writeTestFiles(t, files)

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

// writeTestFiles writes the files to a temporary directory and returns it. {dir} in the files is
// replaced with the temporary directory.
func writeTestFiles(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		content = strings.ReplaceAll(content, "{dir}", dir)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	return dir
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
			// project_dir in the user config tells its paths apart from the root config's
			name: "user config sets url",
			files: map[string]string{
				"mani.yaml": `
projects:
  project1:
    tags: [local]
`,
				"user-config.yaml": `
project_dir: user
projects:
  project1:
    url: git@github.com:owner/project1.git
`,
			},
			expected: "project1",
		},
		{
			name: "project only in the user config",
			files: map[string]string{
				"mani.yaml": "",
				"user-config.yaml": `
project_dir: user
projects:
  project1:
    url: git@github.com:owner/project1.git
`,
			},
			expected: "user/project1",
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
			expected: []string{"failed to evaluate env `TOKEN`", "boom", "--> {dir}/mani.yaml:5"},
		},
		{
			name: "failing env points to the config that sets it",
			files: map[string]string{
				"mani.yaml":   "import: [shared.yaml]\nprojects:\n  app:\n    env:\n      TOKEN: $(exit 1)\n",
				"shared.yaml": "projects:\n  app:\n    path: app\n",
			},
			expected: []string{"failed to evaluate env `TOKEN`", "--> {dir}/mani.yaml:5"},
		},
		{
			name: "command output containing a line number",
			files: map[string]string{
				"mani.yaml": "projects:\n  app:\n    path: app\n    env:\n      TOKEN: \"$(echo 'sh: line 1: foo: command not found'; exit 1)\"\n",
			},
			expected: []string{"failed to evaluate env `TOKEN`", "--> {dir}/mani.yaml:5"},
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
