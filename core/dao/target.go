package dao

import (
	"github.com/alajmo/mani/core"
)

type Target struct {
	Name	 string
	All		 bool
	Projects []string
	Paths    []string
	Tags     []string
	Cwd      bool
}

// Populates TargetList and creates a default target if no default target is set.
func (c *Config) GetTargetList() ([]Target, error) {
	var targets []Target
	count := len(c.Targets.Content)

	for i := 0; i < count; i += 2 {
		target := &Target{}
		err := c.Targets.Content[i+1].Decode(target)
		if err != nil {
			return []Target{}, &core.FailedToParseFile{Name: c.Path, Msg: err}
		}

		target.Name = c.Targets.Content[i].Value
		targets = append(targets, *target)
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
