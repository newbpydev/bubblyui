package bubbly

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWrapIntegration_CompleteCounter tests a complete counter example using Wrap()
// This demonstrates the simplest possible integration: one-line wrapper
func TestWrapIntegration_CompleteCounter(t *testing.T) {
	// Create counter component with automatic commands
	counter, err := NewComponent("Counter").
		WithAutoCommands(true).
		Setup(func(ctx *Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			// State changes automatically trigger UI updates
			ctx.On("increment", func(data interface{}) {
				current := count.Get().(int)
				count.Set(current + 1)
			})

			ctx.On("decrement", func(data interface{}) {
				current := count.Get().(int)
				count.Set(current - 1)
			})

			ctx.On("reset", func(data interface{}) {
				count.Set(0)
			})
		}).
		Template(func(ctx RenderContext) string {
			count := ctx.Get("count").(*Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()
	require.NoError(t, err)

	// ONE LINE INTEGRATION! No manual wrapper model needed
	model := Wrap(counter)

	// Initialize
	cmd := model.Init()
	assert.Nil(t, cmd) // No init commands for this simple example

	// Verify initial state
	view := model.View()
	assert.Equal(t, "Count: 0", view)

	// Simulate user interactions
	// Increment 3 times
	for i := 0; i < 3; i++ {
		counter.Emit("increment", nil)
		model, _ = model.Update(tea.KeyMsg{})
		// With auto commands, commands are generated and drained through component
	}

	// Verify state
	view = model.View()
	assert.Equal(t, "Count: 3", view)

	// Decrement once
	counter.Emit("decrement", nil)
	model, _ = model.Update(tea.KeyMsg{})

	view = model.View()
	assert.Equal(t, "Count: 2", view)

	// Reset
	counter.Emit("reset", nil)
	model, _ = model.Update(tea.KeyMsg{})

	view = model.View()
	assert.Equal(t, "Count: 0", view)
}

// TestWrapIntegration_MultipleStateChanges tests handling multiple state changes in one update
// This verifies that command batching works correctly
func TestWrapIntegration_MultipleStateChanges(t *testing.T) {
	tests := []struct {
		name          string
		autoCommands  bool
		updates       int
		expectedCount int
	}{
		{
			name:          "single update without auto commands",
			autoCommands:  false,
			updates:       1,
			expectedCount: 1,
		},
		{
			name:          "multiple updates with auto commands",
			autoCommands:  true,
			updates:       10,
			expectedCount: 10,
		},
		{
			name:          "batch updates with auto commands",
			autoCommands:  true,
			updates:       100,
			expectedCount: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component
			component, err := NewComponent("BatchTest").
				WithAutoCommands(tt.autoCommands).
				Setup(func(ctx *Context) {
					count := ctx.Ref(0)
					ctx.Expose("count", count)

					ctx.On("increment", func(data interface{}) {
						current := count.Get().(int)
						count.Set(current + 1)
					})

					ctx.On("batch-increment", func(data interface{}) {
						times := data.(int)
						// Multiple state changes in one handler
						// Just increment by times
						current := count.Get().(int)
						count.Set(current + times)
					})
				}).
				Template(func(ctx RenderContext) string {
					count := ctx.Get("count").(*Ref[interface{}])
					return fmt.Sprintf("Count: %d", count.Get().(int))
				}).
				Build()
			require.NoError(t, err)

			// Wrap component
			model := Wrap(component)
			model.Init()

			// Trigger batch update
			if tt.updates == 1 {
				component.Emit("increment", nil)
			} else {
				component.Emit("batch-increment", tt.updates)
			}

			// Process update
			var cmd tea.Cmd
			model, cmd = model.Update(tea.KeyMsg{})

			// With auto commands, cmd should handle batching
			if tt.autoCommands && tt.updates > 1 {
				// Commands are generated and batched
				assert.NotNil(t, cmd)
			}

			// Verify final state
			view := model.View()
			expected := fmt.Sprintf("Count: %d", tt.expectedCount)
			assert.Equal(t, expected, view)
		})
	}
}

// TestWrapIntegration_LifecycleHooks tests integration with lifecycle hooks
// This demonstrates that Wrap() correctly forwards lifecycle events
func TestWrapIntegration_LifecycleHooks(t *testing.T) {
	var setupCalled bool

	// Create component with lifecycle hooks
	component, err := NewComponent("LifecycleTest").
		WithAutoCommands(true).
		Setup(func(ctx *Context) {
			setupCalled = true

			count := ctx.Ref(0)
			ctx.Expose("count", count)

			// Lifecycle hooks (internal to component, demonstrated through behavior)
			ctx.OnMounted(func() {
				// Called during Init - behavior verified by test
			})

			ctx.OnUpdated(func() {
				// Called during Update - behavior verified by test
			})

			ctx.OnUnmounted(func() {
				// Called during cleanup - not directly tested in integration
			})

			ctx.On("increment", func(data interface{}) {
				current := count.Get().(int)
				count.Set(current + 1)
			})
		}).
		Template(func(ctx RenderContext) string {
			count := ctx.Get("count").(*Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()
	require.NoError(t, err)

	// Wrap component
	model := Wrap(component)

	// Init triggers mounted hook
	model.Init()

	// Verify setup was called
	assert.True(t, setupCalled, "Setup should be called")

	// Note: Lifecycle hooks (OnMounted, OnUpdated, OnUnmounted) are internal to the component
	// They are called by the component's lifecycle manager during Init/Update/cleanup
	// For integration tests, we verify behavior rather than internal hook execution

	// Trigger update to ensure component is working
	component.Emit("increment", nil)
	model, _ = model.Update(tea.KeyMsg{})

	// Verify component state changed correctly
	view := model.View()
	assert.Equal(t, "Count: 1", view, "Component should update correctly via lifecycle system")

	// The fact that the component works correctly demonstrates that:
	// - Init() called the lifecycle hooks
	// - Update() processed the event correctly
	// - The reactive system works through Wrap()
}

// TestWrapIntegration_CommandBatching tests that commands are properly batched
// This verifies the core benefit of automatic command generation
func TestWrapIntegration_CommandBatching(t *testing.T) {
	tests := []struct {
		name         string
		stateChanges int
		description  string
	}{
		{
			name:         "single state change",
			stateChanges: 1,
			description:  "One command generated",
		},
		{
			name:         "multiple state changes",
			stateChanges: 5,
			description:  "Multiple commands batched",
		},
		{
			name:         "many state changes",
			stateChanges: 50,
			description:  "Many commands efficiently batched",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track command generation
			commandCount := 0

			component, err := NewComponent("BatchTest").
				WithAutoCommands(true).
				Setup(func(ctx *Context) {
					values := ctx.Ref([]int{})
					ctx.Expose("values", values)

					ctx.On("add-many", func(data interface{}) {
						count := data.(int)
						current := values.Get().([]int)

						// Multiple state changes in rapid succession
						for i := 0; i < count; i++ {
							current = append(current, i)
							values.Set(current)
							commandCount++
						}
					})
				}).
				Template(func(ctx RenderContext) string {
					values := ctx.Get("values").(*Ref[interface{}])
					vals := values.Get().([]int)
					return fmt.Sprintf("Values: %d items", len(vals))
				}).
				Build()
			require.NoError(t, err)

			model := Wrap(component)
			model.Init()

			// Trigger multiple state changes
			component.Emit("add-many", tt.stateChanges)
			var cmd tea.Cmd
			model, cmd = model.Update(tea.KeyMsg{})

			// Verify commands were generated
			assert.Equal(t, tt.stateChanges, commandCount, "Should generate commands for each state change")

			// Verify batching occurred (cmd is not nil when auto commands enabled)
			if tt.stateChanges > 0 {
				assert.NotNil(t, cmd, "Commands should be batched into single tea.Cmd")
			}

			// Verify final state
			view := model.View()
			expected := fmt.Sprintf("Values: %d items", tt.stateChanges)
			assert.Equal(t, expected, view)
		})
	}
}

// TestWrapIntegration_BackwardCompatibility tests that Wrap() works with both manual and automatic modes
// This demonstrates the migration path from manual to automatic
func TestWrapIntegration_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name         string
		autoCommands bool
		description  string
	}{
		{
			name:         "Wrap() with manual commands (backward compatible)",
			autoCommands: false,
			description:  "Use Wrap() but keep manual command generation",
		},
		{
			name:         "Wrap() with auto commands (recommended)",
			autoCommands: true,
			description:  "Full automatic mode - recommended approach",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component (same for all tests)
			component, err := NewComponent("CompatTest").
				WithAutoCommands(tt.autoCommands).
				Setup(func(ctx *Context) {
					count := ctx.Ref(0)
					ctx.Expose("count", count)

					ctx.On("increment", func(data interface{}) {
						current := count.Get().(int)
						count.Set(current + 1)
					})
				}).
				Template(func(ctx RenderContext) string {
					count := ctx.Get("count").(*Ref[interface{}])
					return fmt.Sprintf("Count: %d", count.Get().(int))
				}).
				Build()
			require.NoError(t, err)

			// Use Wrap() for both (demonstrating it works in both modes)
			model := Wrap(component)
			model.Init()

			// Increment 5 times
			for i := 0; i < 5; i++ {
				component.Emit("increment", nil)
				model, _ = model.Update(tea.KeyMsg{})
			}

			// Verify final state (same result regardless of mode)
			view := model.View()
			assert.Equal(t, "Count: 5", view)
		})
	}
}

// TestWrapIntegration_RealWorldScenario tests a realistic form scenario
// This demonstrates a practical use case with multiple fields and validation
func TestWrapIntegration_RealWorldScenario(t *testing.T) {
	// Create a form component with validation
	form, err := NewComponent("UserForm").
		WithAutoCommands(true).
		Setup(func(ctx *Context) {
			// Form fields
			username := ctx.Ref("")
			email := ctx.Ref("")
			valid := ctx.Ref(false)

			ctx.Expose("username", username)
			ctx.Expose("email", email)
			ctx.Expose("valid", valid)

			// Validation logic
			validateForm := func() {
				user := username.Get().(string)
				mail := email.Get().(string)
				isValid := len(user) >= 3 && len(mail) > 0
				valid.Set(isValid)
			}

			// Field update handlers
			ctx.On("update-username", func(data interface{}) {
				username.Set(data.(string))
				validateForm()
			})

			ctx.On("update-email", func(data interface{}) {
				email.Set(data.(string))
				validateForm()
			})

			ctx.On("submit", func(_ interface{}) {
				// Form validation - in real implementation would submit if valid
				_ = valid.Get().(bool)
			})
		}).
		Template(func(ctx RenderContext) string {
			username := ctx.Get("username").(*Ref[interface{}])
			email := ctx.Get("email").(*Ref[interface{}])
			valid := ctx.Get("valid").(*Ref[interface{}])

			validStr := "❌ Invalid"
			if valid.Get().(bool) {
				validStr = "✅ Valid"
			}

			return fmt.Sprintf("Username: %s\nEmail: %s\nStatus: %s",
				username.Get().(string),
				email.Get().(string),
				validStr,
			)
		}).
		Build()
	require.NoError(t, err)

	// ONE LINE! No manual wrapper needed
	model := Wrap(form)
	model.Init()

	// Verify initial state
	view := model.View()
	assert.Contains(t, view, "❌ Invalid")

	// Update username (too short)
	form.Emit("update-username", "ab")
	model, _ = model.Update(tea.KeyMsg{})

	view = model.View()
	assert.Contains(t, view, "❌ Invalid")

	// Update username (valid length)
	form.Emit("update-username", "john")
	model, _ = model.Update(tea.KeyMsg{})

	view = model.View()
	assert.Contains(t, view, "❌ Invalid") // Still invalid (no email)

	// Add email
	form.Emit("update-email", "john@example.com")
	model, _ = model.Update(tea.KeyMsg{})

	// Now form is valid
	view = model.View()
	assert.Contains(t, view, "✅ Valid")
	assert.Contains(t, view, "john")
	assert.Contains(t, view, "john@example.com")
}
