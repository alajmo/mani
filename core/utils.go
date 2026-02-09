package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

const ANSI = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var RE = regexp.MustCompile(ANSI)

func Strip(str string) string {
	return RE.ReplaceAllString(str, "")
}

func Intersection(a []string, b []string) []string {
	var i []string
	for _, s := range a {
		if slices.Contains(b, s) {
			i = append(i, s)
		}
	}

	return i
}

func GetWdRemoteURL(path string) (string, error) {
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); !os.IsNotExist(err) {
		return GetRemoteURL(path)
	}

	return "", nil
}

func GetRemoteURL(path string) (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	return strings.TrimSuffix(string(output), "\n"), nil
}

// GetWorktreeList returns a map of worktrees (absolute path -> branch) for a git repo
// Excludes the main worktree (the repo itself)
func GetWorktreeList(repoPath string) (map[string]string, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	worktrees := make(map[string]string)
	cleanRepoPath := filepath.Clean(repoPath)
	var currentPath string

	for line := range strings.SplitSeq(string(output), "\n") {
		if path, found := strings.CutPrefix(line, "worktree "); found {
			currentPath = filepath.Clean(path)
		} else if branch, found := strings.CutPrefix(line, "branch refs/heads/"); found {
			// Skip the main worktree (same as repoPath)
			if currentPath != cleanRepoPath {
				worktrees[currentPath] = branch
			}
		} else if line == "detached" {
			// Skip detached HEAD worktrees — they have no branch to track
		}
	}

	return worktrees, nil
}

func FindFileInParentDirs(path string, files []string) (string, error) {
	for _, file := range files {
		pathToFile := filepath.Join(path, file)
		if _, err := os.Stat(pathToFile); err == nil {
			return pathToFile, nil
		}
	}

	parentDir := filepath.Dir(path)
	if parentDir == path {
		return "", &ConfigNotFound{files}
	}

	return FindFileInParentDirs(parentDir, files)
}

func GetRelativePath(configDir string, path string) (string, error) {
	relPath, err := filepath.Rel(configDir, path)
	return relPath, err
}

// Get the absolute path
// Need to support following path types:
//
//	lala/land
//	./lala/land
//	../lala/land
//	/lala/land
//	$HOME/lala/land
//	~/lala/land
//	~root/lala/land
func GetAbsolutePath(configDir string, path string, name string) (string, error) {
	path = os.ExpandEnv(path)

	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	homeDir := usr.HomeDir

	// TODO: Remove any .., make path absolute and then cut of configDir
	if path == "~" {
		path = homeDir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homeDir, path[2:])
	} else if len(path) > 0 && filepath.IsAbs(path) { // TODO: Rewrite this
	} else if len(path) > 0 {
		path = filepath.Join(configDir, path)
	} else {
		path = filepath.Join(configDir, name)
	}

	return path, nil
}

// Get the absolute path
// Need to support following path types:
//
//	lala/land
//	./lala/land
//	../lala/land
//	/lala/land
//	$HOME/lala/land
//	~/lala/land
//	~root/lala/land
func ResolveTildePath(path string) (string, error) {
	path = os.ExpandEnv(path)

	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	homeDir := usr.HomeDir

	var p string
	if path == "~" {
		p = homeDir
	} else if strings.HasPrefix(path, "~/") {
		p = filepath.Join(homeDir, path[2:])
	} else {
		p = path
	}

	return p, nil
}

// FormatShell returns the shell program and associated command flag
func FormatShell(shell string) string {
	s := strings.Split(shell, " ")

	if len(s) > 1 { // User provides correct flag, bash -c, /bin/bash -c, /bin/sh -c
		return shell
	} else if strings.Contains(shell, "bash") { // bash, /bin/bash
		return shell + " -c"
	} else if strings.Contains(shell, "zsh") { // zsh, /bin/zsh
		return shell + " -c"
	} else if strings.Contains(shell, "sh") { // sh, /bin/sh
		return shell + " -c"
	} else if strings.Contains(shell, "node") { // node, /bin/node
		return shell + " -e"
	} else if strings.Contains(shell, "python") { // python, /bin/python
		return shell + " -c"
	}
	// TODO: Add fish and other shells

	return shell
}

// FormatShellString returns the shell program (bash,sh,.etc) along with the
// command flag and subsequent commands
// Example:
// "bash", "-c echo hello world"
func FormatShellString(shell string, command string) (string, []string) {
	shellProgram := FormatShell(shell)
	args := strings.SplitN(shellProgram, " ", 2)
	return args[0], append(args[1:], command)
}

// Used when creating pointers to literal. Useful when you want set/unset attributes.
func Ptr[T any](t T) *T {
	return &t
}

func StringsToErrors(str []string) []error {
	errs := []error{}
	for _, s := range str {
		errs = append(errs, errors.New(s))
	}

	return errs
}

func DebugPrint(data any) {
	s, _ := json.MarshalIndent(data, "", "\t")
	fmt.Println()
	fmt.Print(string(s))
	fmt.Println()
}
