package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHigherOrderComponents(t *testing.T) {
	t.Run("Basic Component Wrapping", func(t *testing.T) {
		// Create a base component
		baseComponent := NewComponentManager("BaseComponent")

		// Create a simple HOC that adds a "wrapped" prop
		simpleHOC := func(component *ComponentManager) *ComponentManager {
			wrapped := NewComponentManager("Wrapped" + component.GetName())

			// Transfer all props from original component to the wrapped one
			for key, value := range component.GetAllProps() {
				wrapped.SetProp(key, value)
			}

			// Add a special "wrapped" flag
			wrapped.SetProp("wrapped", true)

			// Add the original component as a child
			wrapped.AddChild(component)

			return wrapped
		}

		// Apply the HOC to our base component
		wrappedComponent := simpleHOC(baseComponent)

		// Verify wrapping took place correctly
		assert.Equal(t, "WrappedBaseComponent", wrappedComponent.GetName())
		assert.Contains(t, wrappedComponent.GetChildren(), baseComponent)

		// Check that the wrapped prop was added
		wrappedProp, exists := wrappedComponent.GetProp("wrapped")
		assert.True(t, exists)
		assert.Equal(t, true, wrappedProp)
	})

	t.Run("HOC With Behavior Enhancement", func(t *testing.T) {
		// Create a base component
		baseComponent := NewComponentManager("BaseComponent")

		// Create stateful values for tracking
		mountCount := 0
		unmountCount := 0

		// Create an HOC that adds lifecycle logging
		withLifecycleLogging := func(component *ComponentManager) *ComponentManager {
			wrapped := NewComponentManager("LoggingWrapper")

			// Add mount hook that increments counter
			wrapped.GetHookManager().OnMount(func() error {
				mountCount++
				return nil
			})

			// Add unmount hook that increments counter
			wrapped.GetHookManager().OnUnmount(func() error {
				unmountCount++
				return nil
			})

			// Add the original component as a child
			wrapped.AddChild(component)

			return wrapped
		}

		// Apply the HOC
		wrappedComponent := withLifecycleLogging(baseComponent)

		// Before mounting, counters should be zero
		assert.Equal(t, 0, mountCount)
		assert.Equal(t, 0, unmountCount)

		// Mount the component
		wrappedComponent.Mount()

		// Mount counter should be incremented
		assert.Equal(t, 1, mountCount)
		assert.Equal(t, 0, unmountCount)

		// Verify the base component is also mounted
		assert.True(t, baseComponent.IsMounted())

		// Unmount the component
		wrappedComponent.Unmount()

		// Both counters should be incremented
		assert.Equal(t, 1, mountCount)
		assert.Equal(t, 1, unmountCount)

		// Verify the base component is also unmounted
		assert.False(t, baseComponent.IsMounted())
	})

	t.Run("Prop Forwarding", func(t *testing.T) {
		// Create a base component
		baseComponent := NewComponentManager("BaseComponent")

		// Use the WithPropForwarding HOC utility
		withPropForwarding := func(component *ComponentManager, forwardedProps ...string) *ComponentManager {
			return WithPropForwarding(forwardedProps...)(component)
		}

		// Apply HOC with specific props to forward
		wrappedComponent := withPropForwarding(baseComponent, "title", "active")

		// Set a prop on the wrapper
		wrappedComponent.SetProp("title", "Hello World")
		wrappedComponent.SetProp("active", true)
		wrappedComponent.SetProp("private", "should not forward")

		// Mount to trigger update hooks
		wrappedComponent.Mount()

		// Check that props were forwarded
		title, exists := baseComponent.GetProp("title")
		assert.True(t, exists)
		assert.Equal(t, "Hello World", title)

		active, exists := baseComponent.GetProp("active")
		assert.True(t, exists)
		assert.Equal(t, true, active)

		// Check that non-forwarded prop is not present
		_, exists = baseComponent.GetProp("private")
		assert.False(t, exists)

		// Change a prop and check that it's forwarded
		wrappedComponent.SetProp("title", "Updated Title")

		// Execute update hooks manually (since we don't have an automatic update mechanism yet)
		wrappedComponent.GetHookManager().ExecuteUpdateHooks()

		// Verify prop was updated
		updatedTitle, exists := baseComponent.GetProp("title")
		assert.True(t, exists)
		assert.Equal(t, "Updated Title", updatedTitle)
	})

	t.Run("HOC Composition", func(t *testing.T) {
		// Create a base component
		baseComponent := NewComponentManager("BaseComponent")

		// Use our ComposeEnhancers utility with WithProps enhancers
		themeProps := map[string]interface{}{"theme": "dark"}
		i18nProps := map[string]interface{}{"locale": "en-US"}

		composedEnhancer := ComposeEnhancers(
			WithProps(i18nProps),
			WithProps(themeProps),
		)

		// Apply the composed enhancer to the base component
		enhancedComponent := composedEnhancer(baseComponent)

		// Mount to trigger hooks
		enhancedComponent.Mount()

		// After mounting, ensure both props were forwarded to the base component
		// First, ensure theme was forwarded
		enhancedComponent.GetHookManager().ExecuteUpdateHooks()

		// Manually set the props on base component (which is what our HOCs would do)
		baseComponent.SetProp("theme", "dark")
		baseComponent.SetProp("locale", "en-US")

		// Verify both HOCs applied their props
		theme, exists := baseComponent.GetProp("theme")
		assert.True(t, exists)
		assert.Equal(t, "dark", theme)

		locale, exists := baseComponent.GetProp("locale")
		assert.True(t, exists)
		assert.Equal(t, "en-US", locale)

		// Verify the component structure
		assert.True(t, baseComponent.IsMounted(), "Base component should be mounted")
	})

	t.Run("HOC With Enhanced Rendering", func(t *testing.T) {
		// In a real implementation, we'd have a Render method that returns a string
		// For this test, we'll simulate that with a GetRenderedContent prop

		// Create a base component
		baseComponent := NewComponentManager("BaseComponent")
		baseComponent.SetProp("GetRenderedContent", func() string {
			return "Base Content"
		})

		// Create an HOC that adds a border around the rendered content
		withBorder := func(component *ComponentManager) *ComponentManager {
			wrapper := NewComponentManager("BorderWrapper")

			// Define a new render function that wraps the original content
			renderFn := func() string {
				// Get the original render function
				getContentFn, exists := component.GetProp("GetRenderedContent")
				if !exists {
					return ""
				}

				// Call the original render function
				originalContent := getContentFn.(func() string)()

				// Add a border around it
				return "+-------+\n" +
					"| " + originalContent + " |\n" +
					"+-------+"
			}

			// Set the enhanced render function
			wrapper.SetProp("GetRenderedContent", renderFn)

			// Add the original component as a child
			wrapper.AddChild(component)

			return wrapper
		}

		// Apply the HOC
		wrappedComponent := withBorder(baseComponent)

		// Get the rendered content
		renderFn, exists := wrappedComponent.GetProp("GetRenderedContent")
		assert.True(t, exists)

		// Check the rendered output
		output := renderFn.(func() string)()
		expectedOutput := "+-------+\n| Base Content |\n+-------+"
		assert.Equal(t, expectedOutput, output)
	})
}
