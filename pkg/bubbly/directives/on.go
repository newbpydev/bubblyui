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
//
// # Future Enhancements
//
// Task 4.2 will add event modifiers:
//   - PreventDefault(): Prevent default browser behavior
//   - StopPropagation(): Stop event bubbling
//   - Once(): Handler executes only once
type OnDirective struct {
	event   string
	handler func(interface{})
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

// Render executes the directive logic and returns content wrapped with event marker.
//
// This method wraps the provided content with an event marker that indicates
// an event handler should be registered for this content. The marker format
// is [Event:eventName]content.
//
// Parameters:
//   - content: The content to wrap with the event marker
//
// Returns:
//   - string: Content wrapped with event marker
//
// Behavior:
//  1. Create event marker with format [Event:eventName]
//  2. Prepend marker to content
//  3. Return combined string
//
// Example:
//
//	directive := On("click", handleClick)
//	output := directive.Render("Click Me")
//	// output: "[Event:click]Click Me"
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
// Future Enhancement:
// In Task 4.2, this will be enhanced to:
//   - Include modifier flags in the marker
//   - Support preventDefault, stopPropagation, once modifiers
//   - Integrate with component event system for actual handler registration
func (d *OnDirective) Render(content string) string {
	// Create event marker with format [Event:eventName]
	marker := fmt.Sprintf("[Event:%s]", d.event)

	// Wrap content with marker
	return marker + content
}
