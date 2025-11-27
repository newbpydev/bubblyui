package testutil

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestNewMockComponent tests mock component creation with default values
func TestNewMockComponent(t *testing.T) {
	tests := []struct {
		name           string
		componentName  string
		expectedID     string
		expectedOutput string
	}{
		{
			name:           "simple name",
			componentName:  "Button",
			expectedID:     "mock-Button",
			expectedOutput: "Mock<Button>",
		},
		{
			name:           "multi-word name",
			componentName:  "TodoList",
			expectedID:     "mock-TodoList",
			expectedOutput: "Mock<TodoList>",
		},
		{
			name:           "empty name",
			componentName:  "",
			expectedID:     "mock-",
			expectedOutput: "Mock<>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockComponent(tt.componentName)

			assert.Equal(t, tt.componentName, mock.Name())
			assert.Equal(t, tt.expectedID, mock.ID())
			assert.Equal(t, tt.expectedOutput, mock.View())
			assert.False(t, mock.IsInitialized())
			assert.Equal(t, 0, mock.GetUpdateCallCount())
			assert.Equal(t, 1, mock.GetViewCallCount()) // View() called once above
		})
	}
}

// TestMockComponent_ComponentInterface verifies mock implements Component interface
func TestMockComponent_ComponentInterface(t *testing.T) {
	var _ bubbly.Component = (*MockComponent)(nil)

	mock := NewMockComponent("Test")

	// Verify all Component methods are callable
	assert.Equal(t, "Test", mock.Name())
	assert.Equal(t, "mock-Test", mock.ID())
	assert.Nil(t, mock.Props())
	assert.NotNil(t, mock.KeyBindings())
	assert.Equal(t, "", mock.HelpText())
	assert.False(t, mock.IsInitialized())

	mock.Emit("test", nil)
	mock.On("test", func(data interface{}) {})
}

// TestMockComponent_Init tests Init() method and tracking
func TestMockComponent_Init(t *testing.T) {
	mock := NewMockComponent("Test")

	// Initially not initialized
	assert.False(t, mock.IsInitialized())

	// Call Init
	cmd := mock.Init()

	// Should be initialized now
	assert.True(t, mock.IsInitialized())
	assert.Nil(t, cmd)

	// Multiple calls should still show initialized
	mock.Init()
	assert.True(t, mock.IsInitialized())
}

// TestMockComponent_Update tests Update() method and call tracking
func TestMockComponent_Update(t *testing.T) {
	tests := []struct {
		name          string
		updateCount   int
		expectedCalls int
	}{
		{
			name:          "no updates",
			updateCount:   0,
			expectedCalls: 0,
		},
		{
			name:          "single update",
			updateCount:   1,
			expectedCalls: 1,
		},
		{
			name:          "multiple updates",
			updateCount:   5,
			expectedCalls: 5,
		},
		{
			name:          "many updates",
			updateCount:   100,
			expectedCalls: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockComponent("Test")

			// Call Update multiple times
			for i := 0; i < tt.updateCount; i++ {
				model, cmd := mock.Update(tea.KeyMsg{})
				assert.Equal(t, mock, model)
				assert.Nil(t, cmd)
			}

			// Verify call count
			assert.Equal(t, tt.expectedCalls, mock.GetUpdateCallCount())
		})
	}
}

// TestMockComponent_View tests View() method and call tracking
func TestMockComponent_View(t *testing.T) {
	tests := []struct {
		name          string
		viewOutput    string
		viewCount     int
		expectedCalls int
	}{
		{
			name:          "default output",
			viewOutput:    "Mock<Test>",
			viewCount:     1,
			expectedCalls: 1,
		},
		{
			name:          "custom output",
			viewOutput:    "Custom View",
			viewCount:     1,
			expectedCalls: 1,
		},
		{
			name:          "multiple views",
			viewOutput:    "Test Output",
			viewCount:     3,
			expectedCalls: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockComponent("Test")

			// Set custom output if specified
			if tt.viewOutput != "Mock<Test>" {
				mock.SetViewOutput(tt.viewOutput)
			}

			// Call View multiple times
			for i := 0; i < tt.viewCount; i++ {
				output := mock.View()
				assert.Equal(t, tt.viewOutput, output)
			}

			// Verify call count
			assert.Equal(t, tt.expectedCalls, mock.GetViewCallCount())
		})
	}
}

// TestMockComponent_Emit tests Emit() method and call tracking
func TestMockComponent_Emit(t *testing.T) {
	tests := []struct {
		name          string
		events        []string
		expectedCalls map[string]int
	}{
		{
			name:   "single event",
			events: []string{"click"},
			expectedCalls: map[string]int{
				"click": 1,
			},
		},
		{
			name:   "multiple same events",
			events: []string{"click", "click", "click"},
			expectedCalls: map[string]int{
				"click": 3,
			},
		},
		{
			name:   "different events",
			events: []string{"click", "submit", "cancel"},
			expectedCalls: map[string]int{
				"click":  1,
				"submit": 1,
				"cancel": 1,
			},
		},
		{
			name:   "mixed events",
			events: []string{"click", "click", "submit", "click"},
			expectedCalls: map[string]int{
				"click":  3,
				"submit": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockComponent("Test")

			// Emit events
			for _, event := range tt.events {
				mock.Emit(event, nil)
			}

			// Verify call counts
			for event, expectedCount := range tt.expectedCalls {
				assert.Equal(t, expectedCount, mock.GetEmitCallCount(event))
			}
		})
	}
}

// TestMockComponent_On tests On() method and call tracking
func TestMockComponent_On(t *testing.T) {
	tests := []struct {
		name          string
		registrations []string
		expectedCalls map[string]int
	}{
		{
			name:          "single handler",
			registrations: []string{"click"},
			expectedCalls: map[string]int{
				"click": 1,
			},
		},
		{
			name:          "multiple handlers same event",
			registrations: []string{"click", "click", "click"},
			expectedCalls: map[string]int{
				"click": 3,
			},
		},
		{
			name:          "different events",
			registrations: []string{"click", "submit", "cancel"},
			expectedCalls: map[string]int{
				"click":  1,
				"submit": 1,
				"cancel": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockComponent("Test")

			// Register handlers
			for _, event := range tt.registrations {
				mock.On(event, func(data interface{}) {})
			}

			// Verify call counts
			for event, expectedCount := range tt.expectedCalls {
				assert.Equal(t, expectedCount, mock.GetOnCallCount(event))
			}
		})
	}
}

// TestMockComponent_EventHandlers tests that registered handlers are called
func TestMockComponent_EventHandlers(t *testing.T) {
	mock := NewMockComponent("Test")

	// Track handler calls
	clickCount := 0
	submitCount := 0

	// Register handlers
	mock.On("click", func(data interface{}) {
		clickCount++
	})
	mock.On("click", func(data interface{}) {
		clickCount++
	})
	mock.On("submit", func(data interface{}) {
		submitCount++
	})

	// Emit events
	mock.Emit("click", nil)
	mock.Emit("submit", nil)
	mock.Emit("click", nil)

	// Verify handlers were called
	assert.Equal(t, 4, clickCount)  // 2 handlers × 2 emits
	assert.Equal(t, 1, submitCount) // 1 handler × 1 emit
}

// TestMockComponent_Props tests Props() method and configuration
func TestMockComponent_Props(t *testing.T) {
	type ButtonProps struct {
		Label string
		Value int
	}

	tests := []struct {
		name          string
		props         interface{}
		expectedProps interface{}
	}{
		{
			name:          "nil props",
			props:         nil,
			expectedProps: nil,
		},
		{
			name:          "string props",
			props:         "test",
			expectedProps: "test",
		},
		{
			name:          "struct props",
			props:         ButtonProps{Label: "Click", Value: 42},
			expectedProps: ButtonProps{Label: "Click", Value: 42},
		},
		{
			name:          "map props",
			props:         map[string]interface{}{"key": "value"},
			expectedProps: map[string]interface{}{"key": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockComponent("Test")

			// Set props
			mock.SetProps(tt.props)

			// Verify props
			assert.Equal(t, tt.expectedProps, mock.Props())
		})
	}
}

// TestMockComponent_KeyBindings tests KeyBindings() method and configuration
func TestMockComponent_KeyBindings(t *testing.T) {
	mock := NewMockComponent("Test")

	// Initially empty
	bindings := mock.KeyBindings()
	assert.NotNil(t, bindings)
	assert.Empty(t, bindings)

	// Set custom bindings
	customBindings := map[string][]bubbly.KeyBinding{
		" ": {{Event: "increment", Description: "Increment counter"}},
		"r": {{Event: "reset", Description: "Reset counter"}},
	}
	mock.SetKeyBindings(customBindings)

	// Verify bindings
	bindings = mock.KeyBindings()
	assert.Equal(t, customBindings, bindings)
}

// TestMockComponent_HelpText tests HelpText() method and configuration
func TestMockComponent_HelpText(t *testing.T) {
	tests := []struct {
		name         string
		helpText     string
		expectedText string
	}{
		{
			name:         "empty help text",
			helpText:     "",
			expectedText: "",
		},
		{
			name:         "simple help text",
			helpText:     "space: increment",
			expectedText: "space: increment",
		},
		{
			name:         "multi-key help text",
			helpText:     "space: increment • r: reset • q: quit",
			expectedText: "space: increment • r: reset • q: quit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockComponent("Test")

			// Set help text
			mock.SetHelpText(tt.helpText)

			// Verify help text
			assert.Equal(t, tt.expectedText, mock.HelpText())
		})
	}
}

// TestMockComponent_Reset tests Reset() method
func TestMockComponent_Reset(t *testing.T) {
	mock := NewMockComponent("Test")

	// Perform various operations
	mock.Init()
	mock.Update(tea.KeyMsg{})
	mock.Update(tea.KeyMsg{})
	mock.View()
	mock.Emit("click", nil)
	mock.On("click", func(data interface{}) {})

	// Verify state before reset
	assert.True(t, mock.IsInitialized())
	assert.Equal(t, 2, mock.GetUpdateCallCount())
	assert.Equal(t, 1, mock.GetViewCallCount())
	assert.Equal(t, 1, mock.GetEmitCallCount("click"))
	assert.Equal(t, 1, mock.GetOnCallCount("click"))

	// Reset
	mock.Reset()

	// Verify state after reset
	assert.False(t, mock.IsInitialized())
	assert.Equal(t, 0, mock.GetUpdateCallCount())
	assert.Equal(t, 0, mock.GetViewCallCount())
	assert.Equal(t, 0, mock.GetEmitCallCount("click"))
	assert.Equal(t, 0, mock.GetOnCallCount("click"))

	// Name, ID, and configured values should remain
	assert.Equal(t, "Test", mock.Name())
	assert.Equal(t, "mock-Test", mock.ID())
}

// TestMockComponent_AssertInitCalled tests AssertInitCalled assertion
func TestMockComponent_AssertInitCalled(t *testing.T) {
	t.Run("init called - passes", func(t *testing.T) {
		mock := NewMockComponent("Test")
		mock.Init()

		// Should not fail
		mock.AssertInitCalled(t)
	})

	t.Run("init not called - fails", func(t *testing.T) {
		mock := NewMockComponent("Test")

		// Create mock testing.T to capture error
		mockT := &mockTestingT{}
		mock.AssertInitCalled(mockT)

		// Should have failed
		assert.True(t, mockT.failed)
		assert.Len(t, mockT.errors, 1)
		assert.Contains(t, mockT.errors[0], "Init() was not called")
	})
}

// TestMockComponent_AssertInitNotCalled tests AssertInitNotCalled assertion
func TestMockComponent_AssertInitNotCalled(t *testing.T) {
	t.Run("init not called - passes", func(t *testing.T) {
		mock := NewMockComponent("Test")

		// Should not fail
		mock.AssertInitNotCalled(t)
	})

	t.Run("init called - fails", func(t *testing.T) {
		mock := NewMockComponent("Test")
		mock.Init()

		// Create mock testing.T to capture error
		mockT := &mockTestingT{}
		mock.AssertInitNotCalled(mockT)

		// Should have failed
		assert.True(t, mockT.failed)
		assert.Len(t, mockT.errors, 1)
		assert.Contains(t, mockT.errors[0], "Init() was called but should not have been")
	})
}

// TestMockComponent_AssertUpdateCalled tests AssertUpdateCalled assertion
func TestMockComponent_AssertUpdateCalled(t *testing.T) {
	tests := []struct {
		name          string
		updateCount   int
		expectedCount int
		shouldFail    bool
	}{
		{
			name:          "correct count",
			updateCount:   3,
			expectedCount: 3,
			shouldFail:    false,
		},
		{
			name:          "zero count",
			updateCount:   0,
			expectedCount: 0,
			shouldFail:    false,
		},
		{
			name:          "incorrect count - too few",
			updateCount:   1,
			expectedCount: 3,
			shouldFail:    true,
		},
		{
			name:          "incorrect count - too many",
			updateCount:   5,
			expectedCount: 2,
			shouldFail:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockComponent("Test")

			// Call Update
			for i := 0; i < tt.updateCount; i++ {
				mock.Update(tea.KeyMsg{})
			}

			// Create mock testing.T to capture error
			mockT := &mockTestingT{}
			mock.AssertUpdateCalled(mockT, tt.expectedCount)

			// Verify result
			assert.Equal(t, tt.shouldFail, mockT.failed)
			if tt.shouldFail {
				assert.Len(t, mockT.errors, 1)
				assert.Contains(t, mockT.errors[0], "Update()")
			}
		})
	}
}

// TestMockComponent_AssertViewCalled tests AssertViewCalled assertion
func TestMockComponent_AssertViewCalled(t *testing.T) {
	tests := []struct {
		name          string
		viewCount     int
		expectedCount int
		shouldFail    bool
	}{
		{
			name:          "correct count",
			viewCount:     2,
			expectedCount: 2,
			shouldFail:    false,
		},
		{
			name:          "zero count",
			viewCount:     0,
			expectedCount: 0,
			shouldFail:    false,
		},
		{
			name:          "incorrect count",
			viewCount:     1,
			expectedCount: 3,
			shouldFail:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockComponent("Test")

			// Call View
			for i := 0; i < tt.viewCount; i++ {
				mock.View()
			}

			// Create mock testing.T to capture error
			mockT := &mockTestingT{}
			mock.AssertViewCalled(mockT, tt.expectedCount)

			// Verify result
			assert.Equal(t, tt.shouldFail, mockT.failed)
			if tt.shouldFail {
				assert.Len(t, mockT.errors, 1)
				assert.Contains(t, mockT.errors[0], "View()")
			}
		})
	}
}

// TestMockComponent_AssertEmitCalled tests AssertEmitCalled assertion
func TestMockComponent_AssertEmitCalled(t *testing.T) {
	tests := []struct {
		name          string
		event         string
		emitCount     int
		expectedCount int
		shouldFail    bool
	}{
		{
			name:          "correct count",
			event:         "click",
			emitCount:     2,
			expectedCount: 2,
			shouldFail:    false,
		},
		{
			name:          "zero count",
			event:         "click",
			emitCount:     0,
			expectedCount: 0,
			shouldFail:    false,
		},
		{
			name:          "incorrect count",
			event:         "click",
			emitCount:     1,
			expectedCount: 3,
			shouldFail:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockComponent("Test")

			// Emit events
			for i := 0; i < tt.emitCount; i++ {
				mock.Emit(tt.event, nil)
			}

			// Create mock testing.T to capture error
			mockT := &mockTestingT{}
			mock.AssertEmitCalled(mockT, tt.event, tt.expectedCount)

			// Verify result
			assert.Equal(t, tt.shouldFail, mockT.failed)
			if tt.shouldFail {
				assert.Len(t, mockT.errors, 1)
				assert.Contains(t, mockT.errors[0], "Emit")
			}
		})
	}
}

// TestMockComponent_AssertOnCalled tests AssertOnCalled assertion
func TestMockComponent_AssertOnCalled(t *testing.T) {
	tests := []struct {
		name          string
		event         string
		onCount       int
		expectedCount int
		shouldFail    bool
	}{
		{
			name:          "correct count",
			event:         "click",
			onCount:       2,
			expectedCount: 2,
			shouldFail:    false,
		},
		{
			name:          "zero count",
			event:         "click",
			onCount:       0,
			expectedCount: 0,
			shouldFail:    false,
		},
		{
			name:          "incorrect count",
			event:         "click",
			onCount:       1,
			expectedCount: 3,
			shouldFail:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockComponent("Test")

			// Register handlers
			for i := 0; i < tt.onCount; i++ {
				mock.On(tt.event, func(data interface{}) {})
			}

			// Create mock testing.T to capture error
			mockT := &mockTestingT{}
			mock.AssertOnCalled(mockT, tt.event, tt.expectedCount)

			// Verify result
			assert.Equal(t, tt.shouldFail, mockT.failed)
			if tt.shouldFail {
				assert.Len(t, mockT.errors, 1)
				assert.Contains(t, mockT.errors[0], "On")
			}
		})
	}
}

// TestMockComponent_ThreadSafety tests concurrent access to mock component
func TestMockComponent_ThreadSafety(t *testing.T) {
	mock := NewMockComponent("Test")

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			mock.Init()
			mock.Update(tea.KeyMsg{})
			mock.View()
			mock.Emit("test", nil)
			mock.On("test", func(data interface{}) {})
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify state is consistent (no panics)
	assert.True(t, mock.IsInitialized())
	assert.Equal(t, 10, mock.GetUpdateCallCount())
	assert.Equal(t, 10, mock.GetViewCallCount())
	assert.Equal(t, 10, mock.GetEmitCallCount("test"))
	assert.Equal(t, 10, mock.GetOnCallCount("test"))
}

// Note: mockTestingT is defined in assertions_state_test.go and reused here
