package directives

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDirectiveInterface verifies that the base Directive interface is defined
func TestDirectiveInterface(t *testing.T) {
	t.Run("interface_exists", func(t *testing.T) {
		// This test verifies that the Directive interface exists
		// by attempting to use it as a type constraint
		var _ Directive = (*testDirective)(nil)
	})
}

// TestConditionalDirectiveInterface verifies the ConditionalDirective interface
func TestConditionalDirectiveInterface(t *testing.T) {
	t.Run("interface_exists", func(t *testing.T) {
		// Verify ConditionalDirective interface exists and extends Directive
		var _ ConditionalDirective = (*testConditionalDirective)(nil)
		var _ Directive = (*testConditionalDirective)(nil)
	})

	t.Run("elseif_chaining", func(t *testing.T) {
		// Verify ElseIf returns ConditionalDirective for chaining
		directive := &testConditionalDirective{}
		result := directive.ElseIf(true, func() string { return "test" })
		assert.NotNil(t, result)
		assert.Implements(t, (*ConditionalDirective)(nil), result)
	})

	t.Run("else_chaining", func(t *testing.T) {
		// Verify Else returns ConditionalDirective for chaining
		directive := &testConditionalDirective{}
		result := directive.Else(func() string { return "test" })
		assert.NotNil(t, result)
		assert.Implements(t, (*ConditionalDirective)(nil), result)
	})
}

// TestDirectiveRender verifies that directives can render output
func TestDirectiveRender(t *testing.T) {
	tests := []struct {
		name      string
		directive Directive
		expected  string
	}{
		{
			name:      "simple_directive",
			directive: &testDirective{output: "Hello"},
			expected:  "Hello",
		},
		{
			name:      "empty_directive",
			directive: &testDirective{output: ""},
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.directive.Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test helper types
type testDirective struct {
	output string
}

func (d *testDirective) Render() string {
	return d.output
}

type testConditionalDirective struct {
	testDirective
}

func (d *testConditionalDirective) ElseIf(condition bool, then func() string) ConditionalDirective {
	return d
}

func (d *testConditionalDirective) Else(then func() string) ConditionalDirective {
	return d
}
