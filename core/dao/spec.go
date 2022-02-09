package dao

import (
	"github.com/alajmo/mani/core"
)

type Spec struct {
	Name        string
	Output      string
	Parallel    bool
	IgnoreError bool `yaml:"ignore_error"`
	OmitEmpty   bool `yaml:"omit_empty"`
}

// Populates SpecList and creates a default spec if no default spec is set.
func (c *Config) GetSpecList() ([]Spec, error) {
	var specs []Spec
	count := len(c.Specs.Content)

	for i := 0; i < count; i += 2 {
		spec := &Spec{}
		err := c.Specs.Content[i+1].Decode(spec)
		if err != nil {
			return []Spec{}, &core.FailedToParseFile{Name: c.Path, Msg: err}
		}

		spec.Name = c.Specs.Content[i].Value
		specs = append(specs, *spec)
	}

	return specs, nil
}

func (c Config) GetSpec(name string) (*Spec, error) {
	for _, spec := range c.SpecList {
		if name == spec.Name {
			return &spec, nil
		}
	}

	return nil, &core.SpecNotFound{Name: name}
}

func (c Config) GetSpecNames() []string {
	names := []string{}
	for _, spec := range c.SpecList {
		names = append(names, spec.Name)
	}

	return names
}
