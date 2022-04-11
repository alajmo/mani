package exec

import (
	"fmt"
	"sync"
	"io"
	"strings"
	"os"
	"golang.org/x/term"

	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/alajmo/mani/core"
)

func (exec *Exec) Text() {
	task := exec.Task
	clients := exec.Clients

	prefixMaxLen := calcMaxPrefixLength(clients)

	var wg sync.WaitGroup
	for i, c := range clients {
		wg.Add(1)

		if task.SpecData.Parallel {
			go func(i int, c Client) {
				defer wg.Done()
				exec.TextWork(i, prefixMaxLen)
			}(i, c)
		} else {
			func(i int, c Client) {
				defer wg.Done()
				exec.TextWork(i, prefixMaxLen)
			}(i, c)
		}
	}

	wg.Wait()
}

func (exec *Exec) TextWork(rIndex int, prefixMaxLen int) {
	client := exec.Clients[rIndex]
	task := exec.Task
	prefix := getPrefixer(client, rIndex, prefixMaxLen, task.ThemeData.Text.Prefix, task.SpecData.Parallel)

	var numTasks int
	if task.Cmd != "" {
		numTasks = len(task.Commands) + 1
	} else {
		numTasks = len(task.Commands)
	}

	var wg sync.WaitGroup
	for j, cmd := range task.Commands {
		err := RunTextCmd(rIndex, j, numTasks, client, cmd.Desc, cmd.Name, prefix, cmd.Shell, cmd.EnvList, cmd.Cmd, task.SpecData.Parallel, &wg)
		if err != nil && !task.SpecData.IgnoreError {
			fmt.Println(err)
		}

		if err != nil {
			fmt.Println(err)
		}
	}

	if task.Cmd != "" {
		err := RunTextCmd(rIndex, len(task.Commands), numTasks, client, task.Desc, task.Name, prefix, task.Shell, task.EnvList, task.Cmd, task.SpecData.Parallel, &wg)
		if err != nil {
			fmt.Println(err)
		}
	}

	wg.Wait()
}

func RunTextCmd(
	rIndex int,
	cIndex int,
	numTasks int,
	c Client,
	desc string,
	name string,
	prefix string,
	shell string,
	env []string,
	cmd string,
	parallel bool,
	wg *sync.WaitGroup,
) error {
	err := c.Run(shell, env, cmd)
	if err != nil {
		return err
	}

	if !parallel {
		printHeader(cIndex, numTasks, name, desc)
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
		os.Exit(1)
	}

	return nil
}

func printHeader(i int, numTasks int, name string, desc string) {
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
		prefixPart2 = fmt.Sprintf("[%s | %s]", text.Bold.Sprintf(prefixName), desc)
	} else {
		prefixPart2 = fmt.Sprintf("[%s]", text.Bold.Sprintf(prefixName))
	}

	width, _, err := term.GetSize(0)
	core.CheckIfError(err)

	header = fmt.Sprintf("%s %s", prefixPart1, prefixPart2)
	headerLength := len(core.Strip(header))
	header = fmt.Sprintf("\n%s %s\n", header, strings.Repeat("*", width - headerLength - 1))
	fmt.Println(header)
}

func getPrefixer(client Client, i, prefixMaxLen int, includePrefix bool, parallel bool) string {
	if !includePrefix {
		return ""
	}

	prefix := client.Prefix()
	prefixLen := len(prefix)
	prefixColor := core.COLOR_INDEX[i % len(core.COLOR_INDEX)]
	if parallel && len(prefix) < prefixMaxLen { // Left padding.
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
