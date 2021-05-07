package test

import (
	"testing"
)

var infoTests = []TemplateTest {
	{ "mani.yaml", "mani.yaml", "info/simple.golden", false, "info" },
	{ "mani.yaml", "mani.yaml", "info/config.golden", false, "info -c ./mani.yaml" },
}

func TestInfoCmd(t *testing.T) {
	for _, tt := range infoTests {
		t.Run(tt.Tmpl, func (t *testing.T) {
			Run(t, tt)
		})
	}
}
