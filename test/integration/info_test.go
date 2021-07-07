package integration

import (
	"testing"
)

var infoTests = []TemplateTest{
	{
		TestName:   "Print info",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani info",
		Golden:     "info/simple",
		WantErr:    false,
	},

	{
		TestName:   "Print info when specifying config file",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani info -c ./mani.yaml",
		Golden:     "info/config",
		WantErr:    false,
	},

	{
		TestName:   "Print info when omitting config file and no config found",
		InputFiles: []string{},
		TestCmd:    "cd /tmp && mani info",
		Golden:     "info/no-config",
		WantErr:    false,
	},
}

func TestInfoCmd(t *testing.T) {
	for _, tt := range infoTests {
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, tt)
		})
	}
}
