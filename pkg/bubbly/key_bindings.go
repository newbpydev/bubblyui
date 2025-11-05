package bubbly

// KeyBinding represents a declarative key-to-event mapping.
// It allows components to register keyboard shortcuts that automatically
// emit events when the corresponding keys are pressed.
//
// Key bindings provide a declarative way to handle keyboard input without
// manually processing tea.KeyMsg in the Update() method. This is especially
// useful for:
//   - Creating consistent keyboard shortcuts across components
//   - Auto-generating help text from key descriptions
//   - Supporting mode-based input (navigation vs typing)
//   - Reducing boilerplate in component code
//
// Example:
//
//	component := NewComponent("Counter").
//	    WithKeyBinding("space", "increment", "Increment counter").
//	    WithKeyBinding("ctrl+c", "quit", "Quit application").
//	    Setup(func(ctx *Context) {
//	        count := ctx.Ref(0)
//	        ctx.On("increment", func(_ interface{}) {
//	            count.Set(count.Get().(int) + 1)
//	        })
//	    }).
//	    Build()
//
// Conditional bindings for mode-based input:
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
type KeyBinding struct {
	// Key is the keyboard key or key combination (e.g., "space", "ctrl+c", "up").
	// The format follows Bubbletea's tea.KeyMsg.String() convention:
	//   - Single keys: "a", "space", "enter", "esc"
	//   - Arrow keys: "up", "down", "left", "right"
	//   - Ctrl combinations: "ctrl+c", "ctrl+d"
	//   - Alt combinations: "alt+a", "alt+enter"
	//   - Special keys: "tab", "backspace", "delete"
	Key string

	// Event is the name of the event to emit when the key is pressed.
	// This should match an event handler registered via ctx.On() in Setup().
	//
	// Special event names:
	//   - "quit": Automatically returns tea.Quit to exit the application
	Event string

	// Description is a human-readable description of what the key does.
	// This is used for auto-generating help text and documentation.
	// Example: "Increment counter", "Save and exit", "Toggle selection"
	Description string

	// Data is optional data to pass to the event handler.
	// This allows the same event to be triggered with different data
	// based on which key was pressed.
	//
	// Example:
	//   WithKeyBinding("1", "selectTab", "Tab 1").Data = 0
	//   WithKeyBinding("2", "selectTab", "Tab 2").Data = 1
	Data interface{}

	// Condition is an optional function that determines if the binding is active.
	// If nil, the binding is always active.
	// If not nil, the binding only triggers when Condition() returns true.
	//
	// This is useful for mode-based input where the same key does different
	// things depending on application state (e.g., navigation vs input mode).
	//
	// Example:
	//   Condition: func() bool { return !inputMode }
	Condition func() bool
}
