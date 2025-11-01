package directives

import "fmt"

// OnDirective implements declarative event handling for template elements.
//
// The On directive provides a way to attach event handlers to rendered content
// in a declarative manner. It wraps content with event markers that can be
// processed by the component system to register actual event handlers.
//
// # Basic Usage
//
//	On("click", func(data interface{}) {
//	    // Handle click event
//	}).Render("Click Me")
//	// Renders: [Event:click]Click Me
//
// # Event Types
//
// The On directive supports any event name string:
//   - "click": Mouse click events
//   - "keypress": Keyboard events
//   - "submit": Form submission
//   - "change": Input change events
//   - Custom event names for component-specific events
//
// # Handler Function
//
// The handler receives arbitrary data associated with the event:
//
//	On("click", func(data interface{}) {
//	    if clickData, ok := data.(*ClickData); ok {
//	        fmt.Printf("Clicked at: %d, %d\n", clickData.X, clickData.Y)
//	    }
//	})
//
// # Event Modifiers
//
// The On directive supports fluent modifier methods for controlling event behavior:
//
//	On("submit", handleSubmit).
//	    PreventDefault().
//	    StopPropagation().
//	    Render("Submit Form")
//	// Renders: [Event:submit:prevent:stop]Submit Form
//
// Available modifiers:
//   - PreventDefault(): Prevents the default action associated with the event
//   - StopPropagation(): Stops the event from bubbling up the component tree
//   - Once(): Handler executes only once, then is automatically removed
//
// # Integration with Component System
//
// In a real component, the On directive would integrate with the component's
// event system. The event markers in the rendered output would be processed
// to register handlers with the component:
//
//	Setup(func(ctx *Context) {
//	    ctx.On("click", func(data interface{}) {
//	        // This handler would be registered by the component system
//	        // when it processes the [Event:click] markers in the template
//	    })
//	})
//
//	Template(func(ctx RenderContext) string {
//	    return On("click", handleClick).Render("Button")
//	})
//
// # Multiple Events
//
// Multiple On directives can be chained on the same content:
//
//	content := "Button"
//	content = On("click", handleClick).Render(content)
//	content = On("hover", handleHover).Render(content)
//	// Renders: [Event:hover][Event:click]Button
//
// # Composition with Other Directives
//
// On directives compose naturally with other directives:
//
//	If(condition, func() string {
//	    return On("click", handler).Render("Conditional Button")
//	}).Render()
//
//	ForEach(items, func(item string, i int) string {
//	    return On("click", handler).Render(item)
//	}).Render()
//
// # Purity
//
// The directive is pure - it has no side effects and only wraps content with
// event markers. The actual event handler registration happens in the component
// system when processing the rendered output.
type OnDirective struct {
	event           string
	handler         func(interface{})
	preventDefault  bool
	stopPropagation bool
	once            bool
}

// On creates a new event handling directive for the given event name and handler.
//
// The On function is the entry point for creating event directives. It accepts
// an event name (e.g., "click", "submit", "keypress") and a handler function
// that will be called when the event occurs.
//
// Parameters:
//   - event: The event name to listen for (e.g., "click", "submit")
//   - handler: Function to execute when event occurs, receives event data
//
// Returns:
//   - *OnDirective: A new On directive that can be rendered
//
// Example:
//
//	On("click", func(data interface{}) {
//	    fmt.Println("Button clicked!")
//	}).Render("Click Me")
//	// Renders: [Event:click]Click Me
//
// Event Names:
// Any string can be used as an event name. Common conventions:
//   - "click": Mouse click
//   - "submit": Form submission
//   - "change": Input value change
//   - "keypress": Keyboard input
//   - "hover": Mouse hover
//   - Custom names for component-specific events
//
// Handler Signature:
// The handler receives interface{} data that should be type-asserted:
//
//	On("click", func(data interface{}) {
//	    if clickData, ok := data.(*ClickData); ok {
//	        // Use clickData
//	    }
//	})
//
// Nil Handler:
// Passing nil as the handler is allowed but not recommended. The directive
// will still render the event marker, but no handler will be registered.
//
// Empty Event Name:
// Empty event names are allowed but not recommended. They will render as
// [Event:] which may not be processed correctly by the component system.
func On(event string, handler func(interface{})) *OnDirective {
	return &OnDirective{
		event:   event,
		handler: handler,
	}
}

// PreventDefault sets the preventDefault modifier on the directive.
//
// This modifier indicates that the default action associated with the event
// should be prevented. In a TUI context, this might prevent default key
// handling or other built-in behaviors.
//
// Returns:
//   - *OnDirective: The same directive instance for method chaining
//
// Example:
//
//	On("submit", handleSubmit).PreventDefault().Render("Submit")
//	// Renders: [Event:submit:prevent]Submit
//
// The method is idempotent - calling it multiple times has the same effect
// as calling it once. It modifies the directive in place and returns the
// same instance to enable fluent chaining.
//
// Fluent API:
//
//	On("click", handler).
//	    PreventDefault().
//	    StopPropagation().
//	    Once().
//	    Render("Button")
func (d *OnDirective) PreventDefault() *OnDirective {
	d.preventDefault = true
	return d
}

// StopPropagation sets the stopPropagation modifier on the directive.
//
// This modifier indicates that the event should not bubble up to parent
// components. This is useful when you want to handle an event locally
// without triggering parent handlers.
//
// Returns:
//   - *OnDirective: The same directive instance for method chaining
//
// Example:
//
//	On("click", handleClick).StopPropagation().Render("Button")
//	// Renders: [Event:click:stop]Button
//
// The method is idempotent - calling it multiple times has the same effect
// as calling it once. It modifies the directive in place and returns the
// same instance to enable fluent chaining.
//
// Use Cases:
//   - Preventing parent click handlers from firing
//   - Isolating event handling to a specific component
//   - Implementing modal dialogs that don't propagate clicks
//
// Fluent API:
//
//	On("submit", handler).
//	    PreventDefault().
//	    StopPropagation().
//	    Render("Form")
func (d *OnDirective) StopPropagation() *OnDirective {
	d.stopPropagation = true
	return d
}

// Once sets the once modifier on the directive.
//
// This modifier indicates that the event handler should only execute once,
// then be automatically removed. This is useful for one-time actions like
// initialization, welcome screens, or confirmation dialogs.
//
// Returns:
//   - *OnDirective: The same directive instance for method chaining
//
// Example:
//
//	On("click", showWelcome).Once().Render("Show Welcome")
//	// Renders: [Event:click:once]Show Welcome
//	// Handler will only fire on first click, then be removed
//
// The method is idempotent - calling it multiple times has the same effect
// as calling it once. It modifies the directive in place and returns the
// same instance to enable fluent chaining.
//
// Use Cases:
//   - One-time initialization handlers
//   - Welcome screens or tutorials
//   - Confirmation dialogs
//   - Resource cleanup on first interaction
//
// Cleanup:
// The component system will automatically remove the handler after it
// executes once, preventing memory leaks and ensuring the handler doesn't
// fire multiple times.
//
// Fluent API:
//
//	On("init", initialize).Once().Render("Initialize")
func (d *OnDirective) Once() *OnDirective {
	d.once = true
	return d
}

// Render executes the directive logic and returns content wrapped with event marker.
//
// This method wraps the provided content with an event marker that indicates
// an event handler should be registered for this content. The marker format
// includes the event name and any active modifiers.
//
// Parameters:
//   - content: The content to wrap with the event marker
//
// Returns:
//   - string: Content wrapped with event marker
//
// Marker Format:
//   - No modifiers: [Event:eventName]content
//   - With modifiers: [Event:eventName:modifier1:modifier2]content
//
// Modifier Markers:
//   - preventDefault: "prevent"
//   - stopPropagation: "stop"
//   - once: "once"
//
// Behavior:
//  1. Build event marker starting with [Event:eventName]
//  2. Append modifier markers in consistent order (prevent, stop, once)
//  3. Prepend complete marker to content
//  4. Return combined string
//
// Example:
//
//	directive := On("click", handleClick)
//	output := directive.Render("Click Me")
//	// output: "[Event:click]Click Me"
//
// With Modifiers:
//
//	On("submit", handleSubmit).
//	    PreventDefault().
//	    StopPropagation().
//	    Render("Submit")
//	// output: "[Event:submit:prevent:stop]Submit"
//
// Multiple Events:
// When multiple On directives are applied, markers stack:
//
//	content := "Button"
//	content = On("click", h1).Render(content)
//	content = On("hover", h2).Render(content)
//	// Result: "[Event:hover][Event:click]Button"
//
// Empty Content:
// Empty content is allowed and will render just the marker:
//
//	On("click", handler).Render("")
//	// Result: "[Event:click]"
//
// The method is pure and idempotent - calling it multiple times with the same
// content produces the same result. It does not modify the handler or any state.
//
// Modifier Order:
// Modifiers always appear in the same order (prevent, stop, once) regardless
// of the order they were called. This ensures consistent marker format for
// parsing by the component system.
func (d *OnDirective) Render(content string) string {
	// Start with base event marker
	marker := fmt.Sprintf("[Event:%s", d.event)

	// Append modifiers in consistent order
	if d.preventDefault {
		marker += ":prevent"
	}
	if d.stopPropagation {
		marker += ":stop"
	}
	if d.once {
		marker += ":once"
	}

	// Close marker
	marker += "]"

	// Wrap content with marker
	return marker + content
}
