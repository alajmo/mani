package dao

import (
	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

type Target struct {
	Name     string   `yaml:"name"`
	All      bool     `yaml:"all"`
	Projects []string `yaml:"projects"`
	Paths    []string `yaml:"paths"`
	Tags     []string `yaml:"tags"`
	Cwd      bool     `yaml:"cwd"`

	context     string
	contextLine int
}

func (t *Target) GetContext() string {
	return t.context
}

func (t *Target) GetContextLine() int {
	return t.contextLine
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
			targetError := ResourceErrors[Target]{Resource: target, Errors: core.StringsToErrors(err.(*yaml.TypeError).Errors)}
			targetErrors = append(targetErrors, targetError)
			continue
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
