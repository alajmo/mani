package test

import (
	"testing"
)

var infoTests = []TemplateTest {
	{
		"Print info",
		{ "mani.yaml" },
		"",
		"info",

		"info/simple",
		"",

		false,

	BootstrapCmds  []string
	TestCmd        string

	Output         []string
	StdOut         []string

	WantErr      bool
	},

	{ "mani.yaml", "Print info when specifying config file", "info/config.golden", false, "info -c ./mani.yaml" },
	{ "", "Print no info when not found any mani config", "info/simple.golden", false, "info" },
}

func TestInfoCmd(t *testing.T) {
	for _, tt := range infoTests {
		t.Run(tt.TestName, func (t *testing.T) {
			Run(t, tt)
		})
	}
}
