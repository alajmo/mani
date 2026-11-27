package exec

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

// TaskResult is the json/yaml record for one project.
type TaskResult struct {
	Project string `json:"project" yaml:"project"`
	// Names of the commands that make up the run, same as the table columns
	Tasks []string `json:"tasks" yaml:"tasks"`
	// Output lines of all commands, per command stdout followed by stderr
	Output []string `json:"output" yaml:"output"`
	// Exit code of the first failing command, 0 if all commands succeeded and
	// -1 if a command could not be started or was terminated by a signal
	ExitCode int `json:"exit_code" yaml:"exit_code"`
}

// Structured runs the tasks and writes one record per project as json or yaml.
//
// Records are written as each project finishes, json as one object per line
// and yaml as one document per project. The exception is json when not running
// in parallel, where all records are written as a single array at the end.
//
// Nothing but the records is written to the writer (no spinner or headers), so
// that the output can be parsed.
func (exec *Exec) Structured(dryRun bool, output string, writer io.Writer) error {
	spec := exec.Tasks[0].SpecData
	results := make([]TaskResult, len(exec.Clients))

	var stream *print.StructuredStream
	if spec.Parallel || output == "yaml" {
		stream = print.NewStructuredStream(output, writer)
	}

	var writeErr error
	var writeErrOnce sync.Once
	work := func(i int, wg *core.SizedWaitGroup) {
		defer wg.Done()
		results[i] = exec.StructuredWork(i, dryRun)
		if stream == nil {
			return
		}
		if err := stream.Write(results[i]); err != nil {
			writeErrOnce.Do(func() { writeErr = err })
		}
	}

	wg := core.NewSizedWaitGroup(spec.Forks)
	for i := range exec.Clients {
		wg.Add()
		if spec.Parallel {
			go work(i, &wg)
		} else {
			work(i, &wg)
		}
	}
	wg.Wait()

	if stream == nil {
		return print.PrintStructured(results, output, writer)
	}

	if err := stream.Close(); err != nil {
		return err
	}

	return writeErr
}

// StructuredWork runs all commands of a task for a single project.
func (exec *Exec) StructuredWork(rIndex int, dryRun bool) TaskResult {
	client := exec.Clients[rIndex]
	task := exec.Tasks[rIndex]
	cmds := taskCommands(task)

	result := TaskResult{
		Project: client.Name,
		Tasks:   make([]string, len(cmds)),
		Output:  []string{},
	}

	for i, cmd := range cmds {
		if cmd.Name != "" {
			result.Tasks[i] = cmd.Name
		} else {
			result.Tasks[i] = "output"
		}
	}

	for _, cmd := range cmds {
		lines, exitCode, err := runStructuredCmd(client, dryRun, cmd)
		result.Output = append(result.Output, lines...)
		if exitCode != 0 && result.ExitCode == 0 {
			result.ExitCode = exitCode
		}

		if err != nil && !task.SpecData.IgnoreErrors {
			break
		}
	}

	return result
}

// taskCommands returns the commands of a task in the order they are run: the
// sub-commands followed by the task's own command.
func taskCommands(task dao.Task) []dao.Command {
	cmds := slices.Clone(task.Commands)
	if task.Cmd != "" {
		cmd := task.ConvertTaskToCommand()
		cmd.TTY = task.TTY
		cmds = append(cmds, cmd)
	}

	return cmds
}

// runStructuredCmd runs a single command and returns its output lines and exit code.
func runStructuredCmd(client Client, dryRun bool, cmd dao.Command) ([]string, int, error) {
	if dryRun {
		return splitLines([]byte(cmd.Cmd)), 0, nil
	}

	if cmd.TTY {
		// Replaces the current process on success
		if err := ExecTTY(cmd.Cmd, cmd.EnvList); err != nil {
			return []string{err.Error()}, -1, err
		}
		return nil, 0, nil
	}

	err := client.Run(cmd.ShellProgram, dao.MergeEnvs(client.Env, cmd.EnvList), cmd.CmdArg)
	if err != nil {
		return []string{err.Error()}, -1, err
	}

	// Both pipes have to be drained before waiting on the command
	var stdout, stderr []byte
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		stdout = readAll(client.Stdout())
	}()
	go func() {
		defer wg.Done()
		stderr = readAll(client.Stderr())
	}()
	wg.Wait()

	lines := append(splitLines(stdout), splitLines(stderr)...)

	if err := client.Wait(); err != nil {
		// A command that exits non-zero is reported via the exit code, and not
		// with Go's "exit status N" in the output (see table.go).
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return lines, exitErr.ExitCode(), err
		}

		return append(lines, err.Error()), -1, err
	}

	return lines, 0, nil
}

func readAll(reader io.Reader) []byte {
	out, err := io.ReadAll(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	return out
}

// splitLines splits command output into lines, without the trailing newline.
func splitLines(out []byte) []string {
	// Invalid UTF-8 can't be represented in json, and yaml encodes the whole line as binary
	s := strings.ToValidUTF8(string(out), "�")
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return nil
	}

	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}

	return lines
}
