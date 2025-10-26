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
	value := count.Get()  // Read: 0
	count.Set(42)         // Write: 42

Create a computed value:

	doubled := bubbly.NewComputed(func() int {
	    return count.Get() * 2
	})
	result := doubled.Get()  // Automatically recomputes when count changes

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
	            m.count.Set(m.count.Get() + 1)
	        }
	    }
	    return m, nil
	}

	func (m model) View() string {
	    return fmt.Sprintf("Count: %d", m.count.Get())
	}

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

  - Ref.Get(): ~26 ns/op with zero allocations
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

  - Ref.Get() and Ref.Set() use RWMutex for concurrent access
  - Computed.Get() uses double-checked locking for cache validation
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
	    m.count.Set(m.count.Get() + 1)
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
