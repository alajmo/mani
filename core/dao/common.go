package dao

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gookit/color"
	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

// Resource Errors

type ResourceErrors[T any] struct {
	Resource *T
	Errors   []error
}

type Resource interface {
	GetContext() string
	GetContextLine() int
}

func FormatErrors(re Resource, errs []error) error {
	var msg = ""
	partsRe := regexp.MustCompile(`line (\d*): (.*)`)

	context := re.GetContext()

	var errPrefix = color.FgRed.Sprintf("error")
	var ptrPrefix = color.FgBlue.Sprintf("-->")
	for _, err := range errs {
		match := partsRe.FindStringSubmatch(err.Error())
		// In-case matching fails, return unformatted error
		if len(match) != 3 {
			contextLine := re.GetContextLine()

			if contextLine == -1 {
				msg = fmt.Sprintf("%s%s: %s\n  %s %s\n\n", msg, errPrefix, err, ptrPrefix, context)
			} else {
				msg = fmt.Sprintf("%s%s: %s\n  %s %s:%d\n\n", msg, errPrefix, err, ptrPrefix, context, contextLine)
			}
		} else {
			msg = fmt.Sprintf("%s%s: %s\n  %s %s:%s\n\n", msg, errPrefix, match[2], ptrPrefix, context, match[1])
		}
	}

	if msg != "" {
		return &core.ConfigErr{Msg: msg}
	}

	return nil
}

// ENV

func ParseNodeEnv(node yaml.Node) []string {
	var envs []string
	count := len(node.Content)

	for i := 0; i < count; i += 2 {
		env := fmt.Sprintf("%v=%v", node.Content[i].Value, node.Content[i+1].Value)
		envs = append(envs, env)
	}

	return envs
}

func EvaluateEnv(envList []string) ([]string, error) {
	var envs []string

	for _, arg := range envList {
		kv := strings.SplitN(arg, "=", 2)

		if val, hasPrefix := strings.CutPrefix(kv[1], "$("); hasPrefix {
			if cmdStr, hasSuffix := strings.CutSuffix(val, ")"); hasSuffix {
				cmd := exec.Command("sh", "-c", cmdStr)
				cmd.Env = os.Environ()
				out, err := cmd.CombinedOutput()
				if err != nil {
					return envs, &core.ConfigEnvFailed{Name: kv[0], Err: string(out)}
				}

				envs = append(envs, fmt.Sprintf("%v=%v", kv[0], string(out)))
				continue
			}
		}
		envs = append(envs, fmt.Sprintf("%v=%v", kv[0], kv[1]))
	}

	return envs, nil
}

// MergeEnvs Merges environment variables.
// Priority is from highest to lowest (1st env takes precedence over the last entry).
func MergeEnvs(envs ...[]string) []string {
	var mergedEnvs []string
	args := make(map[string]bool)

	for _, part := range envs {
		for _, elem := range part {
			elem = strings.TrimSuffix(elem, "\n")

			kv := strings.SplitN(elem, "=", 2)
			_, ok := args[kv[0]]

			if !ok {
				mergedEnvs = append(mergedEnvs, elem)
				args[kv[0]] = true
			}
		}
	}

	return mergedEnvs
}
