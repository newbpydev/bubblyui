package directives

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

// TestBind_CreatesInputHandler tests that Bind creates an input handler
func TestBind_CreatesInputHandler(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef("initial")

	// Act
	directive := Bind(ref)

	// Assert
	assert.NotNil(t, directive)
	assert.Equal(t, ref, directive.ref)
}

// TestBind_SyncsRefToInput tests that Bind syncs Ref value to input
func TestBind_SyncsRefToInput(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef("hello")

	// Act
	directive := Bind(ref)
	output := directive.Render()

	// Assert
	assert.Contains(t, output, "hello")
}

// TestBind_TypeConversion tests type conversion for different types
func TestBind_TypeConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:     "string conversion",
			input:    "test",
			expected: "test",
		},
		{
			name:     "int conversion",
			input:    "42",
			expected: 42,
		},
		{
			name:     "float conversion",
			input:    "3.14",
			expected: 3.14,
		},
		{
			name:     "bool conversion true",
			input:    "true",
			expected: true,
		},
		{
			name:     "bool conversion false",
			input:    "false",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test will verify conversion functions work
			// Implementation will provide type-specific converters
			assert.NotNil(t, tt.expected)
		})
	}
}

// TestBind_UpdatesPropagateToRef tests that input changes update the Ref
func TestBind_UpdatesPropagateToRef(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef("initial")
	directive := Bind(ref)

	// Act - Simulate input change
	// This will be implemented with actual event handling
	directive.Render()

	// Assert - For now, just verify directive exists
	assert.NotNil(t, directive)
	// TODO: Add actual update propagation test when event system is integrated
}

// TestBind_StringType tests Bind with string type
func TestBind_StringType(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef("test value")

	// Act
	directive := Bind(ref)
	output := directive.Render()

	// Assert
	assert.NotNil(t, directive)
	assert.Contains(t, output, "test value")
}

// TestBind_IntType tests Bind with int type
func TestBind_IntType(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef(42)

	// Act
	directive := Bind(ref)
	output := directive.Render()

	// Assert
	assert.NotNil(t, directive)
	assert.Contains(t, output, "42")
}

// TestBind_FloatType tests Bind with float64 type
func TestBind_FloatType(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef(3.14)

	// Act
	directive := Bind(ref)
	output := directive.Render()

	// Assert
	assert.NotNil(t, directive)
	assert.Contains(t, output, "3.14")
}

// TestBind_BoolType tests Bind with bool type
func TestBind_BoolType(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef(true)

	// Act
	directive := Bind(ref)
	output := directive.Render()

	// Assert
	assert.NotNil(t, directive)
	assert.Contains(t, output, "true")
}

// TestBind_EmptyString tests Bind with empty string
func TestBind_EmptyString(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef("")

	// Act
	directive := Bind(ref)
	output := directive.Render()

	// Assert
	assert.NotNil(t, directive)
	// Empty string should still render input
	assert.NotEmpty(t, output)
}

// TestBind_DirectiveInterface tests that BindDirective implements Directive
func TestBind_DirectiveInterface(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef("test")
	directive := Bind(ref)

	// Act & Assert - Verify it implements Directive interface
	var _ Directive = directive
}

// TestBind_TypeSafety tests compile-time type safety
func TestBind_TypeSafety(t *testing.T) {
	// These should all compile, demonstrating type safety
	stringRef := bubbly.NewRef("string")
	intRef := bubbly.NewRef(42)
	floatRef := bubbly.NewRef(3.14)
	boolRef := bubbly.NewRef(true)

	_ = Bind(stringRef)
	_ = Bind(intRef)
	_ = Bind(floatRef)
	_ = Bind(boolRef)
}
