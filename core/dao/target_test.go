package dao

import (
	"reflect"
	"testing"

	"github.com/alajmo/mani/core"
)

func TestTarget_GetContext(t *testing.T) {
	target := Target{
		Name:        "test-target",
		context:     "/path/to/config",
		contextLine: 42,
	}

	if target.GetContext() != "/path/to/config" {
		t.Errorf("expected context '/path/to/config', got %q", target.GetContext())
	}

	if target.GetContextLine() != 42 {
		t.Errorf("expected context line 42, got %d", target.GetContextLine())
	}
}

func TestTarget_GetTargetList(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		expectedCount int
		expectError   bool
	}{
		{
			name: "empty target list",
			config: Config{
				TargetList: []Target{},
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "multiple valid targets",
			config: Config{
				TargetList: []Target{
					{
						Name:     "target1",
						Projects: []string{"proj1", "proj2"},
						Tags:     []string{"frontend"},
					},
					{
						Name:     "target2",
						Projects: []string{"proj3"},
						Tags:     []string{"backend"},
					},
				},
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "target with all flag",
			config: Config{
				TargetList: []Target{
					{
						Name: "all-target",
						All:  true,
					},
				},
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "target with paths",
			config: Config{
				TargetList: []Target{
					{
						Name:  "path-target",
						Paths: []string{"path1", "path2"},
					},
				},
			},
			expectedCount: 1,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targets := tt.config.TargetList

			if len(targets) != tt.expectedCount {
				t.Errorf("expected %d targets, got %d", tt.expectedCount, len(targets))
			}
		})
	}
}

func TestTarget_GetTarget(t *testing.T) {
	config := Config{
		TargetList: []Target{
			{
				Name:     "frontend",
				Projects: []string{"web", "mobile"},
				Tags:     []string{"frontend"},
			},
			{
				Name:     "backend",
				Projects: []string{"api", "worker"},
				Tags:     []string{"backend"},
			},
		},
	}

	tests := []struct {
		name         string
		targetName   string
		expectError  bool
		expectedTags []string
	}{
		{
			name:         "existing target",
			targetName:   "frontend",
			expectError:  false,
			expectedTags: []string{"frontend"},
		},
		{
			name:         "non-existing target",
			targetName:   "nonexistent",
			expectError:  true,
			expectedTags: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := config.GetTarget(tt.targetName)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if _, ok := err.(*core.TargetNotFound); !ok {
					t.Errorf("expected TargetNotFound error, got %T", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(target.Tags, tt.expectedTags) {
				t.Errorf("expected tags %v, got %v", tt.expectedTags, target.Tags)
			}
		})
	}
}

func TestTarget_GetTargetNames(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		expectedNames []string
	}{
		{
			name: "multiple targets",
			config: Config{
				TargetList: []Target{
					{Name: "target1"},
					{Name: "target2"},
					{Name: "target3"},
				},
			},
			expectedNames: []string{"target1", "target2", "target3"},
		},
		{
			name: "empty target list",
			config: Config{
				TargetList: []Target{},
			},
			expectedNames: []string{},
		},
		{
			name: "single target",
			config: Config{
				TargetList: []Target{
					{Name: "solo-target"},
				},
			},
			expectedNames: []string{"solo-target"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names := tt.config.GetTargetNames()

			if !reflect.DeepEqual(names, tt.expectedNames) {
				t.Errorf("expected names %v, got %v", tt.expectedNames, names)
			}
		})
	}
}
