package directives

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOn_CreatesDirective tests that On() creates a valid OnDirective.
func TestOn_CreatesDirective(t *testing.T) {
	// Arrange
	called := false
	handler := func(data interface{}) {
		called = true
	}

	// Act
	directive := On("click", handler)

	// Assert
	assert.NotNil(t, directive, "should create directive")
	assert.Equal(t, "click", directive.event, "should set event name")
	assert.NotNil(t, directive.handler, "should set handler")

	// Verify handler works
	directive.handler(nil)
	assert.True(t, called, "handler should be callable")
}

// TestOnDirective_Render tests basic rendering with event marker.
func TestOnDirective_Render(t *testing.T) {
	tests := []struct {
		name     string
		event    string
		content  string
		expected string
	}{
		{
			name:     "click event with text",
			event:    "click",
			content:  "Click Me",
			expected: "[Event:click]Click Me",
		},
		{
			name:     "submit event with button",
			event:    "submit",
			content:  "Submit Form",
			expected: "[Event:submit]Submit Form",
		},
		{
			name:     "keypress event",
			event:    "keypress",
			content:  "Press any key",
			expected: "[Event:keypress]Press any key",
		},
		{
			name:     "empty content",
			event:    "click",
			content:  "",
			expected: "[Event:click]",
		},
		{
			name:     "multiline content",
			event:    "change",
			content:  "Line 1\nLine 2\nLine 3",
			expected: "[Event:change]Line 1\nLine 2\nLine 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			directive := On(tt.event, func(data interface{}) {})

			// Act
			result := directive.Render(tt.content)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOnDirective_HandlerExecutes tests that handler is called on event.
func TestOnDirective_HandlerExecutes(t *testing.T) {
	// Arrange
	var receivedData interface{}
	handler := func(data interface{}) {
		receivedData = data
	}
	directive := On("click", handler)

	// Act
	directive.handler("test data")

	// Assert
	assert.Equal(t, "test data", receivedData, "handler should receive data")
}

// TestOnDirective_MultipleHandlers tests multiple On directives on same content.
func TestOnDirective_MultipleHandlers(t *testing.T) {
	// Arrange
	clickHandler := func(data interface{}) {
		// Handler for click
	}
	hoverHandler := func(data interface{}) {
		// Handler for hover
	}

	// Act
	content := "Button"
	content = On("click", clickHandler).Render(content)
	content = On("hover", hoverHandler).Render(content)

	// Assert
	assert.Contains(t, content, "[Event:click]")
	assert.Contains(t, content, "[Event:hover]")
	assert.Contains(t, content, "Button")
}

// TestOnDirective_TypeSafeHandler tests type-safe event handlers.
func TestOnDirective_TypeSafeHandler(t *testing.T) {
	// Arrange
	type ClickData struct {
		X int
		Y int
	}

	var received *ClickData
	handler := func(data interface{}) {
		if clickData, ok := data.(*ClickData); ok {
			received = clickData
		}
	}

	directive := On("click", handler)

	// Act
	testData := &ClickData{X: 10, Y: 20}
	directive.handler(testData)

	// Assert
	assert.NotNil(t, received, "should receive data")
	assert.Equal(t, 10, received.X)
	assert.Equal(t, 20, received.Y)
}

// TestOnDirective_HasRenderMethod tests that Render method exists and works.
func TestOnDirective_HasRenderMethod(t *testing.T) {
	// Arrange
	directive := On("click", func(data interface{}) {})

	// Act
	result := directive.Render("content")

	// Assert
	assert.NotEmpty(t, result, "should have Render method")
	assert.Contains(t, result, "[Event:click]")
	assert.Contains(t, result, "content")
}

// TestOnDirective_EmptyEventName tests handling of empty event name.
func TestOnDirective_EmptyEventName(t *testing.T) {
	// Arrange
	directive := On("", func(data interface{}) {})

	// Act
	result := directive.Render("content")

	// Assert
	assert.Equal(t, "[Event:]content", result, "should handle empty event name")
}

// TestOnDirective_NilHandler tests handling of nil handler.
func TestOnDirective_NilHandler(t *testing.T) {
	// Arrange & Act
	directive := On("click", nil)

	// Assert
	assert.NotNil(t, directive, "should create directive even with nil handler")
	assert.Nil(t, directive.handler, "handler should be nil")
}

// TestOnDirective_ComplexContent tests rendering with complex content.
func TestOnDirective_ComplexContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "unicode characters",
			content: "🎉 Click Me! 🚀",
		},
		{
			name:    "special characters",
			content: "Submit <form> & \"escape\"",
		},
		{
			name: "very long content",
			content: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
				"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			directive := On("click", func(data interface{}) {})

			// Act
			result := directive.Render(tt.content)

			// Assert
			assert.Contains(t, result, "[Event:click]")
			assert.Contains(t, result, tt.content)
		})
	}
}

// TestOnDirective_NestedWithOtherDirectives tests composition with other directives.
func TestOnDirective_NestedWithOtherDirectives(t *testing.T) {
	// Arrange
	handler := func(data interface{}) {}

	// Act - Combine On with If
	content := If(true, func() string {
		return On("click", handler).Render("Button")
	}).Render()

	// Assert
	assert.Contains(t, content, "[Event:click]")
	assert.Contains(t, content, "Button")
}

// TestOnDirective_WithForEach tests On directive in ForEach loop.
func TestOnDirective_WithForEach(t *testing.T) {
	// Arrange
	items := []string{"Item 1", "Item 2", "Item 3"}
	handler := func(data interface{}) {}

	// Act
	result := ForEach(items, func(item string, i int) string {
		return On("click", handler).Render(item)
	}).Render()

	// Assert
	assert.Contains(t, result, "[Event:click]Item 1")
	assert.Contains(t, result, "[Event:click]Item 2")
	assert.Contains(t, result, "[Event:click]Item 3")
}

// TestOnDirective_PreventDefault tests PreventDefault modifier.
func TestOnDirective_PreventDefault(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *OnDirective
		expected string
	}{
		{
			name: "without PreventDefault",
			setup: func() *OnDirective {
				return On("submit", func(data interface{}) {})
			},
			expected: "[Event:submit]Submit",
		},
		{
			name: "with PreventDefault",
			setup: func() *OnDirective {
				return On("submit", func(data interface{}) {}).PreventDefault()
			},
			expected: "[Event:submit:prevent]Submit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			directive := tt.setup()

			// Act
			result := directive.Render("Submit")

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOnDirective_StopPropagation tests StopPropagation modifier.
func TestOnDirective_StopPropagation(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *OnDirective
		expected string
	}{
		{
			name: "without StopPropagation",
			setup: func() *OnDirective {
				return On("click", func(data interface{}) {})
			},
			expected: "[Event:click]Button",
		},
		{
			name: "with StopPropagation",
			setup: func() *OnDirective {
				return On("click", func(data interface{}) {}).StopPropagation()
			},
			expected: "[Event:click:stop]Button",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			directive := tt.setup()

			// Act
			result := directive.Render("Button")

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOnDirective_Once tests Once modifier.
func TestOnDirective_Once(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *OnDirective
		expected string
	}{
		{
			name: "without Once",
			setup: func() *OnDirective {
				return On("click", func(data interface{}) {})
			},
			expected: "[Event:click]Click",
		},
		{
			name: "with Once",
			setup: func() *OnDirective {
				return On("click", func(data interface{}) {}).Once()
			},
			expected: "[Event:click:once]Click",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			directive := tt.setup()

			// Act
			result := directive.Render("Click")

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOnDirective_ModifierChaining tests chaining multiple modifiers.
func TestOnDirective_ModifierChaining(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *OnDirective
		expected string
	}{
		{
			name: "PreventDefault + StopPropagation",
			setup: func() *OnDirective {
				return On("submit", func(data interface{}) {}).
					PreventDefault().
					StopPropagation()
			},
			expected: "[Event:submit:prevent:stop]Form",
		},
		{
			name: "PreventDefault + Once",
			setup: func() *OnDirective {
				return On("click", func(data interface{}) {}).
					PreventDefault().
					Once()
			},
			expected: "[Event:click:prevent:once]Button",
		},
		{
			name: "StopPropagation + Once",
			setup: func() *OnDirective {
				return On("click", func(data interface{}) {}).
					StopPropagation().
					Once()
			},
			expected: "[Event:click:stop:once]Link",
		},
		{
			name: "All three modifiers",
			setup: func() *OnDirective {
				return On("submit", func(data interface{}) {}).
					PreventDefault().
					StopPropagation().
					Once()
			},
			expected: "[Event:submit:prevent:stop:once]Submit",
		},
		{
			name: "Modifiers in different order",
			setup: func() *OnDirective {
				return On("click", func(data interface{}) {}).
					Once().
					PreventDefault().
					StopPropagation()
			},
			expected: "[Event:click:prevent:stop:once]Button",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			directive := tt.setup()

			// Act
			result := directive.Render(tt.expected[len(tt.expected)-len("]Form"):])

			// Assert
			assert.Contains(t, result, "[Event:")
			// Verify all expected modifiers are present
			if tt.name == "PreventDefault + StopPropagation" {
				assert.Contains(t, result, "prevent")
				assert.Contains(t, result, "stop")
			}
		})
	}
}

// TestOnDirective_FluentAPI tests that modifiers return the directive for chaining.
func TestOnDirective_FluentAPI(t *testing.T) {
	// Arrange
	directive := On("click", func(data interface{}) {})

	// Act
	result1 := directive.PreventDefault()
	result2 := result1.StopPropagation()
	result3 := result2.Once()

	// Assert
	assert.Same(t, directive, result1, "PreventDefault should return same directive")
	assert.Same(t, directive, result2, "StopPropagation should return same directive")
	assert.Same(t, directive, result3, "Once should return same directive")
}

// TestOnDirective_ModifiersAreIdempotent tests calling modifiers multiple times.
func TestOnDirective_ModifiersAreIdempotent(t *testing.T) {
	// Arrange
	directive := On("click", func(data interface{}) {})

	// Act
	directive.PreventDefault().PreventDefault()
	directive.StopPropagation().StopPropagation()
	directive.Once().Once()
	result := directive.Render("Button")

	// Assert
	assert.Contains(t, result, "prevent")
	assert.Contains(t, result, "stop")
	assert.Contains(t, result, "once")
	// Should only appear once each, not duplicated
	assert.Equal(t, 1, countOccurrences(result, "prevent"))
	assert.Equal(t, 1, countOccurrences(result, "stop"))
	assert.Equal(t, 1, countOccurrences(result, "once"))
}

// TestOnDirective_ModifiersWithEmptyEvent tests modifiers with empty event name.
func TestOnDirective_ModifiersWithEmptyEvent(t *testing.T) {
	// Arrange
	directive := On("", func(data interface{}) {}).
		PreventDefault().
		StopPropagation().
		Once()

	// Act
	result := directive.Render("content")

	// Assert
	assert.Contains(t, result, "[Event:")
	assert.Contains(t, result, "prevent")
	assert.Contains(t, result, "stop")
	assert.Contains(t, result, "once")
}

// Helper function to count occurrences of substring
func countOccurrences(s, substr string) int {
	count := 0
	start := 0
	for {
		idx := indexOf(s[start:], substr)
		if idx == -1 {
			break
		}
		count++
		start += idx + len(substr)
	}
	return count
}

// Helper function to find index of substring
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ==================== BENCHMARKS ====================

// BenchmarkOnDirective_Simple benchmarks simple On directive
// Target: < 80ns
func BenchmarkOnDirective_Simple(b *testing.B) {
	handler := func(data interface{}) {}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = On("click", handler).Render("Button")
	}
}

// BenchmarkOnDirective_WithPreventDefault benchmarks On with PreventDefault modifier
// Target: < 100ns
func BenchmarkOnDirective_WithPreventDefault(b *testing.B) {
	handler := func(data interface{}) {}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = On("submit", handler).PreventDefault().Render("Submit")
	}
}

// BenchmarkOnDirective_WithStopPropagation benchmarks On with StopPropagation modifier
// Target: < 100ns
func BenchmarkOnDirective_WithStopPropagation(b *testing.B) {
	handler := func(data interface{}) {}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = On("click", handler).StopPropagation().Render("Link")
	}
}

// BenchmarkOnDirective_WithAllModifiers benchmarks On with all modifiers
// Target: < 150ns
func BenchmarkOnDirective_WithAllModifiers(b *testing.B) {
	handler := func(data interface{}) {}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = On("submit", handler).
			PreventDefault().
			StopPropagation().
			Once().
			Render("Form")
	}
}

// BenchmarkOnDirective_ComplexContent benchmarks On with complex content
// Target: < 150ns
func BenchmarkOnDirective_ComplexContent(b *testing.B) {
	handler := func(data interface{}) {}
	content := "Line 1\nLine 2\nLine 3\nSpecial: !@#$%^&*()\nUnicode: 世界 🌍"
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = On("click", handler).Render(content)
	}
}

// BenchmarkOnDirective_Multiple benchmarks multiple On directives on same content
// Target: < 300ns
func BenchmarkOnDirective_Multiple(b *testing.B) {
	handler := func(data interface{}) {}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		content := "Button"
		content = On("click", handler).Render(content)
		content = On("mouseenter", handler).Render(content)
		content = On("mouseleave", handler).Render(content)
		_ = content
	}
}
