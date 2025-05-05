package bubble

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/core"
	"github.com/stretchr/testify/assert"
)

// TestBasicIntegration is a simplified test that verifies the core integration
// between BubblyUI components and the Bubble Tea model wrapper.
// This avoids trying to test the full Bubble Tea runtime which can be difficult
// in a test environment due to terminal UI interactions.
func TestBasicIntegration(t *testing.T) {
	t.Run("Component Rendering Through Bubble Model", func(t *testing.T) {
		// Create a simple component with render functionality
		root := core.NewComponentManager("TestComponent")

		// Use a simple render function for testing
		root.SetProp("render", func() string {
			return "Test Component Rendered"
		})

		// Create a bubble model in test mode
		model := NewBubbleModel(root, WithTestMode())

		// Initialize the model
		model.Init()

		// Verify that View returns the rendered content
		output := model.View()
		assert.Contains(t, output, "Test Component Rendered")
	})

	t.Run("Component State Updates", func(t *testing.T) {
		// Create a component with state that can be updated
		root := core.NewComponentManager("StatefulComponent")

		// Initialize state
		counter := 0
		root.SetProp("counter", counter)

		// Set up render function that uses state
		root.SetProp("render", func() string {
			value, _ := root.GetProp("counter")
			return "Counter: " + string(rune(value.(int)+'0'))
		})

		// Create a bubble model
		model := NewBubbleModel(root, WithTestMode())
		model.Init()

		// Initial state check
		output := model.View()
		assert.Contains(t, output, "Counter: 0")

		// Update state
		root.SetProp("counter", 1)

		// Check updated output
		updatedOutput := model.View()
		assert.Contains(t, updatedOutput, "Counter: 1")
	})

	t.Run("Component Lifecycle Integration", func(t *testing.T) {
		// Create a component with mount/unmount hooks
		root := core.NewComponentManager("LifecycleComponent")

		// Track mount state
		mounted := false
		unmounted := false

		// Add lifecycle hooks
		root.GetHookManager().OnMount(func() error {
			mounted = true
			return nil
		})

		root.GetHookManager().OnUnmount(func() error {
			unmounted = true
			return nil
		})

		// Create model and check initial state
		model := NewBubbleModel(root, WithTestMode())
		assert.False(t, mounted)
		assert.False(t, unmounted)

		// Initialize (which should mount the component)
		model.Init()
		assert.True(t, mounted)
		assert.False(t, unmounted)

		// Custom unmount message
		type UnmountTestMsg struct{}

		// Process an unmount message
		_, _ = model.updateTestMode(UnmountMsg{})

		// Component should be unmounted
		assert.True(t, unmounted)
	})
}
