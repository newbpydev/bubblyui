package bubbly

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// componentIDCounter is an atomic counter for generating unique component IDs.
// It's incremented each time a new component is created.
var componentIDCounter atomic.Uint64

// Component is the core interface for BubblyUI components.
// It extends Bubbletea's tea.Model interface with additional methods
// for component identification, props management, and event handling.
//
// Components encapsulate:
//   - State: Reactive values (Refs, Computed) managed internally
//   - Props: Immutable configuration passed from parent
//   - Events: Custom events emitted to parent components
//   - Template: Rendering logic that produces UI strings
//   - Children: Nested child components
//
// Components integrate seamlessly with Bubbletea's Elm architecture:
//   - Init(): Initialize component and run setup function
//   - Update(msg): Handle Bubbletea messages and trigger event handlers
//   - View(): Render component using template function
//
// Example:
//
//	type ButtonProps struct {
//	    Label string
//	}
//
//	button := NewComponent("Button").
//	    Props(ButtonProps{Label: "Click me"}).
//	    Template(func(ctx RenderContext) string {
//	        props := ctx.Props().(ButtonProps)
//	        return props.Label
//	    }).
//	    Build()
type Component interface {
	tea.Model

	// Name returns the component's name (e.g., "Button", "Counter").
	// This is primarily used for debugging and component identification.
	Name() string

	// ID returns the component's unique instance identifier.
	// Each component instance gets a unique ID generated at creation time.
	// This is useful for tracking components in a tree and debugging.
	ID() string

	// Props returns the component's props (configuration data).
	// Props are immutable from the component's perspective and are
	// passed down from parent components.
	//
	// The returned value should be type-asserted to the expected props type:
	//
	//	props := component.Props().(ButtonProps)
	Props() interface{}

	// Emit sends a custom event with associated data.
	// Events bubble up to parent components that have registered handlers.
	//
	// Example:
	//
	//	component.Emit("submit", FormData{...})
	Emit(event string, data interface{})

	// On registers an event handler for the specified event name.
	// Multiple handlers can be registered for the same event.
	//
	// Example:
	//
	//	component.On("click", func(data interface{}) {
	//	    // Handle click event
	//	})
	On(event string, handler EventHandler)

	// KeyBindings returns the component's registered key bindings.
	// This is primarily used for:
	//   - Auto-generating help text
	//   - Debugging key binding configuration
	//   - Introspection by parent components
	//
	// The returned map is keyed by the keyboard key string (e.g., "space", "ctrl+c")
	// and contains a slice of KeyBinding structs for that key. Multiple bindings
	// per key are supported for conditional/mode-based input.
	//
	// Example:
	//
	//	bindings := component.KeyBindings()
	//	for key, bindingList := range bindings {
	//	    for _, binding := range bindingList {
	//	        fmt.Printf("%s: %s\n", key, binding.Description)
	//	    }
	//	}
	KeyBindings() map[string][]KeyBinding

	// HelpText generates a formatted help string from registered key bindings.
	// It automatically creates user-facing documentation by extracting key
	// descriptions and formatting them in a consistent way.
	//
	// The generated format is: "key: description • key: description"
	// - Keys are sorted alphabetically for consistency
	// - Duplicate keys show only the first description
	// - Empty descriptions are skipped
	// - Separator is " • " (bullet point with spaces)
	//
	// This eliminates the need to manually maintain help text that can
	// get out of sync with actual key bindings.
	//
	// Example:
	//
	//	component := NewComponent("Counter").
	//	    WithKeyBinding("space", "increment", "Increment counter").
	//	    WithKeyBinding("ctrl+c", "quit", "Quit application").
	//	    Build()
	//
	//	helpText := component.HelpText()
	//	// Returns: "ctrl+c: Quit application • space: Increment counter"
	//
	// Integration with templates:
	//
	//	Template(func(ctx RenderContext) string {
	//	    comp := ctx.Component()
	//	    return fmt.Sprintf("%s\n\nHelp: %s", content, comp.HelpText())
	//	})
	HelpText() string

	// IsInitialized returns whether the component has been initialized.
	// A component is considered initialized after Init() has been called
	// and the setup function has executed.
	//
	// This method is thread-safe and can be called concurrently.
	//
	// Use cases:
	//   - Auto-initialization in Context.ExposeComponent() (Task 10.2)
	//   - Checking component state before operations
	//   - Debugging component lifecycle
	//
	// Example:
	//
	//	component := NewComponent("Counter").Build()
	//	fmt.Println(component.IsInitialized()) // false
	//
	//	component.Init()
	//	fmt.Println(component.IsInitialized()) // true
	//
	//	component.Init() // Safe to call again - idempotent
	//	fmt.Println(component.IsInitialized()) // still true
	IsInitialized() bool
}

// MessageHandler is a function that handles Bubbletea messages before key bindings.
// It provides an escape hatch for complex message handling scenarios that declarative
// key bindings cannot cover, such as:
//   - Custom Bubbletea message types
//   - Window resize events (tea.WindowSizeMsg)
//   - Mouse events (tea.MouseMsg)
//   - Complex conditional logic
//   - Dynamic key interpretation
//
// The handler receives:
//   - comp: The Component instance (can call Emit(), access state, etc.)
//   - msg: The raw Bubbletea message
//
// The handler can:
//   - Emit events to the component via comp.Emit()
//   - Return a tea.Cmd to be executed (or nil for no command)
//   - Access component state and props
//
// The handler is called BEFORE key binding processing in Update(), allowing
// it to handle messages that key bindings don't support. Commands returned by
// the handler are automatically batched with other commands.
//
// Example with custom message type:
//
//	type CustomDataMsg struct {
//	    Data string
//	}
//
//	component := NewComponent("Dashboard").
//	    WithMessageHandler(func(comp Component, msg tea.Msg) tea.Cmd {
//	        switch msg := msg.(type) {
//	        case CustomDataMsg:
//	            comp.Emit("dataReceived", msg.Data)
//	            return nil
//	        case tea.WindowSizeMsg:
//	            comp.Emit("resize", msg)
//	            return nil
//	        }
//	        return nil
//	    }).
//	    Build()
//
// The handler coexists with key bindings and other message processing:
//   - Handler called first (can intercept any message)
//   - Then key bindings processed
//   - Then state change messages processed
//   - All commands batched automatically
type MessageHandler func(comp Component, msg tea.Msg) tea.Cmd

// componentImpl is the internal implementation of the Component interface.
// It is unexported to enforce the use of the ComponentBuilder for creation.
//
// The struct holds all component state including:
//   - Identification (name, id)
//   - Configuration (props, setup, template)
//   - State (internal state map for exposed values)
//   - Relationships (parent, children)
//   - Event system (handlers map)
//   - Lifecycle (mounted flag)
//
// Note: Some fields are currently unused as they will be implemented in later tasks:
//   - setup: Task 1.3 (Bubbletea Model Implementation)
//   - state: Task 3.1 (Setup Context Implementation)
//   - parent, children: Task 5.1 (Children Management)
//   - mounted: Task 1.3 (Bubbletea Model Implementation)
type componentImpl struct {
	// Identification
	name string // Component name (e.g., "Button")
	id   string // Unique instance ID

	// Configuration
	props interface{} // Props passed from parent
	//nolint:unused // Will be used in Task 1.3
	setup    SetupFunc  // Setup function (runs once on Init)
	template RenderFunc // Template function (runs on every View)

	// State
	//nolint:unused // Will be used in Task 3.1
	state   map[string]interface{} // Exposed state (Refs, Computed, etc.)
	stateMu sync.RWMutex           // Protects state map for concurrent access

	// Relationships
	//nolint:unused // Will be used in Task 5.1
	parent *componentImpl // Parent component (for inject tree traversal and event bubbling)
	//nolint:unused // Will be used in Task 5.1
	children   []Component  // Child components
	childrenMu sync.RWMutex // Protects children slice

	// Provide/Inject (Composition API)
	provides      map[string]interface{} // Provided values for dependency injection
	providesMu    sync.RWMutex           // Protects provides map
	injectCache   map[string]interface{} // Cache for inject lookups (O(1) after first lookup)
	injectCacheMu sync.RWMutex           // Protects inject cache

	// Event system
	handlersMu sync.RWMutex              // Protects handlers map
	handlers   map[string][]EventHandler // Event name -> handlers

	// Command generation (Automatic Reactive Bridge - Feature 08)
	commandQueue   *CommandQueue    // Queue for pending commands from state changes
	commandGen     CommandGenerator // Generator for creating commands from state changes
	autoCommands   bool             // Whether automatic command generation is enabled
	autoCommandsMu sync.RWMutex     // Protects autoCommands and commandGen fields

	// Template context tracking (for safety checks)
	inTemplate   bool         // Whether currently executing inside template function
	inTemplateMu sync.RWMutex // Protects inTemplate flag

	// Loop detection (Automatic Reactive Bridge - Feature 08)
	loopDetector *loopDetector // Detects infinite command generation loops

	// Debug logging (Automatic Reactive Bridge - Feature 08)
	commandLogger CommandLogger // Logger for command generation debugging

	// Key bindings (Automatic Reactive Bridge - Feature 08, Phase 8)
	keyBindings   map[string][]KeyBinding // Key -> []Binding (supports multiple bindings per key)
	keyBindingsMu sync.RWMutex            // Protects keyBindings map

	// Message handler (Automatic Reactive Bridge - Feature 08, Task 8.4)
	messageHandler MessageHandler // Optional handler for complex message processing

	// Lifecycle
	lifecycle *LifecycleManager // Lifecycle manager for hooks
	//nolint:unused // Will be used in Task 1.3
	mounted bool // Whether component has been initialized

	// Initialization tracking (Task 10.1: Auto-Initialization Enhancement)
	initialized bool       // Whether Init() has been called and setup has executed
	initMu      sync.Mutex // Protects initialized flag for thread-safe initialization
}

// newComponentImpl creates a new component instance with the given name.
// This is an internal constructor used by the ComponentBuilder (Task 2.1).
//
// The constructor:
//   - Generates a unique ID using an atomic counter
//   - Initializes all map and slice fields to prevent nil pointer panics
//   - Sets the component name
//   - Leaves other fields at their zero values (nil, false)
//
// Example:
//
//	c := newComponentImpl("Button")
//	// c.id will be "component-1", "component-2", etc.
//	// c.state, c.handlers, c.children are initialized and empty
func newComponentImpl(name string) *componentImpl {
	// Generate unique ID using atomic counter
	id := componentIDCounter.Add(1)

	return &componentImpl{
		name:          name,
		id:            fmt.Sprintf("component-%d", id),
		state:         make(map[string]interface{}),
		provides:      make(map[string]interface{}),
		injectCache:   make(map[string]interface{}),
		handlers:      make(map[string][]EventHandler),
		children:      []Component{},
		commandQueue:  nil,               // Initialized by Build() when WithAutoCommands(true)
		commandGen:    nil,               // Initialized by Build() when WithAutoCommands(true)
		autoCommands:  false,             // Disabled by default for backward compatibility
		loopDetector:  newLoopDetector(), // Always initialized for loop detection
		commandLogger: nil,               // Initialized by Build() when WithCommandDebug(true)
	}
}

// Name returns the component's name.
func (c *componentImpl) Name() string {
	return c.name
}

// ID returns the component's unique identifier.
func (c *componentImpl) ID() string {
	return c.id
}

// Props returns the component's props.
func (c *componentImpl) Props() interface{} {
	return c.props
}

// KeyBindings returns the component's registered key bindings.
// The returned map is thread-safe (protected by RWMutex) and contains
// all key bindings registered via the builder methods.
//
// This method is primarily used for:
//   - Auto-generating help text from key descriptions
//   - Debugging key binding configuration
//   - Introspection by parent components or testing
//
// Example:
//
//	bindings := component.KeyBindings()
//	for key, bindingList := range bindings {
//	    for _, binding := range bindingList {
//	        fmt.Printf("%s: %s\n", key, binding.Description)
//	    }
//	}
func (c *componentImpl) KeyBindings() map[string][]KeyBinding {
	c.keyBindingsMu.RLock()
	defer c.keyBindingsMu.RUnlock()

	// Return a copy to prevent external modification
	if c.keyBindings == nil {
		return make(map[string][]KeyBinding)
	}

	// Create a shallow copy of the map
	result := make(map[string][]KeyBinding, len(c.keyBindings))
	for key, bindings := range c.keyBindings {
		// Copy the slice to prevent external modification
		bindingsCopy := make([]KeyBinding, len(bindings))
		copy(bindingsCopy, bindings)
		result[key] = bindingsCopy
	}

	return result
}

// HelpText generates a formatted help string from registered key bindings.
// It automatically creates user-facing documentation by extracting key
// descriptions and formatting them consistently.
//
// The method:
//   - Iterates through all key bindings
//   - Extracts non-empty descriptions
//   - Handles duplicate keys (shows first description only)
//   - Sorts keys alphabetically for consistency
//   - Formats as "key: description • key: description"
//
// Thread-safe: Uses RWMutex to safely access key bindings.
//
// Returns:
//   - Empty string if no bindings or all descriptions are empty
//   - Formatted help text with " • " separator between entries
//
// Example:
//
//	component := NewComponent("Counter").
//	    WithKeyBinding("space", "increment", "Increment counter").
//	    WithKeyBinding("ctrl+c", "quit", "Quit application").
//	    Build()
//
//	helpText := component.HelpText()
//	// Returns: "ctrl+c: Quit application • space: Increment counter"
//
// Integration with templates:
//
//	Template(func(ctx RenderContext) string {
//	    comp := ctx.component
//	    return fmt.Sprintf("%s\n\nHelp: %s", content, comp.HelpText())
//	})
func (c *componentImpl) HelpText() string {
	c.keyBindingsMu.RLock()
	defer c.keyBindingsMu.RUnlock()

	// Early return if no bindings
	if c.keyBindings == nil || len(c.keyBindings) == 0 {
		return ""
	}

	// Collect help entries (key: description)
	var helpEntries []string
	seen := make(map[string]bool)

	// Iterate through all key bindings
	for key, bindings := range c.keyBindings {
		// Skip if we've already processed this key (handles duplicates)
		if seen[key] {
			continue
		}

		// Find first binding with non-empty description
		for _, binding := range bindings {
			if binding.Description != "" {
				helpEntries = append(helpEntries, fmt.Sprintf("%s: %s", key, binding.Description))
				seen[key] = true
				break // Only use first description for duplicate keys
			}
		}
	}

	// Return empty string if no descriptions found
	if len(helpEntries) == 0 {
		return ""
	}

	// Sort alphabetically for consistency
	sort.Strings(helpEntries)

	// Join with bullet separator
	return strings.Join(helpEntries, " • ")
}

// Emit sends an event with associated data and bubbles it up to parent components.
// It creates an Event struct with metadata (timestamp, source) and
// calls bubbleEvent to propagate the event through the component tree.
//
// Events automatically bubble from child to parent unless a handler calls
// event.StopPropagation().
//
// Example:
//
//	component.Emit("submit", FormData{Username: "user"})
func (c *componentImpl) Emit(eventName string, data interface{}) {
	// Get event from pool to reduce allocations
	event := eventPool.Get().(*Event)

	// Initialize event with metadata
	event.Name = eventName
	event.Source = c
	event.Data = data
	event.Timestamp = time.Now()
	event.Stopped = false

	// Start event bubbling from this component
	c.bubbleEvent(event)

	// Return event to pool after bubbling completes
	// Safe because bubbling is synchronous (no goroutines)
	event.Name = ""
	event.Source = nil
	event.Data = nil
	event.Stopped = false
	eventPool.Put(event)
}

// On registers an event handler for the specified event name.
// Multiple handlers can be registered for the same event.
//
// Example:
//
//	component.On("click", func(data interface{}) {
//	    // Handle click event
//	})
func (c *componentImpl) On(event string, handler EventHandler) {
	c.registerHandler(event, handler)
	// Track listener for debugging/testing
	globalEventRegistry.trackEventListener(event)
}

// IsInitialized returns whether the component has been initialized.
func (c *componentImpl) IsInitialized() bool {
	c.initMu.Lock()
	defer c.initMu.Unlock()
	return c.initialized
}

// Init implements tea.Model.Init().
// It initializes the component by running the setup function.
//
// The Init method is called once when the component is first initialized
// by the Bubbletea runtime. It:
//   - Executes the setup function (if provided) with a Context
//   - Marks the component as mounted and initialized
//   - Initializes all child components
//   - Returns batched commands from children
//
// The setup function is only executed once, even if Init() is called multiple times.
// This method is thread-safe and idempotent (safe to call multiple times).
func (c *componentImpl) Init() tea.Cmd {
	// Thread-safe check and set of initialized flag
	c.initMu.Lock()
	if c.initialized {
		// Already initialized - return early
		c.initMu.Unlock()
		return nil
	}
	c.initialized = true
	c.initMu.Unlock()

	// Run setup function if provided and not already mounted
	if c.setup != nil && !c.mounted {
		ctx := &Context{component: c}
		c.setup(ctx)
		c.mounted = true
	}

	// Initialize child components
	if len(c.children) > 0 {
		cmds := make([]tea.Cmd, len(c.children))
		for i, child := range c.children {
			cmds[i] = child.Init()
		}
		return tea.Batch(cmds...)
	}

	return nil
}

// Update implements tea.Model.Update().
// It handles incoming Bubbletea messages and updates component state.
//
// The Update method is called for every message in the Bubbletea event loop.
// It:
//   - Processes the incoming message (including StateChangedMsg)
//   - Updates child components with the message
//   - Executes onUpdated lifecycle hooks
//   - Drains pending commands from the command queue
//   - Returns the updated model and batched commands
//
// The onUpdated hooks execute after child updates to ensure state changes
// from children are reflected before hook execution.
//
// For automatic reactive bridge (Feature 08):
//   - StateChangedMsg triggers lifecycle hooks when component ID matches
//   - Command queue is drained and commands are batched with child commands
//   - All commands are returned via tea.Batch for execution by Bubbletea runtime
func (c *componentImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// [NEW] Call message handler first (Automatic Reactive Bridge - Task 8.4)
	// Handler is called BEFORE key bindings to allow complex message processing
	// Handler can return nil (no command) or a tea.Cmd to be batched
	if c.messageHandler != nil {
		if cmd := c.messageHandler(c, msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Process key bindings (Automatic Reactive Bridge - Phase 8)
	// This allows declarative key-to-event mapping without manual Update() logic
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if c.keyBindings != nil {
			c.keyBindingsMu.RLock()
			bindings, found := c.keyBindings[keyMsg.String()]
			c.keyBindingsMu.RUnlock()

			if found {
				// Iterate through bindings for this key
				for _, binding := range bindings {
					// Check condition if set (for mode-based input)
					if binding.Condition != nil && !binding.Condition() {
						continue // Skip this binding, try next
					}

					// Special handling for "quit" event
					if binding.Event == "quit" {
						return c, tea.Quit
					}

					// Emit the bound event with optional data
					c.Emit(binding.Event, binding.Data)

					// First matching binding wins (break after first match)
					// This allows multiple conditional bindings per key
					break
				}
			}
		}
	}

	// Handle StateChangedMsg from automatic reactive bridge
	switch msg := msg.(type) {
	case StateChangedMsg:
		// Only process if this message is for this component
		if msg.ComponentID == c.id {
			// State already updated synchronously by Ref.Set()
			// Execute onUpdated hooks to trigger side effects
			if c.lifecycle != nil {
				c.lifecycle.executeUpdated()
			}
		}
	}

	// Update child components
	if len(c.children) > 0 {
		childCmds := make([]tea.Cmd, len(c.children))
		for i, child := range c.children {
			updatedChild, cmd := child.Update(msg)
			// Update the child in the slice
			if impl, ok := updatedChild.(*componentImpl); ok {
				c.children[i] = impl
			}
			childCmds[i] = cmd
		}

		// Collect child commands
		cmds = append(cmds, childCmds...)

		// Execute onUpdated hooks after child updates (for non-StateChangedMsg)
		// StateChangedMsg already executed hooks above
		if _, isStateChanged := msg.(StateChangedMsg); !isStateChanged {
			if c.lifecycle != nil {
				c.lifecycle.executeUpdated()
			}
		}

		// Reset update counter after each Update() cycle completes
		if c.lifecycle != nil {
			c.lifecycle.resetUpdateCount()
		}

		// Reset loop detector after each Update() cycle completes
		if c.loopDetector != nil {
			c.loopDetector.reset()
		}
	} else {
		// Execute onUpdated hooks for components without children (for non-StateChangedMsg)
		// StateChangedMsg already executed hooks above
		if _, isStateChanged := msg.(StateChangedMsg); !isStateChanged {
			if c.lifecycle != nil {
				c.lifecycle.executeUpdated()
			}
		}

		// Reset update counter after each Update() cycle completes
		if c.lifecycle != nil {
			c.lifecycle.resetUpdateCount()
		}

		// Reset loop detector after each Update() cycle completes
		if c.loopDetector != nil {
			c.loopDetector.reset()
		}
	}

	// Drain pending commands from command queue (automatic reactive bridge)
	if c.commandQueue != nil {
		pendingCmds := c.commandQueue.DrainAll()
		if len(pendingCmds) > 0 {
			cmds = append(cmds, pendingCmds...)
		}
	}

	// Return batched commands (or nil if no commands)
	if len(cmds) == 0 {
		return c, nil
	}

	return c, tea.Batch(cmds...)
}

// View implements tea.Model.View().
// It renders the component's UI using the template function.
//
// The View method:
//   - Executes onMounted lifecycle hooks on first render (if not already mounted)
//   - Marks template context as active (for safety checks)
//   - Calls the template function with a RenderContext
//   - Clears template context (even if template panics via defer)
//   - Returns the rendered string
//
// If no template is provided, it returns an empty string.
//
// Template Context Safety:
// During template rendering, the component tracks that it's in a template context.
// This allows Ref.Set() to detect and prevent illegal state mutations inside templates.
// Templates must be pure functions with no side effects.
func (c *componentImpl) View() string {
	// Execute onMounted hooks on first render
	if c.lifecycle != nil && !c.lifecycle.IsMounted() {
		c.lifecycle.executeMounted()
	}

	if c.template == nil {
		return ""
	}

	// Mark template context as active
	// Use Context to access the methods (though we could access component directly)
	ctx := Context{component: c}
	ctx.enterTemplate()

	// Ensure we exit template context even if template panics
	defer ctx.exitTemplate()

	// Render with RenderContext
	renderCtx := RenderContext{component: c}
	return c.template(renderCtx)
}

// Unmount cleans up the component and its children.
// It executes onUnmounted hooks, cleanup functions, and recursively unmounts children.
//
// The Unmount method should be called when the component is being removed from the UI.
// It:
//   - Executes onUnmounted lifecycle hooks
//   - Executes registered cleanup functions in reverse order (LIFO)
//   - Recursively unmounts all child components
//   - Ensures proper resource cleanup
//
// Cleanup order:
//  1. Parent onUnmounted hooks execute
//  2. Parent cleanup functions execute (reverse order)
//  3. Child components unmount recursively
//
// This ensures that parent cleanup logic runs before children are unmounted,
// allowing parents to perform any necessary coordination before child cleanup.
//
// Example:
//
//	component.Unmount()  // Clean up component and children
func (c *componentImpl) Unmount() {
	// Execute lifecycle cleanup (onUnmounted hooks + cleanup functions)
	if c.lifecycle != nil {
		c.lifecycle.executeUnmounted()
	}

	// CRITICAL FIX: Always clean up event handlers, even if no lifecycle
	// Event handlers are registered on the component, not the lifecycle
	// So they must be cleaned up regardless of whether lifecycle hooks exist
	c.handlersMu.Lock()
	c.handlers = make(map[string][]EventHandler)
	c.handlersMu.Unlock()

	// Unmount children recursively
	for _, child := range c.children {
		if impl, ok := child.(*componentImpl); ok {
			impl.Unmount()
		}
	}
}

// inject walks up the component tree to find a provided value.
// It searches the current component first, then recursively checks parents.
// Returns the provided value if found, otherwise returns the default value.
//
// Performance: O(1) for cached lookups, O(depth) for first lookup.
// The cache is populated on first access and persists for the component lifetime.
// This optimization improves inject performance from ~120ns at depth 10 to ~30ns.
//
// This is used internally by Context.Inject for dependency injection.
func (c *componentImpl) inject(key string, defaultValue interface{}) interface{} {
	// Fast path: Check cache first (O(1))
	c.injectCacheMu.RLock()
	if val, ok := c.injectCache[key]; ok {
		c.injectCacheMu.RUnlock()
		return val
	}
	c.injectCacheMu.RUnlock()

	// Slow path: Tree walk and cache population
	// Check current component's provides map
	c.providesMu.RLock()
	if val, ok := c.provides[key]; ok {
		c.providesMu.RUnlock()

		// Cache the result for O(1) future lookups
		c.injectCacheMu.Lock()
		c.injectCache[key] = val
		c.injectCacheMu.Unlock()

		return val
	}
	c.providesMu.RUnlock()

	// Walk up parent chain
	var result interface{}
	if c.parent != nil {
		result = c.parent.inject(key, defaultValue)
	} else {
		// Not found in tree, use default
		result = defaultValue
	}

	// Cache the result (whether found or default)
	c.injectCacheMu.Lock()
	c.injectCache[key] = result
	c.injectCacheMu.Unlock()

	return result
}

// Context is now defined in context.go (Task 3.1)
// RenderContext is now defined in render_context.go (Task 3.2)
