package integration

import (
	"fmt"
	"testing"
)

func TestVersion(t *testing.T) {
	var cases = []TemplateTest{
		{
			TestName:   "Print version when no mani config is found",
			InputFiles: []string{},
			TestCmd:    "mani --version",
			Ignore:     true,
			WantErr:    false,
		},

		{
			TestName:   "Print version when mani config is found",
			InputFiles: []string{"mani-advanced/mani.yaml"},
			TestCmd:    "mani --version",
			Ignore:     true,
			WantErr:    false,
		},
	}

	for i, tt := range cases {
		cases[i].Golden = fmt.Sprintf("version/golden-%d", i)
		cases[i].Index = i
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, cases[i])
		})
	}
}
