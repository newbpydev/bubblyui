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
