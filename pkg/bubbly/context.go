package bubbly

import (
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// refIDCounter is an atomic counter for generating unique ref IDs.
// It's incremented each time a new ref is created with automatic command generation.
var refIDCounter atomic.Uint64

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
//	        return count.GetTyped().(int) * 2
//	    })
//
//	    // Expose state to template
//	    ctx.Expose("count", count)
//	    ctx.Expose("doubled", doubled)
//
//	    // Register event handlers
//	    ctx.On("increment", func(data interface{}) {
//	        current := count.GetTyped().(int)
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
// When automatic command generation is enabled (component.autoCommands = true),
// this method creates a Ref with a setHook that automatically generates Bubbletea
// commands when Set() is called, triggering UI updates without manual event emission.
//
// When automatic command generation is disabled (default), this method creates
// a standard Ref with no automatic command generation, maintaining backward
// compatibility with existing code.
//
// Template Context Safety:
// All Refs created via ctx.Ref() have a template context checker attached.
// This prevents accidental state mutations inside template functions by panicking
// with a clear error message if Set() is called during template rendering.
//
// Example (manual mode - default):
//
//	count := ctx.Ref(0)
//	count.Set(42)
//	ctx.Emit("update", nil) // Manual emit required
//
// Example (automatic mode):
//
//	// In component builder: .WithAutoCommands(true)
//	count := ctx.Ref(0)
//	count.Set(42) // UI updates automatically!
func (ctx *Context) Ref(value interface{}) *Ref[interface{}] {
	// Create base ref
	ref := NewRef(value)

	// Attach template checker for safety (always, regardless of auto commands)
	// This prevents Ref.Set() calls inside templates
	ref.templateChecker = ctx.InTemplate

	// Check if auto commands enabled (thread-safe read)
	ctx.component.autoCommandsMu.RLock()
	autoEnabled := ctx.component.autoCommands
	commandGen := ctx.component.commandGen
	ctx.component.autoCommandsMu.RUnlock()

	// If auto commands disabled or component not set, return standard ref
	if !autoEnabled || ctx.component == nil {
		return ref
	}

	// Auto commands enabled - attach command generation hook
	// Generate unique ref ID
	refID := refIDCounter.Add(1)
	refIDStr := fmt.Sprintf("ref-%d", refID)

	// Capture context for the hook
	componentID := ctx.component.id
	componentName := ctx.component.name
	queue := ctx.component.commandQueue
	detector := ctx.component.loopDetector
	logger := ctx.component.commandLogger

	// Set the hook that generates commands on Set()
	ref.setHook = func(oldValue, newValue interface{}) {
		// Recover from panics in command generation
		// Value has already been updated before this hook runs, so state update always succeeds
		defer func() {
			if r := recover(); r != nil {
				// Report panic to observability system
				if reporter := observability.GetErrorReporter(); reporter != nil {
					cmdErr := &observability.CommandGenerationError{
						ComponentID: componentID,
						RefID:       refIDStr,
						PanicValue:  r,
					}

					errorCtx := &observability.ErrorContext{
						ComponentName: componentName,
						ComponentID:   componentID,
						EventName:     "command:generation",
						Timestamp:     time.Now(),
						StackTrace:    debug.Stack(),
						Tags: map[string]string{
							"error_type": "command_generation_panic",
							"ref_id":     refIDStr,
						},
						Extra: map[string]interface{}{
							"old_value":                oldValue,
							"new_value":                newValue,
							"panic":                    r,
							"command_generation_error": cmdErr,
						},
					}

					// Report using HandlerPanicError for interface compatibility
					reporter.ReportPanic(&observability.HandlerPanicError{
						ComponentName: errorCtx.ComponentName,
						EventName:     errorCtx.EventName,
						PanicValue:    r,
					}, errorCtx)
				}
				// Continue execution - state update succeeded, only command generation failed
			}
		}()

		// Check for command generation loop
		if detector != nil {
			if err := detector.checkLoop(componentID, refIDStr); err != nil {
				// Report loop detection error to observability system
				if reporter := observability.GetErrorReporter(); reporter != nil {
					errorCtx := &observability.ErrorContext{
						ComponentName: componentName,
						ComponentID:   componentID,
						EventName:     "command:loop_detected",
						Timestamp:     time.Now(),
						StackTrace:    debug.Stack(),
						Tags: map[string]string{
							"error_type": "command_loop",
							"ref_id":     refIDStr,
						},
						Extra: map[string]interface{}{
							"old_value":      oldValue,
							"new_value":      newValue,
							"loop_error":     err.Error(),
							"max_commands":   maxCommandsPerRef,
							"component_name": componentName,
						},
					}

					// Report as error (not panic, since this is a controlled error condition)
					reporter.ReportError(err, errorCtx)
				}

				// Do not generate command - loop detected
				return
			}
		}

		// Log command generation if debug mode enabled
		if logger != nil {
			logger.LogCommand(componentName, componentID, refIDStr, oldValue, newValue)
		}

		// Generate command for the state change
		cmd := commandGen.Generate(
			componentID,
			refIDStr,
			oldValue,
			newValue,
		)

		// Enqueue command for component to return from Update()
		queue.Enqueue(cmd)
	}

	return ref
}

// Computed creates a new computed value that automatically updates
// when its dependencies change.
//
// Example:
//
//	count := ctx.Ref(10)
//	doubled := ctx.Computed(func() interface{} {
//	    return count.GetTyped().(int) * 2
//	})
func (ctx *Context) Computed(fn func() interface{}) *Computed[interface{}] {
	return NewComputed(fn)
}

// Watch registers a callback that is called whenever the given Ref changes.
// The callback receives the new and old values.
// Returns a cleanup function that stops watching when called.
//
// The watcher is automatically registered with the lifecycle manager for
// auto-cleanup when the component unmounts, preventing memory leaks.
//
// Example:
//
//	count := ctx.Ref(0)
//	cleanup := ctx.Watch(count, func(newVal, oldVal interface{}) {
//	    log.Printf("Count: %v -> %v", oldVal, newVal)
//	})
//	// Watcher will be automatically cleaned up on unmount
//	// Or manually: cleanup()
func (ctx *Context) Watch(ref *Ref[interface{}], callback WatchCallback[interface{}]) WatchCleanup {
	// Create lifecycle manager if it doesn't exist
	if ctx.component.lifecycle == nil {
		ctx.component.lifecycle = newLifecycleManager(ctx.component)
	}

	// Create watcher using global Watch function
	cleanup := Watch(ref, callback)

	// Register cleanup with lifecycle manager for auto-cleanup
	ctx.component.lifecycle.registerWatcher(cleanup)

	// Return cleanup function for manual cleanup if needed
	return cleanup
}

// Expose stores a value in the component's state map.
// This makes it accessible in the template via Get().
//
// This is the primary way to share state between the setup function
// and the template function.
//
// CRITICAL FOR DEVTOOLS: When exposing a Ref, this notifies framework hooks
// to register ref ownership, enabling DevTools to accurately track which refs
// belong to which components.
//
// This method is thread-safe and can be called concurrently.
//
// Example:
//
//	count := ctx.Ref(0)
//	ctx.Expose("count", count)
//	// Later in template: count := ctx.Get("count").(*Ref[int])
func (ctx *Context) Expose(key string, value interface{}) {
	ctx.component.stateMu.Lock()
	if ctx.component.state == nil {
		ctx.component.state = make(map[string]interface{})
	}
	ctx.component.state[key] = value
	ctx.component.stateMu.Unlock()

	// Check if value is a Ref and notify hooks for DevTools tracking
	// We use type assertion to detect any Ref[T] type
	// The refID is generated using memory address (same as Set() does)
	switch v := value.(type) {
	case *Ref[int]:
		refID := fmt.Sprintf("ref-%p", v)
		notifyHookRefExposed(ctx.component.id, refID, key)
	case *Ref[string]:
		refID := fmt.Sprintf("ref-%p", v)
		notifyHookRefExposed(ctx.component.id, refID, key)
	case *Ref[bool]:
		refID := fmt.Sprintf("ref-%p", v)
		notifyHookRefExposed(ctx.component.id, refID, key)
	case *Ref[float64]:
		refID := fmt.Sprintf("ref-%p", v)
		notifyHookRefExposed(ctx.component.id, refID, key)
	case *Ref[interface{}]:
		refID := fmt.Sprintf("ref-%p", v)
		notifyHookRefExposed(ctx.component.id, refID, key)
		// Note: This covers the most common types. For other Ref[T] types,
		// they won't be tracked by DevTools ref ownership (but will still work).
		// A future improvement could use reflection to detect all Ref types.
	}
}

// Get retrieves a value from the component's state map.
// Returns nil if the key doesn't exist.
//
// This is typically used in the setup function to access previously
// exposed values, though it's more commonly used in templates.
//
// This method is thread-safe and can be called concurrently.
//
// Example:
//
//	value := ctx.Get("count")
//	if ref, ok := value.(*Ref[int]); ok {
//	    // Use the ref
//	}
func (ctx *Context) Get(key string) interface{} {
	ctx.component.stateMu.RLock()
	defer ctx.component.stateMu.RUnlock()

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
//	    fmt.Printf("Count changed to: %d\n", count.GetTyped())
//	}, count)
func (ctx *Context) OnUpdated(hook func(), deps ...Dependency) {
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

// Provide stores a value in the component's provides map, making it available
// for injection by child components via Inject().
//
// This implements the provide/inject pattern for dependency injection,
// allowing parent components to share values with descendants without
// prop drilling.
//
// Provided values can be of any type, including reactive Refs and Computed values.
// When a reactive value is provided, all injecting components share the same
// instance and see updates automatically.
//
// Example:
//
//	// Parent provides theme
//	theme := ctx.Ref("dark")
//	ctx.Provide("theme", theme)
//
//	// Child injects theme
//	theme := ctx.Inject("theme", ctx.Ref("light")).(*Ref[interface{}])
func (ctx *Context) Provide(key string, value interface{}) {
	ctx.component.providesMu.Lock()
	ctx.component.provides[key] = value
	ctx.component.providesMu.Unlock()
}

// Inject retrieves a value provided by an ancestor component.
// It walks up the component tree looking for the first component that
// provided the specified key.
//
// If the key is not found in any ancestor, the defaultValue is returned.
// This allows components to work standalone with sensible defaults.
//
// The nearest provider wins - if both a parent and grandparent provide
// the same key, the parent's value is returned.
//
// Example:
//
//	// Inject with default
//	theme := ctx.Inject("theme", "light")
//
//	// Inject reactive Ref
//	themeRef := ctx.Inject("theme", ctx.Ref("light")).(*Ref[interface{}])
//
//	// Inject with nil default (optional dependency)
//	user := ctx.Inject("currentUser", nil)
func (ctx *Context) Inject(key string, defaultValue interface{}) interface{} {
	return ctx.component.inject(key, defaultValue)
}

// UseTheme retrieves the theme from parent via injection or returns the default.
// This eliminates the boilerplate of manual inject+expose for theme colors.
//
// The method checks if a parent component has provided a theme via ProvideTheme().
// If found, it returns the injected theme. If not found or if the type assertion
// fails (wrong type provided), it returns the defaultTheme parameter.
//
// This is type-safe - if a parent provides a non-Theme value under the "theme" key,
// the type assertion will fail gracefully and the default will be used.
//
// Usage in child component:
//
//	Setup(func(ctx *Context) {
//	    // Get theme from parent or use default
//	    theme := ctx.UseTheme(bubbly.DefaultTheme)
//
//	    // Use theme colors in styles
//	    titleStyle := lipgloss.NewStyle().Foreground(theme.Primary)
//	    errorStyle := lipgloss.NewStyle().Foreground(theme.Error)
//
//	    // Expose for template if needed
//	    ctx.Expose("theme", theme)
//	})
//
// Example with custom default:
//
//	customDefault := bubbly.Theme{
//	    Primary:    lipgloss.Color("99"),
//	    Secondary:  lipgloss.Color("120"),
//	    // ... other colors
//	}
//	theme := ctx.UseTheme(customDefault)
//
// This method is thread-safe and can be called concurrently.
func (ctx *Context) UseTheme(defaultTheme Theme) Theme {
	if injected := ctx.Inject("theme", nil); injected != nil {
		if theme, ok := injected.(Theme); ok {
			return theme
		}
	}
	return defaultTheme
}

// ProvideTheme provides a theme to all descendant components.
// This is a convenience method that wraps ctx.Provide("theme", theme) to eliminate
// boilerplate when sharing theme colors across the component hierarchy.
//
// Parent components should call this in their Setup function to make the theme
// available to all descendants. Child components can then retrieve the theme using
// ctx.UseTheme(defaultTheme).
//
// The theme is propagated down the component tree via the standard Provide/Inject
// mechanism. Children can override the theme for their subtree by calling ProvideTheme
// with a different theme, which will take precedence for their descendants.
//
// Usage in parent component:
//
//	Setup(func(ctx *Context) {
//	    // Provide custom theme to all descendants
//	    customTheme := bubbly.DefaultTheme
//	    customTheme.Primary = lipgloss.Color("99")  // Override primary color
//	    ctx.ProvideTheme(customTheme)
//	})
//
// Usage in child component:
//
//	Setup(func(ctx *Context) {
//	    // Retrieve theme from parent
//	    theme := ctx.UseTheme(bubbly.DefaultTheme)
//	    titleStyle := lipgloss.NewStyle().Foreground(theme.Primary)
//	})
//
// This method is thread-safe and can be called concurrently.
func (ctx *Context) ProvideTheme(theme Theme) {
	ctx.Provide("theme", theme)
}

// EnableAutoCommands enables automatic command generation for reactive state changes.
// When enabled, calling Ref.Set() automatically generates Bubbletea commands that
// trigger UI updates without manual event emission.
//
// This method is idempotent - calling it multiple times has no adverse effects.
// If no command generator is set, a default generator is automatically assigned.
//
// Example:
//
//	ctx.EnableAutoCommands()
//	count := ctx.Ref(0)
//	count.Set(42) // Automatically triggers UI update
func (ctx *Context) EnableAutoCommands() {
	ctx.component.autoCommandsMu.Lock()
	defer ctx.component.autoCommandsMu.Unlock()

	ctx.component.autoCommands = true

	// Ensure a command generator is set
	if ctx.component.commandGen == nil {
		ctx.component.commandGen = &defaultCommandGenerator{}
	}
}

// DisableAutoCommands disables automatic command generation for reactive state changes.
// When disabled, Ref.Set() calls do not generate commands, and manual event emission
// is required to trigger UI updates.
//
// This is useful for performance-critical code paths where you want to batch multiple
// state changes and trigger a single update manually.
//
// Example:
//
//	ctx.DisableAutoCommands()
//	// Batch updates without triggering commands
//	for i := 0; i < 1000; i++ {
//	    count.Set(i)
//	}
//	ctx.EnableAutoCommands()
//	ctx.Emit("batch-complete", nil) // Single manual update
func (ctx *Context) DisableAutoCommands() {
	ctx.component.autoCommandsMu.Lock()
	defer ctx.component.autoCommandsMu.Unlock()

	ctx.component.autoCommands = false
}

// IsAutoCommandsEnabled returns whether automatic command generation is currently enabled.
// This can be used to check the current state before conditionally enabling/disabling.
//
// Example:
//
//	if !ctx.IsAutoCommandsEnabled() {
//	    ctx.EnableAutoCommands()
//	}
func (ctx *Context) IsAutoCommandsEnabled() bool {
	ctx.component.autoCommandsMu.RLock()
	defer ctx.component.autoCommandsMu.RUnlock()

	return ctx.component.autoCommands
}

// ManualRef creates a reactive reference that never generates automatic commands,
// regardless of the component's auto-commands setting.
//
// This is useful when you need explicit control over when UI updates occur, even
// in a component that has automatic command generation enabled globally.
//
// The returned Ref behaves exactly like a standard Ref - you must manually emit
// events to trigger UI updates after calling Set().
//
// Example:
//
//	// Component has auto-commands enabled
//	ctx.EnableAutoCommands()
//
//	// But this ref needs manual control
//	internalCounter := ctx.ManualRef(0)
//	internalCounter.Set(100) // No command generated
//	ctx.Emit("update", nil)  // Manual update required
func (ctx *Context) ManualRef(value interface{}) *Ref[interface{}] {
	// Lock to safely read and modify autoCommands
	ctx.component.autoCommandsMu.Lock()
	wasAuto := ctx.component.autoCommands
	ctx.component.autoCommands = false
	ctx.component.autoCommandsMu.Unlock()

	// Create ref (will be standard ref without command generation)
	ref := ctx.Ref(value)

	// Restore auto commands state
	ctx.component.autoCommandsMu.Lock()
	ctx.component.autoCommands = wasAuto
	ctx.component.autoCommandsMu.Unlock()

	return ref
}

// SetCommandGenerator sets a custom command generator for this component.
// The generator is used to create Bubbletea commands when reactive state changes.
//
// This allows you to customize how commands are generated, for example to add
// logging, metrics, or custom message types.
//
// Setting a nil generator is allowed and will use the component's current generator
// (or none if not set). To reset to the default generator, call EnableAutoCommands().
//
// Example:
//
//	type LoggingGenerator struct{}
//
//	func (g *LoggingGenerator) Generate(componentID, refID string, oldValue, newValue interface{}) tea.Cmd {
//	    log.Printf("State change: %s.%s: %v -> %v", componentID, refID, oldValue, newValue)
//	    return func() tea.Msg {
//	        return StateChangedMsg{
//	            ComponentID: componentID,
//	            RefID:       refID,
//	            OldValue:    oldValue,
//	            NewValue:    newValue,
//	        }
//	    }
//	}
//
//	ctx.SetCommandGenerator(&LoggingGenerator{})
func (ctx *Context) SetCommandGenerator(gen CommandGenerator) {
	ctx.component.autoCommandsMu.Lock()
	defer ctx.component.autoCommandsMu.Unlock()

	ctx.component.commandGen = gen
}

// enterTemplate marks that template rendering has started.
// This is used internally to detect and prevent state mutations during template rendering.
// Templates must be pure functions with no side effects.
//
// This method is called by component.View() before executing the template function.
// It is thread-safe and can be called multiple times (for nested templates).
func (ctx *Context) enterTemplate() {
	ctx.component.inTemplateMu.Lock()
	defer ctx.component.inTemplateMu.Unlock()

	ctx.component.inTemplate = true
}

// exitTemplate marks that template rendering has completed.
// This is used internally to reset template context after rendering.
//
// This method is called by component.View() after executing the template function
// (via defer to ensure it runs even if template panics).
// It is thread-safe.
func (ctx *Context) exitTemplate() {
	ctx.component.inTemplateMu.Lock()
	defer ctx.component.inTemplateMu.Unlock()

	ctx.component.inTemplate = false
}

// InTemplate returns whether the context is currently executing inside a template.
// This is used internally to detect illegal state mutations during template rendering.
//
// Templates must be pure functions (read-only). Calling Ref.Set() inside a template
// will panic with a clear error message.
//
// This method is thread-safe and uses a read lock, allowing multiple concurrent checks.
//
// Example (internal use by Ref.Set()):
//
//	func (r *Ref[T]) Set(value T) {
//	    if ctx.InTemplate() {
//	        panic("Cannot call Ref.Set() in template - templates must be pure")
//	    }
//	    // ... update value
//	}
func (ctx *Context) InTemplate() bool {
	ctx.component.inTemplateMu.RLock()
	defer ctx.component.inTemplateMu.RUnlock()

	return ctx.component.inTemplate
}

// ExposeComponent exposes a component to the component's state map with automatic initialization.
// This is a convenience method that combines initialization checking and state exposure.
//
// This simplifies child component setup by combining Init(), AddChild(), and Expose() into one call.
// If the component is already initialized, Init() is not called again.
//
// CRITICAL FOR DEVTOOLS: This establishes parent-child relationships by calling AddChild(),
// which notifies framework hooks to build the component tree for Inspector visualization.
//
// If Init() returns a command and the parent component has a command queue,
// the command is automatically enqueued for execution.
//
// Example (before - manual initialization):
//
//	Setup(func(ctx *Context) {
//	    header, _ := CreateHeader(props)
//	    sidebar, _ := CreateSidebar(props)
//
//	    // Manual initialization required
//	    header.Init()
//	    sidebar.Init()
//
//	    ctx.Expose("header", header)
//	    ctx.Expose("sidebar", sidebar)
//	})
//
// Example (after - automatic initialization):
//
//	Setup(func(ctx *Context) {
//	    header, _ := CreateHeader(props)
//	    sidebar, _ := CreateSidebar(props)
//
//	    // Auto-initializes if needed AND establishes parent-child relationship
//	    ctx.ExposeComponent("header", header)
//	    ctx.ExposeComponent("sidebar", sidebar)
//	})
//
// Parameters:
//   - name: The key to use in the state map
//   - comp: The component to initialize and expose
//
// Returns:
//   - error: Returns error if comp is nil or if AddChild fails, otherwise nil
func (ctx *Context) ExposeComponent(name string, comp Component) error {
	if comp == nil {
		return fmt.Errorf("cannot expose nil component")
	}

	// Auto-initialize if not already initialized
	if !comp.IsInitialized() {
		cmd := comp.Init()

		// Queue Init() command if one was returned and parent has command queue
		if cmd != nil && ctx.component.commandQueue != nil {
			ctx.component.commandQueue.Enqueue(cmd)
		}
	}

	// CRITICAL FIX 2: Establish parent-child relationship
	// This enables DevTools to build accurate component tree
	if err := ctx.component.AddChild(comp); err != nil {
		return fmt.Errorf("failed to add child component: %w", err)
	}

	// Expose to context using existing Expose method
	ctx.Expose(name, comp)
	return nil
}
