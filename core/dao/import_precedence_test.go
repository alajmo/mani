package dao

import (
	"maps"
	"testing"
)

func TestImportGraph_Closure(t *testing.T) {
	closure := ImportGraph{
		"a": {"b"},
		"b": {"c"},
		"d": {"a", "d"},
	}.closure()

	tests := []struct {
		a, b     string
		expected bool
	}{
		{"a", "b", true},
		{"a", "c", true},
		{"b", "a", false},
		{"c", "a", false},
		{"a", "a", false},
		{"d", "c", true},
		{"d", "d", true}, // Cyclic imports are reported separately
		{"x", "a", false},
	}
	for _, tt := range tests {
		if got := closure.imports(tt.a, tt.b); got != tt.expected {
			t.Errorf("imports(%s, %s): expected %v, got %v", tt.a, tt.b, tt.expected, got)
		}
	}
}

func TestImportGraph_WithUserConfigs(t *testing.T) {
	// The user config is an invisible import of every other config, and its own imports of other
	// configs don't count
	imports := ImportGraph{
		"root":   {"shared", "user"},
		"shared": {"lib"},
		"user":   {"shared", "extra", "root"},
	}.withUserConfigs(map[string]bool{"user": true, "extra": true})

	tests := []struct {
		a, b     string
		expected bool
	}{
		{"root", "shared", true},
		{"shared", "lib", true},
		{"shared", "user", true},
		{"lib", "extra", true},
		{"user", "extra", true},
		{"user", "shared", false},
		{"extra", "user", false},
		{"shared", "root", false},
		{"root", "root", false},
		{"user", "user", false},
	}
	for _, tt := range tests {
		if got := imports(tt.a, tt.b); got != tt.expected {
			t.Errorf("imports(%s, %s): expected %v, got %v", tt.a, tt.b, tt.expected, got)
		}
	}
}

func TestImportGraph_UserConfigs(t *testing.T) {
	// The user configs are the configs the root config only imports through the user config
	graph := ImportGraph{
		"root":   {"a", "shared"},
		"a":      {"user"},
		"user":   {"shared", "extra", "root"},
		"extra":  {"nested"},
		"shared": {"lib"},
	}

	expected := map[string]bool{"user": true, "extra": true, "nested": true}
	if got := graph.userConfigs("root", "user"); !maps.Equal(got, expected) {
		t.Errorf("expected user configs %v, got %v", expected, got)
	}

	if got := graph.userConfigs("root", "root"); len(got) != 0 {
		t.Errorf("expected no user configs when the user config is the root config, got %v", got)
	}
}
