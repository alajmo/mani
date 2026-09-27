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
	// Errors from the YAML parser contain the line of the error. Only the start of the message is
	// matched, since other errors can contain "line N:" in values from the config or command output.
	partsRe := regexp.MustCompile(`^(?:yaml: )?(?:unmarshal errors:\n\s*)?line (\d+): (.*)`)

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

// validateEnv returns an error if an env field isn't a mapping of names to single values
func validateEnv(env *yaml.Node) error {
	node := resolveAlias(env)
	if node.Kind == 0 || node.Tag == "!!null" {
		return nil
	}
	if node.Kind != yaml.MappingNode {
		return &core.InvalidEnv{}
	}

	for i := 0; i+1 < len(node.Content); i += 2 {
		key, value := node.Content[i], node.Content[i+1]
		if key.Tag == "!!merge" {
			for _, mapping := range mergedMappings(value) {
				if err := validateEnv(mapping); err != nil {
					return err
				}
			}
			continue
		}
		if resolveAlias(value).Kind != yaml.ScalarNode {
			return &core.InvalidEnv{Name: key.Value}
		}
	}

	return nil
}

// envList returns the env in an env field of the config file context, checked with validateEnv.
// Like when decoding, env set in the mapping takes precedence over env from merge keys (<<), and
// earlier merged mappings take precedence over later ones. Merged env takes the position of the
// merge key.
func envList(env *yaml.Node, context string) []Env {
	node := resolveAlias(env)

	set := make(map[string]bool)
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Tag != "!!merge" {
			set[node.Content[i].Value] = true
		}
	}

	var envs []Env
	for i := 0; i+1 < len(node.Content); i += 2 {
		key, value := node.Content[i], node.Content[i+1]
		if key.Tag != "!!merge" {
			envs = append(envs, Env{
				Name:        key.Value,
				Value:       resolveAlias(value).Value,
				context:     context,
				contextLine: key.Line,
			})
			continue
		}

		for _, mapping := range mergedMappings(value) {
			for _, merged := range envList(mapping, context) {
				if !set[merged.Name] {
					set[merged.Name] = true
					envs = append(envs, merged)
				}
			}
		}
	}

	return envs
}

// ParseNodeEnv returns an env field, checked with validateEnv, in the form [key=value, key1=$(echo 123)]
func ParseNodeEnv(node yaml.Node) []string {
	var envs []string
	for _, env := range envList(&node, "") {
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
