package bubbly

import (
	"bytes"
	"errors"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
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

// trackingContext represents a single level of dependency tracking.
type trackingContext struct {
	dep  Dependency
	deps []Dependency
}

// trackingState holds the tracking state for a single goroutine.
// Each goroutine has its own isolated tracking state to prevent contention.
type trackingState struct {
	mu    sync.Mutex
	stack []*trackingContext
}

// getGoroutineID returns the ID of the current goroutine.
// This is extracted from the runtime stack trace and is used for per-goroutine tracking.
// Note: This is an internal implementation detail and should not be exposed publicly.
func getGoroutineID() uint64 {
	// Get stack trace with format: "goroutine 123 [running]:\n..."
	buf := make([]byte, 64)
	n := runtime.Stack(buf, false)
	buf = buf[:n]

	// Find "goroutine " prefix
	const prefix = "goroutine "
	idx := bytes.Index(buf, []byte(prefix))
	if idx == -1 {
		return 0
	}

	// Skip "goroutine " and parse the number
	buf = buf[idx+len(prefix):]

	// Find the space after the number
	spaceIdx := bytes.IndexByte(buf, ' ')
	if spaceIdx == -1 {
		return 0
	}

	// Parse the goroutine ID
	id, err := strconv.ParseUint(string(buf[:spaceIdx]), 10, 64)
	if err != nil {
		return 0
	}

	return id
}

// DepTracker manages dependency tracking during computed value evaluation.
// It maintains per-goroutine tracking state to prevent contention between goroutines.
// Each goroutine has its own isolated stack of tracking contexts.
//
// The tracker uses sync.Map for goroutine-local storage, eliminating the global
// mutex bottleneck that caused deadlocks with 100+ concurrent goroutines.
//
// Performance optimization: Uses atomic counter as fast-path filter to avoid
// expensive getGoroutineID() calls when no tracking is active anywhere.
type DepTracker struct {
	states         sync.Map     // map[uint64]*trackingState - goroutine ID -> tracking state
	activeTrackers atomic.Int32 // Count of active tracking contexts across all goroutines
}

// globalTracker is the global dependency tracker instance.
// Each goroutine will have its own tracking state managed through this instance.
var globalTracker = &DepTracker{}

// getOrCreateState returns the tracking state for the current goroutine,
// creating it if it doesn't exist.
func (dt *DepTracker) getOrCreateState() *trackingState {
	gid := getGoroutineID()

	// Try to load existing state
	if state, ok := dt.states.Load(gid); ok {
		return state.(*trackingState)
	}

	// Create new state
	state := &trackingState{
		stack: nil,
	}

	// Store and return (may race with another goroutine, but that's fine)
	actual, _ := dt.states.LoadOrStore(gid, state)
	return actual.(*trackingState)
}

// BeginTracking starts tracking dependencies for the current evaluation.
// It should be called before evaluating a computed function.
// Returns an error if circular dependency is detected or max depth is exceeded.
func (dt *DepTracker) BeginTracking(dep Dependency) error {
	state := dt.getOrCreateState()
	state.mu.Lock()
	defer state.mu.Unlock()

	// Check for circular dependency
	for _, ctx := range state.stack {
		if ctx.dep == dep {
			return ErrCircularDependency
		}
	}

	// Check max depth
	if len(state.stack) >= MaxDependencyDepth {
		return ErrMaxDepthExceeded
	}

	// Push new tracking context onto stack
	state.stack = append(state.stack, &trackingContext{
		dep:  dep,
		deps: nil,
	})

	// Increment global active trackers counter (atomic fast-path optimization)
	dt.activeTrackers.Add(1)

	return nil
}

// Track records a dependency access during evaluation.
// This is called by Ref.GetTyped() when dependency tracking is active.
func (dt *DepTracker) Track(dep Dependency) {
	state := dt.getOrCreateState()
	state.mu.Lock()
	defer state.mu.Unlock()

	if len(state.stack) == 0 {
		return
	}

	// Get current tracking context (top of stack)
	ctx := state.stack[len(state.stack)-1]

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
// If this is the last tracking context for the goroutine, the state is cleaned up
// to prevent memory leaks.
func (dt *DepTracker) EndTracking() []Dependency {
	gid := getGoroutineID()
	state := dt.getOrCreateState()
	state.mu.Lock()
	defer state.mu.Unlock()

	if len(state.stack) == 0 {
		return nil
	}

	// Pop from stack
	ctx := state.stack[len(state.stack)-1]
	state.stack = state.stack[:len(state.stack)-1]

	// Decrement global active trackers counter (atomic fast-path optimization)
	dt.activeTrackers.Add(-1)

	// Clean up goroutine state if stack is now empty (prevent memory leaks)
	if len(state.stack) == 0 {
		state.mu.Unlock() // Unlock before deleting
		dt.states.Delete(gid)
		state.mu.Lock() // Re-lock for defer
	}

	return ctx.deps
}

// IsTracking returns true if dependency tracking is currently active
// for the current goroutine.
//
// Performance optimization: Uses atomic fast-path check to avoid expensive
// getGoroutineID() call when no tracking is active anywhere in the system.
// This reduces Ref.GetTyped() from ~4600ns to ~26ns (178x faster, zero allocations).
func (dt *DepTracker) IsTracking() bool {
	// Fast path: if no trackers are active anywhere, return false immediately
	// This avoids the expensive getGoroutineID() call (runtime.Stack allocation)
	if dt.activeTrackers.Load() == 0 {
		return false
	}

	// Slow path: tracking is active somewhere, check if it's this goroutine
	gid := getGoroutineID()
	state, ok := dt.states.Load(gid)
	if !ok {
		return false
	}

	ts := state.(*trackingState)
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return len(ts.stack) > 0
}

// Reset clears all tracking state. This is primarily for testing.
func (dt *DepTracker) Reset() {
	dt.states.Range(func(key, value interface{}) bool {
		dt.states.Delete(key)
		return true
	})
}
