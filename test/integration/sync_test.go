package integration

import (
	"fmt"
	"testing"
)

func TestSync(t *testing.T) {
	var cases = []TemplateTest{
		{
			TestName:   "Throw error when trying to sync a non-existing mani repository",
			InputFiles: []string{},
			TestCmd: `
			mani sync
			`,
			WantErr: true,
		},

		{
			TestName:   "Should sync",
			InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
			TestCmd: `
			mani sync
			`,
			WantErr: false,
		},
	}

	for i, tt := range cases {
		cases[i].Golden = fmt.Sprintf("sync/golden-%d", i)
		cases[i].Index = i
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, cases[i])
		})
	}
}
