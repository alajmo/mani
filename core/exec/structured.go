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

	// A tty command replaces the mani process, so there would be no records
	if !dryRun {
		for _, cmd := range taskCommands(exec.Tasks[0]) {
			if cmd.TTY {
				return &core.OutputTTYError{Output: output}
			}
		}
	}

	// Projects that don't exist are skipped with ignore-non-existing, without
	// it they have already been rejected. They don't get a record, since a
	// record would look like a task that failed or one that ran without output.
	var clients []int
	for i, client := range exec.Clients {
		if spec.IgnoreNonExisting {
			if _, err := os.Stat(client.Path); os.IsNotExist(err) {
				continue
			}
		}
		clients = append(clients, i)
	}

	// Records that are streamed are not kept, only the json array needs all of them
	var stream *print.StructuredStream
	var results []TaskResult
	if spec.Parallel || output == "yaml" {
		stream = print.NewStructuredStream(output, writer)
	} else {
		results = make([]TaskResult, len(clients))
	}

	var writeErr error
	var writeErrOnce sync.Once
	work := func(n int, wg *core.SizedWaitGroup) {
		defer wg.Done()
		result := exec.StructuredWork(clients[n], dryRun)
		if stream == nil {
			results[n] = result
			return
		}
		if err := stream.Write(result); err != nil {
			writeErrOnce.Do(func() { writeErr = err })
		}
	}

	wg := core.NewSizedWaitGroup(spec.Forks)
	for n := range clients {
		wg.Add()
		if spec.Parallel {
			go work(n, &wg)
		} else {
			work(n, &wg)
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
		Tasks:   commandNames(task),
		Output:  []string{},
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

// commandNames returns the names of the commands of a task, same as the table
// columns: unnamed commands are called output.
func commandNames(task dao.Task) []string {
	cmds := taskCommands(task)
	names := make([]string, len(cmds))
	for i, cmd := range cmds {
		if cmd.Name != "" {
			names[i] = cmd.Name
		} else {
			names[i] = "output"
		}
	}

	return names
}

// runStructuredCmd runs a single command and returns its output lines and exit code.
func runStructuredCmd(client Client, dryRun bool, cmd dao.Command) ([]string, int, error) {
	if dryRun {
		return splitLines([]byte(cmd.Cmd)), 0, nil
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
	// No output is no lines, while a single newline is one blank line
	if len(out) == 0 {
		return nil
	}

	// Invalid UTF-8 can't be represented in json, and yaml encodes the whole line as binary
	s := strings.ToValidUTF8(string(out), "�")
	lines := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	for i := range lines {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}

	return lines
}
