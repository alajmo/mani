package dao

import (
	"errors"
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

// Errors from the YAML parser contain the line of the error. Only the start of the message is
// matched, since other errors can contain "line N:" in values from the config or command output.
var yamlErrorLine = regexp.MustCompile(`^(?:yaml: )?(?:unmarshal errors:\n\s*)?line (\d+): (.*)`)

// decodeErrors returns the errors of decoding YAML, one for each field that failed to decode
func decodeErrors(err error) []error {
	if typeErr, ok := errors.AsType[*yaml.TypeError](err); ok {
		return core.StringsToErrors(typeErr.Errors)
	}
	return []error{err}
}

func FormatErrors(re Resource, errs []error) error {
	var msg = ""
	context := re.GetContext()

	var errPrefix = color.FgRed.Sprintf("error")
	var ptrPrefix = color.FgBlue.Sprintf("-->")
	for _, err := range errs {
		match := yamlErrorLine.FindStringSubmatch(err.Error())

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

// Env is an environment variable defined in a config, in the env section of the config, a project
// or a task
type Env struct {
	Name  string
	Value string

	context     string
	contextLine int
}

func (e *Env) GetContext() string {
	return e.context
}

func (e *Env) GetContextLine() int {
	return e.contextLine
}

func (e *Env) String() string {
	return fmt.Sprintf("%v=%v", e.Name, e.Value)
}

// parseEnv returns the env in an env field of the config file context, which must be a mapping of
// names to single values. Merge keys (<<) are supported, see mappingPairs.
func parseEnv(env *yaml.Node, context string) ([]Env, error) {
	pairs, ok := mappingPairs(env)
	if !ok {
		return nil, &core.InvalidEnv{}
	}

	var envs []Env
	for _, p := range pairs {
		if p.value.Kind != yaml.ScalarNode {
			return nil, &core.InvalidEnv{Name: p.key.Value}
		}
		envs = append(envs, Env{Name: p.key.Value, Value: p.value.Value, context: context, contextLine: p.key.Line})
	}

	return envs, nil
}

// ParseNodeEnv returns an env field in the form [key=value, key1=$(echo 123)]. Env fields are
// checked when the config is loaded, so an invalid one is ignored here.
func ParseNodeEnv(node yaml.Node) []string {
	var envs []string
	parsed, _ := parseEnv(&node, "")
	for _, env := range parsed {
		envs = append(envs, env.String())
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
