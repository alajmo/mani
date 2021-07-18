package core

import (
	"path/filepath"
	"fmt"
	"strings"
	"os"
	"os/exec"
	// "gopkg.in/yaml.v3"
)

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

			out, err := exec.Command("sh", "-c", kv[1]).Output()
			if err != nil {
				return envs, &ConfigEnvFailed { Name: kv[0], Err: err }
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
// 3. Global Env
func MergeEnv(userEnv []string, cmdEnv []string, globalEnv []string) []string {
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

		if  !ok {
			envs = append(envs, elem)
			args[kv[0]] = true
		}
	}

	for _, elem := range globalEnv {
		elem = strings.TrimSuffix(elem, "\n") 

		kv := strings.SplitN(elem, "=", 2)
		_, ok := args[kv[0]]

		if  !ok {
			envs = append(envs, elem)
			args[kv[0]] = true
		}
	}

	return envs
}
