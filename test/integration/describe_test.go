package integration

import (
	"testing"
)

var describeTests = []TemplateTest{
	// Projects
	{
		TestName:   "Describe 0 projects when there's 0 projects",
		InputFiles: []string{"mani-empty/mani.yaml"},
		TestCmd:    "mani describe projects",
		Golden:     "describe/projects-empty",
		WantErr:    false,
	},
	{
		TestName:   "Describe 0 projects on non-existent tag",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani describe projects --tags lala",
		Golden:     "describe/projects-with-1-tag-non-existing-empty",
		WantErr:    false,
	},
	{
		TestName:   "Describe 0 projects on 2 non-matching tags",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani describe projects --tags frontend,tmux",
		Golden:     "describe/projects-with-2-tags-empty",
		WantErr:    false,
	},
	{
		TestName:   "Describe all projects",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani describe projects",
		Golden:     "describe/projects",
		WantErr:    false,
	},
	{
		TestName:   "Describe projects matching 1 tag",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani describe projects --tags frontend",
		Golden:     "describe/projects-with-1-tag",
		WantErr:    false,
	},
	{
		TestName:   "Describe projects matching multiple tags",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani describe projects --tags tmux,frontend",
		Golden:     "describe/projects-with-2-tags",
		WantErr:    false,
	},
	{
		TestName:   "Describe 1 project",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani describe projects pinto",
		Golden:     "describe/projects-1-args",
		WantErr:    false,
	},

	// Commands
	{
		TestName:   "Describe 0 commands when no commands exists ",
		InputFiles: []string{"mani-no-commands/mani.yaml"},
		TestCmd:    "mani describe commands",
		Golden:     "describe/commands-empty",
		WantErr:    false,
	},
	{
		TestName:   "Describe all commands",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani describe commands",
		Golden:     "describe/commands",
		WantErr:    false,
	},
	{
		TestName:   "Describe 1 commands",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani describe commands status",
		Golden:     "describe/commands-1-arg",
		WantErr:    false,
	},
}

func TestDescribeCmd(t *testing.T) {
	for _, tt := range describeTests {
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, tt)
		})
	}
}
