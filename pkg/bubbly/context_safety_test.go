package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTemplateContextDetection verifies that Ref.Set() panics when called inside a template
func TestTemplateContextDetection(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*Context)
		template      func(RenderContext) string
		shouldPanic   bool
		panicContains string
	}{
		{
			name: "panic_on_set_in_template",
			setup: func(ctx *Context) {
				count := ctx.Ref(0)
				ctx.Expose("count", count)
			},
			template: func(ctx RenderContext) string {
				count := ctx.Get("count").(*Ref[interface{}])
				// This should panic - cannot mutate state in template
				count.Set(42)
				return "Should not reach here"
			},
			shouldPanic:   true,
			panicContains: "Ref.Set", // Just check that message contains key terms
		},
		{
			name: "normal_set_outside_template_works",
			setup: func(ctx *Context) {
				count := ctx.Ref(0)
				ctx.Expose("count", count)

				// This should work - Set() outside template
				ctx.On("increment", func(data interface{}) {
					count.Set(count.Get().(int) + 1)
				})
			},
			template: func(ctx RenderContext) string {
				count := ctx.Get("count").(*Ref[interface{}])
				return "Count: " + string(rune('0'+count.Get().(int)))
			},
			shouldPanic: false,
		},
		{
			name: "get_in_template_allowed",
			setup: func(ctx *Context) {
				count := ctx.Ref(42)
				ctx.Expose("count", count)
			},
			template: func(ctx RenderContext) string {
				count := ctx.Get("count").(*Ref[interface{}])
				// Get() is allowed in templates (read-only)
				val := count.Get().(int)
				return "Value: " + string(rune('0'+val))
			},
			shouldPanic: false,
		},
		{
			name: "multiple_gets_in_template_allowed",
			setup: func(ctx *Context) {
				count := ctx.Ref(5)
				name := ctx.Ref("test")
				ctx.Expose("count", count)
				ctx.Expose("name", name)
			},
			template: func(ctx RenderContext) string {
				count := ctx.Get("count").(*Ref[interface{}])
				name := ctx.Get("name").(*Ref[interface{}])
				// Multiple Get() calls allowed
				return name.Get().(string) + ": " + string(rune('0'+count.Get().(int)))
			},
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build component
			builder := NewComponent("TestComponent").
				Setup(tt.setup).
				Template(tt.template)

			component, err := builder.Build()
			require.NoError(t, err)

			// Initialize component
			component.Init()

			// Test View() which will execute template
			if tt.shouldPanic {
				// Verify it panics
				var panicValue interface{}
				assert.Panics(t, func() {
					defer func() {
						if r := recover(); r != nil {
							panicValue = r
							panic(r) // Re-panic so assert.Panics detects it
						}
					}()
					component.View()
				}, "Expected panic when calling Set() in template")

				// Verify panic message contains expected keywords
				if panicValue != nil {
					panicMsg := panicValue.(string)
					assert.Contains(t, panicMsg, tt.panicContains, "Panic message should contain expected keywords")
				}
			} else {
				assert.NotPanics(t, func() {
					view := component.View()
					assert.NotEmpty(t, view, "Should render successfully")
				}, "Should not panic for read-only operations in template")
			}
		})
	}
}

// TestTemplateContextLifecycle verifies enter/exit template context tracking
func TestTemplateContextLifecycle(t *testing.T) {
	tests := []struct {
		name         string
		operation    func(*Context) bool
		expectedBool bool
		description  string
	}{
		{
			name: "not_in_template_initially",
			operation: func(ctx *Context) bool {
				return ctx.InTemplate()
			},
			expectedBool: false,
			description:  "Context should not be in template initially",
		},
		{
			name: "in_template_after_enter",
			operation: func(ctx *Context) bool {
				ctx.enterTemplate()
				return ctx.InTemplate()
			},
			expectedBool: true,
			description:  "Context should be in template after enterTemplate()",
		},
		{
			name: "not_in_template_after_exit",
			operation: func(ctx *Context) bool {
				ctx.enterTemplate()
				ctx.exitTemplate()
				return ctx.InTemplate()
			},
			expectedBool: false,
			description:  "Context should not be in template after exitTemplate()",
		},
		{
			name: "multiple_exit_safe",
			operation: func(ctx *Context) bool {
				ctx.enterTemplate()
				ctx.exitTemplate()
				ctx.exitTemplate() // Extra exit is safe (no-op)
				return ctx.InTemplate()
			},
			expectedBool: false,
			description:  "Multiple exits should be safe and result in false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a component to get a context
			component, err := NewComponent("TestComponent").
				Setup(func(ctx *Context) {
					ctx.Expose("result", tt.operation(ctx))
				}).
				Template(func(ctx RenderContext) string {
					return "test"
				}).
				Build()

			require.NoError(t, err)
			component.Init()

			result := component.(*componentImpl).state["result"]
			assert.Equal(t, tt.expectedBool, result, tt.description)
		})
	}
}

// TestTemplateContextThreadSafety verifies thread-safe template context tracking
func TestTemplateContextThreadSafety(t *testing.T) {
	// Create component with auto commands (uses goroutines internally)
	component, err := NewComponent("TestComponent").
		WithAutoCommands(true).
		Setup(func(ctx *Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx RenderContext) string {
			// Just read - no mutations
			count := ctx.Get("count").(*Ref[interface{}])
			return "Count: " + string(rune('0'+count.Get().(int)))
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// Call View() multiple times concurrently
	// This tests that inTemplate flag is properly protected by mutex
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			_ = component.View()
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we got here without data races, the test passes
}

// TestTemplateContextPanicMessage verifies clear error messages
func TestTemplateContextPanicMessage(t *testing.T) {
	component, err := NewComponent("TestComponent").
		Setup(func(ctx *Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx RenderContext) string {
			count := ctx.Get("count").(*Ref[interface{}])
			count.Set(100) // Should panic
			return "unreachable"
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// Capture panic value
	var panicValue interface{}
	func() {
		defer func() {
			panicValue = recover()
		}()
		component.View()
	}()

	// Verify panic occurred
	require.NotNil(t, panicValue, "Expected panic when calling Set() in template")

	// Verify error message is helpful
	panicMsg := panicValue.(string)
	assert.Contains(t, panicMsg, "Ref.Set()", "Error should mention Ref.Set()")
	assert.Contains(t, panicMsg, "template", "Error should mention template")
	assert.Contains(t, panicMsg, "pure", "Error should mention purity requirement")
}

// TestTemplateContextAfterPanic verifies state is clean after panic
func TestTemplateContextAfterPanic(t *testing.T) {
	component, err := NewComponent("TestComponent").
		Setup(func(c *Context) {
			count := c.Ref(0)
			c.Expose("count", count)
		}).
		Template(func(rc RenderContext) string {
			count := rc.Get("count").(*Ref[interface{}])
			count.Set(100) // Will panic
			return "unreachable"
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// First View() should panic
	assert.Panics(t, func() {
		component.View()
	})

	// After panic, template context should be clean (exitTemplate called via defer)
	// We can't directly test this without exposing internal state,
	// but we can verify a second call with fixed template works

	// Update template to not call Set()
	component.(*componentImpl).template = func(rc RenderContext) string {
		count := rc.Get("count").(*Ref[interface{}])
		return "Count: " + string(rune('0'+count.Get().(int)))
	}

	// This should work now
	assert.NotPanics(t, func() {
		view := component.View()
		assert.Contains(t, view, "Count")
	})
}

// TestTemplateContextMultipleComponents verifies isolation between components
func TestTemplateContextMultipleComponents(t *testing.T) {
	// Component 1: Tries to mutate in template (should panic)
	component1, _ := NewComponent("Component1").
		Setup(func(ctx *Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx RenderContext) string {
			count := ctx.Get("count").(*Ref[interface{}])
			count.Set(1) // Should panic
			return "unreachable"
		}).
		Build()

	// Component 2: Normal read-only template (should work)
	component2, _ := NewComponent("Component2").
		Setup(func(ctx *Context) {
			value := ctx.Ref(42)
			ctx.Expose("value", value)
		}).
		Template(func(ctx RenderContext) string {
			value := ctx.Get("value").(*Ref[interface{}])
			return "Value: " + string(rune('0'+value.Get().(int)))
		}).
		Build()

	component1.Init()
	component2.Init()

	// Component 1 should panic
	assert.Panics(t, func() {
		component1.View()
	})

	// Component 2 should still work (isolation)
	assert.NotPanics(t, func() {
		view := component2.View()
		assert.Contains(t, view, "Value")
	})
}
