package dao

import (
	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

type Spec struct {
	Name         string `yaml:"name"`
	Output       string `yaml:"output"`
	Parallel     bool   `yaml:"parallel"`
	IgnoreErrors bool   `yaml:"ignore_errors"`
	OmitEmpty    bool   `yaml:"omit_empty"`

	context     string
	contextLine int
}

func (s *Spec) GetContext() string {
	return s.context
}

func (s *Spec) GetContextLine() int {
	return s.contextLine
}

// Populates SpecList and creates a default spec if no default spec is set.
func (c *Config) GetSpecList() ([]Spec, []ResourceErrors[Spec]) {
	var specs []Spec
	count := len(c.Specs.Content)

	specErrors := []ResourceErrors[Spec]{}
	foundErrors := false
	for i := 0; i < count; i += 2 {
		spec := &Spec{
			Name:        c.Specs.Content[i].Value,
			context:     c.Path,
			contextLine: c.Specs.Content[i].Line,
		}

		err := c.Specs.Content[i+1].Decode(spec)
		if err != nil {
			foundErrors = true
			specError := ResourceErrors[Spec]{Resource: spec, Errors: core.StringsToErrors(err.(*yaml.TypeError).Errors)}
			specErrors = append(specErrors, specError)
			continue
		}
		specs = append(specs, *spec)
	}

	if foundErrors {
		return specs, specErrors
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
