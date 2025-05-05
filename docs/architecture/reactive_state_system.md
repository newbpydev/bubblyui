# Reactive State System Design

## Overview

This document outlines the design for BubblyUI's reactive state system, inspired by Solid.js signals. The goal is to create a fine-grained reactivity model that can efficiently update only the parts of the UI that need to change when state is modified.

## Core Concepts

### 1. Signals as Primitive

Signals form the foundation of our reactive system. A signal is a container for a value that can notify subscribers when the value changes.

```go
// Signal is a generic container for reactive values
type Signal[T any] struct {
    value      T
    dependencies map[string]struct{}
    subscribers map[string]func(T)
    id         string
    mutex      sync.RWMutex
}

// NewSignal creates a new signal with an initial value
func NewSignal[T any](initialValue T) *Signal[T] {
    return &Signal[T]{
        value:      initialValue,
        dependencies: make(map[string]struct{}),
        subscribers: make(map[string]func(T)),
        id:         uuid.New().String(),
    }
}
```

### 2. Dependency Tracking

The reactivity system requires automatic dependency tracking. When a component reads a signal's value, the signal should be registered as a dependency of the current tracking context.

```go
// Global tracking context
var (
    currentTracker     *DependencyTracker
    trackerMutex       sync.Mutex
)

// DependencyTracker tracks signal dependencies
type DependencyTracker struct {
    dependencies map[string]struct{}
    parent       *DependencyTracker
}

// WithTracking establishes a dependency tracking context 
// and executes the given function within that context
func WithTracking(fn func()) map[string]struct{} {
    tracker := &DependencyTracker{
        dependencies: make(map[string]struct{}),
    }
    
    trackerMutex.Lock()
    previousTracker := currentTracker
    currentTracker = tracker
    trackerMutex.Unlock()
    
    // Execute the function with dependency tracking
    fn()
    
    trackerMutex.Lock()
    currentTracker = previousTracker
    trackerMutex.Unlock()
    
    return tracker.dependencies
}

// Value returns the current value of the signal 
// and registers itself as a dependency
func (s *Signal[T]) Value() T {
    s.mutex.RLock()
    value := s.value
    s.mutex.RUnlock()
    
    // Record dependency
    trackerMutex.Lock()
    if currentTracker != nil {
        currentTracker.dependencies[s.id] = struct{}{}
        s.addDependent(currentTracker)
    }
    trackerMutex.Unlock()
    
    return value
}
```

### 3. Update Propagation

When a signal's value changes, all dependent computations and components need to be notified so they can update.

```go
// SetValue updates the signal value and notifies subscribers
func (s *Signal[T]) SetValue(newValue T) {
    s.mutex.Lock()
    
    // Skip if value hasn't changed (using deep equality)
    if reflect.DeepEqual(s.value, newValue) {
        s.mutex.Unlock()
        return
    }
    
    oldValue := s.value
    s.value = newValue
    subs := s.subscribers
    s.mutex.Unlock()
    
    // Notify all subscribers in a separate goroutine
    // to prevent blocking the caller
    go func(subs map[string]func(T), val T) {
        for _, notify := range subs {
            notify(val)
        }
    }(subs, newValue)
    
    // Schedule UI update
    bubblyui.ScheduleUpdate()
}
```

### 4. Batched Updates

To optimize performance, we'll use batched updates to coalesce multiple signal changes into a single UI update.

```go
var (
    updateScheduled   bool
    updateMutex       sync.Mutex
    pendingUpdates    map[string]struct{}
)

// ScheduleUpdate ensures that the UI will be updated after signal changes
func ScheduleUpdate() {
    updateMutex.Lock()
    defer updateMutex.Unlock()
    
    if !updateScheduled {
        updateScheduled = true
        
        // Use a microtask-like mechanism to batch updates
        go func() {
            // Small delay to collect multiple updates
            time.Sleep(16 * time.Millisecond) // ~60fps
            
            updateMutex.Lock()
            updateScheduled = false
            updates := pendingUpdates
            pendingUpdates = make(map[string]struct{})
            updateMutex.Unlock()
            
            // Process all pending updates
            processUpdates(updates)
        }()
    }
}
```

### 5. Derived Signals (Computed/Memo)

Derived signals (similar to Solid.js's `createMemo`) compute a value based on other signals and automatically update when dependencies change.

```go
// Computed is a signal whose value is derived from other signals
type Computed[T any] struct {
    Signal[T]
    compute       func() T
    dependencies  map[string]struct{}
}

// NewComputed creates a new computed signal
func NewComputed[T any](compute func() T) *Computed[T] {
    computed := &Computed[T]{
        compute:      compute,
        dependencies: make(map[string]struct{}),
    }
    
    // Initialize by running the computation once
    initialValue := computed.updateValue()
    computed.Signal = *NewSignal(initialValue)
    
    return computed
}

// updateValue recomputes the value and updates dependencies
func (c *Computed[T]) updateValue() T {
    // Track dependencies while computing
    deps := WithTracking(func() {
        newValue := c.compute()
        c.SetValue(newValue)
    })
    
    // Update dependency registrations
    c.updateDependencies(deps)
    
    return c.Value()
}
```

### 6. Effects for Side Effects

Effects allow for the execution of side effects when signals change, similar to React's `useEffect` or Solid.js's `createEffect`.

```go
// Effect represents a side effect that runs when its dependencies change
type Effect struct {
    id          string
    fn          func()
    dependencies map[string]struct{}
    cleanup     func()
}

// NewEffect creates a new effect
func NewEffect(fn func()) *Effect {
    effect := &Effect{
        id:          uuid.New().String(),
        fn:          fn,
        dependencies: make(map[string]struct{}),
    }
    
    // Run the effect once to establish initial dependencies
    effect.run()
    
    return effect
}

// run executes the effect and tracks dependencies
func (e *Effect) run() {
    // Run cleanup if present
    if e.cleanup != nil {
        e.cleanup()
        e.cleanup = nil
    }
    
    // Track dependencies while running
    deps := WithTracking(func() {
        // Capture cleanup function if provided
        var cleanupFn func()
        
        // Create scope with function to register cleanup
        cleanupScope := func(cleanup func()) {
            cleanupFn = cleanup
        }
        
        // Call effect with cleanup scope
        e.fn()
        
        // Store cleanup if provided
        e.cleanup = cleanupFn
    })
    
    // Update dependency registrations
    e.updateDependencies(deps)
}
```

## Signal Propagation Model

The signal propagation follows a directed graph-like model:

1. Signals hold references to dependents (computed values and effects)
2. When a signal changes, it notifies all its direct dependents
3. Updated computed values then notify their dependents, continuing the chain
4. Effects are always terminal nodes in the dependency graph

This creates a propagation flow:

```
Signal → Computed → Computed → ... → Effect
   ↓         ↓          ↓
 Effect    Effect     Effect
```

### Cycle Detection

To prevent infinite update loops, we'll implement cycle detection in the dependency graph:

```go
// Detect cycles in dependency graph
func detectCycles(startNode string, visited map[string]bool, stack map[string]bool, graph map[string][]string) bool {
    visited[startNode] = true
    stack[startNode] = true
    
    for _, neighbor := range graph[startNode] {
        if !visited[neighbor] {
            if detectCycles(neighbor, visited, stack, graph) {
                return true
            }
        } else if stack[neighbor] {
            return true // Cycle detected
        }
    }
    
    stack[startNode] = false
    return false
}
```

## Performance Optimization Strategies

1. **Equality Checking**: Skip updates when values haven't changed
2. **Batched Updates**: Coalesce multiple signal changes into single UI updates
3. **Lazy Evaluation**: Defer computation of derived values until needed
4. **Memoization**: Cache results of expensive computations
5. **Update Prioritization**: Process high-priority UI elements first

## Testing Approach

Testing will focus on:

1. **Signal Creation**: Verify signals are created correctly with initial values
2. **Dependency Tracking**: Ensure dependencies are properly tracked when values are read
3. **Update Propagation**: Confirm signals notify dependents when values change
4. **Derived Values**: Validate computed values update when dependencies change
5. **Batching**: Verify multiple updates are batched appropriately
6. **Performance Benchmarks**: Measure update propagation time with various graph complexities

## Next Steps

1. Implement a proof-of-concept signal implementation for testing
2. Develop test cases for the reactivity system
3. Benchmark performance with different update strategies
4. Integrate with the component rendering system
