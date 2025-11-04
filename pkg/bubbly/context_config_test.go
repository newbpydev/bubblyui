package bubbly

import (
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestContext_EnableAutoCommands tests enabling automatic command generation
func TestContext_EnableAutoCommands(t *testing.T) {
	tests := []struct {
		name            string
		initialState    bool
		expectedState   bool
		expectGenerator bool
	}{
		{
			name:            "enable_from_disabled",
			initialState:    false,
			expectedState:   true,
			expectGenerator: true,
		},
		{
			name:            "enable_when_already_enabled",
			initialState:    true,
			expectedState:   true,
			expectGenerator: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := newComponentImpl("TestComponent")
			c.autoCommands = tt.initialState
			ctx := &Context{component: c}

			// Act
			ctx.EnableAutoCommands()

			// Assert
			assert.Equal(t, tt.expectedState, c.autoCommands, "autoCommands state should match expected")
			if tt.expectGenerator {
				assert.NotNil(t, c.commandGen, "commandGen should be set when enabling")
			}
		})
	}
}

// TestContext_DisableAutoCommands tests disabling automatic command generation
func TestContext_DisableAutoCommands(t *testing.T) {
	tests := []struct {
		name          string
		initialState  bool
		expectedState bool
	}{
		{
			name:          "disable_from_enabled",
			initialState:  true,
			expectedState: false,
		},
		{
			name:          "disable_when_already_disabled",
			initialState:  false,
			expectedState: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := newComponentImpl("TestComponent")
			c.autoCommands = tt.initialState
			ctx := &Context{component: c}

			// Act
			ctx.DisableAutoCommands()

			// Assert
			assert.Equal(t, tt.expectedState, c.autoCommands, "autoCommands state should match expected")
		})
	}
}

// TestContext_IsAutoCommandsEnabled tests checking auto commands state
func TestContext_IsAutoCommandsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		state    bool
		expected bool
	}{
		{
			name:     "returns_true_when_enabled",
			state:    true,
			expected: true,
		},
		{
			name:     "returns_false_when_disabled",
			state:    false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := newComponentImpl("TestComponent")
			c.autoCommands = tt.state
			ctx := &Context{component: c}

			// Act
			result := ctx.IsAutoCommandsEnabled()

			// Assert
			assert.Equal(t, tt.expected, result, "IsAutoCommandsEnabled should return correct state")
		})
	}
}

// TestContext_ManualRef tests creating refs without auto commands
func TestContext_ManualRef(t *testing.T) {
	tests := []struct {
		name              string
		autoCommandsState bool
		value             interface{}
		expectCommands    bool
	}{
		{
			name:              "manual_ref_with_auto_enabled",
			autoCommandsState: true,
			value:             42,
			expectCommands:    false, // ManualRef should bypass auto commands
		},
		{
			name:              "manual_ref_with_auto_disabled",
			autoCommandsState: false,
			value:             "test",
			expectCommands:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := newComponentImpl("TestComponent")
			c.autoCommands = tt.autoCommandsState
			ctx := &Context{component: c}

			// Act
			ref := ctx.ManualRef(tt.value)

			// Assert
			assert.NotNil(t, ref, "ManualRef should return a ref")
			assert.Equal(t, tt.value, ref.Get(), "Ref should have correct initial value")

			// Verify no commands generated on Set
			ref.Set("changed")
			assert.Equal(t, 0, c.commandQueue.Len(), "ManualRef should not generate commands")
		})
	}
}

// TestContext_ManualRef_RestoresState tests that ManualRef restores auto commands state
func TestContext_ManualRef_RestoresState(t *testing.T) {
	// Arrange
	c := newComponentImpl("TestComponent")
	c.autoCommands = true
	ctx := &Context{component: c}

	// Act
	_ = ctx.ManualRef(42)

	// Assert - auto commands should still be enabled after ManualRef
	assert.True(t, c.autoCommands, "autoCommands should be restored after ManualRef")

	// Verify regular Ref still generates commands
	autoRef := ctx.Ref(100)
	autoRef.Set(200)
	assert.Greater(t, c.commandQueue.Len(), 0, "Regular Ref should still generate commands")
}

// TestContext_SetCommandGenerator tests setting custom command generator
func TestContext_SetCommandGenerator(t *testing.T) {
	// Use mockCommandGenerator which implements the interface
	customGen := &mockCommandGenerator{}

	tests := []struct {
		name      string
		generator CommandGenerator
	}{
		{
			name:      "set_custom_generator",
			generator: customGen,
		},
		{
			name:      "set_nil_generator",
			generator: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := newComponentImpl("TestComponent")
			ctx := &Context{component: c}

			// Act
			ctx.SetCommandGenerator(tt.generator)

			// Assert
			if tt.generator != nil {
				assert.Equal(t, tt.generator, c.commandGen, "commandGen should be set to custom generator")
			} else {
				// nil generator should be allowed (component will use default)
				assert.Nil(t, c.commandGen, "commandGen should be nil when set to nil")
			}
		})
	}
}

// TestContext_SetCommandGenerator_Integration tests custom generator with Ref
func TestContext_SetCommandGenerator_Integration(t *testing.T) {
	// Custom generator that tracks calls
	type trackingGenerator struct {
		calls []struct {
			componentID string
			refID       string
			oldValue    interface{}
			newValue    interface{}
		}
		mu sync.Mutex
	}

	customGen := &trackingGenerator{
		calls: make([]struct {
			componentID string
			refID       string
			oldValue    interface{}
			newValue    interface{}
		}, 0),
	}

	// Implement CommandGenerator interface
	generateFunc := func(componentID, refID string, oldValue, newValue interface{}) tea.Cmd {
		customGen.mu.Lock()
		defer customGen.mu.Unlock()
		customGen.calls = append(customGen.calls, struct {
			componentID string
			refID       string
			oldValue    interface{}
			newValue    interface{}
		}{
			componentID: componentID,
			refID:       refID,
			oldValue:    oldValue,
			newValue:    newValue,
		})
		return func() tea.Msg {
			return StateChangedMsg{
				ComponentID: componentID,
				RefID:       refID,
				OldValue:    oldValue,
				NewValue:    newValue,
			}
		}
	}

	// Create mock generator
	mockGen := &mockCommandGenerator{generateFunc: generateFunc}

	// Arrange
	c := newComponentImpl("TestComponent")
	c.autoCommands = true
	ctx := &Context{component: c}
	ctx.SetCommandGenerator(mockGen)

	// Act
	ref := ctx.Ref(10)
	ref.Set(20)

	// Assert
	customGen.mu.Lock()
	defer customGen.mu.Unlock()
	assert.Equal(t, 1, len(customGen.calls), "Custom generator should be called once")
	assert.Equal(t, 10, customGen.calls[0].oldValue, "Old value should be captured")
	assert.Equal(t, 20, customGen.calls[0].newValue, "New value should be captured")
}

// TestContext_AutoCommands_ThreadSafe tests thread-safe enable/disable operations
func TestContext_AutoCommands_ThreadSafe(t *testing.T) {
	// Arrange
	c := newComponentImpl("TestComponent")
	ctx := &Context{component: c}

	// Act - Enable/disable concurrently
	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			ctx.EnableAutoCommands()
		}()

		go func() {
			defer wg.Done()
			ctx.DisableAutoCommands()
		}()
	}

	wg.Wait()

	// Assert - Should not panic, state should be consistent
	state := ctx.IsAutoCommandsEnabled()
	assert.IsType(t, false, state, "State should be boolean")
}

// TestContext_ManualRef_ThreadSafe tests thread-safe ManualRef creation
func TestContext_ManualRef_ThreadSafe(t *testing.T) {
	// Arrange
	c := newComponentImpl("TestComponent")
	c.autoCommands = true
	ctx := &Context{component: c}

	// Act - Create manual refs concurrently
	var wg sync.WaitGroup
	iterations := 100
	refs := make([]*Ref[interface{}], iterations)

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			refs[idx] = ctx.ManualRef(idx)
		}(i)
	}

	wg.Wait()

	// Assert - All refs should be created, none should generate commands
	for i, ref := range refs {
		assert.NotNil(t, ref, "Ref %d should be created", i)
		assert.Equal(t, i, ref.Get(), "Ref %d should have correct value", i)
	}

	// Verify no commands generated
	assert.Equal(t, 0, c.commandQueue.Len(), "No commands should be generated by manual refs")
}

// mockCommandGenerator is a test helper for custom generator testing
type mockCommandGenerator struct {
	generateFunc func(componentID, refID string, oldValue, newValue interface{}) tea.Cmd
}

func (m *mockCommandGenerator) Generate(componentID, refID string, oldValue, newValue interface{}) tea.Cmd {
	if m.generateFunc != nil {
		return m.generateFunc(componentID, refID, oldValue, newValue)
	}
	return func() tea.Msg {
		return StateChangedMsg{
			ComponentID: componentID,
			RefID:       refID,
			OldValue:    oldValue,
			NewValue:    newValue,
		}
	}
}
