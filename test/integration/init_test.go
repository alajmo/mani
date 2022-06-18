package integration

import (
	"fmt"
	"testing"
)

func TestInit(t *testing.T) {
	var cases = []TemplateTest{
		{
			TestName:   "Initialize mani in empty directory",
			InputFiles: []string{},
			TestCmd:    "mani init",
			WantErr:    false,
		},

		{
			TestName:   "Initialize mani with auto-discovery",
			InputFiles: []string{},
			TestCmd: `
			(mkdir -p dashgrid && touch dashgrid/empty);
			(mkdir -p tap-report && touch tap-report/empty && cd tap-report && git init && git remote add origin https://github.com/alajmo/tap-report);
			(mkdir -p nested/template-generator && touch nested/template-generator/empty && cd nested/template-generator && git init && git remote add origin https://github.com/alajmo/template-generator);
			(mkdir nameless && touch nameless/empty);
			(git init && git remote add origin https://github.com/alajmo/pinto)
			mani init
			`,
			WantErr: false,
		},

		{
			TestName:   "Throw error when initialize in existing mani directory",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani init",
			WantErr:    true,
		},
	}

	for i, tt := range cases {
		cases[i].Golden = fmt.Sprintf("init/golden-%d", i)
		cases[i].Index = i
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, cases[i])
		})
	}
}
