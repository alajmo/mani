package dao

import (
	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

type Spec struct {
	Name              string `yaml:"name"`
	Output            string `yaml:"output"`
	Parallel          bool   `yaml:"parallel"`
	IgnoreErrors      bool   `yaml:"ignore_errors"`
	IgnoreNonExisting bool   `yaml:"ignore_non_existing"`
	OmitEmptyRows     bool   `yaml:"omit_empty_rows"`
	OmitEmptyColumns  bool   `yaml:"omit_empty_columns"`
	ClearOutput       bool   `yaml:"clear_output"`
	Forks             uint32 `yaml:"forks"`

	context     string
	contextLine int
	node        *yaml.Node // definition in the config, used to merge specs defined in multiple configs
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
	pairs, err := sectionPairs(&c.Specs, "specs")
	if err != nil {
		return specs, []ResourceErrors[Spec]{{Resource: &Spec{context: c.Path, contextLine: -1}, Errors: []error{err}}}
	}

	specErrors := []ResourceErrors[Spec]{}
	foundErrors := false
	for _, p := range pairs {
		spec := &Spec{
			Name:        p.key.Value,
			context:     c.Path,
			contextLine: p.key.Line,
			node:        p.value,
		}

		err := p.value.Decode(spec)
		if err != nil {
			foundErrors = true
			specError := ResourceErrors[Spec]{Resource: spec, Errors: decodeErrors(err)}
			specErrors = append(specErrors, specError)
			continue
		}

		switch spec.Output {
		case "", "table", "stream", "html", "markdown":
		default:
			foundErrors = true
			specError := ResourceErrors[Spec]{
				Resource: spec,
				Errors:   []error{&core.SpecOutputError{Name: spec.Name, Output: spec.Output}},
			}
			specErrors = append(specErrors, specError)
		}

		if spec.Forks == 0 {
			spec.Forks = 4
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
