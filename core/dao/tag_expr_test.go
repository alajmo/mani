package dao

import (
	"strings"
	"testing"
)

func TestTagExpression(t *testing.T) {
	projects := []Project{
		{
			Name: "Project A",
			Tags: []string{"active", "git", "frontend"},
		},
		{
			Name: "Project B",
			Tags: []string{"active", "sake", "backend"},
		},
	}

	// Test cases for valid expressions
	validTests := []struct {
		name     string
		expr     string
		project  string
		expected bool
	}{
		{"simple AND", "active && git", "Project A", true},
		{"simple AND false", "active && git", "Project B", false},
		{"simple OR", "git || sake", "Project A", true},
		{"simple OR", "git || sake", "Project B", true},
		{"nested AND-OR", "((active && git) || (sake && backend))", "Project A", true},
		{"nested AND-OR", "((active && git) || (sake && backend))", "Project B", true},
		{"parentheses precedence", "(active && (git || sake))", "Project A", true},
		{"parentheses precedence", "(active && (git || sake))", "Project B", true},
		{"complex expression", "((active && git) || (active && sake)) && (frontend || backend)", "Project A", true},
		{"complex expression", "((active && git) || (active && sake)) && (frontend || backend)", "Project B", true},
		{"NOT operator", "!(active && (git || sake))", "Project A", false},
		{"NOT operator", "!(active && (git || sake))", "Project B", false},
		{"triple nested", "(((active && git) || sake) && backend)", "Project A", false},
		{"triple nested", "(((active && git) || sake) && backend)", "Project B", true},
	}

	t.Run("valid expressions", func(t *testing.T) {
		for _, tt := range validTests {
			t.Run(tt.name, func(t *testing.T) {
				var proj Project
				for _, p := range projects {
					if p.Name == tt.project {
						proj = p
						break
					}
				}

				result, err := evaluateExpression(&proj, tt.expr)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expression %q on project %q: got %v, want %v",
						tt.expr, tt.project, result, tt.expected)
				}
			})
		}
	})

	// Test cases for invalid expressions
	invalidTests := []struct {
		name        string
		expr        string
		expectedErr string
	}{
		{"empty expression", "", "empty expression"},
		{"operator without operands", "&&", "unexpected token"},
		{"missing right operand", "tag &&", "missing right operand"},
		{"missing left operand", "&& tag", "unexpected token"},
		{"empty parentheses", "()", "empty parentheses"},
		{"unmatched parenthesis", "((tag)", "missing closing parenthesis"},
		{"missing operator", "tag tag", "unexpected token"},
		{"double operator", "tag && && tag", "unexpected token"},
		{"NOT without operand", "!", "missing operand after NOT"},
		{"invalid tag character", "tag-with-invalid#", "unexpected character"},
	}

	t.Run("invalid expressions", func(t *testing.T) {
		for _, tt := range invalidTests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := evaluateExpression(&projects[0], tt.expr)
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedErr)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("expected error containing %q, got %q", tt.expectedErr, err.Error())
				}
			})
		}
	})
}
