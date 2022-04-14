package exec

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/alajmo/mani/core"
)

// Client is a wrapper over the SSH connection/sessions.
type Client struct {
	Name	string
	User    string
	Path    string
	Env		[]string

	// Env		[]string
	// Shell	string
	// Cmd		string

	cmd     *exec.Cmd
	stdout  io.Reader
	stderr  io.Reader
	running bool
}

func (c *Client) Run(shell string, env []string, cmdStr string) error {
	var err error
	if c.running {
		return fmt.Errorf("Command already running")
	}

	shellProgram, args := core.FormatShellString(shell, cmdStr)
	cmd := exec.Command(shellProgram, args...)
	cmd.Dir = c.Path
	cmd.Env = append(os.Environ(), env...)

	c.cmd = cmd

	c.stdout, err = cmd.StdoutPipe()
	if err != nil {
		return err
	}

	c.stderr, err = cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := c.cmd.Start(); err != nil {
		return err
	}

	c.running = true

	return nil
}

func (c *Client) Wait() error {
	if !c.running {
		return fmt.Errorf("Trying to wait on stopped command")
	}

	err := c.cmd.Wait()
	c.running = false

	return err
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) Stderr() io.Reader {
	return c.stderr
}

func (c *Client) Stdout() io.Reader {
	return c.stdout
}

func (c *Client) Prefix() (string) {
	return c.Name
}

