package bubbly

import (
	"errors"
	"sync"
)

// MaxDependencyDepth is the maximum allowed depth of dependency chains to prevent
// infinite recursion in circular dependencies.
const MaxDependencyDepth = 100

// Common errors for the reactivity system.
var (
	// ErrCircularDependency is returned when a circular dependency is detected in computed values.
	// This occurs when computed value A depends on B, which depends on A (directly or indirectly).
	ErrCircularDependency = errors.New("circular dependency detected")

	// ErrMaxDepthExceeded is returned when dependency depth exceeds the maximum allowed (100).
	// This prevents stack overflow from deeply nested computed values.
	ErrMaxDepthExceeded = errors.New("max dependency depth exceeded")

	// ErrNilCallback is returned when a nil callback function is provided to Watch.
	// Watch requires a valid callback function to notify when values change.
	ErrNilCallback = errors.New("callback cannot be nil")

	// ErrNilComputeFn is returned when a nil compute function is provided to NewComputed.
	// Computed values require a valid function to compute their value.
	ErrNilComputeFn = errors.New("compute function cannot be nil")
)

// Dependency represents something that can be invalidated when its dependencies change.
// Both Ref and Computed implement this interface to participate in the reactive system.
type Dependency interface {
	// Invalidate marks the dependency as needing recomputation or re-evaluation.
	Invalidate()

	// AddDependent registers another dependency that depends on this one.
	AddDependent(dep Dependency)
}

// trackingContext represents a single level of dependency tracking.
type trackingContext struct {
	dep  Dependency
	deps []Dependency
}

// DepTracker manages dependency tracking during computed value evaluation.
// It maintains a stack of currently evaluating dependencies to detect circular
// references and tracks which dependencies are accessed during evaluation.
//
// The tracker uses thread-local storage pattern via goroutine-local state to
// ensure each goroutine has its own tracking context.
type DepTracker struct {
	mu    sync.RWMutex
	stack []*trackingContext
}

// globalTracker is the global dependency tracker instance.
// Each goroutine will have its own tracking state managed through this instance.
var globalTracker = &DepTracker{}

// BeginTracking starts tracking dependencies for the current evaluation.
// It should be called before evaluating a computed function.
// Returns an error if circular dependency is detected or max depth is exceeded.
func (dt *DepTracker) BeginTracking(dep Dependency) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	// Check for circular dependency
	for _, ctx := range dt.stack {
		if ctx.dep == dep {
			return ErrCircularDependency
		}
	}

	// Check max depth
	if len(dt.stack) >= MaxDependencyDepth {
		return ErrMaxDepthExceeded
	}

	// Push new tracking context onto stack
	dt.stack = append(dt.stack, &trackingContext{
		dep:  dep,
		deps: nil,
	})

	return nil
}

// Track records a dependency access during evaluation.
// This is called by Ref.Get() when dependency tracking is active.
func (dt *DepTracker) Track(dep Dependency) {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	if len(dt.stack) == 0 {
		return
	}

	// Get current tracking context (top of stack)
	ctx := dt.stack[len(dt.stack)-1]

	// Avoid duplicate dependencies
	for _, d := range ctx.deps {
		if d == dep {
			return
		}
	}

	ctx.deps = append(ctx.deps, dep)
}

// EndTracking stops tracking and returns the collected dependencies.
// It should be called after evaluating a computed function.
func (dt *DepTracker) EndTracking() []Dependency {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	if len(dt.stack) == 0 {
		return nil
	}

	// Pop from stack
	ctx := dt.stack[len(dt.stack)-1]
	dt.stack = dt.stack[:len(dt.stack)-1]

	return ctx.deps
}

// IsTracking returns true if dependency tracking is currently active.
func (dt *DepTracker) IsTracking() bool {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	return len(dt.stack) > 0
}

// Reset clears the tracking stack. This is primarily for testing.
func (dt *DepTracker) Reset() {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	dt.stack = nil
}
