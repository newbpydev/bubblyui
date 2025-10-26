package bubbly

// ComponentBuilder provides a fluent API for creating components.
// It implements the builder pattern to make component creation
// readable and type-safe.
//
// The builder:
//   - Stores a reference to the component being built
//   - Tracks validation errors during configuration
//   - Provides chainable methods for setting component properties
//   - Validates configuration before building the final component
//
// Example:
//
//	component := NewComponent("Button").
//	    Props(ButtonProps{Label: "Click me"}).
//	    Setup(func(ctx *Context) {
//	        // Initialize state
//	    }).
//	    Template(func(ctx RenderContext) string {
//	        return "Hello"
//	    }).
//	    Build()
type ComponentBuilder struct {
	// component is the component being built.
	// It's created immediately in NewComponent() and configured
	// through the builder methods.
	component *componentImpl

	// errors tracks validation errors encountered during configuration.
	// Errors are accumulated and checked in Build().
	errors []error
}

// NewComponent creates a new ComponentBuilder for building a component.
// This is the entry point for creating components using the fluent API.
//
// The function:
//   - Creates a new component instance with the given name
//   - Initializes the builder with empty error tracking
//   - Returns the builder ready for method chaining
//
// Example:
//
//	builder := NewComponent("Button")
//	// Now chain configuration methods...
//	builder.Props(...).Setup(...).Template(...).Build()
//
// Parameters:
//   - name: The component name (e.g., "Button", "Counter", "Form")
//
// Returns:
//   - *ComponentBuilder: A builder instance ready for configuration
func NewComponent(name string) *ComponentBuilder {
	return &ComponentBuilder{
		component: newComponentImpl(name),
		errors:    []error{},
	}
}
