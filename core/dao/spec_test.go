package dao

import (
	"reflect"
	"testing"

	"github.com/alajmo/mani/core"
)

func TestSpec_GetContext(t *testing.T) {
	spec := Spec{
		Name:        "test-spec",
		context:     "/path/to/config",
		contextLine: 42,
	}

	if spec.GetContext() != "/path/to/config" {
		t.Errorf("expected context '/path/to/config', got %q", spec.GetContext())
	}

	if spec.GetContextLine() != 42 {
		t.Errorf("expected context line 42, got %d", spec.GetContextLine())
	}
}

func TestSpec_GetSpecList(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		expectedCount int
		expectError   bool
	}{
		{
			name: "empty spec list",
			config: Config{
				SpecList: []Spec{},
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "valid specs",
			config: Config{
				SpecList: []Spec{
					{
						Name:     "spec1",
						Output:   "table",
						Parallel: true,
						Forks:    4,
					},
					{
						Name:   "spec2",
						Output: "stream",
						Forks:  8,
					},
				},
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "spec with defaults",
			config: Config{
				SpecList: []Spec{
					{
						Name:   "default-spec",
						Output: "table",
						Forks:  4,
					},
				},
			},
			expectedCount: 1,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specs := tt.config.SpecList

			if len(specs) != tt.expectedCount {
				t.Errorf("expected %d specs, got %d", tt.expectedCount, len(specs))
			}
		})
	}
}

func TestSpec_GetSpec(t *testing.T) {
	config := Config{
		SpecList: []Spec{
			{
				Name:   "spec1",
				Output: "table",
				Forks:  4,
			},
			{
				Name:   "spec2",
				Output: "stream",
				Forks:  8,
			},
		},
	}

	tests := []struct {
		name          string
		specName      string
		expectError   bool
		expectedForks uint32
	}{
		{
			name:          "existing spec",
			specName:      "spec1",
			expectError:   false,
			expectedForks: 4,
		},
		{
			name:          "non-existing spec",
			specName:      "nonexistent",
			expectError:   true,
			expectedForks: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := config.GetSpec(tt.specName)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if _, ok := err.(*core.SpecNotFound); !ok {
					t.Errorf("expected SpecNotFound error, got %T", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if spec.Forks != tt.expectedForks {
				t.Errorf("expected forks %d, got %d", tt.expectedForks, spec.Forks)
			}
		})
	}
}

func TestSpec_GetSpecNames(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		expectedNames []string
	}{
		{
			name: "multiple specs",
			config: Config{
				SpecList: []Spec{
					{Name: "spec1"},
					{Name: "spec2"},
					{Name: "spec3"},
				},
			},
			expectedNames: []string{"spec1", "spec2", "spec3"},
		},
		{
			name: "empty spec list",
			config: Config{
				SpecList: []Spec{},
			},
			expectedNames: []string{},
		},
		{
			name: "single spec",
			config: Config{
				SpecList: []Spec{
					{Name: "solo-spec"},
				},
			},
			expectedNames: []string{"solo-spec"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names := tt.config.GetSpecNames()

			if !reflect.DeepEqual(names, tt.expectedNames) {
				t.Errorf("expected names %v, got %v", tt.expectedNames, names)
			}
		})
	}
}
