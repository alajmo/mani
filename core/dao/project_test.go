package dao

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

func TestProject_GetValue(t *testing.T) {
	project := Project{
		Name:    "test-project",
		Path:    "/path/to/project",
		RelPath: "relative/path",
		Desc:    "Test description",
		URL:     "https://example.com",
		Tags:    []string{"frontend", "api"},
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "get project name",
			key:      "Project",
			expected: "test-project",
		},
		{
			name:     "get project path",
			key:      "Path",
			expected: "/path/to/project",
		},
		{
			name:     "get relative path",
			key:      "RelPath",
			expected: "relative/path",
		},
		{
			name:     "get description",
			key:      "Desc",
			expected: "Test description",
		},
		{
			name:     "get url",
			key:      "Url",
			expected: "https://example.com",
		},
		{
			name:     "get tags",
			key:      "Tag",
			expected: "frontend, api",
		},
		{
			name:     "get invalid key",
			key:      "InvalidKey",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := project.GetValue(tt.key, 0)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestProject_GetProjectsByName(t *testing.T) {
	config := Config{
		ProjectList: []Project{
			{Name: "project1", Path: "/path/1"},
			{Name: "project2", Path: "/path/2"},
			{Name: "project3", Path: "/path/3"},
		},
	}

	tests := []struct {
		name          string
		projectNames  []string
		expectError   bool
		expectedCount int
	}{
		{
			name:          "find existing projects",
			projectNames:  []string{"project1", "project2"},
			expectError:   false,
			expectedCount: 2,
		},
		{
			name:          "find non-existing project",
			projectNames:  []string{"project1", "nonexistent"},
			expectError:   true,
			expectedCount: 0,
		},
		{
			name:          "empty project names",
			projectNames:  []string{},
			expectError:   false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projects, err := config.GetProjectsByName(tt.projectNames)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(projects) != tt.expectedCount {
				t.Errorf("expected %d projects, got %d", tt.expectedCount, len(projects))
			}
			if err != nil && !tt.expectError {
				if _, ok := err.(*core.ProjectNotFound); !ok {
					t.Errorf("expected ProjectNotFound error, got %T", err)
				}
			}
		})
	}
}

func TestProject_GetProjectsByTags(t *testing.T) {
	config := Config{
		ProjectList: []Project{
			{Name: "project1", Tags: []string{"frontend", "react"}},
			{Name: "project2", Tags: []string{"backend", "api"}},
			{Name: "project3", Tags: []string{"frontend", "vue"}},
		},
	}

	tests := []struct {
		name          string
		tags          []string
		expectError   bool
		expectedNames []string
	}{
		{
			name:          "find projects with existing tag",
			tags:          []string{"frontend"},
			expectError:   false,
			expectedNames: []string{"project1", "project3"},
		},
		{
			name:          "find projects with multiple tags",
			tags:          []string{"frontend", "react"},
			expectError:   false,
			expectedNames: []string{"project1"},
		},
		{
			name:          "find projects with non-existing tag",
			tags:          []string{"nonexistent"},
			expectError:   true,
			expectedNames: []string{},
		},
		{
			name:          "empty tags",
			tags:          []string{},
			expectError:   false,
			expectedNames: []string{"project1", "project2", "project3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projects, err := config.GetProjectsByTags(tt.tags)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			gotNames := getProjectNames(projects)
			if !equalStringSlices(gotNames, tt.expectedNames) {
				t.Errorf("expected projects %v, got %v", tt.expectedNames, gotNames)
			}
		})
	}
}

func TestProject_GetProjectsByPath(t *testing.T) {
	config := Config{
		Dir: "/base",
		ProjectList: []Project{
			{Name: "project1", Path: "/base/frontend/app1", RelPath: "frontend/app1"},
			{Name: "project2", Path: "/base/backend/api", RelPath: "backend/api"},
			{Name: "project3", Path: "/base/frontend/app2", RelPath: "frontend/app2"},
			{Name: "project4", Path: "/base/frontend/nested/app3", RelPath: "frontend/nested/app3"},
		},
	}

	tests := []struct {
		name          string
		paths         []string
		expectError   bool
		expectedNames []string
	}{
		{
			name:          "find projects in frontend path",
			paths:         []string{"frontend"},
			expectError:   false,
			expectedNames: []string{"project1", "project3", "project4"},
		},
		{
			name:          "find projects with specific path",
			paths:         []string{"frontend/app1"},
			expectError:   false,
			expectedNames: []string{"project1"},
		},
		{
			name:          "find projects with single-level glob (1)",
			paths:         []string{"*/app*"},
			expectError:   false,
			expectedNames: []string{"project1", "project3"},
		},
		{
			name:          "find projects with single-level glob (2)",
			paths:         []string{"*/app?"},
			expectError:   false,
			expectedNames: []string{"project1", "project3"},
		},
		{
			name:          "find projects with double-star glob (1)",
			paths:         []string{"frontend/**/app*"},
			expectError:   false,
			expectedNames: []string{"project1", "project3", "project4"},
		},
		{
			name:          "find projects with double-star glob (2)",
			paths:         []string{"frontend/**/app?"},
			expectError:   false,
			expectedNames: []string{"project1", "project3", "project4"},
		},
		{
			name:          "find projects with double-star glob (3)",
			paths:         []string{"frontend/**/**/app?"},
			expectError:   false,
			expectedNames: []string{"project1", "project3", "project4"},
		},
		{
			name:          "find projects with double-star glob (4)",
			paths:         []string{"**/app?"},
			expectError:   false,
			expectedNames: []string{"project1", "project3", "project4"},
		},
		{
			name:          "find projects with non-existing path",
			paths:         []string{"nonexistent"},
			expectError:   true,
			expectedNames: []string{},
		},
		{
			name:          "empty paths",
			paths:         []string{},
			expectError:   false,
			expectedNames: []string{"project1", "project2", "project3", "project4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projects, err := config.GetProjectsByPath(tt.paths)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			gotNames := getProjectNames(projects)
			if !equalStringSlices(gotNames, tt.expectedNames) {
				t.Errorf("expected projects %v, got %v", tt.expectedNames, gotNames)
			}
		})
	}
}

func TestProject_TestAddToTree(t *testing.T) {
	tests := []struct {
		name          string
		nodes         []TNode
		expectedPaths []string
	}{
		{
			name: "simple tree",
			nodes: []TNode{
				{Name: "app1", Path: "frontend/app1"},
				{Name: "app2", Path: "frontend/app2"},
				{Name: "api", Path: "backend/api"},
			},
			expectedPaths: []string{"frontend", "backend"},
		},
		{
			name: "nested tree",
			nodes: []TNode{
				{Name: "app1", Path: "frontend/web/app1"},
				{Name: "app2", Path: "frontend/mobile/app2"},
			},
			expectedPaths: []string{"frontend"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tree []TreeNode
			for _, node := range tt.nodes {
				tree = AddToTree(tree, node)
			}

			paths := getTreePaths(tree)
			if !equalStringSlices(paths, tt.expectedPaths) {
				t.Errorf("expected paths %v, got %v", tt.expectedPaths, paths)
			}
		})
	}
}

func TestProject_GetIntersectProjects(t *testing.T) {
	config := Config{
		ProjectList: []Project{
			{Name: "project1", Tags: []string{"frontend"}},
			{Name: "project2", Tags: []string{"backend"}},
			{Name: "project3", Tags: []string{"frontend", "api"}},
		},
	}

	tests := []struct {
		name          string
		inputs        [][]Project
		expectedNames []string
	}{
		{
			name: "intersect frontend and api projects",
			inputs: [][]Project{
				{{Name: "project1"}, {Name: "project3"}}, // frontend projects
				{{Name: "project3"}},                     // api projects
			},
			expectedNames: []string{"project3"},
		},
		{
			name: "no intersection",
			inputs: [][]Project{
				{{Name: "project1"}},
				{{Name: "project2"}},
			},
			expectedNames: []string{},
		},
		{
			name:          "empty input",
			inputs:        [][]Project{},
			expectedNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.GetIntersectProjects(tt.inputs...)

			gotNames := getProjectNames(result)
			if !equalStringSlices(gotNames, tt.expectedNames) {
				t.Errorf("expected projects %v, got %v", tt.expectedNames, gotNames)
			}
		})
	}
}

func TestParseWorktrees(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		expected    []Worktree
		expectError bool
	}{
		{
			name: "worktree with path and branch",
			yaml: `
- path: feature-branch
  branch: feature/awesome
`,
			expected: []Worktree{
				{Path: "feature-branch", Branch: "feature/awesome"},
			},
		},
		{
			name: "multiple worktrees",
			yaml: `
- path: feature-branch
  branch: feature/awesome
- path: staging
  branch: staging
`,
			expected: []Worktree{
				{Path: "feature-branch", Branch: "feature/awesome"},
				{Path: "staging", Branch: "staging"},
			},
		},
		{
			name: "worktree without branch defaults to path basename",
			yaml: `
- path: hotfix
`,
			expected: []Worktree{
				{Path: "hotfix", Branch: "hotfix"},
			},
		},
		{
			name: "worktree with nested path defaults branch to basename",
			yaml: `
- path: worktrees/feature
`,
			expected: []Worktree{
				{Path: "worktrees/feature", Branch: "feature"},
			},
		},
		{
			name:     "empty worktrees",
			yaml:     ``,
			expected: []Worktree{},
		},
		{
			name: "worktree without path returns error",
			yaml: `
- branch: feat
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node yaml.Node
			if err := yaml.Unmarshal([]byte(tt.yaml), &node); err != nil {
				t.Fatalf("failed to parse yaml: %v", err)
			}

			// Handle empty YAML case
			if len(node.Content) == 0 {
				result, err := ParseWorktrees(yaml.Node{})
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(result) != 0 {
					t.Errorf("expected empty worktrees, got %v", result)
				}
				return
			}

			result, err := ParseWorktrees(*node.Content[0])

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d worktrees, got %d", len(tt.expected), len(result))
				return
			}

			for i, wt := range result {
				if wt.Path != tt.expected[i].Path {
					t.Errorf("worktree[%d].Path: expected %q, got %q", i, tt.expected[i].Path, wt.Path)
				}
				if wt.Branch != tt.expected[i].Branch {
					t.Errorf("worktree[%d].Branch: expected %q, got %q", i, tt.expected[i].Branch, wt.Branch)
				}
			}
		})
	}
}

func TestProject_GetValue_Worktrees(t *testing.T) {
	tests := []struct {
		name     string
		project  Project
		expected string
	}{
		{
			name: "project with worktrees",
			project: Project{
				Name: "test-project",
				WorktreeList: []Worktree{
					{Path: "feature", Branch: "feature/test"},
					{Path: "staging", Branch: "staging"},
				},
			},
			expected: "feature/test, staging",
		},
		{
			name: "project without worktrees",
			project: Project{
				Name:         "test-project",
				WorktreeList: []Worktree{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.project.GetValue("worktrees", 0)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
