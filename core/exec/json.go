package exec

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

// TaskResult represents the structured output for a single task execution
type TaskResult struct {
	Project  string   `json:"project" yaml:"project"`
	Tasks    []string `json:"tasks" yaml:"tasks"`
	Output   []string `json:"output" yaml:"output"`
	ExitCode int      `json:"exit_code" yaml:"exit_code"`
}

func (exec *Exec) JSON(runFlags *core.RunFlags, outputFormat string, writer io.Writer) []TaskResult {
	task := exec.Tasks[0]
	clients := exec.Clients
	projects := exec.Projects
	isParallel := task.SpecData.Parallel

	// Collect all unique task names
	taskNames := make([]string, 0)
	taskSet := make(map[string]bool)
	for _, t := range exec.Tasks {
		if !taskSet[t.Name] {
			taskSet[t.Name] = true
			taskNames = append(taskNames, t.Name)
		}
	}

	// No spinner for structured output formats - it would interfere with JSON/YAML parsing
	// Just handle interrupt signal for clean exit
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		os.Exit(0)
	}()

	results := make([]TaskResult, len(projects))
	var dataMutex = sync.RWMutex{}
	var outputMutex = sync.Mutex{}

	// Initialize results with project names and all task names
	for i, p := range projects {
		results[i] = TaskResult{
			Project:  p.Name,
			Tasks:    taskNames,
			Output:   []string{},
			ExitCode: 0,
		}
	}

	wg := core.NewSizedWaitGroup(task.SpecData.Forks)
	for i, c := range clients {
		wg.Add()
		if isParallel {
			go func(i int, c Client, wg *core.SizedWaitGroup) {
				defer wg.Done()
				_ = exec.JSONWork(i, runFlags.DryRun, results, &dataMutex)
				// Stream output immediately in parallel mode
				outputMutex.Lock()
				var streamErr error
				if outputFormat == "json" {
					streamErr = PrintJSONStream(results[i], writer)
				} else if outputFormat == "yaml" {
					streamErr = PrintYAMLStream(results[i], writer)
				}
				if streamErr != nil {
					fmt.Fprintf(os.Stderr, "%v", streamErr)
				}
				outputMutex.Unlock()
			}(i, c, &wg)
		} else {
			func(i int, c Client, wg *core.SizedWaitGroup) {
				defer wg.Done()
				_ = exec.JSONWork(i, runFlags.DryRun, results, &dataMutex)
			}(i, c, &wg)
		}
	}
	wg.Wait()

	// Return results for non-parallel mode (they will be printed by caller)
	// For parallel mode, results were already streamed
	return results
}

func (exec *Exec) JSONWork(rIndex int, dryRun bool, results []TaskResult, dataMutex *sync.RWMutex) error {
	client := exec.Clients[rIndex]
	task := exec.Tasks[rIndex]

	var output []string
	var exitCode int

	for _, cmd := range task.Commands {
		if cmd.TTY {
			return ExecTTY(cmd.Cmd, cmd.EnvList)
		}

		out, code, err := RunJSONCmd(JSONCmd{
			client: client,
			dryRun: dryRun,
			shell:  cmd.ShellProgram,
			env:    cmd.EnvList,
			cmd:    cmd.Cmd,
			cmdArr: cmd.CmdArg,
		})

		output = append(output, out...)
		if code != 0 {
			exitCode = code
		}

		if err != nil && !task.SpecData.IgnoreErrors {
			dataMutex.Lock()
			results[rIndex].Output = output
			results[rIndex].ExitCode = exitCode
			dataMutex.Unlock()
			return err
		}
	}

	if task.Cmd != "" {
		if task.TTY {
			return ExecTTY(task.Cmd, task.EnvList)
		}

		out, code, err := RunJSONCmd(JSONCmd{
			client: client,
			dryRun: dryRun,
			shell:  task.ShellProgram,
			env:    task.EnvList,
			cmd:    task.Cmd,
			cmdArr: task.CmdArg,
		})

		output = append(output, out...)
		if code != 0 {
			exitCode = code
		}

		if err != nil && !task.SpecData.IgnoreErrors {
			dataMutex.Lock()
			results[rIndex].Output = output
			results[rIndex].ExitCode = exitCode
			dataMutex.Unlock()
			return err
		}
	}

	dataMutex.Lock()
	results[rIndex].Output = output
	results[rIndex].ExitCode = exitCode
	dataMutex.Unlock()

	return nil
}

type JSONCmd struct {
	client Client
	dryRun bool
	shell  string
	env    []string
	cmd    string
	cmdArr []string
}

func RunJSONCmd(j JSONCmd) ([]string, int, error) {
	combinedEnvs := dao.MergeEnvs(j.client.Env, j.env)

	if j.dryRun {
		return []string{j.cmd}, 0, nil
	}

	err := j.client.Run(j.shell, combinedEnvs, j.cmdArr)
	if err != nil {
		return []string{}, 1, err
	}

	var outputLines []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Read STDOUT
	wg.Add(1)
	go func(client Client) {
		defer wg.Done()
		out, err := io.ReadAll(client.Stdout())
		if err != nil && err != io.EOF {
			return
		}
		outStr := strings.TrimSuffix(string(out), "\n")
		if outStr != "" {
			lines := strings.Split(outStr, "\n")
			mu.Lock()
			outputLines = append(outputLines, lines...)
			mu.Unlock()
		}
	}(j.client)

	// Read STDERR
	wg.Add(1)
	go func(client Client) {
		defer wg.Done()
		out, err := io.ReadAll(client.Stderr())
		if err != nil && err != io.EOF {
			return
		}
		outStr := strings.TrimSuffix(string(out), "\n")
		if outStr != "" {
			lines := strings.Split(outStr, "\n")
			mu.Lock()
			outputLines = append(outputLines, lines...)
			mu.Unlock()
		}
	}(j.client)

	wg.Wait()

	exitCode := 0
	if err := j.client.Wait(); err != nil {
		exitCode = 1
		outputLines = append(outputLines, err.Error())
		return outputLines, exitCode, err
	}

	return outputLines, exitCode, nil
}

// PrintJSON outputs the results as JSON
func PrintJSON(results []TaskResult, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

// PrintJSONStream outputs each result as a single-line JSON object (for parallel/streaming)
func PrintJSONStream(result TaskResult, writer io.Writer) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	fmt.Fprintf(writer, "%s\n", data)
	return nil
}

// PrintYAML outputs the results as YAML with document separators
func PrintYAML(results []TaskResult, writer io.Writer) error {
	for i, result := range results {
		// Add document separator before each document (except the first)
		if i > 0 {
			fmt.Fprintf(writer, "---\n")
		}
		encoder := yaml.NewEncoder(writer)
		encoder.SetIndent(2)
		if err := encoder.Encode(result); err != nil {
			return err
		}
		encoder.Close()
	}
	return nil
}

// PrintYAMLStream outputs a single result as a YAML document with separator (for parallel/streaming)
func PrintYAMLStream(result TaskResult, writer io.Writer) error {
	fmt.Fprintf(writer, "---\n")
	encoder := yaml.NewEncoder(writer)
	encoder.SetIndent(2)
	if err := encoder.Encode(result); err != nil {
		return err
	}
	return encoder.Close()
}
