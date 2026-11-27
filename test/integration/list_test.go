package integration

import (
	"fmt"
	"testing"
)

func TestList(t *testing.T) {
	var cases = []TemplateTest{
		// Projects
		{
			TestName:   "List 0 projects",
			InputFiles: []string{"mani-empty/mani.yaml"},
			TestCmd:    "mani list projects",
			WantErr:    false,
		},
		{
			TestName:   "List 0 projects on non-existent tag",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list projects --tags lala",
			WantErr:    true,
		},
		{
			TestName:   "List 0 projects on 2 non-matching tags",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list projects --tags frontend,cli",
			WantErr:    false,
		},
		{
			TestName:   "List multiple projects",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list projects",
			WantErr:    false,
		},
		{
			TestName:   "List only project names and no description/tags",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list projects --output table --headers project",
			WantErr:    false,
		},
		{
			TestName:   "List projects matching 1 tag",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list projects --tags frontend",
			WantErr:    false,
		},
		{
			TestName:   "List projects matching multiple tags",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list projects --tags misc,frontend",
			WantErr:    false,
		},
		{
			TestName:   "List two projects",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list projects pinto dashgrid",
			WantErr:    false,
		},
		{
			TestName:   "List projects matching 1 dir",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list projects --paths frontend",
			WantErr:    false,
		},
		{
			TestName:   "List 0 projects with no matching paths",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list projects --paths hello",
			WantErr:    true,
		},

		{
			TestName:   "List empty projects tree",
			InputFiles: []string{"mani-empty/mani.yaml"},
			TestCmd:    "mani list projects --tree",
			WantErr:    false,
		},
		{
			TestName:   "List full tree",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list projects --tree",
			WantErr:    false,
		},
		{
			TestName:   "List tree filtered on tag",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list projects --tree --tags frontend",
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

		// JSON/YAML
		{
			TestName:   "List projects as json",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list projects --output json",
			WantErr:    false,
		},
		{
			TestName:   "List projects as yaml",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list projects --tags frontend --output yaml",
			WantErr:    false,
		},
		{
			TestName:   "List 0 projects as json",
			InputFiles: []string{"mani-empty/mani.yaml"},
			TestCmd:    "mani list projects --output json",
			WantErr:    false,
		},
		{
			TestName:   "List tags as json",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list tags --output json",
			WantErr:    false,
		},
		{
			TestName:   "List tags as yaml",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list tags frontend --output yaml",
			WantErr:    false,
		},
		{
			TestName:   "List 0 tags as yaml",
			InputFiles: []string{"mani-empty/mani.yaml"},
			TestCmd:    "mani list tags --output yaml",
			WantErr:    false,
		},
		{
			TestName:   "List tasks as json",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list tasks fetch submarine --output json",
			WantErr:    false,
		},
		{
			TestName:   "List tasks as yaml",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list tasks fetch submarine --output yaml",
			WantErr:    false,
		},

		// Invalid output
		{
			TestName:   "Fail to list projects with an invalid output",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list projects --output JSON",
			WantErr:    true,
		},
		{
			TestName:   "Fail to list tags with an invalid output",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list tags --output stream",
			WantErr:    true,
		},
		{
			TestName:   "Fail to list tasks with an invalid output",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list tasks --output jsonl",
			WantErr:    true,
		},

		{
			TestName:   "List a tag once when provided twice",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani list tags frontend misc frontend --output json",
			WantErr:    false,
		},
	}

	for i, tt := range cases {
		cases[i].Golden = fmt.Sprintf("list/golden-%d", i)
		cases[i].Index = i

		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, cases[i])
		})
	}
}
