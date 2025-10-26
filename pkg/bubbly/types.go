package bubbly

// SetupFunc is a function type that initializes component state and behavior.
// It receives a Context that provides access to reactive primitives (Ref, Computed, Watch),
// event handling (On, Emit), and component data (Props, Children).
//
// The setup function is called once during component initialization (Init phase)
// and is where you should:
//   - Create reactive state using ctx.Ref() and ctx.Computed()
//   - Register event handlers using ctx.On()
//   - Set up watchers using ctx.Watch()
//   - Expose state to the template using ctx.Expose()
//
// Example:
//
//	Setup(func(ctx *Context) {
//	    count := ctx.Ref(0)
//	    ctx.Expose("count", count)
//	    ctx.On("increment", func(data interface{}) {
//	        count.Set(count.Get() + 1)
//	    })
//	})
type SetupFunc func(ctx *Context)

// RenderFunc is a function type that generates the component's visual output.
// It receives a RenderContext that provides read-only access to component state,
// props, and children.
//
// The render function should:
//   - Access state using ctx.Get()
//   - Access props using ctx.Props()
//   - Render children using ctx.RenderChild()
//   - Use Lipgloss for styling
//   - Return a string representing the UI
//
// The render function is called on every View() cycle and should be pure
// (no side effects, same input produces same output).
//
// Example:
//
//	Template(func(ctx RenderContext) string {
//	    count := ctx.Get("count").(*Ref[int])
//	    return fmt.Sprintf("Count: %d", count.Get())
//	})
type RenderFunc func(ctx RenderContext) string

// EventHandler is a function type that handles component events.
// It receives arbitrary data associated with the event.
//
// Event handlers are registered using ctx.On() in the setup function
// and are triggered when events are emitted using ctx.Emit() or c.Emit().
//
// The data parameter type depends on what the event emitter provides.
// Handlers should perform type assertions to access the data safely.
//
// Example:
//
//	ctx.On("submit", func(data interface{}) {
//	    if formData, ok := data.(FormData); ok {
//	        // Handle form submission
//	    }
//	})
type EventHandler func(data interface{})
