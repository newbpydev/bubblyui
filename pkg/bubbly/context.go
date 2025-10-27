package bubbly

import "fmt"

// Context provides the API available during component setup.
// It allows components to create reactive state, register event handlers,
// and access component data during the initialization phase.
//
// The Context is passed to the Setup function and provides access to:
//   - Reactive primitives (Ref, Computed, Watch)
//   - Event handling (On, Emit)
//   - Component data (Props, Children)
//   - State exposure (Expose, Get)
//
// Example usage in a Setup function:
//
//	Setup(func(ctx *Context) {
//	    // Create reactive state
//	    count := ctx.Ref(0)
//	    doubled := ctx.Computed(func() interface{} {
//	        return count.Get().(int) * 2
//	    })
//
//	    // Expose state to template
//	    ctx.Expose("count", count)
//	    ctx.Expose("doubled", doubled)
//
//	    // Register event handlers
//	    ctx.On("increment", func(data interface{}) {
//	        current := count.Get().(int)
//	        count.Set(current + 1)
//	    })
//
//	    // Watch for changes
//	    ctx.Watch(count, func(oldVal, newVal interface{}) {
//	        log.Printf("Count changed: %v -> %v", oldVal, newVal)
//	    })
//
//	    // Access props
//	    props := ctx.Props()
//
//	    // Access children
//	    children := ctx.Children()
//	})
type Context struct {
	component *componentImpl
}

// Ref creates a new reactive reference with the given initial value.
// The returned Ref can be used to get and set values reactively.
//
// Example:
//
//	count := ctx.Ref(0)
//	count.Set(42)
//	value := count.Get()
func (ctx *Context) Ref(value interface{}) *Ref[interface{}] {
	return NewRef(value)
}

// Computed creates a new computed value that automatically updates
// when its dependencies change.
//
// Example:
//
//	count := ctx.Ref(10)
//	doubled := ctx.Computed(func() interface{} {
//	    return count.Get().(int) * 2
//	})
func (ctx *Context) Computed(fn func() interface{}) *Computed[interface{}] {
	return NewComputed(fn)
}

// Watch registers a callback that is called whenever the given Ref changes.
// The callback receives the new and old values.
//
// Example:
//
//	count := ctx.Ref(0)
//	ctx.Watch(count, func(newVal, oldVal interface{}) {
//	    log.Printf("Count: %v -> %v", oldVal, newVal)
//	})
func (ctx *Context) Watch(ref *Ref[interface{}], callback WatchCallback[interface{}]) {
	Watch(ref, callback)
}

// Expose stores a value in the component's state map, making it accessible
// in the template function via RenderContext.Get().
//
// This is the primary way to share state between the setup function
// and the template function.
//
// Example:
//
//	count := ctx.Ref(0)
//	ctx.Expose("count", count)
//	// Later in template: count := ctx.Get("count").(*Ref[int])
func (ctx *Context) Expose(key string, value interface{}) {
	if ctx.component.state == nil {
		ctx.component.state = make(map[string]interface{})
	}
	ctx.component.state[key] = value
}

// Get retrieves a value from the component's state map.
// Returns nil if the key doesn't exist.
//
// This is typically used in the setup function to access previously
// exposed values, though it's more commonly used in templates.
//
// Example:
//
//	value := ctx.Get("count")
//	if ref, ok := value.(*Ref[int]); ok {
//	    // Use the ref
//	}
func (ctx *Context) Get(key string) interface{} {
	if ctx.component.state == nil {
		return nil
	}
	return ctx.component.state[key]
}

// On registers an event handler for the specified event name.
// Multiple handlers can be registered for the same event.
//
// Event handlers are called when the event is emitted via Emit()
// or when the component receives the event from a child.
//
// Example:
//
//	ctx.On("submit", func(data interface{}) {
//	    if formData, ok := data.(FormData); ok {
//	        // Handle form submission
//	    }
//	})
func (ctx *Context) On(event string, handler EventHandler) {
	ctx.component.On(event, handler)
}

// Emit sends a custom event with associated data.
// All registered handlers for this event will be called.
//
// Events can be used for internal component logic or to communicate
// with parent components.
//
// Example:
//
//	ctx.Emit("submit", FormData{
//	    Username: "user",
//	    Password: "pass",
//	})
func (ctx *Context) Emit(event string, data interface{}) {
	ctx.component.Emit(event, data)
}

// Props returns the component's props (configuration data).
// Props are immutable from the component's perspective and are
// passed down from parent components.
//
// The returned value should be type-asserted to the expected props type.
//
// Example:
//
//	props := ctx.Props().(ButtonProps)
//	label := props.Label
func (ctx *Context) Props() interface{} {
	return ctx.component.Props()
}

// Children returns the component's child components.
// This allows the setup function to access and interact with children,
// such as registering event handlers on them.
//
// Example:
//
//	children := ctx.Children()
//	for _, child := range children {
//	    child.On("click", func(data interface{}) {
//	        // Handle child click
//	    })
//	}
func (ctx *Context) Children() []Component {
	return ctx.component.children
}

// OnMounted registers a hook that executes after the component is mounted.
// The hook runs once, after the first render.
//
// If the component is already mounted when this is called, the hook
// executes immediately.
//
// Example:
//
//	ctx.OnMounted(func() {
//	    // Initialize data, start timers, etc.
//	    fmt.Println("Component mounted!")
//	})
func (ctx *Context) OnMounted(hook func()) {
	if ctx.component.lifecycle == nil {
		ctx.component.lifecycle = newLifecycleManager(ctx.component)
	}

	// Generate unique ID for this hook
	id := hookIDCounter.Add(1)

	// Get current number of hooks for order
	order := len(ctx.component.lifecycle.hooks["mounted"])

	// Register the hook
	ctx.component.lifecycle.registerHook("mounted", lifecycleHook{
		id:       fmt.Sprintf("hook-%d", id),
		callback: hook,
		order:    order,
	})
}

// OnUpdated registers a hook that executes after the component updates.
// If dependencies are provided, the hook only runs when those dependencies change.
// If no dependencies are provided, the hook runs on every update.
//
// Example:
//
//	count := ctx.Ref(0)
//	ctx.OnUpdated(func() {
//	    fmt.Printf("Count changed to: %d\n", count.Get())
//	}, count)
func (ctx *Context) OnUpdated(hook func(), deps ...*Ref[any]) {
	if ctx.component.lifecycle == nil {
		ctx.component.lifecycle = newLifecycleManager(ctx.component)
	}

	// Generate unique ID for this hook
	id := hookIDCounter.Add(1)

	// Get current number of hooks for order
	order := len(ctx.component.lifecycle.hooks["updated"])

	// Create hook with dependencies
	h := lifecycleHook{
		id:           fmt.Sprintf("hook-%d", id),
		callback:     hook,
		dependencies: deps,
		order:        order,
	}

	// Capture initial values of dependencies
	if len(deps) > 0 {
		h.lastValues = make([]any, len(deps))
		for i, dep := range deps {
			h.lastValues[i] = dep.Get()
		}
	}

	// Register the hook
	ctx.component.lifecycle.registerHook("updated", h)
}

// OnUnmounted registers a hook that executes when the component is unmounted.
// This is the place to perform cleanup operations.
//
// Example:
//
//	ctx.OnUnmounted(func() {
//	    // Cleanup resources
//	    fmt.Println("Component unmounting!")
//	})
func (ctx *Context) OnUnmounted(hook func()) {
	if ctx.component.lifecycle == nil {
		ctx.component.lifecycle = newLifecycleManager(ctx.component)
	}

	// Generate unique ID for this hook
	id := hookIDCounter.Add(1)

	// Get current number of hooks for order
	order := len(ctx.component.lifecycle.hooks["unmounted"])

	// Register the hook
	ctx.component.lifecycle.registerHook("unmounted", lifecycleHook{
		id:       fmt.Sprintf("hook-%d", id),
		callback: hook,
		order:    order,
	})
}

// OnBeforeUpdate registers a hook that executes before the component updates.
// This is optional and used for advanced use cases.
//
// Example:
//
//	ctx.OnBeforeUpdate(func() {
//	    fmt.Println("About to update...")
//	})
func (ctx *Context) OnBeforeUpdate(hook func()) {
	if ctx.component.lifecycle == nil {
		ctx.component.lifecycle = newLifecycleManager(ctx.component)
	}

	// Generate unique ID for this hook
	id := hookIDCounter.Add(1)

	// Get current number of hooks for order
	order := len(ctx.component.lifecycle.hooks["beforeUpdate"])

	// Register the hook
	ctx.component.lifecycle.registerHook("beforeUpdate", lifecycleHook{
		id:       fmt.Sprintf("hook-%d", id),
		callback: hook,
		order:    order,
	})
}

// OnBeforeUnmount registers a hook that executes before the component unmounts.
// This is optional and used for advanced use cases.
//
// Example:
//
//	ctx.OnBeforeUnmount(func() {
//	    fmt.Println("About to unmount...")
//	})
func (ctx *Context) OnBeforeUnmount(hook func()) {
	if ctx.component.lifecycle == nil {
		ctx.component.lifecycle = newLifecycleManager(ctx.component)
	}

	// Generate unique ID for this hook
	id := hookIDCounter.Add(1)

	// Get current number of hooks for order
	order := len(ctx.component.lifecycle.hooks["beforeUnmount"])

	// Register the hook
	ctx.component.lifecycle.registerHook("beforeUnmount", lifecycleHook{
		id:       fmt.Sprintf("hook-%d", id),
		callback: hook,
		order:    order,
	})
}

// OnCleanup registers a cleanup function that executes when the component unmounts.
// Cleanup functions are executed in reverse order (LIFO).
//
// Example:
//
//	ticker := time.NewTicker(time.Second)
//	ctx.OnCleanup(func() {
//	    ticker.Stop()
//	})
func (ctx *Context) OnCleanup(cleanup CleanupFunc) {
	if ctx.component.lifecycle == nil {
		ctx.component.lifecycle = newLifecycleManager(ctx.component)
	}

	// Register cleanup function
	ctx.component.lifecycle.cleanups = append(ctx.component.lifecycle.cleanups, cleanup)
}
