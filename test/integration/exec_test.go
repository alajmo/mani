package integration

import (
	"fmt"
	"testing"
)

func TestExec(t *testing.T) {
	var cases = []TemplateTest{
		{
			TestName:   "Should fail to exec when no configuration file found",
			InputFiles: []string{},
			TestCmd: `
				mani exec --all -o table ls
			`,
			WantErr: true,
		},

		{
			TestName:   "Should exec in zero projects",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
				mani sync
				mani exec -o table ls
			`,
			WantErr: false,
		},

		{
			TestName:   "Should exec in all projects",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
				mani sync
				mani exec --all -o table ls
			`,
			WantErr: false,
		},

		{
			TestName:   "Should exec when filtered on project name",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
				mani sync
				mani exec -o table --projects pinto ls
			`,
			WantErr: false,
		},

		{
			TestName:   "Should exec when filtered on tags",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
				mani sync
				mani exec -o table --tags frontend ls
			`,
			WantErr: false,
		},

		{
			TestName:   "Should exec when filtered on cwd",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
				mani sync
				cd template-generator
				mani exec -o table --cwd pwd
			`,
			WantErr: false,
		},

		{
			TestName:   "Should dry run exec",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
				mani sync
				mani exec -o table --dry-run --projects template-generator pwd
			`,
			WantErr: false,
		},
	}

	for i, tt := range cases {
		cases[i].Golden = fmt.Sprintf("exec/golden-%d", i)
		cases[i].Index = i
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, cases[i])
		})
	}
}
