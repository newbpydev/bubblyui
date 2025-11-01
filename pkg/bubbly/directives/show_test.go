package directives

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestShowDirective_Visible tests basic Show directive visibility
func TestShowDirective_Visible(t *testing.T) {
	tests := []struct {
		name     string
		visible  bool
		content  func() string
		expected string
	}{
		{
			name:     "visible true shows content",
			visible:  true,
			content:  func() string { return "visible content" },
			expected: "visible content",
		},
		{
			name:     "visible false hides content",
			visible:  false,
			content:  func() string { return "hidden content" },
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Show(tt.visible, tt.content).Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestShowDirective_WithTransition tests transition functionality
func TestShowDirective_WithTransition(t *testing.T) {
	tests := []struct {
		name     string
		visible  bool
		content  func() string
		expected string
	}{
		{
			name:     "visible true with transition shows content",
			visible:  true,
			content:  func() string { return "content" },
			expected: "content",
		},
		{
			name:     "visible false with transition shows hidden marker",
			visible:  false,
			content:  func() string { return "content" },
			expected: "[Hidden]content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Show(tt.visible, tt.content).WithTransition().Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestShowDirective_WithoutTransition tests default behavior without transition
func TestShowDirective_WithoutTransition(t *testing.T) {
	tests := []struct {
		name     string
		visible  bool
		content  func() string
		expected string
	}{
		{
			name:     "visible true shows content",
			visible:  true,
			content:  func() string { return "content" },
			expected: "content",
		},
		{
			name:     "visible false returns empty",
			visible:  false,
			content:  func() string { return "content" },
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Show(tt.visible, tt.content).Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestShowDirective_Nested tests nested Show directives
func TestShowDirective_Nested(t *testing.T) {
	tests := []struct {
		name         string
		outerVisible bool
		innerVisible bool
		expected     string
	}{
		{
			name:         "both visible",
			outerVisible: true,
			innerVisible: true,
			expected:     "inner content",
		},
		{
			name:         "outer visible inner hidden",
			outerVisible: true,
			innerVisible: false,
			expected:     "",
		},
		{
			name:         "outer hidden inner visible",
			outerVisible: false,
			innerVisible: true,
			expected:     "",
		},
		{
			name:         "both hidden",
			outerVisible: false,
			innerVisible: false,
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Show(tt.outerVisible, func() string {
				return Show(tt.innerVisible, func() string {
					return "inner content"
				}).Render()
			}).Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestShowDirective_ComplexContent tests with complex string content
func TestShowDirective_ComplexContent(t *testing.T) {
	tests := []struct {
		name     string
		visible  bool
		content  string
		expected string
	}{
		{
			name:     "multiline content visible",
			visible:  true,
			content:  "Line 1\nLine 2\nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "multiline content hidden",
			visible:  false,
			content:  "Line 1\nLine 2\nLine 3",
			expected: "",
		},
		{
			name:     "special characters visible",
			visible:  true,
			content:  "Special: !@#$%^&*()",
			expected: "Special: !@#$%^&*()",
		},
		{
			name:     "unicode content visible",
			visible:  true,
			content:  "Hello 世界 🌍",
			expected: "Hello 世界 🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Show(tt.visible, func() string { return tt.content }).Render()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestShowDirective_EmptyContent tests edge cases with empty content
func TestShowDirective_EmptyContent(t *testing.T) {
	t.Run("empty content visible", func(t *testing.T) {
		result := Show(true, func() string { return "" }).Render()
		assert.Equal(t, "", result)
	})

	t.Run("empty content hidden", func(t *testing.T) {
		result := Show(false, func() string { return "" }).Render()
		assert.Equal(t, "", result)
	})

	t.Run("empty content with transition visible", func(t *testing.T) {
		result := Show(true, func() string { return "" }).WithTransition().Render()
		assert.Equal(t, "", result)
	})

	t.Run("empty content with transition hidden", func(t *testing.T) {
		result := Show(false, func() string { return "" }).WithTransition().Render()
		assert.Equal(t, "[Hidden]", result)
	})
}

// TestShowDirective_FluentAPI tests method chaining
func TestShowDirective_FluentAPI(t *testing.T) {
	t.Run("chaining WithTransition returns same directive", func(t *testing.T) {
		directive := Show(true, func() string { return "test" })
		result := directive.WithTransition()
		assert.Equal(t, directive, result, "WithTransition should return self for chaining")
	})

	t.Run("multiple WithTransition calls work", func(t *testing.T) {
		result := Show(false, func() string { return "test" }).
			WithTransition().
			WithTransition().
			Render()
		assert.Equal(t, "[Hidden]test", result)
	})
}

// TestShowDirective_DirectiveInterface tests interface compliance
func TestShowDirective_DirectiveInterface(t *testing.T) {
	t.Run("implements Directive interface", func(t *testing.T) {
		var _ Directive = Show(true, func() string { return "test" })
	})

	t.Run("can be used as Directive type", func(t *testing.T) {
		var directive Directive = Show(true, func() string { return "content" })
		result := directive.Render()
		assert.Equal(t, "content", result)
	})
}

// TestShowDirective_Performance tests performance characteristics
func TestShowDirective_Performance(t *testing.T) {
	t.Run("content function not called when hidden without transition", func(t *testing.T) {
		called := false
		Show(false, func() string {
			called = true
			return "content"
		}).Render()
		assert.False(t, called, "content function should not be called when hidden without transition")
	})

	t.Run("content function called when hidden with transition", func(t *testing.T) {
		called := false
		Show(false, func() string {
			called = true
			return "content"
		}).WithTransition().Render()
		assert.True(t, called, "content function should be called when hidden with transition")
	})

	t.Run("content function called when visible", func(t *testing.T) {
		called := false
		Show(true, func() string {
			called = true
			return "content"
		}).Render()
		assert.True(t, called, "content function should be called when visible")
	})
}

// TestShowDirective_WithIfDirective tests composition with If directive
func TestShowDirective_WithIfDirective(t *testing.T) {
	t.Run("Show wrapping If", func(t *testing.T) {
		result := Show(true, func() string {
			return If(true, func() string {
				return "both true"
			}).Else(func() string {
				return "if false"
			}).Render()
		}).Render()
		assert.Equal(t, "both true", result)
	})

	t.Run("If wrapping Show", func(t *testing.T) {
		result := If(true, func() string {
			return Show(true, func() string {
				return "both true"
			}).Render()
		}).Else(func() string {
			return "if false"
		}).Render()
		assert.Equal(t, "both true", result)
	})
}

// ==================== BENCHMARKS ====================

// BenchmarkShowDirective_Visible benchmarks Show with visible=true
// Target: < 50ns
func BenchmarkShowDirective_Visible(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Show(true, func() string {
			return "content"
		}).Render()
	}
}

// BenchmarkShowDirective_Hidden benchmarks Show with visible=false
// Target: < 50ns
func BenchmarkShowDirective_Hidden(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Show(false, func() string {
			return "content"
		}).Render()
	}
}

// BenchmarkShowDirective_WithTransitionVisible benchmarks Show with transition and visible=true
// Target: < 100ns
func BenchmarkShowDirective_WithTransitionVisible(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Show(true, func() string {
			return "content"
		}).WithTransition().Render()
	}
}

// BenchmarkShowDirective_WithTransitionHidden benchmarks Show with transition and visible=false
// Target: < 100ns
func BenchmarkShowDirective_WithTransitionHidden(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Show(false, func() string {
			return "content"
		}).WithTransition().Render()
	}
}

// BenchmarkShowDirective_ComplexContent benchmarks Show with complex string content
// Target: < 100ns
func BenchmarkShowDirective_ComplexContent(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	content := "Line 1\nLine 2\nLine 3\nSpecial: !@#$%^&*()\nUnicode: 世界 🌍"
	for i := 0; i < b.N; i++ {
		_ = Show(true, func() string {
			return content
		}).Render()
	}
}

// BenchmarkShowDirective_Nested benchmarks nested Show directives
// Target: < 300ns
func BenchmarkShowDirective_Nested(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Show(true, func() string {
			return Show(true, func() string {
				return "nested"
			}).Render()
		}).Render()
	}
}
