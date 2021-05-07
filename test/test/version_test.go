package test

import (
	"testing"
)

var versionTests = []TemplateTest {
	{ "", "Print version when no mani config is found", "version/empty.golden", false, "version" },
	{ "mani.yaml", "Print version when mani config is found", "version/simple.golden", false, "version" },
}

func TestVersion(t *testing.T) {
	for _, tt := range versionTests {
		t.Run(tt.Tmpl, func (t *testing.T) {
			Run(t, tt)
		})
	}
}

