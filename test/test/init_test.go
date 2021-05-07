package test

import (
	"testing"
)

var initTests = []TemplateTest {
	{ "", "mani.yaml", "init/simple.golden", false, "init" },
}

func TestInit(t *testing.T) {
	for _, tt := range initTests {
		t.Run(tt.Tmpl, func (t *testing.T) {
			Run(t, tt)
		})
	}
}

