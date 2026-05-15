package exec

import (
	"bytes"
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPrintJSON(t *testing.T) {
	tests := []struct {
		name     string
		results  []TaskResult
		wantErr  bool
		validate func(t *testing.T, output []byte)
	}{
		{
			name: "single result",
			results: []TaskResult{
				{
					Project:  "project1",
					Tasks:    []string{"task1"},
					Output:   []string{"line1", "line2"},
					ExitCode: 0,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, output []byte) {
				var results []TaskResult
				if err := json.Unmarshal(output, &results); err != nil {
					t.Errorf("failed to unmarshal JSON: %v", err)
					return
				}
				if len(results) != 1 {
					t.Errorf("expected 1 result, got %d", len(results))
					return
				}
				if results[0].Project != "project1" {
					t.Errorf("expected project 'project1', got %q", results[0].Project)
				}
				if results[0].ExitCode != 0 {
					t.Errorf("expected exit code 0, got %d", results[0].ExitCode)
				}
			},
		},
		{
			name: "multiple results",
			results: []TaskResult{
				{
					Project:  "project1",
					Tasks:    []string{"task1"},
					Output:   []string{"success"},
					ExitCode: 0,
				},
				{
					Project:  "project2",
					Tasks:    []string{"task1"},
					Output:   []string{"error", "failed"},
					ExitCode: 1,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, output []byte) {
				var results []TaskResult
				if err := json.Unmarshal(output, &results); err != nil {
					t.Errorf("failed to unmarshal JSON: %v", err)
					return
				}
				if len(results) != 2 {
					t.Errorf("expected 2 results, got %d", len(results))
					return
				}
				if results[1].ExitCode != 1 {
					t.Errorf("expected exit code 1 for second result, got %d", results[1].ExitCode)
				}
			},
		},
		{
			name: "multiple tasks",
			results: []TaskResult{
				{
					Project:  "project1",
					Tasks:    []string{"echo", "pwd"},
					Output:   []string{"hello", "/home/user"},
					ExitCode: 0,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, output []byte) {
				var results []TaskResult
				if err := json.Unmarshal(output, &results); err != nil {
					t.Errorf("failed to unmarshal JSON: %v", err)
					return
				}
				if len(results) != 1 {
					t.Errorf("expected 1 result, got %d", len(results))
					return
				}
				if len(results[0].Tasks) != 2 {
					t.Errorf("expected 2 tasks, got %d", len(results[0].Tasks))
					return
				}
				if results[0].Tasks[0] != "echo" || results[0].Tasks[1] != "pwd" {
					t.Errorf("expected tasks [echo, pwd], got %v", results[0].Tasks)
				}
			},
		},
		{
			name:    "empty results",
			results: []TaskResult{},
			wantErr: false,
			validate: func(t *testing.T, output []byte) {
				var results []TaskResult
				if err := json.Unmarshal(output, &results); err != nil {
					t.Errorf("failed to unmarshal JSON: %v", err)
					return
				}
				if len(results) != 0 {
					t.Errorf("expected 0 results, got %d", len(results))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintJSON(tt.results, &buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrintJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil {
				tt.validate(t, buf.Bytes())
			}
		})
	}
}

func TestPrintYAML(t *testing.T) {
	tests := []struct {
		name     string
		results  []TaskResult
		wantErr  bool
		validate func(t *testing.T, output []byte)
	}{
		{
			name: "single result",
			results: []TaskResult{
				{
					Project:  "project1",
					Tasks:    []string{"task1"},
					Output:   []string{"line1", "line2"},
					ExitCode: 0,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, output []byte) {
				var result TaskResult
				decoder := yaml.NewDecoder(bytes.NewReader(output))
				if err := decoder.Decode(&result); err != nil {
					t.Errorf("failed to unmarshal YAML: %v", err)
					return
				}
				if result.Project != "project1" {
					t.Errorf("expected project 'project1', got %q", result.Project)
				}
				if result.ExitCode != 0 {
					t.Errorf("expected exit code 0, got %d", result.ExitCode)
				}
			},
		},
		{
			name: "multiple results as YAML documents",
			results: []TaskResult{
				{
					Project:  "project1",
					Tasks:    []string{"task1"},
					Output:   []string{"success"},
					ExitCode: 0,
				},
				{
					Project:  "project2",
					Tasks:    []string{"task1"},
					Output:   []string{"error"},
					ExitCode: 1,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, output []byte) {
				var results []TaskResult
				decoder := yaml.NewDecoder(bytes.NewReader(output))
				for {
					var result TaskResult
					if err := decoder.Decode(&result); err != nil {
						break
					}
					results = append(results, result)
				}
				if len(results) != 2 {
					t.Errorf("expected 2 results, got %d", len(results))
					return
				}
				if results[1].ExitCode != 1 {
					t.Errorf("expected exit code 1 for second result, got %d", results[1].ExitCode)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintYAML(tt.results, &buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrintYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil {
				tt.validate(t, buf.Bytes())
			}
		})
	}
}

func TestPrintJSONStream(t *testing.T) {
	tests := []struct {
		name    string
		result  TaskResult
		wantErr bool
	}{
		{
			name: "single result streamed",
			result: TaskResult{
				Project:  "project1",
				Tasks:    []string{"task1"},
				Output:   []string{"output line"},
				ExitCode: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintJSONStream(tt.result, &buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrintJSONStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Check that output is a single line (streaming format)
			output := buf.String()
			if len(output) == 0 {
				t.Error("expected non-empty output")
				return
			}
			if output[len(output)-1] != '\n' {
				t.Error("expected output to end with newline")
			}
			// Verify it's valid JSON
			var result TaskResult
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Errorf("output is not valid JSON: %v", err)
			}
		})
	}
}

func TestPrintYAMLStream(t *testing.T) {
	tests := []struct {
		name    string
		result  TaskResult
		wantErr bool
	}{
		{
			name: "single result streamed",
			result: TaskResult{
				Project:  "project1",
				Tasks:    []string{"task1"},
				Output:   []string{"output line"},
				ExitCode: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintYAMLStream(tt.result, &buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrintYAMLStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Verify it's valid YAML
			var result TaskResult
			if err := yaml.Unmarshal(buf.Bytes(), &result); err != nil {
				t.Errorf("output is not valid YAML: %v", err)
			}
			if result.Project != tt.result.Project {
				t.Errorf("expected project %q, got %q", tt.result.Project, result.Project)
			}
		})
	}
}
