package dao

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEnv_ParseNodeEnv(t *testing.T) {
	tests := []struct {
		name     string
		input    yaml.Node
		expected []string
	}{
		{
			name: "basic env variables",
			input: yaml.Node{
				Content: []*yaml.Node{
					{Value: "KEY1"},
					{Value: "value1"},
					{Value: "KEY2"},
					{Value: "value2"},
				},
			},
			expected: []string{
				"KEY1=value1",
				"KEY2=value2",
			},
		},
		{
			name: "empty env",
			input: yaml.Node{
				Content: []*yaml.Node{},
			},
			expected: []string{},
		},
		{
			name: "env with special characters",
			input: yaml.Node{
				Content: []*yaml.Node{
					{Value: "PATH"},
					{Value: "/usr/bin:/bin"},
					{Value: "URL"},
					{Value: "http://example.com"},
				},
			},
			expected: []string{
				"PATH=/usr/bin:/bin",
				"URL=http://example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseNodeEnv(tt.input)
			if !equalStringSlices(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEnv_MergeEnvs(t *testing.T) {
	tests := []struct {
		name     string
		inputs   [][]string
		expected []string
	}{
		{
			name: "basic merge",
			inputs: [][]string{
				{"KEY1=value1", "KEY2=value2"},
				{"KEY3=value3"},
			},
			expected: []string{
				"KEY1=value1",
				"KEY2=value2",
				"KEY3=value3",
			},
		},
		{
			name: "override priority",
			inputs: [][]string{
				{"KEY1=override"},
				{"KEY1=original", "KEY2=value2"},
			},
			expected: []string{
				"KEY1=override",
				"KEY2=value2",
			},
		},
		{
			name: "empty inputs",
			inputs: [][]string{
				{},
				{},
			},
			expected: []string{},
		},
		{
			name: "with newline characters",
			inputs: [][]string{
				{"KEY1=value1\n", "KEY2=value2\n"},
				{"KEY3=value3\n"},
			},
			expected: []string{
				"KEY1=value1",
				"KEY2=value2",
				"KEY3=value3",
			},
		},
		{
			name: "complex values",
			inputs: [][]string{
				{"PATH=/usr/bin:/bin", "URL=http://example.com"},
				{"DEBUG=true", "PATH=/custom/path"},
			},
			expected: []string{
				"PATH=/usr/bin:/bin",
				"URL=http://example.com",
				"DEBUG=true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeEnvs(tt.inputs...)
			if !equalStringSlices(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
