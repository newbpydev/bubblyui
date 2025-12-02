// Package bubblyui provides a Vue-inspired TUI framework for Go.
//
// BubblyUI brings reactive state management and component-based architecture
// to terminal applications built on Bubbletea. It offers type-safe reactive
// primitives, a powerful component system, lifecycle hooks, and composables.
//
// # Quick Start
//
//	import "github.com/newbpydev/bubblyui"
//
//	func main() {
//	    counter, _ := bubblyui.NewComponent("Counter").
//	        Setup(func(ctx *bubblyui.Context) {
//	            count := ctx.Ref(0)
//	            ctx.Expose("count", count)
//	        }).
//	        Template(func(ctx bubblyui.RenderContext) string {
//	            return fmt.Sprintf("Count: %v", ctx.Get("count"))
//	        }).
//	        Build()
//
//	    bubblyui.Run(counter)
//	}
//
// # Core Types
//
// The following types are re-exported from pkg/bubbly for convenience:
//   - Component: A BubblyUI component instance
//   - ComponentBuilder: Fluent API for constructing components
//   - Ref[T]: A reactive reference holding a mutable value
//   - Computed[T]: A derived reactive value that auto-updates
//   - Context: The component setup context
//   - RenderContext: The template rendering context
//
// # Subpackages
//
// For additional functionality, import the subpackages directly:
//
//	import "github.com/newbpydev/bubblyui/pkg/bubbly/router"      // Navigation
//	import "github.com/newbpydev/bubblyui/pkg/bubbly/composables" // Composables
//	import "github.com/newbpydev/bubblyui/pkg/bubbly/directives"  // Directives
//	import "github.com/newbpydev/bubblyui/pkg/components"          // UI components
package bubblyui

import "github.com/newbpydev/bubblyui/pkg/bubbly"

// =============================================================================
// Core Types - Re-exported for convenient access
// =============================================================================

// Component represents a BubblyUI component with reactive state,
// lifecycle hooks, and template rendering. It extends Bubbletea's tea.Model
// interface with additional methods for component identification, props
// management, and event handling.
type Component = bubbly.Component

// ComponentBuilder provides a fluent API for constructing components.
// Use NewComponent() to create a builder, then chain configuration methods
// and call Build() to create the component.
type ComponentBuilder = bubbly.ComponentBuilder

// Context provides access to component state and utilities during setup.
// It allows creating reactive state (Ref, Computed), registering lifecycle
// hooks, and exposing values for template rendering.
type Context = bubbly.Context

// RenderContext provides access to component state during template rendering.
// It allows retrieving exposed values, accessing props, and getting the
// component instance.
type RenderContext = bubbly.RenderContext

// =============================================================================
// Reactive Primitives
// =============================================================================

// Ref is a reactive reference that holds a mutable value of type T.
// Changes to the value automatically trigger dependent computations and watchers.
//
// Example:
//
//	count := bubblyui.NewRef(0)
//	count.Set(count.Get() + 1)
type Ref[T any] = bubbly.Ref[T]

// Computed is a derived reactive value that automatically recomputes
// when its dependencies change. Computed values are cached and only
// recalculate when necessary.
//
// Example:
//
//	count := bubblyui.NewRef(10)
//	doubled := bubblyui.NewComputed(func() int {
//	    return count.Get() * 2
//	})
type Computed[T any] = bubbly.Computed[T]

// =============================================================================
// Runner Options
// =============================================================================

// RunOption configures the behavior of the Run function.
// Use With* functions to create options.
type RunOption = bubbly.RunOption

// =============================================================================
// Core Functions
// =============================================================================

// NewComponent creates a new ComponentBuilder with the given name.
// Use the builder's fluent API to configure the component, then call Build().
//
// Example:
//
//	component, err := bubblyui.NewComponent("Counter").
//	    Setup(func(ctx *bubblyui.Context) {
//	        // Initialize reactive state
//	    }).
//	    Template(func(ctx bubblyui.RenderContext) string {
//	        return "Hello, World!"
//	    }).
//	    Build()
var NewComponent = bubbly.NewComponent

// NewRef creates a new reactive reference with the given initial value.
// The reference is type-safe thanks to Go generics.
//
// Example:
//
//	name := bubblyui.NewRef("Alice")
//	count := bubblyui.NewRef(0)
//	items := bubblyui.NewRef([]string{"a", "b"})
func NewRef[T any](value T) *Ref[T] {
	return bubbly.NewRef(value)
}

// NewComputed creates a new computed value with the given computation function.
// The computed value automatically tracks dependencies accessed within the
// function and recomputes when they change.
//
// Example:
//
//	total := bubblyui.NewComputed(func() int {
//	    return price.Get() * quantity.Get()
//	})
func NewComputed[T any](fn func() T) *Computed[T] {
	return bubbly.NewComputed(fn)
}

// Watch creates a watcher that executes the callback when the watched value changes.
// Returns a cleanup function that stops the watcher when called.
//
// Example:
//
//	cleanup := bubblyui.Watch(count, func(newVal, oldVal int) {
//	    fmt.Printf("Count changed from %d to %d\n", oldVal, newVal)
//	})
//	defer cleanup()
func Watch[T any](ref *Ref[T], callback func(newVal, oldVal T)) func() {
	return bubbly.Watch(ref, callback)
}

// WatchEffect creates a side-effect watcher that tracks dependencies automatically.
// The effect runs immediately and re-runs whenever any accessed reactive value changes.
// Returns a cleanup function that stops the effect when called.
//
// Example:
//
//	cleanup := bubblyui.WatchEffect(func() {
//	    fmt.Printf("Current count: %d\n", count.Get())
//	})
//	defer cleanup()
var WatchEffect = bubbly.WatchEffect

// Run starts the Bubbletea application with the given component.
// This is the main entry point for BubblyUI applications.
// Options can be passed to configure the application behavior.
//
// Example:
//
//	err := bubblyui.Run(app,
//	    bubblyui.WithAltScreen(),
//	    bubblyui.WithMouseAllMotion(),
//	)
var Run = bubbly.Run

// =============================================================================
// Run Options - Screen and Display
// =============================================================================

// WithAltScreen enables the alternate screen buffer for full-screen applications.
// This is the most common option for TUI applications.
var WithAltScreen = bubbly.WithAltScreen

// WithFPS sets the target frames per second for rendering.
// Default is 60 FPS. Higher values provide smoother animations but use more CPU.
var WithFPS = bubbly.WithFPS

// WithReportFocus enables focus reporting.
// The program will receive messages when the terminal gains/loses focus.
var WithReportFocus = bubbly.WithReportFocus

// =============================================================================
// Run Options - Mouse Support
// =============================================================================

// WithMouseAllMotion enables mouse support with all motion events.
// This captures all mouse movements, clicks, and scroll events.
var WithMouseAllMotion = bubbly.WithMouseAllMotion

// WithMouseCellMotion enables mouse support with cell motion events.
// This captures mouse events only when the mouse moves between cells.
var WithMouseCellMotion = bubbly.WithMouseCellMotion

// =============================================================================
// Run Options - Input/Output
// =============================================================================

// WithInput sets a custom input source for the program.
// By default, the program reads from os.Stdin.
var WithInput = bubbly.WithInput

// WithOutput sets a custom output destination for the program.
// By default, the program writes to os.Stdout.
var WithOutput = bubbly.WithOutput

// WithInputTTY forces the program to use a TTY for input.
// This is useful when running in non-interactive environments.
var WithInputTTY = bubbly.WithInputTTY

// WithEnvironment sets custom environment variables for the program.
// This is useful for controlling terminal behavior.
var WithEnvironment = bubbly.WithEnvironment

// =============================================================================
// Run Options - Context and Lifecycle
// =============================================================================

// WithContext sets a context for the program.
// The program will exit when the context is canceled.
var WithContext = bubbly.WithContext

// WithoutBracketedPaste disables bracketed paste mode.
// This is useful for terminals that don't support bracketed paste.
var WithoutBracketedPaste = bubbly.WithoutBracketedPaste

// WithoutSignalHandler disables the default signal handler.
// Use this if you want to handle signals (like SIGINT) manually.
var WithoutSignalHandler = bubbly.WithoutSignalHandler

// WithoutCatchPanics disables panic catching.
// Use this during development to see full panic stack traces.
var WithoutCatchPanics = bubbly.WithoutCatchPanics

// =============================================================================
// Run Options - Async Refresh
// =============================================================================

// WithAsyncRefresh enables async refresh with the specified interval.
// This is useful for components that update from goroutines.
var WithAsyncRefresh = bubbly.WithAsyncRefresh

// WithoutAsyncAutoDetect disables automatic async detection.
// By default, Run() auto-detects if async refresh is needed.
var WithoutAsyncAutoDetect = bubbly.WithoutAsyncAutoDetect
