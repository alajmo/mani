package exec

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/print"
)

func (exec *Exec) Text(dryRun bool) {
	task := exec.Task
	clients := exec.Clients

	prefixMaxLen := calcMaxPrefixLength(clients)

	var wg sync.WaitGroup
	for i, c := range clients {
		wg.Add(1)

		if task.SpecData.Parallel {
			go func(i int, c Client) {
				defer wg.Done()
				exec.TextWork(i, prefixMaxLen, dryRun)
			}(i, c)
		} else {
			func(i int, c Client) {
				defer wg.Done()
				exec.TextWork(i, prefixMaxLen, dryRun)
			}(i, c)
		}
	}

	wg.Wait()
}

func (exec *Exec) TextWork(rIndex int, prefixMaxLen int, dryRun bool) {
	client := exec.Clients[rIndex]
	task := exec.Task

	prefix := getPrefixer(client, rIndex, prefixMaxLen, task.ThemeData.Text.Prefix, task.ThemeData.Text.Header, task.ThemeData.Text.Colors, task.SpecData.Parallel)

	var numTasks int
	if task.Cmd != "" {
		numTasks = len(task.Commands) + 1
	} else {
		numTasks = len(task.Commands)
	}

	var wg sync.WaitGroup
	for j, cmd := range task.Commands {
		err := RunTextCmd(rIndex, j, numTasks, client, dryRun, task.ThemeData.Text.Header, task.ThemeData.Text.HeaderChar, cmd.Desc, cmd.Name, prefix, cmd.Shell, cmd.EnvList, cmd.Cmd, task.SpecData.Parallel, &wg)
		if err != nil && !task.SpecData.IgnoreError {
			return
		}
	}

	if task.Cmd != "" {
		_ = RunTextCmd(rIndex, len(task.Commands), numTasks, client, dryRun, task.ThemeData.Text.Header, task.ThemeData.Text.HeaderChar, task.Desc, task.Name, prefix, task.Shell, task.EnvList, task.Cmd, task.SpecData.Parallel, &wg)
	}

	wg.Wait()
}

func RunTextCmd(
	rIndex int,
	cIndex int,
	numTasks int,
	c Client,
	dryRun bool,
	header bool,
	headerChar string,
	desc string,
	name string,
	prefix string,
	shell string,
	env []string,
	cmd string,
	parallel bool,
	wg *sync.WaitGroup,
) error {
	combinedEnvs := core.MergeEnvs(c.Env, env)

	if header && !parallel {
		printHeader(cIndex, numTasks, name, desc, headerChar)
	}

	if dryRun {
		printCmd(prefix, cmd)
		return nil
	}

	err := c.Run(shell, combinedEnvs, cmd)
	if err != nil {
		return err
	}

	// Copy over commands STDOUT.
	go func(c Client) {
		defer wg.Done()
		_, err := io.Copy(os.Stdout, core.NewPrefixer(c.Stdout(), prefix))
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}(c)
	wg.Add(1)

	// Copy over tasks's STDERR.
	go func(c Client) {
		defer wg.Done()
		_, err := io.Copy(os.Stdout, core.NewPrefixer(c.Stderr(), prefix))
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}(c)
	wg.Add(1)

	wg.Wait()

	if err := c.Wait(); err != nil {
		fmt.Printf("%s%s\n", prefix, err.Error())
		return err
	}

	return nil
}

func printHeader(i int, numTasks int, name string, desc string, headerChar string) {
	var header string

	var prefixName string
	if name == "" {
		prefixName = "Command"
	} else {
		prefixName = name
	}

	var prefixPart1 string
	if numTasks > 1 {
		prefixPart1 = fmt.Sprintf("%s (%d/%d)", text.Bold.Sprintf("TASK"), i + 1, numTasks)
	} else {
		prefixPart1 = fmt.Sprintf("%s", text.Bold.Sprintf("TASK"))
	}

	var prefixPart2 string
	if desc != "" {
		prefixPart2 = fmt.Sprintf("[%s: %s]", text.Bold.Sprintf(prefixName), desc)
	} else {
		prefixPart2 = fmt.Sprintf("[%s]", text.Bold.Sprintf(prefixName))
	}

	width, _, err := term.GetSize(0)
	core.CheckIfError(err)

	header = fmt.Sprintf("%s %s", prefixPart1, prefixPart2)
	headerLength := len(core.Strip(header))

	if headerChar != "" {
		header = fmt.Sprintf("\n%s %s\n", header, strings.Repeat(headerChar, width - headerLength - 1))
	} else {
		header = fmt.Sprintf("\n%s\n", header)
	}
	fmt.Println(header)
}

func getPrefixer(client Client, i, prefixMaxLen int, includePrefix bool, includeHeader bool, colors []string, parallel bool) string {
	if !includePrefix {
		return ""
	}

	prefix := client.Prefix()
	prefixLen := len(prefix)
	prefixColor := print.GetFg(colors[i % len(colors)])
	if (!includeHeader || parallel) && len(prefix) < prefixMaxLen { // Left padding.
		prefixString := prefix + strings.Repeat(" ", prefixMaxLen-prefixLen) + " | "
		prefix = fmt.Sprintf("%s", prefixColor.Sprintf(prefixString))
	} else {
		prefixString := prefix + " | "
		prefix = fmt.Sprintf("%s", prefixColor.Sprintf(prefixString))
	}

	return prefix
}

func calcMaxPrefixLength(clients []Client) int {
	var prefixMaxLen int = 0
	for _, c := range clients {
		prefix := c.Prefix()
		if len(prefix) > prefixMaxLen {
			prefixMaxLen = len(prefix)
		}
	}

	return prefixMaxLen
}


func printCmd(prefix string, cmd string) {
	scanner := bufio.NewScanner(strings.NewReader(cmd))
	for scanner.Scan() {
		fmt.Printf("%s%s\n", prefix, scanner.Text())
	}
}
