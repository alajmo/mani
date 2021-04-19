package integration

import (
	"testing"
)

var initTests = []TemplateTest{
	{
		TestName:   "Initialize mani in empty directory",
		InputFiles: []string{},
		TestCmd:    "mani init",
		Golden:     "init/empty",
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
		Golden:  "init/auto-discovery",
		WantErr: false,
	},

	{
		TestName:   "Throw error when initialize in existing mani directory",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani init",
		Golden:     "init/existing",
		WantErr:    true,
	},
}

func TestInitCmd(t *testing.T) {
	for _, tt := range initTests {
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, tt)
		})
	}
}
