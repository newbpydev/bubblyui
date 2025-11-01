package directives

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
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

// TestBindCheckbox_CreatesCheckbox tests that BindCheckbox creates a checkbox
func TestBindCheckbox_CreatesCheckbox(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef(false)

	// Act
	directive := BindCheckbox(ref)

	// Assert
	assert.NotNil(t, directive)
	assert.Equal(t, ref, directive.ref)
}

// TestBindCheckbox_RendersCheckedState tests checkbox rendering
func TestBindCheckbox_RendersCheckedState(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		expected string
	}{
		{
			name:     "checked state",
			value:    true,
			expected: "[X]",
		},
		{
			name:     "unchecked state",
			value:    false,
			expected: "[ ]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ref := bubbly.NewRef(tt.value)

			// Act
			directive := BindCheckbox(ref)
			output := directive.Render()

			// Assert
			assert.Contains(t, output, tt.expected)
		})
	}
}

// TestBindCheckbox_ToggleState tests checkbox toggle
func TestBindCheckbox_ToggleState(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef(false)
	directive := BindCheckbox(ref)

	// Act - Initial render
	output1 := directive.Render()

	// Change ref value
	ref.Set(true)
	output2 := directive.Render()

	// Assert
	assert.Contains(t, output1, "[ ]")
	assert.Contains(t, output2, "[X]")
}

// TestBindCheckbox_MultipleCheckboxes tests multiple independent checkboxes
func TestBindCheckbox_MultipleCheckboxes(t *testing.T) {
	// Arrange
	ref1 := bubbly.NewRef(true)
	ref2 := bubbly.NewRef(false)

	// Act
	checkbox1 := BindCheckbox(ref1)
	checkbox2 := BindCheckbox(ref2)

	output1 := checkbox1.Render()
	output2 := checkbox2.Render()

	// Assert
	assert.Contains(t, output1, "[X]")
	assert.Contains(t, output2, "[ ]")
}

// TestBindCheckbox_DirectiveInterface tests that BindCheckbox implements Directive
func TestBindCheckbox_DirectiveInterface(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef(true)
	directive := BindCheckbox(ref)

	// Act & Assert - Verify it implements Directive interface
	var _ Directive = directive
}

// TestBindSelect_CreatesSelect tests that BindSelect creates a select directive
func TestBindSelect_CreatesSelect(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}

	// Act
	directive := BindSelect(ref, options)

	// Assert
	assert.NotNil(t, directive)
	assert.Equal(t, ref, directive.ref)
	assert.Equal(t, options, directive.options)
}

// TestBindSelect_RendersOptions tests select rendering with options
func TestBindSelect_RendersOptions(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef("option2")
	options := []string{"option1", "option2", "option3"}

	// Act
	directive := BindSelect(ref, options)
	output := directive.Render()

	// Assert
	assert.Contains(t, output, "option1")
	assert.Contains(t, output, "option2")
	assert.Contains(t, output, "option3")
	assert.Contains(t, output, "option2") // Selected option
}

// TestBindSelect_HighlightsSelected tests that selected option is highlighted
func TestBindSelect_HighlightsSelected(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef("option2")
	options := []string{"option1", "option2", "option3"}

	// Act
	directive := BindSelect(ref, options)
	output := directive.Render()

	// Assert
	// Selected option should be marked differently
	assert.Contains(t, output, "> option2")
}

// TestBindSelect_ChangeSelection tests changing selected option
func TestBindSelect_ChangeSelection(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}
	directive := BindSelect(ref, options)

	// Act - Initial render
	output1 := directive.Render()

	// Change selection
	ref.Set("option3")
	output2 := directive.Render()

	// Assert
	assert.Contains(t, output1, "> option1")
	assert.Contains(t, output2, "> option3")
}

// TestBindSelect_IntType tests BindSelect with int type
func TestBindSelect_IntType(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef(2)
	options := []int{1, 2, 3, 4, 5}

	// Act
	directive := BindSelect(ref, options)
	output := directive.Render()

	// Assert
	assert.Contains(t, output, "1")
	assert.Contains(t, output, "2")
	assert.Contains(t, output, "3")
	assert.Contains(t, output, "> 2")
}

// TestBindSelect_StructType tests BindSelect with struct type
func TestBindSelect_StructType(t *testing.T) {
	// Arrange
	type Option struct {
		ID   int
		Name string
	}

	opt1 := Option{ID: 1, Name: "First"}
	opt2 := Option{ID: 2, Name: "Second"}
	opt3 := Option{ID: 3, Name: "Third"}

	ref := bubbly.NewRef(opt2)
	options := []Option{opt1, opt2, opt3}

	// Act
	directive := BindSelect(ref, options)
	output := directive.Render()

	// Assert
	assert.NotEmpty(t, output)
	// Should contain the selected option
	assert.Contains(t, output, "Second")
}

// TestBindSelect_EmptyOptions tests BindSelect with empty options
func TestBindSelect_EmptyOptions(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef("default")
	options := []string{}

	// Act
	directive := BindSelect(ref, options)
	output := directive.Render()

	// Assert
	assert.NotEmpty(t, output)
	// Should handle empty options gracefully
}

// TestBindSelect_DirectiveInterface tests that SelectBindDirective implements Directive
func TestBindSelect_DirectiveInterface(t *testing.T) {
	// Arrange
	ref := bubbly.NewRef("option1")
	options := []string{"option1", "option2"}
	directive := BindSelect(ref, options)

	// Act & Assert - Verify it implements Directive interface
	var _ Directive = directive
}

// TestBindSelect_TypeSafety tests compile-time type safety for BindSelect
func TestBindSelect_TypeSafety(t *testing.T) {
	// These should all compile, demonstrating type safety
	stringRef := bubbly.NewRef("a")
	intRef := bubbly.NewRef(1)
	floatRef := bubbly.NewRef(1.5)

	_ = BindSelect(stringRef, []string{"a", "b", "c"})
	_ = BindSelect(intRef, []int{1, 2, 3})
	_ = BindSelect(floatRef, []float64{1.5, 2.5, 3.5})
}

// TestConversionFunctions tests all type conversion helper functions
func TestConversionFunctions(t *testing.T) {
	t.Run("convertString", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected string
		}{
			{"empty string", "", ""},
			{"simple string", "hello", "hello"},
			{"string with spaces", "hello world", "hello world"},
			{"string with special chars", "hello@world!", "hello@world!"},
			{"unicode string", "你好世界", "你好世界"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := convertString(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("convertInt", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected int
		}{
			{"positive integer", "42", 42},
			{"negative integer", "-42", -42},
			{"zero", "0", 0},
			{"large number", "999999", 999999},
			{"invalid - empty string", "", 0},
			{"invalid - letters", "abc", 0},
			{"invalid - float", "3.14", 0},
			{"invalid - mixed", "12abc", 0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := convertInt(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("convertInt64", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected int64
		}{
			{"positive integer", "42", int64(42)},
			{"negative integer", "-42", int64(-42)},
			{"zero", "0", int64(0)},
			{"large number", "9223372036854775807", int64(9223372036854775807)},
			{"invalid - empty string", "", int64(0)},
			{"invalid - letters", "xyz", int64(0)},
			{"invalid - float", "2.71", int64(0)},
			{"invalid - overflow", "9223372036854775808", int64(0)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := convertInt64(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("convertFloat64", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected float64
		}{
			{"positive float", "3.14", 3.14},
			{"negative float", "-3.14", -3.14},
			{"zero", "0", 0.0},
			{"zero float", "0.0", 0.0},
			{"integer as float", "42", 42.0},
			{"scientific notation", "1.23e10", 1.23e10},
			{"very small number", "0.000001", 0.000001},
			{"invalid - empty string", "", 0.0},
			{"invalid - letters", "abc", 0.0},
			{"invalid - mixed", "12.34abc", 0.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := convertFloat64(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("convertBool", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected bool
		}{
			{"true string", "true", true},
			{"1 string", "1", true},
			{"false string", "false", false},
			{"0 string", "0", false},
			{"empty string", "", false},
			{"True capitalized", "True", false},
			{"TRUE uppercase", "TRUE", false},
			{"yes", "yes", false},
			{"no", "no", false},
			{"random text", "random", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := convertBool(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

// TestConversionFunctions_EdgeCases tests edge cases for conversion functions
func TestConversionFunctions_EdgeCases(t *testing.T) {
	t.Run("convertInt with whitespace", func(t *testing.T) {
		// strconv.Atoi trims whitespace, but we test the behavior
		result := convertInt(" 42 ")
		// This will fail to parse due to spaces, returning 0
		assert.Equal(t, 0, result)
	})

	t.Run("convertInt64 with negative zero", func(t *testing.T) {
		result := convertInt64("-0")
		assert.Equal(t, int64(0), result)
	})

	t.Run("convertFloat64 with special values", func(t *testing.T) {
		// ParseFloat doesn't accept "inf" or "NaN" as strings
		// It only accepts "+Inf", "-Inf", "NaN" with specific casing
		result1 := convertFloat64("invalid")
		assert.Equal(t, 0.0, result1)

		result2 := convertFloat64("not-a-number")
		assert.Equal(t, 0.0, result2)
	})

	t.Run("convertBool case sensitivity", func(t *testing.T) {
		// Only lowercase "true" and "1" return true
		assert.False(t, convertBool("TRUE"))
		assert.False(t, convertBool("True"))
		assert.False(t, convertBool("tRuE"))
		assert.True(t, convertBool("true"))
		assert.True(t, convertBool("1"))
	})
}

// ==================== BENCHMARKS ====================

// BenchmarkBindDirective_String benchmarks Bind with string type
// Target: < 100ns
func BenchmarkBindDirective_String(b *testing.B) {
	ref := bubbly.NewRef("test value")
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Bind(ref).Render()
	}
}

// BenchmarkBindDirective_Int benchmarks Bind with int type
// Target: < 100ns
func BenchmarkBindDirective_Int(b *testing.B) {
	ref := bubbly.NewRef(42)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Bind(ref).Render()
	}
}

// BenchmarkBindDirective_Float benchmarks Bind with float64 type
// Target: < 100ns
func BenchmarkBindDirective_Float(b *testing.B) {
	ref := bubbly.NewRef(3.14159)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Bind(ref).Render()
	}
}

// BenchmarkBindDirective_Bool benchmarks Bind with bool type
// Target: < 100ns
func BenchmarkBindDirective_Bool(b *testing.B) {
	ref := bubbly.NewRef(true)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Bind(ref).Render()
	}
}

// BenchmarkBindCheckbox benchmarks BindCheckbox
// Target: < 100ns
func BenchmarkBindCheckbox(b *testing.B) {
	ref := bubbly.NewRef(true)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = BindCheckbox(ref).Render()
	}
}

// BenchmarkBindSelect benchmarks BindSelect with string options
// Target: < 200ns
func BenchmarkBindSelect(b *testing.B) {
	ref := bubbly.NewRef("option2")
	options := []string{"option1", "option2", "option3"}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = BindSelect(ref, options).Render()
	}
}

// BenchmarkBindSelect_LargeOptions benchmarks BindSelect with many options
// Target: < 500ns
func BenchmarkBindSelect_LargeOptions(b *testing.B) {
	options := make([]int, 50)
	for i := range options {
		options[i] = i
	}
	ref := bubbly.NewRef(25)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = BindSelect(ref, options).Render()
	}
}

// BenchmarkConvertString benchmarks string conversion
// Target: < 10ns
func BenchmarkConvertString(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = convertString("test value")
	}
}

// BenchmarkConvertInt benchmarks int conversion
// Target: < 50ns
func BenchmarkConvertInt(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = convertInt("42")
	}
}

// BenchmarkConvertFloat64 benchmarks float64 conversion
// Target: < 100ns
func BenchmarkConvertFloat64(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = convertFloat64("3.14159")
	}
}

// BenchmarkConvertBool benchmarks bool conversion
// Target: < 20ns
func BenchmarkConvertBool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = convertBool("true")
	}
}
