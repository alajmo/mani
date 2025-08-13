package exec

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func TestRemoveOrphanedProjects(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mani-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock mani.yaml in the temp directory
	configPath := filepath.Join(tempDir, "mani.yaml")
	configContent := `projects:
  active-project:
    path: active-project
    url: https://github.com/example/active-project.git

tasks:
  hello:
    desc: Print Hello World
    cmd: echo "Hello World"`

	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create active project directory (should not be removed)
	activeProjectDir := filepath.Join(tempDir, "active-project")
	err = os.MkdirAll(filepath.Join(activeProjectDir, ".git"), 0755)
	if err != nil {
		t.Fatalf("Failed to create active project: %v", err)
	}

	// Create orphaned project directory (should be removed)
	orphanedProjectDir := filepath.Join(tempDir, "orphaned-project")
	err = os.MkdirAll(filepath.Join(orphanedProjectDir, ".git"), 0755)
	if err != nil {
		t.Fatalf("Failed to create orphaned project: %v", err)
	}

	// Create non-git directory (should not be removed)
	nonGitDir := filepath.Join(tempDir, "non-git-dir")
	err = os.MkdirAll(nonGitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create non-git directory: %v", err)
	}

	// Create hidden directory (should not be removed)
	hiddenDir := filepath.Join(tempDir, ".hidden-dir")
	err = os.MkdirAll(hiddenDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create hidden directory: %v", err)
	}

	// Create config and project list
	config := &dao.Config{
		Path: configPath,
		ProjectList: []dao.Project{
			{
				Name: "active-project",
				Path: "active-project",
				Url:  "https://github.com/example/active-project.git",
			},
		},
		RemoveOrphaned: &[]bool{true}[0],
	}

	// Test the function (we need to mock the user input)
	// Since we can't easily mock stdin in a unit test, we'll test the directory detection logic
	// by checking what directories exist before and after

	// Check initial state
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp dir: %v", err)
	}

	var dirNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirNames = append(dirNames, entry.Name())
		}
	}

	// Should have: active-project, orphaned-project, non-git-dir, .hidden-dir
	expectedDirs := []string{"active-project", "orphaned-project", "non-git-dir", ".hidden-dir"}
	for _, expectedDir := range expectedDirs {
		found := false
		for _, dirName := range dirNames {
			if dirName == expectedDir {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected directory %s not found. Found: %v", expectedDir, dirNames)
		}
	}

	// Test that we can identify orphaned projects correctly
	// We'll extract the logic that identifies orphaned directories for testing
	t.Run("IdentifyOrphanedProjects", func(t *testing.T) {
		testIdentifyOrphanedProjects(t, config, tempDir)
	})

	t.Run("RemoveOrphanedProjectsWithAutoConfirm", func(t *testing.T) {
		testRemoveOrphanedProjectsWithAutoConfirm(t, config, tempDir)
	})
}

// Helper function to test the orphaned project identification logic
func testIdentifyOrphanedProjects(t *testing.T, config *dao.Config, configDir string) {
	configDir = filepath.Dir(config.Path)
	// Get all project paths that should exist based on current configuration
	activeProjectPaths := make(map[string]bool)
	for _, project := range config.ProjectList {
		projectPath, err := core.GetAbsolutePath(configDir, project.Path, project.Name)
		if err != nil {
			t.Fatalf("Failed to resolve project path: %v", err)
		}
		activeProjectPaths[projectPath] = true
	}

	// Find all directories in the config directory
	var orphanedPaths []string

	entries, err := os.ReadDir(configDir)
	if err != nil {
		t.Fatalf("Failed to read config directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		fullPath := filepath.Join(configDir, entry.Name())

		// Only consider directories that have a .git folder (git repositories)
		gitDir := filepath.Join(fullPath, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			continue // Not a git repository, skip
		}

		// Check if this path is in our active projects
		if !activeProjectPaths[fullPath] {
			orphanedPaths = append(orphanedPaths, fullPath)
		}
	}

	// Should identify only the orphaned-project as orphaned
	if len(orphanedPaths) != 1 {
		t.Errorf("Expected 1 orphaned project, found %d: %v", len(orphanedPaths), orphanedPaths)
		return
	}

	var expectedOrphanedPath string
	expectedOrphanedPath, err = core.GetAbsolutePath(configDir, "orphaned-project", "orphaned-project")
	if err != nil {
		t.Fatalf("Failed to resolve expected orphaned project path: %v", err)
	}
	if orphanedPaths[0] != expectedOrphanedPath {
		t.Errorf("Expected orphaned project at %s, found %s", expectedOrphanedPath, orphanedPaths[0])
	}
}

// Helper function to test the actual removal with auto-confirmation
func testRemoveOrphanedProjectsWithAutoConfirm(t *testing.T, config *dao.Config, tempDir string) {
	configDir := filepath.Dir(config.Path)
	// Verify orphaned project exists before removal
	orphanedPath := filepath.Join(configDir, "orphaned-project")
	if _, err := os.Stat(orphanedPath); os.IsNotExist(err) {
		t.Fatalf("Orphaned project should exist before removal")
	}

	// Call the function with auto-confirmation (no user input required)
	err := removeOrphanedProjectsWithConfirm(config, config.ProjectList, false)
	if err != nil {
		t.Errorf("removeOrphanedProjectsWithConfirm should not error: %v", err)
		return
	}

	// Verify orphaned project was removed
	if _, err := os.Stat(orphanedPath); !os.IsNotExist(err) {
		t.Error("Orphaned project should have been removed")
	}

	// Verify active project still exists
	activePath := filepath.Join(configDir, "active-project")
	if _, err := os.Stat(activePath); os.IsNotExist(err) {
		t.Error("Active project should not have been removed")
	}

	// Verify non-git directory still exists
	nonGitPath := filepath.Join(configDir, "non-git-dir")
	if _, err := os.Stat(nonGitPath); os.IsNotExist(err) {
		t.Error("Non-git directory should not have been removed")
	}

	// Verify hidden directory still exists
	hiddenPath := filepath.Join(configDir, ".hidden-dir")
	if _, err := os.Stat(hiddenPath); os.IsNotExist(err) {
		t.Error("Hidden directory should not have been removed")
	}
}

func TestRemoveOrphanedProjectsNoOrphans(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mani-test-no-orphans-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock mani.yaml in the temp directory
	configPath := filepath.Join(tempDir, "mani.yaml")
	configContent := `projects:
  active-project:
    path: active-project
    url: https://github.com/example/active-project.git`

	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	configDir := filepath.Dir(configPath)
	// Create active project directory (should not be removed)
	activeProjectDir := filepath.Join(configDir, "active-project")
	err = os.MkdirAll(filepath.Join(activeProjectDir, ".git"), 0755)
	if err != nil {
		t.Fatalf("Failed to create active project: %v", err)
	}

	// Create config
	config := &dao.Config{
		Path: configPath,
		ProjectList: []dao.Project{
			{
				Name: "active-project",
				Path: "active-project",
				Url:  "https://github.com/example/active-project.git",
			},
		},
		RemoveOrphaned: &[]bool{false}[0],
	}

	// Call the function - should return without error and not remove anything
	err = removeOrphanedProjectsWithConfirm(config, config.ProjectList, false)
	if err != nil {
		t.Errorf("RemoveOrphanedProjects should not error when no orphans exist: %v", err)
	}

	// Verify active project still exists
	if _, err := os.Stat(activeProjectDir); os.IsNotExist(err) {
		t.Error("Active project directory should not have been removed")
	}
}
