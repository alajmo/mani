package core

import (
	"path/filepath"
	"fmt"
	"strings"
	"os"
	"os/exec"
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

func EvaluateEnv(envMap map[string]string) ([]string, error) {
	var envs []string

	for k, v := range envMap {
		if strings.HasPrefix(v, "$(") && strings.HasSuffix(v, ")") {
			v = strings.TrimPrefix(v, "$(")
			v = strings.TrimSuffix(v, ")")

			out, err := exec.Command("sh", "-c", v).Output()
			if err != nil {
				return envs, &ConfigEnvFailed { Name: k, Err: err }
			}

			envs = append(envs, fmt.Sprintf("%v=%v", k, string(out)))
		} else {
			envs = append(envs, fmt.Sprintf("%v=%v", k, v))
		}
	}

	return envs, nil
}

// Order of preference (highest to lowest):
// 1. User argument
// 2. Command Env
// 3. Global Env
func MergeEnv(userEnv []string, cmdEnv []string, globalEnv []string) map[string]string {
	args := make(map[string]string)

	// User Env
	for _, arg := range userEnv {
		kv := strings.SplitN(arg, "=", 2)
		args[kv[0]] = kv[1]
	}

	// Command Env
	for _, arg := range cmdEnv {
		kv := strings.SplitN(arg, "=", 2)
		if args[kv[0]] == "" {
			args[kv[0]] = kv[1]
		}
	}

	// Global Env
	for _, arg := range globalEnv {
		kv := strings.SplitN(arg, "=", 2)
		if args[kv[0]] == "" {
			args[kv[0]] = kv[1]
		}
	}

	return args
}
