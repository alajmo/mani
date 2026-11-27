package print

import (
	"bytes"
	"sync"
	"testing"

	"github.com/alajmo/mani/core/dao"
)

type record struct {
	Name  string   `json:"name" yaml:"name"`
	Lines []string `json:"lines" yaml:"lines"`
}

func TestPrintStructured(t *testing.T) {
	records := []record{
		{Name: "a", Lines: []string{"<b> & c"}},
		{Name: "b", Lines: []string{}},
	}

	tests := []struct {
		name   string
		items  []record
		output string
		want   string
	}{
		{
			name:   "json is an indented array without html escaping",
			items:  records,
			output: "json",
			want: `[
  {
    "name": "a",
    "lines": [
      "<b> & c"
    ]
  },
  {
    "name": "b",
    "lines": []
  }
]
`,
		},
		{
			name:   "json without items is an empty array",
			items:  nil,
			output: "json",
			want:   "[]\n",
		},
		{
			name:   "yaml is one document per item",
			items:  records,
			output: "yaml",
			want: `name: a
lines:
  - <b> & c
---
name: b
lines: []
`,
		},
		{
			name:   "yaml without items is empty",
			items:  nil,
			output: "yaml",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintStructured(tt.items, tt.output, &buf); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestStructuredStream(t *testing.T) {
	tests := []struct {
		output string
		want   string
	}{
		{
			output: "json",
			want: `{"name":"a","lines":["<b> & c"]}
{"name":"b","lines":[]}
`,
		},
		{
			output: "yaml",
			want: `name: a
lines:
  - <b> & c
---
name: b
lines: []
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.output, func(t *testing.T) {
			var buf bytes.Buffer
			stream := NewStructuredStream(tt.output, &buf)
			for _, r := range []record{{Name: "a", Lines: []string{"<b> & c"}}, {Name: "b", Lines: []string{}}} {
				if err := stream.Write(r); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
			if err := stream.Close(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

// Records written concurrently must not be interleaved.
func TestStructuredStreamConcurrent(t *testing.T) {
	var buf bytes.Buffer
	stream := NewStructuredStream("json", &buf)

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = stream.Write(record{Name: "a", Lines: []string{"1", "2"}})
		}()
	}
	wg.Wait()

	want := `{"name":"a","lines":["1","2"]}` + "\n"
	lines := bytes.SplitAfter(buf.Bytes(), []byte("\n"))
	lines = lines[:len(lines)-1] // Remainder after the last newline
	if len(lines) != 50 {
		t.Fatalf("got %d lines, want 50", len(lines))
	}
	for _, line := range lines {
		if string(line) != want {
			t.Fatalf("got line %q, want %q", line, want)
		}
	}
}

func TestListOutput(t *testing.T) {
	var buf bytes.Buffer

	projects := ProjectsOutput([]dao.Project{
		{Name: "a", Path: "/tmp/a", RelPath: "a", Desc: "desc", URL: "url", Tags: []string{"x"}},
		{Name: "b"},
	})
	if err := PrintStructured(projects, "json", &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `[
  {
    "name": "a",
    "path": "/tmp/a",
    "rel_path": "a",
    "description": "desc",
    "url": "url",
    "tags": [
      "x"
    ]
  },
  {
    "name": "b",
    "path": "",
    "rel_path": "",
    "description": "",
    "url": "",
    "tags": []
  }
]
`
	if got := buf.String(); got != want {
		t.Errorf("projects, got:\n%s\nwant:\n%s", got, want)
	}

	buf.Reset()
	tasks := TasksOutput([]dao.Task{
		{Name: "a", Desc: "desc", SpecData: dao.Spec{Name: "spec"}, TargetData: dao.Target{Name: "target"}},
	})
	if err := PrintStructured(tasks, "yaml", &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want = `name: a
description: desc
spec: spec
target: target
`
	if got := buf.String(); got != want {
		t.Errorf("tasks, got:\n%s\nwant:\n%s", got, want)
	}

	buf.Reset()
	tags := TagsOutput([]dao.Tag{{Name: "x", Projects: []string{"a"}}, {Name: "y"}})
	if err := PrintStructured(tags, "json", &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want = `[
  {
    "name": "x",
    "projects": [
      "a"
    ]
  },
  {
    "name": "y",
    "projects": []
  }
]
`
	if got := buf.String(); got != want {
		t.Errorf("tags, got:\n%s\nwant:\n%s", got, want)
	}
}
