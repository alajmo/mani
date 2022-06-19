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
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func (exec *Exec) Text(dryRun bool) {
	clients := exec.Clients

	prefixMaxLen := calcMaxPrefixLength(clients)

	wg := core.NewSizedWaitGroup(20)
	for i, c := range clients {
		task := exec.Tasks[i]
		wg.Add()

		if task.SpecData.Parallel {
			go func(i int, c Client, wg *core.SizedWaitGroup) {
				defer wg.Done()
				_ = exec.TextWork(i, prefixMaxLen, dryRun)
			}(i, c, &wg)
		} else {
			func(i int, c Client, wg *core.SizedWaitGroup) {
				defer wg.Done()
				_ = exec.TextWork(i, prefixMaxLen, dryRun)
			}(i, c, &wg)
		}
	}

	wg.Wait()
}

func (exec *Exec) TextWork(rIndex int, prefixMaxLen int, dryRun bool) error {
	client := exec.Clients[rIndex]
	task := exec.Tasks[rIndex]

	prefix := getPrefixer(client, rIndex, prefixMaxLen, task.ThemeData.Text, task.SpecData.Parallel)

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

		err := RunTextCmd(args, task.ThemeData.Text, prefix, task.SpecData.Parallel, &wg)
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

		err := RunTextCmd(args, task.ThemeData.Text, prefix, task.SpecData.Parallel, &wg)
		if err != nil && !task.SpecData.IgnoreErrors {
			return err
		}
	}

	wg.Wait()

	return nil
}

func RunTextCmd(t TableCmd, textStyle dao.Text, prefix string, parallel bool, wg *sync.WaitGroup) error {
	combinedEnvs := dao.MergeEnvs(t.client.Env, t.env)

	if textStyle.Header && !parallel {
		printHeader(t.cIndex, t.numTasks, t.name, t.desc, textStyle)
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
			_, err = io.Copy(os.Stdout, core.NewPrefixer(client.Stdout(), prefix))
		} else {
			_, err = io.Copy(os.Stdout, client.Stdout())
		}

		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}(t.client)
	wg.Add(1)

	// Copy over tasks's STDERR.
	go func(client Client) {
		defer wg.Done()
		var err error
		if prefix != "" {
			_, err = io.Copy(os.Stderr, core.NewPrefixer(client.Stderr(), prefix))
		} else {
			_, err = io.Copy(os.Stderr, client.Stderr())
		}

		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}(t.client)
	wg.Add(1)

	wg.Wait()

	if err := t.client.Wait(); err != nil {
		if prefix != "" {
			fmt.Printf("%s%s\n", prefix, err.Error())
		} else {
			fmt.Printf("%s\n", err.Error())
		}

		return err
	}

	return nil
}

func printHeader(i int, numTasks int, name string, desc string, ts dao.Text) {
	var header string

	var prefixName string
	if name == "" {
		prefixName = "Command"
	} else {
		prefixName = name
	}

	var prefixPart1 string
	if numTasks > 1 {
		prefixPart1 = fmt.Sprintf("%s (%d/%d)", text.Bold.Sprintf(ts.HeaderPrefix), i+1, numTasks)
	} else {
		prefixPart1 = text.Bold.Sprintf(ts.HeaderPrefix)
	}

	var prefixPart2 string
	if desc != "" {
		prefixPart2 = fmt.Sprintf("[%s: %s]", text.Bold.Sprintf(prefixName), desc)
	} else {
		prefixPart2 = fmt.Sprintf("[%s]", text.Bold.Sprintf(prefixName))
	}

	width, _, _ := term.GetSize(0)

	if prefixPart1 != "" {
		header = fmt.Sprintf("%s %s", prefixPart1, prefixPart2)
	} else {
		header = prefixPart2
	}
	headerLength := len(core.Strip(header))

	if width > 0 && ts.HeaderChar != "" {
		header = fmt.Sprintf("\n%s %s\n", header, strings.Repeat(ts.HeaderChar, width-headerLength-1))
	} else {
		header = fmt.Sprintf("\n%s\n", header)
	}
	fmt.Println(header)
}

func getPrefixer(client Client, i, prefixMaxLen int, textStyle dao.Text, parallel bool) string {
	if !textStyle.Prefix {
		return ""
	}

	prefix := client.Prefix()
	prefixLen := len(prefix)
	var prefixColor *text.Color
	if len(textStyle.PrefixColors) < 1 {
		prefixColor = print.GetFg("")
	} else {
		prefixColor = print.GetFg(textStyle.PrefixColors[i%len(textStyle.PrefixColors)])
	}

	if (!textStyle.Header || parallel) && len(prefix) < prefixMaxLen { // Left padding.
		prefixString := prefix + strings.Repeat(" ", prefixMaxLen-prefixLen) + " | "
		if prefixColor != nil {
			prefix = prefixColor.Sprintf(prefixString)
		} else {
			prefix = prefixString
		}
	} else {
		prefixString := prefix + " | "
		if prefixColor != nil {
			prefix = prefixColor.Sprintf(prefixString)
		} else {
			prefix = prefixString
		}
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
