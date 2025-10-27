package bubbly

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
