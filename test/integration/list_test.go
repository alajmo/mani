package integration

import (
	"testing"
)

var listTests = []TemplateTest{
	// Projects
	{
		TestName:   "List 0 projects",
		InputFiles: []string{"mani-empty/mani.yaml"},
		TestCmd:    "mani list projects",
		Golden:     "list/projects-empty",
		WantErr:    false,
	},
	{
		TestName:   "List 0 projects on non-existent tag",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list projects --tags lala",
		Golden:     "list/projects-with-1-tag-non-existing-empty",
		WantErr:    false,
	},
	{
		TestName:   "List 0 projects on 2 non-matching tags",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list projects --tags frontend,tmux",
		Golden:     "list/projects-with-2-tags-empty",
		WantErr:    false,
	},
	{
		TestName:   "List multiple projects",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list projects",
		Golden:     "list/projects",
		WantErr:    false,
	},
	{
		TestName:   "List only project names and no description/tags",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list projects --format table --no-headers --no-borders --headers name",
		Golden:     "list/project-names",
		WantErr:    false,
	},
	{
		TestName:   "List projects matching 1 tag",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list projects --tags frontend",
		Golden:     "list/projects-with-1-tag",
		WantErr:    false,
	},
	{
		TestName:   "List projects matching multiple tags",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list projects --tags tmux,frontend",
		Golden:     "list/projects-with-2-tags",
		WantErr:    false,
	},
	{
		TestName:   "List two projects",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list projects pinto dashgrid",
		Golden:     "list/projects-2-args",
		WantErr:    false,
	},

	// Tags
	{
		TestName:   "List all tags",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list tags",
		Golden:     "list/tags",
		WantErr:    false,
	},
	{
		TestName:   "List tags matching one project",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list tags --projects pinto",
		Golden:     "list/tags-with-1-project",
		WantErr:    false,
	},
	{
		TestName:   "List tags matching multiple projects",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list tags --projects pinto,dashgrid",
		Golden:     "list/tags-with-2-projects",
		WantErr:    false,
	},
	{
		TestName:   "List tags matching non-existent project",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list tags --projects lala",
		Golden:     "list/tags-with-1-project-non-existing-empty",
		WantErr:    false,
	},
	{
		TestName:   "List two tags",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list tags frontend misc",
		Golden:     "list/tags-2-args",
		WantErr:    false,
	},

	// Commands
	{
		TestName:   "List 0 commands when no commands exists ",
		InputFiles: []string{"mani-no-commands/mani.yaml"},
		TestCmd:    "mani list commands",
		Golden:     "list/commands-empty",
		WantErr:    false,
	},
	{
		TestName:   "List all commands",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list commands",
		Golden:     "list/commands",
		WantErr:    false,
	},
	{
		TestName:   "List two args",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list commands fetch status",
		Golden:     "list/commands-2-args",
		WantErr:    false,
	},
}

func TestListCmd(t *testing.T) {
	for _, tt := range listTests {
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, tt)
		})
	}
}
