package test

import (
	"testing"
)

var listTests = []TemplateTest {
	// Projects
	{ "mani-empty.yaml", "List 0 projects", "list/projects-empty.golden", false, "list projects" },
	{ "mani.yaml", "List multiple projects", "list/projects.golden", false, "list projects" },
	{ "mani.yaml", "List only project names and no description/tags", "list/projects-raw.golden", false, "list projects --list-raw" },
	{ "mani.yaml", "List projects matching 1 tag", "list/projects-with-1-tag.golden", false, "list projects --tags frontend" },
	{ "mani.yaml", "List projects matching multiple tags", "list/projects-with-2-tags.golden", false, "list projects --tags tmux,frontend" },
	{ "mani.yaml", "List 0 projects on non-existent tag", "list/projects-with-1-tag-non-existing-empty.golden", false, "list projects --tags lala" },
	{ "mani.yaml", "List 0 projects on 2 non-matching tags", "list/projects-with-2-tags-empty.golden", false, "list projects --tags frontend,tmux" },

	// Tags
	{ "mani.yaml", "List all tags", "list/tags.golden", false, "list tags" },
	{ "mani.yaml", "List tags matching one project", "list/tags-with-1-project.golden", false, "list tags --projects pinto" },
	{ "mani.yaml", "List tags matching multiple projects", "list/tags-with-2-projects.golden", false, "list tags --projects pinto,dashgrid" },
	{ "mani.yaml", "List tags matching non-existent project", "list/tags-with-1-project-non-existing-empty.golden", false, "list tags --projects lala" },

	// Commands
	{ "mani-no-commands.yaml", "List 0 commands when no commands exists ", "list/commands-empty.golden", false, "list commands" },
	{ "mani.yaml", "List all commands", "list/commands.golden", false, "list commands" },
}

func TestListCmd(t *testing.T) {
	for _, tt := range listTests {
		t.Run(tt.Tmpl, func (t *testing.T) {
			Run(t, tt)
		})
	}
}
