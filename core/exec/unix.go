//go:build !windows
// +build !windows

package exec

import (
	"os"
	"os/exec"

	"golang.org/x/sys/unix"
)

func ExecTTY(cmd string, envs []string) error {
	shell := "bash"
	foundShell, found := os.LookupEnv("SHELL")
	if found {
		shell = foundShell
	}

	execBin, err := exec.LookPath(shell)
	if err != nil {
		return err
	}

	userEnv := append(os.Environ(), envs...)
	err = unix.Exec(execBin, []string{shell, "-c", cmd}, userEnv)
	if err != nil {
		return err
	}

	return nil
}
