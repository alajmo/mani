package test

import (
	"testing"
)

var versionTests = []TemplateTest {
	{ "", "mani.yaml", "version/empty.golden", false, "version" },
	{ "mani.yaml", "mani.yaml", "version/simple.golden", false, "version" },
}

func TestVersion(t *testing.T) {
	for _, tt := range versionTests {
		t.Run(tt.Tmpl, func (t *testing.T) {
			Run(t, tt)
		})
	}
}

