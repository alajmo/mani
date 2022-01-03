package integration

import (
	"testing"
)

var syncTests = []TemplateTest{
	{
		TestName:   "Throw error when trying to sync a non-existing mani repository",
		InputFiles: []string{},
		TestCmd: `
			mani sync --parallel=false
		`,
		Golden:  "sync/empty",
		WantErr: true,
	},

	{
		TestName:   "Should sync",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync --parallel=false
		`,
		Golden:  "sync/simple",
		WantErr: false,
	},
}

func TestSyncCmd(t *testing.T) {
	for _, tt := range syncTests {
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, tt)
		})
	}
}
