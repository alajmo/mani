package integration

import (
	"fmt"
	"testing"
)

func TestRun(t *testing.T) {
	var cases = []TemplateTest{
		{
			TestName:   "Should fail to run when no configuration file found",
			InputFiles: []string{},
			TestCmd: `
			mani run pwd --all
			`,
			WantErr: true,
		},

		{
			TestName:   "Should run in zero projects",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
			mani sync
			mani run pwd -o table
			`,
			WantErr: false,
		},

		{
			TestName:   "Should run in all projects",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
			mani sync
			mani run --all pwd
			`,
			WantErr: false,
		},

		{
			TestName:   "Should run when filtered on project",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
			mani sync
			mani run -o table --projects pinto pwd
			`,
			WantErr: false,
		},

		{
			TestName:   "Should run when filtered on tags",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
			mani sync
			mani run -o table --tags frontend pwd
			`,
			WantErr: false,
		},

		{
			TestName:   "Should run when filtered on cwd",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
			mani sync
			cd template-generator
			mani run -o table --cwd pwd
			`,
			WantErr: false,
		},

		{
			TestName:   "Should run on default tags",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
			mani sync
			mani run -o table default-tags
			`,
			WantErr: false,
		},

		{
			TestName:   "Should run on default projects",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
			mani sync
			mani run -o table default-projects
			`,
			WantErr: false,
		},

		{
			TestName:   "Should print table when output set to table in task",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
			mani sync
			mani run default-output -p dashgrid
			`,
			WantErr: false,
		},

		{
			TestName:   "Should dry run",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
			mani sync
			mani run --dry-run --projects template-generator -o table pwd
			`,
			WantErr: false,
		},

		{
			TestName:   "Should run multiple commands",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
			mani sync
			mani run pwd multi -o table --all
			`,
			WantErr: false,
		},

		{
			TestName:   "Should run sub-commands",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
			mani sync
			mani run submarine --all
			`,
			WantErr: false,
		},
	}

	for i, tt := range cases {
		cases[i].Golden = fmt.Sprintf("run/golden-%d", i)
		cases[i].Index = i
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, cases[i])
		})
	}
}
