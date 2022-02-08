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
		TestCmd:    "mani list projects --output table --no-headers --no-borders --headers name",
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
		TestCmd:    "mani list projects --tags misc,frontend",
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
	{
		TestName:   "List projects matching 1 dir",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list projects --paths frontend",
		Golden:     "list/projects-with-1-paths",
		WantErr:    false,
	},
	{
		TestName:   "List 0 projects with no matching paths",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list projects --paths hello",
		Golden:     "list/projects-0-paths",
		WantErr:    false,
	},

	{
		TestName:   "List empty projects tree",
		InputFiles: []string{"mani-empty/mani.yaml"},
		TestCmd:    "mani list projects --tree",
		Golden:     "list/projects-empty-tree",
		WantErr:    false,
	},
	{
		TestName:   "List full tree",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list projects --tree",
		Golden:     "list/projects-full-tree",
		WantErr:    false,
	},
	{
		TestName:   "List tree filtered on tag",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list projects --tree --tags frontend",
		Golden:     "list/projects-tags-tree",
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
		TestName:   "List two tags",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list tags frontend misc",
		Golden:     "list/tags-2-args",
		WantErr:    false,
	},

	// Tasks
	{
		TestName:   "List 0 tasks when no tasks exists ",
		InputFiles: []string{"mani-no-tasks/mani.yaml"},
		TestCmd:    "mani list tasks",
		Golden:     "list/tasks-empty",
		WantErr:    false,
	},
	{
		TestName:   "List all tasks",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list tasks",
		Golden:     "list/tasks",
		WantErr:    false,
	},
	{
		TestName:   "List two args",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani list tasks fetch status",
		Golden:     "list/tasks-2-args",
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
