/*
Package bubbly provides a Vue-inspired reactive state management system for Go TUI applications.

# Overview

Bubbly offers type-safe reactive primitives built on Go generics that integrate seamlessly
with the Bubbletea framework's Elm architecture. It enables automatic UI updates through
reactive references, computed values, and watchers.

# Core Concepts

Bubbly provides three main reactive primitives:

  - Ref[T]: A reactive reference that holds a mutable value
  - Computed[T]: A derived value that automatically recomputes when dependencies change
  - Watch: Observe changes to reactive values and execute callbacks

All primitives are thread-safe and use efficient locking strategies for concurrent access.

# Quick Start

Create a reactive reference:

	count := bubbly.NewRef(0)
	value := count.GetTyped()  // Read: 0
	count.Set(42)         // Write: 42

Create a computed value:

	doubled := bubbly.NewComputed(func() int {
	    return count.GetTyped() * 2
	})
	result := doubled.GetTyped()  // Automatically recomputes when count changes

Watch for changes:

	cleanup := bubbly.Watch(count, func(newVal, oldVal int) {
	    fmt.Printf("Count changed: %d → %d\n", oldVal, newVal)
	})
	defer cleanup()  // Stop watching when done

# Integration with Bubbletea

Bubbly integrates naturally with Bubbletea's Update/View cycle:

	type model struct {
	    count *bubbly.Ref[int]
	}

	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	    switch msg := msg.(type) {
	    case tea.KeyMsg:
	        if msg.String() == "+" {
	            m.count.Set(m.count.GetTyped() + 1)
	        }
	    }
	    return m, nil
	}

	func (m model) View() string {
	    return fmt.Sprintf("Count: %d", m.count.GetTyped())
	}

# Component Model

BubblyUI provides a Vue-inspired component system that wraps Bubbletea's Model-Update-View
pattern with a declarative, composable API. Components encapsulate state, behavior, and
presentation into reusable units.

# Creating Components

Use the ComponentBuilder fluent API to create components:

	component, err := bubbly.NewComponent("Button").
	    Props(ButtonProps{Label: "Click me"}).
	    Setup(func(ctx *bubbly.Context) {
	        // Initialize state and handlers
	    }).
	    Template(func(ctx bubbly.RenderContext) string {
	        return "[Button]"
	    }).
	    Build()

The builder validates configuration and returns an error if required fields are missing.

# Props System

Props are immutable configuration data passed to components:

	type ButtonProps struct {
	    Label    string
	    Disabled bool
	}

	component, _ := bubbly.NewComponent("Button").
	    Props(ButtonProps{Label: "Submit", Disabled: false}).
	    Template(func(ctx bubbly.RenderContext) string {
	        props := ctx.Props().(ButtonProps)
	        return props.Label
	    }).
	    Build()

Props are read-only from the component's perspective. Use reactive state (Ref) for mutable data.

# Setup Function

The Setup function initializes component state and registers event handlers:

	Setup(func(ctx *bubbly.Context) {
	    // Create reactive state
	    count := ctx.Ref(0)

	    // Expose to template
	    ctx.Expose("count", count)

	    // Register event handler
	    ctx.On("increment", func(data interface{}) {
	        count.Set(count.GetTyped().(int) + 1)
	    })
	})

The Setup function runs once during component initialization (Init() call).

# Template Function

The Template function defines how the component renders:

	Template(func(ctx bubbly.RenderContext) string {
	    // Access exposed state
	    count := ctx.Get("count").(*bubbly.Ref[interface{}])

	    // Access props
	    props := ctx.Props().(ButtonProps)

	    // Render children
	    for _, child := range ctx.Children() {
	        output += ctx.RenderChild(child)
	    }

	    return fmt.Sprintf("Count: %d", count.GetTyped().(int))
	})

Templates are called on every View() invocation and should be pure functions.

# Event System

Components communicate through events:

Emitting events:

	component.Emit("buttonClicked", map[string]interface{}{
	    "timestamp": time.Now(),
	})

Handling events:

	component.On("buttonClicked", func(data interface{}) {
	    fmt.Println("Button was clicked!")
	})

Events bubble up from child to parent components automatically.

# Component Composition

Nest components to build complex UIs:

	child := bubbly.NewComponent("Child").
	    Template(func(ctx bubbly.RenderContext) string {
	        return "Child"
	    }).
	    Build()

	parent := bubbly.NewComponent("Parent").
	    Children(child).
	    Template(func(ctx bubbly.RenderContext) string {
	        output := "Parent:\n"
	        for _, c := range ctx.Children() {
	            output += ctx.RenderChild(c) + "\n"
	        }
	        return output
	    }).
	    Build()

Children are initialized and updated automatically by the parent component.

# Complete Component Example

A stateful counter component:

	counter, _ := bubbly.NewComponent("Counter").
	    Setup(func(ctx *bubbly.Context) {
	        // Reactive state
	        count := ctx.Ref(0)
	        ctx.Expose("count", count)

	        // Event handlers
	        ctx.On("increment", func(data interface{}) {
	            count.Set(count.GetTyped().(int) + 1)
	        })
	        ctx.On("decrement", func(data interface{}) {
	            count.Set(count.GetTyped().(int) - 1)
	        })
	    }).
	    Template(func(ctx bubbly.RenderContext) string {
	        count := ctx.Get("count").(*bubbly.Ref[interface{}])
	        return fmt.Sprintf("Count: %d", count.GetTyped().(int))
	    }).
	    Build()

	// Use with Bubbletea
	counter.Init()
	counter.Emit("increment", nil)
	fmt.Println(counter.View()) // "Count: 1"

# Advanced Features

Deep Watching:

Watch nested struct changes using reflection-based comparison:

	user := bubbly.NewRef(User{Name: "John", Age: 30})
	bubbly.Watch(user, func(n, o User) {
	    fmt.Println("User changed")
	}, bubbly.WithDeep())

Custom Comparators:

Provide custom comparison logic for complex types:

	comparator := func(a, b User) bool {
	    return a.Name == b.Name  // Only compare names
	}
	bubbly.Watch(user, callback, bubbly.WithDeepCompare(comparator))

Async Flush Modes:

Batch multiple changes into a single callback execution:

	bubbly.Watch(count, callback, bubbly.WithFlush("post"))
	count.Set(1)  // Queued
	count.Set(2)  // Replaces previous
	count.Set(3)  // Replaces previous
	bubbly.FlushWatchers()  // Executes callback once with final value (3)

Immediate Execution:

Execute callback immediately on watcher creation:

	bubbly.Watch(count, callback, bubbly.WithImmediate())

# Performance

Bubbly is designed for high performance:

  - Ref.GetTyped(): ~26 ns/op with zero allocations
  - Ref.Set(): ~38 ns/op with zero allocations (no watchers)
  - Thread-safe with RWMutex for read-heavy workloads
  - Computed values cache results until dependencies change
  - Watchers are notified outside locks to prevent deadlocks

# Error Handling

Bubbly panics on programming errors to catch bugs early:

  - ErrNilCallback: Watch() called with nil callback
  - ErrNilComputeFn: NewComputed() called with nil function
  - ErrCircularDependency: Circular dependency detected in computed values
  - ErrMaxDepthExceeded: Dependency chain exceeds 100 levels

# Thread Safety

All operations are thread-safe:

  - Ref.GetTyped() and Ref.Set() use RWMutex for concurrent access
  - Computed.GetTyped() uses double-checked locking for cache validation
  - Watch callbacks are executed outside locks to prevent deadlocks
  - Multiple watchers can safely observe the same Ref

# Migration from Manual State

Before (manual state):

	type model struct {
	    count int
	}

	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	    m.count++
	    return m, nil
	}

After (reactive state):

	type model struct {
	    count *bubbly.Ref[int]
	}

	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	    m.count.Set(m.count.GetTyped() + 1)
	    return m, nil
	}

Benefits:
  - Automatic dependency tracking
  - Computed values update automatically
  - Watchers enable side effects
  - Type-safe with generics
  - Thread-safe by default

# Examples

See the example_test.go file for runnable examples demonstrating:
  - Basic Ref operations
  - Computed value chains
  - Watcher patterns
  - Deep watching
  - Flush modes
  - Bubbletea integration

# Package Structure

The package is organized into focused files:

  - ref.go: Reactive references (Ref[T])
  - computed.go: Computed values (Computed[T])
  - watch.go: Watchers and options
  - tracker.go: Dependency tracking system
  - scheduler.go: Async flush scheduler
  - errors.go: Error definitions (in tracker.go)

# Design Philosophy

Bubbly follows these principles:

  - Type Safety: Leverage Go generics for compile-time type checking
  - Simplicity: Clean, intuitive API inspired by Vue 3
  - Performance: Optimize hot paths with zero allocations
  - Safety: Thread-safe by default, panic on programming errors
  - Integration: Seamless integration with Bubbletea's architecture

# Compatibility

  - Requires Go 1.22+ (generics)
  - Compatible with Bubbletea v0.24+
  - No external dependencies beyond standard library

# License

See the LICENSE file in the repository root.
*/
package bubbly
