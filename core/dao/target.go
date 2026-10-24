package dao

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"

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
	node        *yaml.Node // definition in the config, used to merge targets defined in multiple configs
	dir         string     // Directory that Paths are relative to, empty for the root config, see override
}

func (t *Target) GetContext() string {
	return t.context
}

func (t *Target) GetContextLine() int {
	return t.contextLine
}

// override sets the projects, paths, tags and tag expression that are set, for instance from flags,
// replacing the target's. Paths that are set are relative to the root config.
func (t *Target) override(projects []string, paths []string, tags []string, tagsExpr string) {
	if len(projects) > 0 {
		t.Projects = projects
	}

	if len(paths) > 0 {
		t.Paths = paths
		t.dir = ""
	}

	if len(tags) > 0 {
		t.Tags = tags
	}

	if tagsExpr != "" {
		t.TagsExpr = tagsExpr
	}
}

// targetValue returns the value of a target definition to compare it with other definitions. Its
// paths are relative to the directory dir, see pathsDir, so the same paths in configs in different
// directories are different values.
func targetValue(node *yaml.Node, dir string) string {
	value := nodeString(node)
	if dir == "" {
		return value
	}

	pairs, _ := mappingPairs(node)
	for _, p := range pairs {
		if p.key.Value == "paths" && p.value.Kind == yaml.SequenceNode && len(p.value.Content) > 0 {
			return fmt.Sprintf("%s relative to %s", value, dir)
		}
	}

	return value
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
	pairs, err := sectionPairs(&c.Targets, "targets")
	if err != nil {
		return targets, []ResourceErrors[Target]{{Resource: &Target{context: c.Path, contextLine: -1}, Errors: []error{err}}}
	}

	targetErrors := []ResourceErrors[Target]{}
	foundErrors := false
	for _, p := range pairs {
		target := &Target{
			Name:        p.key.Value,
			context:     c.Path,
			contextLine: p.key.Line,
			node:        p.value,
		}

		err := p.value.Decode(target)
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
