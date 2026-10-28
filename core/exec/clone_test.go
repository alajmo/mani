package exec

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/alajmo/mani/core/dao"
)

// mkRepo creates dir (and its parents) as a git repository: with a .git
// directory, or with a .git file like worktrees and submodules have.
func mkRepo(t *testing.T, dir string, gitFile bool) {
	t.Helper()
	if gitFile {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: elsewhere\n"), 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
}

func TestFindOrphanedProjects(t *testing.T) {
	configDir := t.TempDir()

	// The config directory is itself a git repository and must never be reported
	mkRepo(t, configDir, false)

	mkRepo(t, filepath.Join(configDir, "active"), false)
	mkRepo(t, filepath.Join(configDir, "orphan"), false)
	mkRepo(t, filepath.Join(configDir, "group", "nested-active"), false)
	mkRepo(t, filepath.Join(configDir, "group", "nested-orphan"), false)
	mkRepo(t, filepath.Join(configDir, "active-wt"), true)

	// Kept: deeper than any project, a worktree or submodule (.git file), or in a hidden directory
	mkRepo(t, filepath.Join(configDir, "group", "deeper", "deep-repo"), false)
	mkRepo(t, filepath.Join(configDir, "other-worktree"), true)
	mkRepo(t, filepath.Join(configDir, ".cache", "repo"), false)

	// A repository inside a project (submodule, vendored checkout) is the project's business
	mkRepo(t, filepath.Join(configDir, "active", "vendor", "lib"), false)

	// Directories that aren't git repositories are never reported
	if err := os.MkdirAll(filepath.Join(configDir, "notes", "drafts"), 0755); err != nil {
		t.Fatal(err)
	}

	projects := []dao.Project{
		{Name: "active", Path: filepath.Join(configDir, "active"), WorktreeList: []dao.Worktree{{Path: "../active-wt", Branch: "dev"}}},
		{Name: "nested-active", Path: filepath.Join(configDir, "group", "nested-active")},
		{Name: "not-cloned", Path: filepath.Join(configDir, "group", "not-cloned")},
	}

	got, err := findOrphanedProjects(configDir, projects)
	if err != nil {
		t.Fatalf("findOrphanedProjects: %v", err)
	}

	want := []string{
		filepath.Join(configDir, "group", "nested-orphan"),
		filepath.Join(configDir, "orphan"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("orphans:\n got  %v\n want %v", got, want)
	}
}

func TestFindOrphanedProjectsRelativeAndDefaultPaths(t *testing.T) {
	configDir := t.TempDir()
	mkRepo(t, filepath.Join(configDir, "by-name"), false)
	mkRepo(t, filepath.Join(configDir, "sub", "by-path"), false)
	mkRepo(t, filepath.Join(configDir, "sub", "by-path-wt"), false)
	mkRepo(t, filepath.Join(configDir, "sub", "orphan"), false)

	projects := []dao.Project{
		{Name: "by-name"},
		{Name: "x", Path: "sub/by-path", WorktreeList: []dao.Worktree{{Path: filepath.Join(configDir, "sub", "by-path-wt")}}},
	}

	got, err := findOrphanedProjects(configDir, projects)
	if err != nil {
		t.Fatalf("findOrphanedProjects: %v", err)
	}
	want := []string{filepath.Join(configDir, "sub", "orphan")}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("orphans:\n got  %v\n want %v", got, want)
	}
}

func TestFindOrphanedProjectsWithoutProjectsBelowConfigDir(t *testing.T) {
	configDir := t.TempDir()
	mkRepo(t, filepath.Join(configDir, "repo"), false)

	projects := []dao.Project{{Name: "elsewhere", Path: filepath.Join(t.TempDir(), "elsewhere")}}

	got, err := findOrphanedProjects(configDir, projects)
	if err != nil {
		t.Fatalf("findOrphanedProjects: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no orphans when no project lives below the config directory, got %v", got)
	}
}

func TestFindOrphanedProjectsMissingConfigDir(t *testing.T) {
	_, err := findOrphanedProjects(filepath.Join(t.TempDir(), "missing"), nil)
	if err == nil {
		t.Fatal("expected an error for a missing config directory")
	}
}

func TestRemoveOrphanedProjectsKeepsEverythingWithoutTerminal(t *testing.T) {
	configDir := t.TempDir()
	orphan := filepath.Join(configDir, "orphan")
	mkRepo(t, orphan, false)
	mkRepo(t, filepath.Join(configDir, "active"), false)

	stdin, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer stdin.Close()
	oldStdin := os.Stdin
	os.Stdin = stdin
	defer func() { os.Stdin = oldStdin }()

	config := &dao.Config{
		Path:        filepath.Join(configDir, "mani.yaml"),
		Dir:         configDir,
		ProjectList: []dao.Project{{Name: "active", Path: filepath.Join(configDir, "active")}},
	}
	if err := RemoveOrphanedProjects(config); err != nil {
		t.Fatalf("RemoveOrphanedProjects: %v", err)
	}

	if _, err := os.Stat(orphan); err != nil {
		t.Errorf("orphan was removed without confirmation: %v", err)
	}
}
