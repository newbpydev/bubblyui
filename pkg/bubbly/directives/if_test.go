package directives

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIfDirective_Simple tests basic If directive functionality
func TestIfDirective_Simple(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		thenFunc  func() string
		expected  string
	}{
		{
			name:      "condition true renders then branch",
			condition: true,
			thenFunc:  func() string { return "then branch" },
			expected:  "then branch",
		},
		{
			name:      "condition false renders empty",
			condition: false,
			thenFunc:  func() string { return "then branch" },
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := If(tt.condition, tt.thenFunc).Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIfDirective_WithElse tests If with Else branch
func TestIfDirective_WithElse(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		thenFunc  func() string
		elseFunc  func() string
		expected  string
	}{
		{
			name:      "condition true renders then branch",
			condition: true,
			thenFunc:  func() string { return "then" },
			elseFunc:  func() string { return "else" },
			expected:  "then",
		},
		{
			name:      "condition false renders else branch",
			condition: false,
			thenFunc:  func() string { return "then" },
			elseFunc:  func() string { return "else" },
			expected:  "else",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := If(tt.condition, tt.thenFunc).Else(tt.elseFunc).Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIfDirective_ElseIfChain tests ElseIf chaining
func TestIfDirective_ElseIfChain(t *testing.T) {
	tests := []struct {
		name       string
		condition1 bool
		condition2 bool
		condition3 bool
		expected   string
	}{
		{
			name:       "first condition true",
			condition1: true,
			condition2: false,
			condition3: false,
			expected:   "first",
		},
		{
			name:       "second condition true",
			condition1: false,
			condition2: true,
			condition3: false,
			expected:   "second",
		},
		{
			name:       "third condition true",
			condition1: false,
			condition2: false,
			condition3: true,
			expected:   "third",
		},
		{
			name:       "all conditions false with else",
			condition1: false,
			condition2: false,
			condition3: false,
			expected:   "else",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := If(tt.condition1, func() string { return "first" }).
				ElseIf(tt.condition2, func() string { return "second" }).
				ElseIf(tt.condition3, func() string { return "third" }).
				Else(func() string { return "else" }).
				Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIfDirective_ElseIfWithoutElse tests ElseIf without final Else
func TestIfDirective_ElseIfWithoutElse(t *testing.T) {
	tests := []struct {
		name       string
		condition1 bool
		condition2 bool
		expected   string
	}{
		{
			name:       "first condition true",
			condition1: true,
			condition2: false,
			expected:   "first",
		},
		{
			name:       "second condition true",
			condition1: false,
			condition2: true,
			expected:   "second",
		},
		{
			name:       "all conditions false returns empty",
			condition1: false,
			condition2: false,
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := If(tt.condition1, func() string { return "first" }).
				ElseIf(tt.condition2, func() string { return "second" }).
				Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIfDirective_Nested tests nested If directives
func TestIfDirective_Nested(t *testing.T) {
	tests := []struct {
		name      string
		outerCond bool
		innerCond bool
		expected  string
	}{
		{
			name:      "both conditions true",
			outerCond: true,
			innerCond: true,
			expected:  "inner true",
		},
		{
			name:      "outer true inner false",
			outerCond: true,
			innerCond: false,
			expected:  "inner false",
		},
		{
			name:      "outer false",
			outerCond: false,
			innerCond: true,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := If(tt.outerCond, func() string {
				return If(tt.innerCond, func() string {
					return "inner true"
				}).Else(func() string {
					return "inner false"
				}).Render()
			}).Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIfDirective_EmptyConditions tests edge cases with empty/nil functions
func TestIfDirective_EmptyConditions(t *testing.T) {
	t.Run("empty then function returns empty", func(t *testing.T) {
		result := If(true, func() string { return "" }).Render()
		assert.Equal(t, "", result)
	})

	t.Run("empty else function returns empty", func(t *testing.T) {
		result := If(false, func() string { return "then" }).
			Else(func() string { return "" }).
			Render()
		assert.Equal(t, "", result)
	})

	t.Run("multiple elseif with empty returns", func(t *testing.T) {
		result := If(false, func() string { return "first" }).
			ElseIf(true, func() string { return "" }).
			Else(func() string { return "else" }).
			Render()
		assert.Equal(t, "", result)
	})
}

// TestIfDirective_ComplexContent tests with complex string content
func TestIfDirective_ComplexContent(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		expected  string
	}{
		{
			name:      "multiline content",
			condition: true,
			expected:  "Line 1\nLine 2\nLine 3",
		},
		{
			name:      "content with special characters",
			condition: true,
			expected:  "Special: !@#$%^&*()",
		},
		{
			name:      "unicode content",
			condition: true,
			expected:  "Hello 世界 🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := If(tt.condition, func() string { return tt.expected }).Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIfDirective_ConditionalDirectiveInterface tests interface compliance
func TestIfDirective_ConditionalDirectiveInterface(t *testing.T) {
	t.Run("implements ConditionalDirective interface", func(t *testing.T) {
		var _ ConditionalDirective = If(true, func() string { return "test" })
	})

	t.Run("implements Directive interface", func(t *testing.T) {
		var _ Directive = If(true, func() string { return "test" })
	})
}
