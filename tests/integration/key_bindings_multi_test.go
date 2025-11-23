package integration

import (
	"fmt"
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestMultiKeyBinding_AllKeysTriggerEvent tests that all bound keys trigger the same event.
// This verifies the core functionality of WithMultiKeyBindings - multiple keys → same event.
func TestMultiKeyBinding_AllKeysTriggerEvent(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		keyMsg   tea.KeyMsg
		expected int
	}{
		{
			name:     "up arrow increments",
			key:      "up",
			keyMsg:   tea.KeyMsg{Type: tea.KeyUp},
			expected: 1,
		},
		{
			name:     "k key increments",
			key:      "k",
			keyMsg:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			expected: 1,
		},
		{
			name:     "plus key increments",
			key:      "+",
			keyMsg:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component with multi-key binding
			component, err := bubbly.NewComponent("Counter").
				WithMultiKeyBindings("increment", "Increment counter", "up", "k", "+").
				Setup(func(ctx *bubbly.Context) {
					count := ctx.Ref(0)
					ctx.Expose("count", count)

					ctx.On("increment", func(_ interface{}) {
						c := count.Get().(int)
						count.Set(c + 1)
					})
				}).
				Template(func(ctx bubbly.RenderContext) string {
					count := ctx.Get("count").(*bubbly.Ref[interface{}])
					return fmt.Sprintf("Count: %d", count.Get().(int))
				}).
				Build()

			require.NoError(t, err)
			require.NotNil(t, component)

			// Init component
			component.Init()

			// Verify initial state
			assert.Equal(t, "Count: 0", component.View())

			// Send key message
			model, _ := component.Update(tt.keyMsg)
			component = model.(bubbly.Component)

			// Verify count incremented via View()
			assert.Equal(t, fmt.Sprintf("Count: %d", tt.expected), component.View(), "Key %s should increment counter", tt.key)
		})
	}
}

// TestMultiKeyBinding_MultipleEvents tests multiple events each with multiple keys.
// This verifies that WithMultiKeyBindings works correctly for different events.
func TestMultiKeyBinding_MultipleEvents(t *testing.T) {
	// Create component with two multi-key bindings
	component, err := bubbly.NewComponent("Counter").
		WithMultiKeyBindings("increment", "Increment counter", "up", "k", "+").
		WithMultiKeyBindings("decrement", "Decrement counter", "down", "j", "-").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			ctx.On("increment", func(_ interface{}) {
				c := count.Get().(int)
				count.Set(c + 1)
			})

			ctx.On("decrement", func(_ interface{}) {
				c := count.Get().(int)
				count.Set(c - 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// Verify initial state
	assert.Equal(t, "Count: 0", component.View())

	// Test increment keys
	incrementKeys := []tea.KeyMsg{
		{Type: tea.KeyUp},
		{Type: tea.KeyRunes, Runes: []rune{'k'}},
		{Type: tea.KeyRunes, Runes: []rune{'+'}},
	}

	for _, keyMsg := range incrementKeys {
		model, _ := component.Update(keyMsg)
		component = model.(bubbly.Component)
	}

	// Verify count is 3 via View()
	assert.Equal(t, "Count: 3", component.View(), "All increment keys should work")

	// Test decrement keys
	decrementKeys := []tea.KeyMsg{
		{Type: tea.KeyDown},
		{Type: tea.KeyRunes, Runes: []rune{'j'}},
		{Type: tea.KeyRunes, Runes: []rune{'-'}},
	}

	for _, keyMsg := range decrementKeys {
		model, _ := component.Update(keyMsg)
		component = model.(bubbly.Component)
	}

	// Verify count is 0 via View()
	assert.Equal(t, "Count: 0", component.View(), "All decrement keys should work")
}

// TestMultiKeyBinding_MixedWithSingleBinding tests mixing WithKeyBinding and WithMultiKeyBindings.
// This verifies backward compatibility and that both methods work together.
func TestMultiKeyBinding_MixedWithSingleBinding(t *testing.T) {
	var mu sync.Mutex
	var incrementCount, resetCount int

	component, err := bubbly.NewComponent("Counter").
		WithKeyBinding("r", "reset", "Reset counter").                          // Single binding
		WithMultiKeyBindings("increment", "Increment counter", "up", "k", "+"). // Multi binding
		WithKeyBinding("esc", "cancel", "Cancel operation").                    // Another single
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			ctx.On("increment", func(_ interface{}) {
				mu.Lock()
				incrementCount++
				mu.Unlock()
				c := count.Get().(int)
				count.Set(c + 1)
			})

			ctx.On("reset", func(_ interface{}) {
				mu.Lock()
				resetCount++
				mu.Unlock()
				count.Set(0)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// Verify initial state
	assert.Equal(t, "Count: 0", component.View())

	// Test multi-key binding
	model, _ := component.Update(tea.KeyMsg{Type: tea.KeyUp})
	component = model.(bubbly.Component)

	mu.Lock()
	assert.Equal(t, 1, incrementCount, "Multi-key binding should work")
	mu.Unlock()

	// Verify count incremented
	assert.Equal(t, "Count: 1", component.View())

	// Test single binding
	model, _ = component.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	component = model.(bubbly.Component)

	mu.Lock()
	assert.Equal(t, 1, resetCount, "Single binding should work")
	mu.Unlock()

	// Verify both methods coexist - reset should have worked
	assert.Equal(t, "Count: 0", component.View(), "Reset should have worked")
}

// TestMultiKeyBinding_ManyKeys tests binding 10+ keys to a single event.
// This verifies there's no artificial limit on the number of keys.
func TestMultiKeyBinding_ManyKeys(t *testing.T) {
	var mu sync.Mutex
	var eventCount int

	// Create component with 12 keys bound to one event
	component, err := bubbly.NewComponent("ManyKeys").
		WithMultiKeyBindings("action", "Perform action",
			"1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "a", "b").
		Setup(func(ctx *bubbly.Context) {
			ctx.On("action", func(_ interface{}) {
				mu.Lock()
				eventCount++
				mu.Unlock()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return ""
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// Test all 12 keys
	keys := []rune{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0', 'a', 'b'}
	for _, key := range keys {
		model, _ := component.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}})
		component = model.(bubbly.Component)
	}

	// Verify all keys triggered the event
	mu.Lock()
	assert.Equal(t, 12, eventCount, "All 12 keys should trigger the event")
	mu.Unlock()
}

// TestMultiKeyBinding_EmptyKeysList tests that empty keys list is a safe no-op.
// This verifies error handling for edge case.
func TestMultiKeyBinding_EmptyKeysList(t *testing.T) {
	// Create component with empty keys list
	component, err := bubbly.NewComponent("EmptyKeys").
		WithMultiKeyBindings("action", "Perform action"). // No keys
		Setup(func(ctx *bubbly.Context) {
			ctx.On("action", func(_ interface{}) {
				t.Error("Event should not fire - no keys bound")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return ""
		}).
		Build()

	// Should build successfully (no-op)
	require.NoError(t, err)
	require.NotNil(t, component)

	component.Init()

	// Send a random key - should not trigger event
	model, _ := component.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	_ = model.(bubbly.Component)

	// Test passes if no error (event handler didn't fire)
}

// TestMultiKeyBinding_EventHandlerExecution tests that event handler logic executes correctly.
// This verifies the handler receives correct data and can perform complex operations.
func TestMultiKeyBinding_EventHandlerExecution(t *testing.T) {
	var mu sync.Mutex
	var executionLog []string

	component, err := bubbly.NewComponent("HandlerTest").
		WithMultiKeyBindings("log", "Log action", "a", "b", "c").
		Setup(func(ctx *bubbly.Context) {
			ctx.On("log", func(data interface{}) {
				mu.Lock()
				executionLog = append(executionLog, "executed")
				mu.Unlock()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return ""
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// Execute handler 3 times with different keys
	keys := []rune{'a', 'b', 'c'}
	for _, key := range keys {
		model, _ := component.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}})
		component = model.(bubbly.Component)
	}

	// Verify handler executed 3 times
	mu.Lock()
	assert.Equal(t, 3, len(executionLog), "Handler should execute for each key press")
	assert.Equal(t, []string{"executed", "executed", "executed"}, executionLog)
	mu.Unlock()
}

// TestMultiKeyBinding_HelpText tests that help text includes all bound keys.
// This verifies KeyBindings() method returns all registered keys.
func TestMultiKeyBinding_HelpText(t *testing.T) {
	component, err := bubbly.NewComponent("HelpTest").
		WithMultiKeyBindings("increment", "Increment counter", "up", "k", "+").
		WithMultiKeyBindings("decrement", "Decrement counter", "down", "j", "-").
		Setup(func(ctx *bubbly.Context) {}).
		Template(func(ctx bubbly.RenderContext) string {
			return ""
		}).
		Build()

	require.NoError(t, err)

	// Get key bindings
	bindings := component.KeyBindings()
	require.NotNil(t, bindings)

	// Verify all increment keys are registered
	incrementKeys := []string{"up", "k", "+"}
	for _, key := range incrementKeys {
		assert.Contains(t, bindings, key, "Key %s should be in bindings", key)
		assert.Len(t, bindings[key], 1, "Key %s should have one binding", key)
		assert.Equal(t, "increment", bindings[key][0].Event, "Key %s should trigger increment event", key)
		assert.Equal(t, "Increment counter", bindings[key][0].Description, "Key %s should have correct description", key)
	}

	// Verify all decrement keys are registered
	decrementKeys := []string{"down", "j", "-"}
	for _, key := range decrementKeys {
		assert.Contains(t, bindings, key, "Key %s should be in bindings", key)
		assert.Len(t, bindings[key], 1, "Key %s should have one binding", key)
		assert.Equal(t, "decrement", bindings[key][0].Event, "Key %s should trigger decrement event", key)
		assert.Equal(t, "Decrement counter", bindings[key][0].Description, "Key %s should have correct description", key)
	}

	// Verify total number of keys (6 keys total)
	assert.Len(t, bindings, 6, "Should have 6 keys registered")
}

// TestMultiKeyBinding_WithAutoCommands tests integration with auto-commands feature.
// This verifies multi-key bindings work with automatic command generation.
func TestMultiKeyBinding_WithAutoCommands(t *testing.T) {
	component, err := bubbly.NewComponent("AutoCounter").
		WithAutoCommands(true).
		WithMultiKeyBindings("increment", "Increment", "up", "k", "+").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			ctx.On("increment", func(_ interface{}) {
				c := count.Get().(int)
				count.Set(c + 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// Verify initial state
	assert.Equal(t, "Count: 0", component.View())

	// Test each key generates command
	keys := []tea.KeyMsg{
		{Type: tea.KeyUp},
		{Type: tea.KeyRunes, Runes: []rune{'k'}},
		{Type: tea.KeyRunes, Runes: []rune{'+'}},
	}

	for i, keyMsg := range keys {
		model, cmd := component.Update(keyMsg)
		component = model.(bubbly.Component)

		// Verify command generated (auto-commands enabled)
		assert.NotNil(t, cmd, "Auto-command should be generated for key %d", i)
	}

	// Verify final count via View()
	assert.Equal(t, "Count: 3", component.View(), "All keys should have incremented")
}
