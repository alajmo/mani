package exec

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"

	"golang.org/x/term"

	"github.com/gookit/color"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func (exec *Exec) Text(
	dryRun bool,
	stdout io.Writer,
	stderr io.Writer,
) {
	task := exec.Tasks[0]
	clients := exec.Clients

	prefixMaxLen := calcMaxPrefixLength(clients)

	wg := core.NewSizedWaitGroup(task.SpecData.Forks)
	for i, c := range clients {
		task := exec.Tasks[i]
		wg.Add()

		if task.SpecData.Parallel {
			go func(i int, c Client, wg *core.SizedWaitGroup) {
				defer wg.Done()
				_ = exec.TextWork(i, prefixMaxLen, dryRun, stdout, stderr)
			}(i, c, &wg)
		} else {
			func(i int, c Client, wg *core.SizedWaitGroup) {
				defer wg.Done()
				_ = exec.TextWork(i, prefixMaxLen, dryRun, stdout, stderr)
			}(i, c, &wg)
		}
	}

	wg.Wait()

	fmt.Fprintf(stdout, "\n")
}

func (exec *Exec) TextWork(
	rIndex int,
	prefixMaxLen int,
	dryRun bool,
	stdout io.Writer,
	stderr io.Writer,
) error {
	client := exec.Clients[rIndex]
	task := exec.Tasks[rIndex]
	prefix := getPrefixer(client, rIndex, prefixMaxLen, task.ThemeData.Stream, task.SpecData.Parallel)

	var numTasks int
	if task.Cmd != "" {
		numTasks = len(task.Commands) + 1
	} else {
		numTasks = len(task.Commands)
	}

	var wg sync.WaitGroup
	for j, cmd := range task.Commands {
		args := TableCmd{
			rIndex:   rIndex,
			cIndex:   j,
			client:   client,
			dryRun:   dryRun,
			shell:    cmd.ShellProgram,
			env:      cmd.EnvList,
			cmd:      cmd.Cmd,
			cmdArr:   cmd.CmdArg,
			desc:     cmd.Desc,
			name:     cmd.Name,
			numTasks: numTasks,
		}

		if cmd.TTY {
			return ExecTTY(cmd.Cmd, cmd.EnvList)
		}

		err := RunTextCmd(args, task.ThemeData.Stream, prefix, task.SpecData.Parallel, &wg, stdout, stderr)
		if err != nil && !task.SpecData.IgnoreErrors {
			return err
		}
	}

	if task.Cmd != "" {
		args := TableCmd{
			rIndex:   rIndex,
			cIndex:   len(task.Commands),
			client:   client,
			dryRun:   dryRun,
			shell:    task.ShellProgram,
			env:      task.EnvList,
			cmd:      task.Cmd,
			cmdArr:   task.CmdArg,
			desc:     task.Desc,
			name:     task.Name,
			numTasks: numTasks,
		}

		if task.TTY {
			return ExecTTY(task.Cmd, task.EnvList)
		}

		err := RunTextCmd(args, task.ThemeData.Stream, prefix, task.SpecData.Parallel, &wg, stdout, stderr)
		if err != nil && !task.SpecData.IgnoreErrors {
			return err
		}
	}

	wg.Wait()

	return nil
}

func RunTextCmd(
	t TableCmd,
	textStyle dao.Stream,
	prefix string,
	parallel bool,
	wg *sync.WaitGroup,
	stdout io.Writer,
	stderr io.Writer,
) error {
	combinedEnvs := dao.MergeEnvs(t.client.Env, t.env)

	if textStyle.Header && !parallel {
		printHeader(stdout, t.cIndex, t.numTasks, t.name, t.desc, textStyle)
	}

	if t.dryRun {
		printCmd(prefix, t.cmd)
		return nil
	}

	err := t.client.Run(t.shell, combinedEnvs, t.cmdArr)
	if err != nil {
		return err
	}

	// Copy over commands STDOUT.
	go func(client Client) {
		defer wg.Done()
		var err error
		if prefix != "" {
			_, err = io.Copy(stdout, core.NewPrefixer(client.Stdout(), prefix))
		} else {
			_, err = io.Copy(stdout, client.Stdout())
		}

		if err != nil && err != io.EOF {
			fmt.Fprintf(stderr, "%s", err)
		}
	}(t.client)
	wg.Add(1)

	// Copy over tasks's STDERR.
	go func(client Client) {
		defer wg.Done()
		var err error
		if prefix != "" {
			_, err = io.Copy(stderr, core.NewPrefixer(client.Stderr(), prefix))
		} else {
			_, err = io.Copy(stderr, client.Stderr())
		}

		if err != nil && err != io.EOF {
			fmt.Fprintf(stderr, "%s", err)
		}
	}(t.client)
	wg.Add(1)

	wg.Wait()

	if err := t.client.Wait(); err != nil {
		if prefix != "" {
			fmt.Fprintf(stderr, "%s%s\n", prefix, err)
		} else {
			fmt.Fprintf(stderr, "%s\n", err)
		}

		return err
	}

	return nil
}

// TASK [pwd] -------------
func printHeader(stdout io.Writer, i int, numTasks int, name string, desc string, ts dao.Stream) {
	var header string

	prefixName := ""
	if name == "" {
		prefixName = color.Bold.Sprint("Command")
	} else {
		prefixName = color.Bold.Sprint(name)
	}

	var prefixPart1 string
	if numTasks > 1 {
		prefixPart1 = fmt.Sprintf("%s (%d/%d)", color.Bold.Sprint(ts.HeaderPrefix), i+1, numTasks)
	} else {
		prefixPart1 = color.Bold.Sprint(ts.HeaderPrefix)
	}

	var prefixPart2 string
	if desc != "" {
		prefixPart2 = fmt.Sprintf("[%s: %s]", prefixName, desc)
	} else {
		prefixPart2 = fmt.Sprintf("[%s]", prefixName)
	}

	width, _, _ := term.GetSize(0)

	if prefixPart1 != "" {
		header = fmt.Sprintf("%s %s", prefixPart1, prefixPart2)
	} else {
		header = prefixPart2
	}
	headerLength := len(core.Strip(header))

	if width > 0 && ts.HeaderChar != "" {
		header = fmt.Sprintf("\n%s %s\n\n", header, strings.Repeat(ts.HeaderChar, width-headerLength-1))
	} else {
		header = fmt.Sprintf("\n%s\n\n", header)
	}
	fmt.Fprint(stdout, header)
}

// mani | /projects/mani
func getPrefixer(client Client, i, prefixMaxLen int, textStyle dao.Stream, parallel bool) string {
	if !textStyle.Prefix {
		return ""
	}

	// Project name color
	var prefixColor color.RGBColor
	if len(textStyle.PrefixColors) < 1 {
		prefixColor = dao.StyleFg("")
	} else {
		fg := textStyle.PrefixColors[i%len(textStyle.PrefixColors)]
		prefixColor = dao.StyleFg(fg)
	}

	prefix := client.Prefix()
	prefixLen := len(prefix)
	// If we don't have a task header or the execution is parallel, then left pad the prefix.
	if (!textStyle.Header || parallel) && len(prefix) < prefixMaxLen { // Left padding.
		prefixString := prefix + strings.Repeat(" ", prefixMaxLen-prefixLen) + " | "
		prefix = prefixColor.Sprint(prefixString)
	} else {
		prefixString := prefix + " | "
		prefix = prefixColor.Sprint(prefixString)
	}

	return prefix
}

func calcMaxPrefixLength(clients []Client) int {
	var prefixMaxLen = 0
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
