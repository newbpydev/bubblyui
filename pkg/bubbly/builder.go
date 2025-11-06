package bubbly

import (
	"errors"
	"fmt"
	"strings"
)

// Validation errors returned by Build().
var (
	// ErrMissingTemplate is returned when Build() is called without setting a template.
	ErrMissingTemplate = errors.New("template is required")
)

// ValidationError represents one or more validation errors encountered during Build().
// It provides detailed information about what went wrong during component validation.
type ValidationError struct {
	// ComponentName is the name of the component that failed validation.
	ComponentName string

	// Errors is the list of validation errors encountered.
	Errors []error
}

// Error implements the error interface for ValidationError.
// It formats the error message to include the component name and all validation errors.
func (e *ValidationError) Error() string {
	var errMsgs []string
	for _, err := range e.Errors {
		errMsgs = append(errMsgs, err.Error())
	}
	return fmt.Sprintf("component '%s' validation failed: %s",
		e.ComponentName,
		strings.Join(errMsgs, "; "))
}

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

	// autoCommands indicates whether automatic command generation is enabled.
	// When true, Build() will initialize the command queue and generator.
	autoCommands bool

	// debugCommands indicates whether command debug logging is enabled.
	// When true, Build() will initialize a command logger.
	debugCommands bool
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

// Props sets the component's props (configuration data).
// Props are immutable from the component's perspective and are
// passed down from parent components.
//
// The props parameter accepts any type, allowing for flexible
// component configuration. Common patterns include:
//   - Struct types for structured props
//   - Primitive types for simple configuration
//   - Maps for dynamic key-value pairs
//
// Example:
//
//	type ButtonProps struct {
//	    Label    string
//	    Disabled bool
//	}
//
//	builder := NewComponent("Button").
//	    Props(ButtonProps{
//	        Label:    "Click me",
//	        Disabled: false,
//	    })
//
// Parameters:
//   - props: The props value (can be any type)
//
// Returns:
//   - *ComponentBuilder: The builder for method chaining
func (b *ComponentBuilder) Props(props interface{}) *ComponentBuilder {
	b.component.props = props
	return b
}

// Setup sets the component's setup function.
// The setup function is called once during component initialization (Init phase)
// and is where you should:
//   - Create reactive state using ctx.Ref() and ctx.Computed()
//   - Register event handlers using ctx.On()
//   - Set up watchers using ctx.Watch()
//   - Expose state to the template using ctx.Expose()
//
// Example:
//
//	builder := NewComponent("Counter").
//	    Setup(func(ctx *Context) {
//	        count := ctx.Ref(0)
//	        ctx.Expose("count", count)
//	        ctx.On("increment", func(data interface{}) {
//	            count.Set(count.GetTyped() + 1)
//	        })
//	    })
//
// Parameters:
//   - fn: The setup function (SetupFunc type)
//
// Returns:
//   - *ComponentBuilder: The builder for method chaining
func (b *ComponentBuilder) Setup(fn SetupFunc) *ComponentBuilder {
	b.component.setup = fn
	return b
}

// Template sets the component's template function.
// The template function generates the component's visual output and is called
// on every View() cycle.
//
// The template function should:
//   - Access state using ctx.GetTyped()
//   - Access props using ctx.Props()
//   - Render children using ctx.RenderChild()
//   - Use Lipgloss for styling
//   - Return a string representing the UI
//   - Be pure (no side effects, same input produces same output)
//
// Example:
//
//	builder := NewComponent("Button").
//	    Template(func(ctx RenderContext) string {
//	        props := ctx.Props().(ButtonProps)
//	        style := lipgloss.NewStyle().Bold(true)
//	        return style.Render(props.Label)
//	    })
//
// Parameters:
//   - fn: The template function (RenderFunc type)
//
// Returns:
//   - *ComponentBuilder: The builder for method chaining
func (b *ComponentBuilder) Template(fn RenderFunc) *ComponentBuilder {
	b.component.template = fn
	return b
}

// Children sets the component's child components.
// Children are nested components that are managed by the parent component's
// lifecycle and can be rendered within the parent's template.
//
// The method accepts a variadic parameter, allowing you to pass:
//   - No children (empty component)
//   - A single child
//   - Multiple children
//
// Example:
//
//	child1 := NewComponent("Child1").Template(...).Build()
//	child2 := NewComponent("Child2").Template(...).Build()
//
//	parent := NewComponent("Parent").
//	    Children(child1, child2).
//	    Template(func(ctx RenderContext) string {
//	        // Render children
//	        outputs := []string{}
//	        for _, child := range ctx.Children() {
//	            outputs = append(outputs, ctx.RenderChild(child))
//	        }
//	        return strings.Join(outputs, "\n")
//	    })
//
// Parameters:
//   - children: Variadic Component parameters
//
// Returns:
//   - *ComponentBuilder: The builder for method chaining
func (b *ComponentBuilder) Children(children ...Component) *ComponentBuilder {
	b.component.children = children

	// Set parent reference for each child
	for _, child := range children {
		if childImpl, ok := child.(*componentImpl); ok {
			childImpl.parent = b.component
		}
	}

	return b
}

// WithAutoCommands enables or disables automatic command generation for the component.
// When enabled, the component will automatically generate Bubbletea commands when
// reactive state changes (via Ref.Set()), eliminating the need for manual Emit() calls.
//
// Automatic command generation provides a Vue-like developer experience where state
// changes trigger UI updates automatically. The component's Build() method will
// initialize the command queue and default command generator when this is enabled.
//
// Example with automatic commands:
//
//	component := NewComponent("Counter").
//	    WithAutoCommands(true).
//	    Setup(func(ctx *Context) {
//	        count := ctx.Ref(0)
//	        ctx.On("increment", func(_ interface{}) {
//	            count.Set(count.Get().(int) + 1)
//	            // UI updates automatically - no manual Emit() needed!
//	        })
//	    }).
//	    Template(func(ctx RenderContext) string {
//	        return fmt.Sprintf("Count: %d", ctx.Get("count"))
//	    }).
//	    Build()
//
// Example with manual control (default):
//
//	component := NewComponent("Counter").
//	    WithAutoCommands(false).  // or omit this line
//	    Setup(func(ctx *Context) {
//	        count := ctx.Ref(0)
//	        ctx.On("increment", func(_ interface{}) {
//	            count.Set(count.Get().(int) + 1)
//	            ctx.Emit("update", nil)  // Manual emit required
//	        })
//	    }).
//	    Build()
//
// Parameters:
//   - enabled: true to enable automatic commands, false to disable (default: false)
//
// Returns:
//   - *ComponentBuilder: The builder for method chaining
func (b *ComponentBuilder) WithAutoCommands(enabled bool) *ComponentBuilder {
	b.autoCommands = enabled
	return b
}

// WithCommandDebug enables or disables debug logging for command generation.
//
// When enabled, all command generation events are logged with detailed information:
//   - Component name and ID
//   - Ref ID
//   - Old and new values
//   - Timestamp
//
// This is extremely useful for:
//   - Understanding reactive update flow
//   - Debugging infinite loop issues
//   - Troubleshooting unexpected UI updates
//   - Performance profiling
//
// When disabled (default), there is zero overhead (no logging calls, no allocations).
//
// Example:
//
//	component := NewComponent("Counter").
//	    WithAutoCommands(true).
//	    WithCommandDebug(true). // Enable debug logging
//	    Setup(...).Build()
//
//	// Logs will show:
//	// [DEBUG] Command Generated | Component: Counter (component-1) | Ref: ref-5 | 0 → 1
func (b *ComponentBuilder) WithCommandDebug(enabled bool) *ComponentBuilder {
	b.debugCommands = enabled
	return b
}

// WithKeyBinding registers a simple key-to-event binding.
// This is a convenience method for the most common case: mapping a key to an event
// with a description for help text generation.
//
// The binding will always be active (no condition). For conditional bindings,
// use WithConditionalKeyBinding instead.
//
// Multiple bindings can be registered for the same key. When a key is pressed,
// the first matching binding (based on condition evaluation) will be used.
//
// Example:
//
//	component := NewComponent("Counter").
//	    WithKeyBinding("space", "increment", "Increment counter").
//	    WithKeyBinding("ctrl+c", "quit", "Quit application").
//	    WithKeyBinding("up", "selectPrevious", "Previous item").
//	    Setup(func(ctx *Context) {
//	        count := ctx.Ref(0)
//	        ctx.On("increment", func(_ interface{}) {
//	            count.Set(count.Get().(int) + 1)
//	        })
//	    }).
//	    Build()
//
// Parameters:
//   - key: The keyboard key (e.g., "space", "ctrl+c", "up")
//   - event: The event name to emit when key is pressed
//   - description: Human-readable description for help text
//
// Returns:
//   - *ComponentBuilder: The builder for method chaining
func (b *ComponentBuilder) WithKeyBinding(key, event, description string) *ComponentBuilder {
	binding := KeyBinding{
		Key:         key,
		Event:       event,
		Description: description,
	}
	return b.WithConditionalKeyBinding(binding)
}

// WithConditionalKeyBinding registers a key binding with optional condition and data.
// This is the full-featured method that supports all KeyBinding fields including
// conditional activation and custom data.
//
// Use this method when you need:
//   - Mode-based input (different behavior based on application state)
//   - Same key doing different things in different modes
//   - Passing custom data to event handlers
//
// Example with mode-based input:
//
//	inputMode := false
//	component := NewComponent("Form").
//	    WithConditionalKeyBinding(KeyBinding{
//	        Key:         "space",
//	        Event:       "toggle",
//	        Description: "Toggle in navigation mode",
//	        Condition:   func() bool { return !inputMode },
//	    }).
//	    WithConditionalKeyBinding(KeyBinding{
//	        Key:         "space",
//	        Event:       "addChar",
//	        Description: "Add space in input mode",
//	        Data:        " ",
//	        Condition:   func() bool { return inputMode },
//	    }).
//	    Build()
//
// Parameters:
//   - binding: The KeyBinding configuration
//
// Returns:
//   - *ComponentBuilder: The builder for method chaining
func (b *ComponentBuilder) WithConditionalKeyBinding(binding KeyBinding) *ComponentBuilder {
	// Initialize keyBindings map in component if needed
	if b.component.keyBindings == nil {
		b.component.keyBindings = make(map[string][]KeyBinding)
	}

	// Append binding to the list for this key
	// Multiple bindings per key are allowed (e.g., mode-based)
	b.component.keyBindings[binding.Key] = append(
		b.component.keyBindings[binding.Key],
		binding,
	)

	return b
}

// WithKeyBindings registers multiple key bindings at once.
// This is a convenience method for batch registration when you have
// a predefined set of key bindings.
//
// The map keys are the keyboard keys, and the values are KeyBinding structs.
// Note that the KeyBinding.Key field should match the map key.
//
// Example:
//
//	bindings := map[string]KeyBinding{
//	    "space": {
//	        Key:         "space",
//	        Event:       "increment",
//	        Description: "Increment counter",
//	    },
//	    "ctrl+c": {
//	        Key:         "ctrl+c",
//	        Event:       "quit",
//	        Description: "Quit application",
//	    },
//	    "up": {
//	        Key:         "up",
//	        Event:       "selectPrevious",
//	        Description: "Previous item",
//	    },
//	}
//
//	component := NewComponent("Counter").
//	    WithKeyBindings(bindings).
//	    Build()
//
// Parameters:
//   - bindings: Map of key to KeyBinding
//
// Returns:
//   - *ComponentBuilder: The builder for method chaining
func (b *ComponentBuilder) WithKeyBindings(bindings map[string]KeyBinding) *ComponentBuilder {
	// Initialize keyBindings map even if bindings is nil
	// This ensures the map is always initialized for consistency
	if b.component.keyBindings == nil {
		b.component.keyBindings = make(map[string][]KeyBinding)
	}

	if bindings == nil {
		return b
	}

	for _, binding := range bindings {
		b.WithConditionalKeyBinding(binding)
	}

	return b
}

// WithMessageHandler registers a custom message handler for complex message processing.
// The message handler provides an escape hatch for scenarios that declarative key bindings
// cannot handle, such as:
//   - Custom Bubbletea message types (e.g., data updates, timers)
//   - Window resize events (tea.WindowSizeMsg)
//   - Mouse events (tea.MouseMsg)
//   - Complex conditional logic
//   - Dynamic message routing
//
// The handler is called BEFORE key binding processing in the component's Update() method,
// allowing it to intercept and handle any message type. Commands returned by the handler
// are automatically batched with other commands from key bindings and state changes.
//
// Handler characteristics:
//   - Receives the component and raw Bubbletea message
//   - Can emit events to the component via comp.Emit()
//   - Can return a tea.Cmd (or nil for no command)
//   - Coexists with key bindings (both can be used together)
//   - Called on every Update() cycle before other processing
//
// Example with custom messages and window resize:
//
//	type DataUpdateMsg struct {
//	    Data []Item
//	}
//
//	component := NewComponent("Dashboard").
//	    WithAutoCommands(true).
//	    WithKeyBinding("r", "refresh", "Refresh data").
//	    WithMessageHandler(func(comp Component, msg tea.Msg) tea.Cmd {
//	        switch msg := msg.(type) {
//	        case DataUpdateMsg:
//	            // Handle custom data update message
//	            comp.Emit("dataReceived", msg.Data)
//	            return nil
//
//	        case tea.WindowSizeMsg:
//	            // Handle window resize
//	            comp.Emit("resize", map[string]int{
//	                "width":  msg.Width,
//	                "height": msg.Height,
//	            })
//	            return nil
//
//	        case tea.MouseMsg:
//	            // Handle mouse click
//	            if msg.Type == tea.MouseLeft {
//	                comp.Emit("click", msg)
//	            }
//	            return nil
//	        }
//	        return nil // Let other processing continue
//	    }).
//	    Setup(func(ctx *Context) {
//	        // Handle semantic events from both key bindings and message handler
//	        ctx.On("dataReceived", func(data interface{}) {
//	            // Update state...
//	        })
//	    }).
//	    Build()
//
// When to use:
//   - Use key bindings for simple key-to-event mapping (covers 90% of cases)
//   - Use message handler for complex message types or dynamic routing
//   - Both can coexist: key bindings for keyboard, handler for everything else
//
// Parameters:
//   - handler: The MessageHandler function
//
// Returns:
//   - *ComponentBuilder: The builder for method chaining
func (b *ComponentBuilder) WithMessageHandler(handler MessageHandler) *ComponentBuilder {
	b.component.messageHandler = handler
	return b
}

// Build validates the component configuration and returns the final Component.
// This is the terminal method in the builder chain that performs validation
// and creates the component instance.
//
// Validation rules:
//   - Template is required (components must have a render function)
//   - All accumulated errors are checked and reported
//
// If validation fails, Build returns nil and an error describing what's wrong.
// If validation succeeds, Build returns the configured Component ready for use.
//
// When WithAutoCommands(true) was called, Build will also:
//   - Initialize the command queue for pending commands
//   - Set up the default command generator
//   - Enable automatic command generation for reactive state changes
//
// Example:
//
//	component, err := NewComponent("Button").
//	    Props(ButtonProps{Label: "Click me"}).
//	    Template(func(ctx RenderContext) string {
//	        props := ctx.Props().(ButtonProps)
//	        return props.Label
//	    }).
//	    Build()
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Use component with Bubbletea
//	p := tea.NewProgram(component)
//	p.Run()
//
// Returns:
//   - Component: The built component (nil if validation fails)
//   - error: Validation error (nil if validation succeeds)
func (b *ComponentBuilder) Build() (Component, error) {
	// Validate required fields
	if b.component.template == nil {
		b.errors = append(b.errors, ErrMissingTemplate)
	}

	// Check for accumulated errors
	if len(b.errors) > 0 {
		return nil, &ValidationError{
			ComponentName: b.component.name,
			Errors:        b.errors,
		}
	}

	// Initialize command infrastructure if automatic commands enabled
	if b.autoCommands {
		b.component.autoCommands = true
		b.component.commandQueue = NewCommandQueue()
		b.component.commandGen = &defaultCommandGenerator{}
	}

	// Initialize command debug logger if debugging enabled
	if b.debugCommands {
		// Use stdout for debug logging
		// Users can redirect stdout or set custom logger via component field
		b.component.commandLogger = newCommandLogger(nil) // nil uses os.Stdout via log package default
	} else {
		// Use no-op logger for zero overhead
		b.component.commandLogger = newNopCommandLogger()
	}

	// Return the component (implements Component interface)
	return b.component, nil
}
