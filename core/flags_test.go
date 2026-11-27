package core

import (
	"errors"
	"testing"
)

func TestCheckOutput(t *testing.T) {
	tests := []struct {
		output  string
		valid   []string
		wantErr string
	}{
		{output: "json", valid: RunOutputs},
		{output: "jsonl", valid: RunOutputs},
		{output: "stream", valid: RunOutputs},
		{output: "yaml", valid: ListOutputs},
		{output: "ndjson", valid: RunOutputs, wantErr: "invalid output `ndjson`, expected one of: stream, table, html, markdown, json, jsonl, yaml"},
		{output: "jsonl", valid: ListOutputs, wantErr: "invalid output `jsonl`, expected one of: table, markdown, html, json, yaml"},
		{output: "JSON", valid: ListOutputs, wantErr: "invalid output `JSON`, expected one of: table, markdown, html, json, yaml"},
		{output: "stream", valid: ListOutputs, wantErr: "invalid output `stream`, expected one of: table, markdown, html, json, yaml"},
		{output: "", valid: RunOutputs, wantErr: "invalid output ``, expected one of: stream, table, html, markdown, json, jsonl, yaml"},
	}

	for _, tt := range tests {
		err := CheckOutput(tt.output, tt.valid)
		if tt.wantErr == "" {
			if err != nil {
				t.Errorf("%q: unexpected error: %v", tt.output, err)
			}
			continue
		}

		var outputErr *OutputError
		if !errors.As(err, &outputErr) || err.Error() != tt.wantErr {
			t.Errorf("%q: got error %v, want %q", tt.output, err, tt.wantErr)
		}
	}
}
