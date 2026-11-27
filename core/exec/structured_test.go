//go:build !windows

package exec

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func shCommand(name string, cmd string) dao.Command {
	return dao.Command{Name: name, Cmd: cmd, ShellProgram: "sh", CmdArg: []string{"-c", cmd}}
}

// newTestExec creates an Exec that runs the task in each of the projects.
func newTestExec(t *testing.T, task dao.Task, projects ...string) *Exec {
	t.Helper()

	task.SpecData.Forks = 4
	exec := &Exec{}
	for _, name := range projects {
		exec.Projects = append(exec.Projects, dao.Project{Name: name})
		exec.Clients = append(exec.Clients, Client{Name: name, Path: t.TempDir()})
		exec.Tasks = append(exec.Tasks, task)
	}

	return exec
}

func TestStructuredWork(t *testing.T) {
	tests := []struct {
		name   string
		task   dao.Task
		dryRun bool
		want   TaskResult
	}{
		{
			name: "single command",
			task: dao.Task{Name: "hello", Cmd: "echo hello", ShellProgram: "sh", CmdArg: []string{"-c", "echo hello"}},
			want: TaskResult{Project: "p", Tasks: []string{"hello"}, Output: []string{"hello"}, ExitCode: 0},
		},
		{
			name: "no output is an empty list",
			task: dao.Task{Commands: []dao.Command{shCommand("quiet", "true")}},
			want: TaskResult{Project: "p", Tasks: []string{"quiet"}, Output: []string{}, ExitCode: 0},
		},
		{
			name: "exit code of the command and no exit status line",
			task: dao.Task{Commands: []dao.Command{shCommand("fail", "exit 3")}},
			want: TaskResult{Project: "p", Tasks: []string{"fail"}, Output: []string{}, ExitCode: 3},
		},
		{
			name: "stdout comes before stderr",
			task: dao.Task{Commands: []dao.Command{shCommand("both", "echo err >&2; echo out")}},
			want: TaskResult{Project: "p", Tasks: []string{"both"}, Output: []string{"out", "err"}, ExitCode: 0},
		},
		{
			name: "lines are split and blank lines kept",
			task: dao.Task{Commands: []dao.Command{shCommand("lines", `printf 'a\r\n\nb\n'`)}},
			want: TaskResult{Project: "p", Tasks: []string{"lines"}, Output: []string{"a", "", "b"}, ExitCode: 0},
		},
		{
			name: "a single blank line is kept",
			task: dao.Task{Commands: []dao.Command{shCommand("blank", "echo"), shCommand("blank-crlf", `printf '\r\n'`)}},
			want: TaskResult{Project: "p", Tasks: []string{"blank", "blank-crlf"}, Output: []string{"", ""}, ExitCode: 0},
		},
		{
			name: "output without a trailing newline",
			task: dao.Task{Commands: []dao.Command{shCommand("partial", `printf 'a\nb'`)}},
			want: TaskResult{Project: "p", Tasks: []string{"partial"}, Output: []string{"a", "b"}, ExitCode: 0},
		},
		{
			name: "invalid utf-8 is replaced",
			task: dao.Task{Commands: []dao.Command{shCommand("binary", `printf 'a\377b\n'`)}},
			want: TaskResult{Project: "p", Tasks: []string{"binary"}, Output: []string{"a�b"}, ExitCode: 0},
		},
		{
			name: "sub-commands run before the task command and unnamed ones are called output",
			task: dao.Task{
				Name: "last", Cmd: "echo 3", ShellProgram: "sh", CmdArg: []string{"-c", "echo 3"},
				Commands: []dao.Command{shCommand("first", "echo 1"), shCommand("", "echo 2")},
			},
			want: TaskResult{Project: "p", Tasks: []string{"first", "output", "last"}, Output: []string{"1", "2", "3"}, ExitCode: 0},
		},
		{
			name: "stops at the first failing command",
			task: dao.Task{Commands: []dao.Command{shCommand("a", "echo a"), shCommand("b", "echo b; exit 2"), shCommand("c", "echo c")}},
			want: TaskResult{Project: "p", Tasks: []string{"a", "b", "c"}, Output: []string{"a", "b"}, ExitCode: 2},
		},
		{
			name: "ignore errors runs all commands and keeps the first failing exit code",
			task: dao.Task{
				SpecData: dao.Spec{IgnoreErrors: true},
				Commands: []dao.Command{shCommand("a", "exit 2"), shCommand("b", "exit 5"), shCommand("c", "echo c")},
			},
			want: TaskResult{Project: "p", Tasks: []string{"a", "b", "c"}, Output: []string{"c"}, ExitCode: 2},
		},
		{
			name:   "dry run returns the commands without running them",
			dryRun: true,
			task:   dao.Task{Commands: []dao.Command{shCommand("a", "exit 2"), shCommand("b", "echo 1\necho 2")}},
			want:   TaskResult{Project: "p", Tasks: []string{"a", "b"}, Output: []string{"exit 2", "echo 1", "echo 2"}, ExitCode: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec := newTestExec(t, tt.task, "p")
			got := exec.StructuredWork(0, tt.dryRun)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestStructuredWorkCommandNotStarted(t *testing.T) {
	cmd := dao.Command{Name: "a", Cmd: "echo a", ShellProgram: "mani-shell-that-does-not-exist", CmdArg: []string{"-c", "echo a"}}
	exec := newTestExec(t, dao.Task{Commands: []dao.Command{cmd, shCommand("b", "echo b")}}, "p")

	got := exec.StructuredWork(0, false)
	if got.ExitCode != -1 {
		t.Errorf("got exit code %d, want -1", got.ExitCode)
	}
	if len(got.Output) != 1 || !strings.Contains(got.Output[0], "mani-shell-that-does-not-exist") {
		t.Errorf("got output %q, want the error from starting the command", got.Output)
	}
}

// A tty command replaces the process, which would leave no records and skip the remaining projects.
func TestStructuredTTY(t *testing.T) {
	tty := shCommand("edit", "echo hi")
	tty.TTY = true

	tasks := map[string]dao.Task{
		"sub-command": {Commands: []dao.Command{shCommand("first", "echo 1"), tty}},
		"task":        {Name: "edit", Cmd: "echo hi", TTY: true, ShellProgram: "sh", CmdArg: []string{"-c", "echo hi"}},
	}

	for name, task := range tasks {
		t.Run(name, func(t *testing.T) {
			for _, output := range []string{"json", "yaml"} {
				var buf bytes.Buffer
				exec := newTestExec(t, task, "a", "b")
				err := exec.Structured(false, output, &buf)

				var ttyErr *core.OutputTTYError
				if !errors.As(err, &ttyErr) || ttyErr.Name != "edit" || ttyErr.Output != output {
					t.Errorf("%s: got error %v, want OutputTTYError for edit", output, err)
				}
				if buf.Len() != 0 {
					t.Errorf("%s: got output %q, want none", output, buf.String())
				}
			}

			// Nothing is run in a dry run
			var buf bytes.Buffer
			exec := newTestExec(t, task, "a")
			if err := exec.Structured(true, "json", &buf); err != nil {
				t.Errorf("dry run: unexpected error: %v", err)
			}
		})
	}
}

func TestStructured(t *testing.T) {
	task := dao.Task{Commands: []dao.Command{shCommand("name", "echo hi")}}

	t.Run("json is an array in project order", func(t *testing.T) {
		var buf bytes.Buffer
		exec := newTestExec(t, task, "a", "b")
		if err := exec.Structured(false, "json", &buf); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := `[
  {
    "project": "a",
    "tasks": [
      "name"
    ],
    "output": [
      "hi"
    ],
    "exit_code": 0
  },
  {
    "project": "b",
    "tasks": [
      "name"
    ],
    "output": [
      "hi"
    ],
    "exit_code": 0
  }
]
`
		if got := buf.String(); got != want {
			t.Errorf("got:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("yaml is one document per project", func(t *testing.T) {
		var buf bytes.Buffer
		exec := newTestExec(t, task, "a", "b")
		if err := exec.Structured(false, "yaml", &buf); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := `project: a
tasks:
  - name
output:
  - hi
exit_code: 0
---
project: b
tasks:
  - name
output:
  - hi
exit_code: 0
`
		if got := buf.String(); got != want {
			t.Errorf("got:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("parallel json is one object per line", func(t *testing.T) {
		parallel := task
		parallel.SpecData.Parallel = true

		var buf bytes.Buffer
		exec := newTestExec(t, parallel, "a", "b", "c")
		if err := exec.Structured(false, "json", &buf); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Projects finish in any order
		got := map[string]bool{}
		lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
		for _, line := range lines {
			var result TaskResult
			if err := json.Unmarshal([]byte(line), &result); err != nil {
				t.Fatalf("line %q is not a json object: %v", line, err)
			}
			want := TaskResult{Project: result.Project, Tasks: []string{"name"}, Output: []string{"hi"}}
			if !reflect.DeepEqual(result, want) {
				t.Errorf("got %+v, want %+v", result, want)
			}
			got[result.Project] = true
		}
		if want := map[string]bool{"a": true, "b": true, "c": true}; !reflect.DeepEqual(got, want) {
			t.Errorf("got projects %v, want %v", got, want)
		}
	})

	t.Run("parallel yaml is one document per project", func(t *testing.T) {
		parallel := task
		parallel.SpecData.Parallel = true

		var buf bytes.Buffer
		exec := newTestExec(t, parallel, "a", "b", "c")
		if err := exec.Structured(false, "yaml", &buf); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		docs := strings.Split(buf.String(), "---\n")
		if len(docs) != 3 {
			t.Fatalf("got %d documents, want 3:\n%s", len(docs), buf.String())
		}
		for _, doc := range docs {
			if !strings.HasPrefix(doc, "project: ") || !strings.HasSuffix(doc, "exit_code: 0\n") {
				t.Errorf("unexpected document:\n%s", doc)
			}
		}
	})
}
