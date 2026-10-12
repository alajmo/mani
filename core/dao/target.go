package dao

import (
	"path/filepath"

	"github.com/alajmo/mani/core"
)

type Target struct {
	Name     string   `yaml:"name"`
	All      bool     `yaml:"all"`
	Projects []string `yaml:"projects"`
	Paths    []string `yaml:"paths"`
	Tags     []string `yaml:"tags"`
	TagsExpr string   `yaml:"tags_expr"`
	Cwd      bool     `yaml:"cwd"`

	context     string
	contextLine int
	dir         string // Directory that paths are relative to, empty for the root config's directory
}

func (t *Target) GetContext() string {
	return t.context
}

func (t *Target) GetContextLine() int {
	return t.contextLine
}

// pathsDir returns the directory that the paths of a target set in the config file context are
// relative to, which is the directory of that config, so that a target selects the same projects
// wherever the config is imported from. The user configs apply to whichever config mani runs with,
// so the paths of their targets are relative to the root config, like --paths. Returns an empty
// string for the root config's directory.
func (c Config) pathsDir(context string) string {
	if context == "" || c.userConfigs[context] || filepath.Dir(context) == c.Dir {
		return ""
	}

	return filepath.Dir(context)
}

// Populates TargetList and creates a default target if no default target is set.
func (c *Config) GetTargetList() ([]Target, []ResourceErrors[Target]) {
	var targets []Target
	count := len(c.Targets.Content)

	targetErrors := []ResourceErrors[Target]{}
	foundErrors := false
	for i := 0; i < count; i += 2 {
		target := &Target{
			Name:        c.Targets.Content[i].Value,
			context:     c.Path,
			contextLine: c.Targets.Content[i].Line,
		}

		err := c.Targets.Content[i+1].Decode(target)
		if err != nil {
			foundErrors = true
			targetError := ResourceErrors[Target]{Resource: target, Errors: decodeErrors(err)}
			targetErrors = append(targetErrors, targetError)
			continue
		}

		if target.TagsExpr != "" {
			valid := validateExpression(target.TagsExpr)
			if valid != nil {
				foundErrors = true
				targetError := ResourceErrors[Target]{
					Resource: target,
					Errors:   []error{&core.TargetTagsExprError{Name: target.Name, Err: valid}},
				}
				targetErrors = append(targetErrors, targetError)
			}
		}

		targets = append(targets, *target)
	}

	if foundErrors {
		return targets, targetErrors
	}

	return targets, nil
}

func (c Config) GetTarget(name string) (*Target, error) {
	for _, target := range c.TargetList {
		if name == target.Name {
			return &target, nil
		}
	}

	return nil, &core.TargetNotFound{Name: name}
}

func (c Config) GetTargetNames() []string {
	names := []string{}
	for _, target := range c.TargetList {
		names = append(names, target.Name)
	}

	return names
}
