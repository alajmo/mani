package integration

import (
	"testing"
)

var versionTests = []TemplateTest{
	{
		TestName:   "Print version when no mani config is found",
		InputFiles: []string{},
		TestCmd:    "mani version",
		Golden:     "version/empty",
		Ignore:     true,
		WantErr:    false,
	},

	{
		TestName:   "Print version when mani config is found",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani version",
		Golden:     "version/simple",
		Ignore:     true,
		WantErr:    false,
	},
}

func TestVersionCmd(t *testing.T) {
	for _, tt := range versionTests {
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, tt)
		})
	}
}
