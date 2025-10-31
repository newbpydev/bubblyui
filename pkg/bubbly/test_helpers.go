package bubbly

// This file provides testing utilities for BubblyUI composables and components.
// These functions are exported to enable testing in external test packages
// while maintaining encapsulation of internal component implementation.

// NewTestContext creates a minimal Context suitable for testing composables.
// The returned context supports all standard Context operations without
// requiring full Bubbletea integration.
//
// This function is primarily used by the btesting package and in component tests.
//
// Example:
//
//	ctx := bubbly.NewTestContext()
//	count := ctx.Ref(0)
//	count.Set(42)
func NewTestContext() *Context {
	comp := newComponentImpl("TestComponent")
	return &Context{component: comp}
}

// TriggerMount executes all onMounted hooks registered on the context.
// This simulates the component mount lifecycle event for testing.
//
// Example:
//
//	ctx := bubbly.NewTestContext()
//	mounted := false
//	ctx.OnMounted(func() { mounted = true })
//	bubbly.TriggerMount(ctx)
//	// mounted == true
func TriggerMount(ctx *Context) {
	if ctx.component.lifecycle != nil {
		ctx.component.lifecycle.executeMounted()
	}
}

// TriggerUpdate executes all onUpdated hooks registered on the context.
// This simulates the component update lifecycle event for testing.
//
// Example:
//
//	ctx := bubbly.NewTestContext()
//	updated := false
//	ctx.OnUpdated(func() { updated = true })
//	bubbly.TriggerUpdate(ctx)
//	// updated == true
func TriggerUpdate(ctx *Context) {
	if ctx.component.lifecycle != nil {
		ctx.component.lifecycle.executeUpdated()
	}
}

// TriggerUnmount executes all onUnmounted hooks registered on the context.
// This simulates the component unmount lifecycle event for testing.
//
// Example:
//
//	ctx := bubbly.NewTestContext()
//	unmounted := false
//	ctx.OnUnmounted(func() { unmounted = true })
//	bubbly.TriggerUnmount(ctx)
//	// unmounted == true
func TriggerUnmount(ctx *Context) {
	if ctx.component.lifecycle != nil {
		ctx.component.lifecycle.executeUnmounted()
	}
}

// SetParent establishes a parent-child relationship between two test contexts.
// This enables testing of provide/inject functionality across component trees.
//
// Example:
//
//	parent := bubbly.NewTestContext()
//	child := bubbly.NewTestContext()
//	bubbly.SetParent(child, parent)
//
//	parent.Provide("theme", "dark")
//	theme := child.Inject("theme", "light")
//	// theme == "dark"
func SetParent(child, parent *Context) {
	child.component.parent = parent.component
}
