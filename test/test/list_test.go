package test

import (
	"testing"
)

var listTests = []TemplateTest {
	// Projects
	{ "mani-empty.yaml", "mani.yaml", "list/projects-empty.golden", false, "list projects" },
	{ "mani.yaml", "mani.yaml", "list/projects.golden", false, "list projects" },
	{ "mani.yaml", "mani.yaml", "list/projects-raw.golden", false, "list projects --list-raw" },
	{ "mani.yaml", "mani.yaml", "list/projects-with-1-tag.golden", false, "list projects --tags frontend" },
	{ "mani.yaml", "mani.yaml", "list/projects-with-2-tags.golden", false, "list projects --tags tmux" },
	{ "mani.yaml", "mani.yaml", "list/projects-with-1-tag-non-existing-empty.golden", false, "list projects --tags lala" },
	{ "mani.yaml", "mani.yaml", "list/projects-with-2-tags-empty.golden", false, "list projects --tags frontend,tmux" },

	// Tags
	{ "mani.yaml", "mani.yaml", "list/tags.golden", false, "list tags" },
	{ "mani.yaml", "mani.yaml", "list/tags-with-1-project.golden", false, "list tags --projects pinto" },
	{ "mani.yaml", "mani.yaml", "list/tags-with-2-projects.golden", false, "list tags --projects pinto,dashgrid" },
	{ "mani.yaml", "mani.yaml", "list/tags-with-1-project-non-existing-empty.golden", false, "list tags --projects lala" },

	// Commands
	{ "mani-no-commands.yaml", "mani.yaml", "list/commands-empty.golden", false, "list commands" },
	{ "mani.yaml", "mani.yaml", "list/commands.golden", false, "list commands" },
}

func TestListCmd(t *testing.T) {
	for _, tt := range listTests {
		t.Run(tt.Tmpl, func (t *testing.T) {
			Run(t, tt)
		})
	}
}
