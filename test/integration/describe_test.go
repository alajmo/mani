package integration

import (
	"fmt"
	"testing"
)

func TestDescribe(t *testing.T) {
	var cases = []TemplateTest{
		// Projects
		{
			TestName:   "Describe 0 projects when there's 0 projects",
			InputFiles: []string{"mani-empty/mani.yaml"},
			TestCmd:    "mani describe projects",
			WantErr:    false,
		},
		{
			TestName:   "Describe 0 projects on non-existent tag",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani describe projects --tags lala",
			WantErr:    true,
		},
		{
			TestName:   "Describe 0 projects on 2 non-matching tags",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani describe projects --tags frontend,cli",
			WantErr:    false,
		},
		{
			TestName:   "Describe all projects",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani describe projects",
			WantErr:    false,
		},
		{
			TestName:   "Describe projects matching 1 tag",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani describe projects --tags frontend",
			WantErr:    false,
		},
		{
			TestName:   "Describe projects matching multiple tags",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani describe projects --tags misc,frontend",
			WantErr:    false,
		},
		{
			TestName:   "Describe 1 project",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani describe projects pinto",
			WantErr:    false,
		},

		// Tasks
		{
			TestName:   "Describe 0 tasks when no tasks exists ",
			InputFiles: []string{"mani-no-tasks/mani.yaml"},
			TestCmd:    "mani describe tasks",
			WantErr:    false,
		},
		{
			TestName:   "Describe all tasks",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani describe tasks",
			WantErr:    false,
		},
		{
			TestName:   "Describe 1 tasks",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani describe tasks status",
			WantErr:    false,
		},
	}

	for i, tt := range cases {
		cases[i].Golden = fmt.Sprintf("describe/golden-%d", i)
		cases[i].Index = i
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, cases[i])
		})
	}
}
