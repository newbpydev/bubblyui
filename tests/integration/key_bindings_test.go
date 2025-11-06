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

// TestKeyBindingsAutoCommands_BasicFlow tests basic key binding → event → Ref.Set() → auto-command → UI update flow
func TestKeyBindingsAutoCommands_BasicFlow(t *testing.T) {
	// Create counter component with key bindings and auto-commands
	component, err := bubbly.NewComponent("Counter").
		WithAutoCommands(true).
		WithKeyBinding(" ", "increment", "Increment counter").
		WithKeyBinding("r", "reset", "Reset counter").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			ctx.On("increment", func(_ interface{}) {
				c := count.Get().(int)
				count.Set(c + 1)
				// No manual Emit() needed - auto-commands handle it
			})

			ctx.On("reset", func(_ interface{}) {
				count.Set(0)
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

	// Simulate space keypress
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	model, cmd := component.Update(spaceMsg)
	component = model.(bubbly.Component)

	// Verify UI updated immediately (state changes are synchronous)
	assert.Equal(t, "Count: 1", component.View())

	// Auto-commands should still be generated for Bubbletea state management
	assert.NotNil(t, cmd, "Expected command from auto-commands")

	// Test reset key
	resetMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	model, cmd = component.Update(resetMsg)
	component = model.(bubbly.Component)

	// Verify reset worked immediately
	assert.Equal(t, "Count: 0", component.View())

	// Command should still be generated
	assert.NotNil(t, cmd, "Expected command from auto-commands")
}

// TestKeyBindingsAutoCommands_MultipleChanges tests multiple rapid state changes batch correctly
func TestKeyBindingsAutoCommands_MultipleChanges(t *testing.T) {
	var mu sync.Mutex
	var commandCount int

	component, err := bubbly.NewComponent("MultiCounter").
		WithAutoCommands(true).
		WithKeyBinding(" ", "increment", "Increment").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			ctx.On("increment", func(_ interface{}) {
				c := count.Get().(int)
				count.Set(c + 1)
				mu.Lock()
				commandCount++
				mu.Unlock()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// Simulate 5 rapid space keypresses
	for i := 0; i < 5; i++ {
		spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
		model, _ := component.Update(spaceMsg)
		component = model.(bubbly.Component)
	}

	// Verify all changes applied immediately
	assert.Equal(t, "Count: 5", component.View())

	// Verify all handlers executed
	mu.Lock()
	assert.Equal(t, 5, commandCount, "All increment handlers should have executed")
	mu.Unlock()
}

// TestKeyBindingsAutoCommands_CommandBatching tests that commands batch correctly in single Update cycle
func TestKeyBindingsAutoCommands_CommandBatching(t *testing.T) {
	component, err := bubbly.NewComponent("BatchCounter").
		WithAutoCommands(true).
		WithKeyBinding(" ", "incrementMultiple", "Increment 3 times").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			ctx.On("incrementMultiple", func(_ interface{}) {
				// Multiple state changes in one event handler
				c := count.Get().(int)
				count.Set(c + 1)
				count.Set(c + 2)
				count.Set(c + 3)
				// All 3 should batch into single Update cycle
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// Initial state
	assert.Equal(t, "Count: 0", component.View())

	// Trigger multiple changes
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	model, cmd := component.Update(spaceMsg)
	component = model.(bubbly.Component)

	// Final state should reflect last Set() call immediately
	assert.Equal(t, "Count: 3", component.View())

	// Commands should still be generated
	assert.NotNil(t, cmd, "Expected batched commands")
}

// TestKeyBindingsAutoCommands_WithWrap tests integration with bubbly.Wrap() helper
func TestKeyBindingsAutoCommands_WithWrap(t *testing.T) {
	component, err := bubbly.NewComponent("WrappedCounter").
		WithAutoCommands(true).
		WithKeyBinding(" ", "increment", "Increment").
		WithKeyBinding("ctrl+c", "quit", "Quit").
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

	// Wrap component
	wrapped := bubbly.Wrap(component)

	// Init
	cmd := wrapped.Init()
	assert.Nil(t, cmd) // No init commands for this simple case

	// Initial view
	assert.Equal(t, "Count: 0", wrapped.View())

	// Simulate space keypress
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	model, cmd := wrapped.Update(spaceMsg)
	wrapped = model.(tea.Model)

	// Verify update happened immediately
	assert.Equal(t, "Count: 1", wrapped.View())

	// Command should still be generated
	assert.NotNil(t, cmd, "Expected command from auto-commands")

	// Test quit key
	quitMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	model, cmd = wrapped.Update(quitMsg)

	// Should return tea.Quit
	assert.NotNil(t, cmd, "Expected quit command")
	// Execute to verify it's tea.Quit
	msg := cmd()
	_, isQuitMsg := msg.(tea.QuitMsg)
	assert.True(t, isQuitMsg, "Expected tea.QuitMsg")
}

// TestKeyBindingsAutoCommands_NoAutoMode tests backward compatibility without auto-commands
func TestKeyBindingsAutoCommands_NoAutoMode(t *testing.T) {
	var eventFired bool
	var mu sync.Mutex

	component, err := bubbly.NewComponent("ManualCounter").
		// WithAutoCommands NOT called - manual mode
		WithKeyBinding(" ", "increment", "Increment").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			ctx.On("increment", func(_ interface{}) {
				mu.Lock()
				eventFired = true
				mu.Unlock()
				c := count.Get().(int)
				count.Set(c + 1)
				// In manual mode, this won't auto-update UI
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// Simulate keypress
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	model, cmd := component.Update(spaceMsg)
	component = model.(bubbly.Component)

	// Event fires synchronously, so no need to sleep
	mu.Lock()
	fired := eventFired
	mu.Unlock()
	assert.True(t, fired, "Event should have fired")

	// State updated immediately (even in manual mode, Set() works)
	assert.Equal(t, "Count: 1", component.View())

	// But no command should be generated (manual mode)
	assert.Nil(t, cmd, "No commands in manual mode")
}
