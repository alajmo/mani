package dao

import (
	"reflect"
	"sort"
	"testing"

	"github.com/alajmo/mani/core"
)

func TestTask_ParseTask(t *testing.T) {
	config := Config{
		Shell: "sh -c",
		SpecList: []Spec{
			DEFAULT_SPEC,
		},
		TargetList: []Target{
			DEFAULT_TARGET,
		},
		ThemeList: []Theme{
			DEFAULT_THEME,
		},
	}

	tests := []struct {
		name          string
		task          Task
		expectError   bool
		expectedShell string
	}{
		{
			name: "basic task parsing",
			task: Task{
				Name:       "test-task",
				Cmd:        "echo hello",
				SpecData:   DEFAULT_SPEC,
				TargetData: DEFAULT_TARGET,
				ThemeData:  DEFAULT_THEME,
			},
			expectError:   false,
			expectedShell: "sh -c",
		},
		{
			name: "custom shell",
			task: Task{
				Name:       "node-task",
				Shell:      "node -e",
				Cmd:        "console.log('hello')",
				SpecData:   DEFAULT_SPEC,
				TargetData: DEFAULT_TARGET,
				ThemeData:  DEFAULT_THEME,
			},
			expectError:   false,
			expectedShell: "node -e",
		},
		{
			name: "with commands",
			task: Task{
				Name: "multi-cmd",
				Commands: []Command{
					{
						Name: "cmd1",
						Cmd:  "echo first",
					},
					{
						Name: "cmd2",
						Cmd:  "echo second",
					},
				},
				SpecData:   DEFAULT_SPEC,
				TargetData: DEFAULT_TARGET,
				ThemeData:  DEFAULT_THEME,
			},
			expectError:   false,
			expectedShell: "sh -c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskErrors := &ResourceErrors[Task]{}
			tt.task.ParseTask(config, taskErrors)

			if tt.expectError && len(taskErrors.Errors) == 0 {
				t.Error("expected errors but got none")
			}
			if !tt.expectError && len(taskErrors.Errors) > 0 {
				t.Errorf("unexpected errors: %v", taskErrors.Errors)
			}
			if tt.task.Shell != tt.expectedShell {
				t.Errorf("expected shell %q, got %q", tt.expectedShell, tt.task.Shell)
			}
		})
	}
}

func TestTask_GetTaskProjects(t *testing.T) {
	config := Config{
		Shell: DEFAULT_SHELL,
		ProjectList: []Project{
			{Name: "proj1", Tags: []string{"frontend"}},
			{Name: "proj2", Tags: []string{"backend"}},
			{Name: "proj3", Tags: []string{"frontend", "api"}},
		},
		SpecList: []Spec{
			DEFAULT_SPEC,
		},
		TargetList: []Target{
			DEFAULT_TARGET,
		},
		ThemeList: []Theme{
			DEFAULT_THEME,
		},
	}

	tests := []struct {
		name          string
		task          *Task
		flags         *core.RunFlags
		setFlags      *core.SetRunFlags
		expectedCount int
		expectError   bool
	}{
		{
			name: "filter by tags",
			task: &Task{
				Name:  "test-task",
				Shell: DEFAULT_SHELL,
				TargetData: Target{
					Name: "default",
					Tags: []string{"frontend"},
				},
				SpecData:  DEFAULT_SPEC,
				ThemeData: DEFAULT_THEME,
			},
			flags:         &core.RunFlags{},
			setFlags:      &core.SetRunFlags{},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "filter by projects",
			task: &Task{
				Name: "test-task",
				TargetData: Target{
					Name:     DEFAULT_TARGET.Name,
					Projects: []string{"proj1", "proj2"},
				},
				SpecData:  DEFAULT_SPEC,
				ThemeData: DEFAULT_THEME,
			},
			flags:         &core.RunFlags{},
			setFlags:      &core.SetRunFlags{},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "override with flag projects",
			task: &Task{
				Name: "test-task",
				TargetData: Target{
					Name:     DEFAULT_TARGET.Name,
					Projects: []string{"proj1"},
				},
				SpecData:  DEFAULT_SPEC,
				ThemeData: DEFAULT_THEME,
			},
			flags: &core.RunFlags{
				Projects: []string{"proj2", "proj3"},
			},
			setFlags:      &core.SetRunFlags{},
			expectedCount: 2,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projects, err := config.GetTaskProjects(tt.task, tt.flags, tt.setFlags)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(projects) != tt.expectedCount {
				t.Errorf("expected %d projects, got %d", tt.expectedCount, len(projects))
			}
		})
	}
}

func TestTask_CmdParse(t *testing.T) {
	config := &Config{
		Shell: DEFAULT_SHELL,
		ProjectList: []Project{
			{Name: "test-project", Path: "/test/path"},
		},
		SpecList: []Spec{
			DEFAULT_SPEC,
		},
		TargetList: []Target{
			DEFAULT_TARGET,
		},
		ThemeList: []Theme{
			DEFAULT_THEME,
		},
	}

	tests := []struct {
		name           string
		cmd            string
		runFlags       *core.RunFlags
		setFlags       *core.SetRunFlags
		expectTasks    int
		expectProjects int
		expectError    bool
	}{
		{
			name: "basic command",
			cmd:  "echo hello",
			runFlags: &core.RunFlags{
				Target:   "default",
				Projects: []string{"test-project"},
			},
			setFlags:       &core.SetRunFlags{},
			expectTasks:    1,
			expectProjects: 1,
			expectError:    false,
		},
		{
			name: "command with no matching projects",
			cmd:  "echo hello",
			runFlags: &core.RunFlags{
				Projects: []string{"non-existent"},
				Target:   "default",
			},
			setFlags:       &core.SetRunFlags{},
			expectTasks:    0,
			expectProjects: 0,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, projects, err := ParseCmd(tt.cmd, tt.runFlags, tt.setFlags, config)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(tasks) != tt.expectTasks {
				t.Errorf("expected %d tasks, got %d", tt.expectTasks, len(tasks))
			}
			if len(projects) != tt.expectProjects {
				t.Errorf("expected %d projects, got %d", tt.expectProjects, len(projects))
			}
		})
	}
}

func TestConfig_FilterProjects(t *testing.T) {
	// Setup test configuration with sample projects
	config := Config{
		ProjectList: []Project{
			{Name: "root", Path: "/path", RelPath: ".", Tags: []string{}},
			{Name: "frontend", Path: "/path/frontend", RelPath: "frontend", Tags: []string{"web", "ui"}},
			{Name: "backend", Path: "/path/backend", RelPath: "backend", Tags: []string{"api", "db"}},
			{Name: "mobile", Path: "/path/mobile", RelPath: "mobile", Tags: []string{"ui", "app"}},
			{Name: "docs", Path: "/path/docs", RelPath: "docs", Tags: []string{"docs"}},
			{Name: "shared", Path: "/path/shared", RelPath: "shared", Tags: []string{"lib", "shared"}},
		},
	}

	tests := []struct {
		name             string
		cwdFlag          bool
		allProjectsFlag  bool
		projectsFlag     []string
		projectPathsFlag []string
		tagsFlag         []string
		tagsExprFlag     string
		expectedCount    int
		expectedNames    []string
		expectError      bool
	}{
		{
			name:            "single project",
			allProjectsFlag: true,
			projectsFlag:    []string{"frontend"},
			tagsFlag:        []string{"ui"},
			expectedCount:   1,
			expectedNames:   []string{"frontend"},
			expectError:     false,
		},
		{
			name:          "filter by project names",
			projectsFlag:  []string{"frontend", "backend"},
			expectedCount: 2,
			expectedNames: []string{"frontend", "backend"},
			expectError:   false,
		},
		{
			name:             "partial path matching",
			projectPathsFlag: []string{"front"}, // Should match 'frontend'
			expectedCount:    1,
			expectedNames:    []string{"frontend"},
			expectError:      false,
		},
		{
			name:          "filter by single tag",
			tagsFlag:      []string{"ui"},
			expectedCount: 2,
			expectedNames: []string{"frontend", "mobile"},
			expectError:   false,
		},
		{
			name:          "filter by multiple tags - intersection",
			tagsFlag:      []string{"ui", "web"},
			expectedCount: 1,
			expectedNames: []string{"frontend"},
			expectError:   false,
		},
		{
			name:             "filter by project paths",
			projectPathsFlag: []string{"frontend"},
			expectedCount:    1,
			expectedNames:    []string{"frontend"},
			expectError:      false,
		},
		{
			name:          "filter by tags expression",
			tagsExprFlag:  "ui && !web",
			expectedCount: 1,
			expectedNames: []string{"mobile"},
			expectError:   false,
		},
		{
			name:          "multiple criteria - intersection",
			projectsFlag:  []string{"frontend", "mobile", "backend"},
			tagsFlag:      []string{"ui"},
			expectedCount: 2,
			expectedNames: []string{"frontend", "mobile"},
			expectError:   false,
		},
		{
			name:          "non-existent project name",
			projectsFlag:  []string{"nonexistent"},
			expectedCount: 0,
			expectError:   true,
		},
		{
			name:          "non-existent tag",
			tagsFlag:      []string{"nonexistent"},
			expectedCount: 0,
			expectedNames: []string{},
			expectError:   true,
		},
		{
			name:          "invalid tags expression",
			tagsExprFlag:  "ui && (NOT", // Invalid syntax
			expectedCount: 0,
			expectError:   true,
		},
		{
			name:          "cwd flag with other flags",
			cwdFlag:       true,
			projectsFlag:  []string{"root"},
			expectedCount: 1,
			expectedNames: []string{""},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projects, err := config.FilterProjects(
				tt.cwdFlag,
				tt.allProjectsFlag,
				tt.projectsFlag,
				tt.projectPathsFlag,
				tt.tagsFlag,
				tt.tagsExprFlag,
			)

			// Check error expectations
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Skip further checks if we expected an error
			if tt.expectError {
				return
			}

			// Check number of projects returned
			if len(projects) != tt.expectedCount {
				t.Errorf("expected %d projects, got %d", tt.expectedCount, len(projects))
			}

			// Check specific projects returned (if specified)
			if tt.expectedNames != nil {
				actualNames := make([]string, len(projects))
				for i, p := range projects {
					actualNames[i] = p.Name
				}

				// Sort both slices to ensure consistent comparison
				sort.Strings(actualNames)
				sort.Strings(tt.expectedNames)

				if !reflect.DeepEqual(actualNames, tt.expectedNames) {
					t.Errorf("expected projects %v, got %v", tt.expectedNames, actualNames)
				}
			}
		})
	}
}
