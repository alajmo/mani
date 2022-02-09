package core

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"gopkg.in/yaml.v3"
	"strings"
	"regexp"
)

var COLOR_INDEX = []int {2, 32, 179, 63, 205}
const ANSI = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
var RE = regexp.MustCompile(ANSI)

type TreeNode struct {
	Name     string
	Children []TreeNode
}

func AddToTree(root []TreeNode, names []string) []TreeNode {
	if len(names) > 0 {
		var i int
		for i = 0; i < len(root); i++ {
			if root[i].Name == names[0] { // already in tree
				break
			}
		}

		if i == len(root) {
			root = append(root, TreeNode{Name: names[0], Children: []TreeNode{}})
		}

		root[i].Children = AddToTree(root[i].Children, names[1:])
	}

	return root
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Intersection(a []string, b []string) []string {
	var i []string
	for _, s := range a {
		if StringInSlice(s, b) {
			i = append(i, s)
		}
	}

	return i
}

func GetWdRemoteUrl(path string) string {
	cwd, err := os.Getwd()
	CheckIfError(err)

	gitDir := filepath.Join(cwd, ".git")
	if _, err := os.Stat(gitDir); !os.IsNotExist(err) {
		return GetRemoteUrl(cwd)
	}

	return ""
}

func GetRemoteUrl(path string) string {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	var url string
	if err != nil {
		url = ""
	} else {
		url = strings.TrimSuffix(string(output), "\n")
	}

	return url
}

func FindFileInParentDirs(path string, files []string) (string, error) {
	for _, file := range files {
		pathToFile := filepath.Join(path, file)

		if _, err := os.Stat(pathToFile); err == nil {
			return pathToFile, nil
		}
	}

	parentDir := filepath.Dir(path)

	// TODO: Check different path if on windows subsystem
	// https://stackoverflow.com/questions/151860/root-folder-equivalent-in-windows/152038
	// https://en.wikipedia.org/wiki/Directory_structure#:~:text=In%20DOS%2C%20Windows%2C%20and%20OS,to%20being%20combined%20as%20one.
	// Seems it's \ in windows
	if parentDir == "/" {
		return "", &ConfigNotFound{files}
	}

	return FindFileInParentDirs(parentDir, files)
}

func EvaluateEnv(envList []string) ([]string, error) {
	var envs []string

	for _, arg := range envList {
		kv := strings.SplitN(arg, "=", 2)

		if strings.HasPrefix(kv[1], "$(") && strings.HasSuffix(kv[1], ")") {
			kv[1] = strings.TrimPrefix(kv[1], "$(")
			kv[1] = strings.TrimSuffix(kv[1], ")")

			cmd := exec.Command("sh", "-c", kv[1])
			cmd.Env = os.Environ()
			out, err := cmd.Output()
			if err != nil {
				return envs, &ConfigEnvFailed{Name: kv[0], Err: err}
			}

			envs = append(envs, fmt.Sprintf("%v=%v", kv[0], string(out)))
		} else {
			envs = append(envs, fmt.Sprintf("%v=%v", kv[0], kv[1]))
		}
	}

	return envs, nil
}

// Order of preference (highest to lowest):
// 1. User argument
// 2. Command Env
// 3. Parent Env
// 4. Global Env
func MergeEnv(userEnv []string, cmdEnv []string, parentEnv []string, globalEnv []string) []string {
	var envs []string
	args := make(map[string]bool)

	// User Env
	for _, elem := range userEnv {
		elem = strings.TrimSuffix(elem, "\n")

		kv := strings.SplitN(elem, "=", 2)
		envs = append(envs, elem)
		args[kv[0]] = true
	}

	// Command Env
	for _, elem := range cmdEnv {
		elem = strings.TrimSuffix(elem, "\n")

		kv := strings.SplitN(elem, "=", 2)
		_, ok := args[kv[0]]

		if !ok {
			envs = append(envs, elem)
			args[kv[0]] = true
		}
	}

	// Parent Env
	for _, elem := range parentEnv {
		elem = strings.TrimSuffix(elem, "\n")

		kv := strings.SplitN(elem, "=", 2)
		_, ok := args[kv[0]]

		if !ok {
			envs = append(envs, elem)
			args[kv[0]] = true
		}
	}

	// Config Env
	for _, elem := range globalEnv {
		elem = strings.TrimSuffix(elem, "\n")

		kv := strings.SplitN(elem, "=", 2)
		_, ok := args[kv[0]]

		if !ok {
			envs = append(envs, elem)
			args[kv[0]] = true
		}
	}

	return envs
}

func DebugPrint(data interface{}) {
	s, _ := json.MarshalIndent(data, "", "\t")
	fmt.Print(string(s))
}

func GetRelativePath(configDir string, path string) (string, error) {
	relPath, err := filepath.Rel(configDir, path)
	return relPath, err
}

// Get the absolute path
// Need to support following path types:
//		lala/land
//		./lala/land
//		../lala/land
//		/lala/land
//		$HOME/lala/land
//		~/lala/land
//		~root/lala/land
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

func Strip(str string) string {
	return RE.ReplaceAllString(str, "")
}

func GetEnv(node yaml.Node) []string {
	var envs []string
	count := len(node.Content)

	for i := 0; i < count; i += 2 {
		env := fmt.Sprintf("%v=%v", node.Content[i].Value, node.Content[i+1].Value)
		envs = append(envs, env)
	}

	return envs
}

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

	return shell
}
