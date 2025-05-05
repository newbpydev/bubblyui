package core

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Signal represents a reactive value container that tracks dependencies
// and notifies subscribers when the value changes.
type Signal[T any] struct {
	value    T
	version  uint64
	mutex    sync.RWMutex
	deps     map[string]Dependency
	equalsFn func(a, b T) bool
	id       string
}

// AddDependency adds a dependency to this signal
func (s *Signal[T]) AddDependency(id string, dep Dependency) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.deps[id] = dep
}

// Dependency represents an entity that depends on a signal
type Dependency interface {
	Notify()
}

// Effect represents a function that will be called when its dependencies change
type Effect struct {
	fn   func()
	deps []string
}

// AsyncEffect represents an effect that handles asynchronous operations
type AsyncEffect struct {
	fn   func()
	deps []string
}

var (
	// Global tracking context
	trackingStack   = []map[string]bool{}
	effectsRegistry = map[string]Dependency{}
	signalRegistry  = make(map[string]any)
	batchDepth      int32
	pendingEffects  = make(map[string]bool)
	idCounter       uint64
	errorHandler    func(error)
	globalMutex     sync.RWMutex
)

// NewSignal creates a new signal with the given initial value
func NewSignal[T any](initialValue T) *Signal[T] {
	return NewSignalWithEquals(initialValue, nil)
}

// NewSignalWithEquals creates a new signal with a custom equality function
func NewSignalWithEquals[T any](initialValue T, equalsFn func(a, b T) bool) *Signal[T] {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	signal := &Signal[T]{
		value:    initialValue,
		version:  0,
		deps:     make(map[string]Dependency),
		equalsFn: equalsFn,
		id:       fmt.Sprintf("signal_%d", atomic.AddUint64(&idCounter, 1)),
	}

	// Register the signal in the global registry
	signalRegistry[signal.id] = signal

	return signal
}

// Value gets the current value of the signal and records the dependency
// in the current tracking context if one exists
func (s *Signal[T]) Value() T {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Record dependency if we're currently tracking
	if len(trackingStack) > 0 {
		globalMutex.Lock()
		if len(trackingStack) > 0 { // Double-check under lock
			trackingContext := trackingStack[len(trackingStack)-1]
			trackingContext[s.id] = true
		}
		globalMutex.Unlock()
	}

	return s.value
}

// Set updates the value of the signal if it's different from the current value
// and notifies all dependencies
func (s *Signal[T]) Set(newValue T) {
	s.mutex.Lock()

	// Check if value is actually changing
	if s.equalsFn != nil {
		if s.equalsFn(s.value, newValue) {
			s.mutex.Unlock()
			return
		}
	} else {
		// For basic types we can do direct comparison
		// For complex types, we need to use reflection or a custom equality function
		// Users should provide custom equality functions for complex types
		isEqual := false

		// Use a deferred recover to handle potential panics from comparing incomparable types
		defer func() {
			if r := recover(); r != nil {
				isEqual = false
			}
		}()

		// Attempt safe comparison using fmt.Sprintf as a simple workaround
		// This is not ideal but will work for most cases
		// In a production system, use reflection or custom equality functions
		valueStr := fmt.Sprintf("%v", s.value)
		newValueStr := fmt.Sprintf("%v", newValue)
		isEqual = (valueStr == newValueStr)

		if isEqual {
			s.mutex.Unlock()
			return
		}
	}

	// For debugging
	// fmt.Printf("Signal %s changing from %v to %v\n", s.id, s.value, newValue)

	// Update the value and increment version
	s.value = newValue
	s.version++

	// Get the dependencies to notify
	deps := make([]Dependency, 0, len(s.deps))
	for _, dep := range s.deps {
		deps = append(deps, dep)
	}

	s.mutex.Unlock()

	// If we're in a batch, defer notifications
	if atomic.LoadInt32(&batchDepth) > 0 {
		globalMutex.Lock()
		for _, dep := range deps {
			switch e := dep.(type) {
			case *Effect:
				// Find the effect ID
				var effectID string
				for id, registered := range effectsRegistry {
					if registered == e {
						effectID = id
						break
					}
				}
				// Only add if we found the effect ID
				if effectID != "" {
					pendingEffects[effectID] = true
				}
			case *AsyncEffect:
				// Find the async effect ID
				var effectID string
				for id, registered := range effectsRegistry {
					if registered == e {
						effectID = id
						break
					}
				}
				// Only add if we found the effect ID
				if effectID != "" {
					pendingEffects[effectID] = true
				}
			}
		}
		globalMutex.Unlock()
	} else {
		// Otherwise notify immediately
		for _, dep := range deps {
			dep.Notify()
		}
	}
}

// StartTracking begins tracking signal dependencies
func StartTracking() {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	// Create a new tracking context
	trackingStack = append(trackingStack, make(map[string]bool))

	// For debugging
	// fmt.Println("Started tracking, depth:", len(trackingStack))
}

// StopTracking stops tracking signal dependencies and returns the collected dependencies
func StopTracking() []string {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	if len(trackingStack) == 0 {
		return []string{}
	}

	deps := trackingStack[len(trackingStack)-1]
	trackingStack = trackingStack[:len(trackingStack)-1]

	result := make([]string, 0, len(deps))
	for dep := range deps {
		result = append(result, dep)
	}

	return result
}

// RegisterEffect registers a function to be called when any of its dependencies change
func RegisterEffect(fn func(), deps []string) string {
	globalMutex.Lock()
	effectID := fmt.Sprintf("effect_%d", atomic.AddUint64(&idCounter, 1))

	effect := &Effect{
		fn:   fn,
		deps: deps,
	}

	effectsRegistry[effectID] = effect
	globalMutex.Unlock()

	// Register this effect with all signals it depends on
	for _, depID := range deps {
		registerDependencyWithSignal(depID, effectID, effect)
	}

	// Run the effect once to initialize - this should run outside of batching
	safelyRunEffect(fn)

	return effectID
}

// RegisterAsyncEffect registers an effect that performs asynchronous operations
func RegisterAsyncEffect(fn func(), deps []string) string {
	globalMutex.Lock()
	effectID := fmt.Sprintf("async_effect_%d", atomic.AddUint64(&idCounter, 1))

	effect := &AsyncEffect{
		fn:   fn,
		deps: deps,
	}

	effectsRegistry[effectID] = effect
	globalMutex.Unlock()

	// Register this effect with all signals it depends on
	for _, depID := range deps {
		registerDependencyWithSignal(depID, effectID, effect)
	}

	// Run the effect once to initialize
	safelyRunEffect(fn)

	return effectID
}

// Batch executes a function in a batched context, deferring signal notifications
// until the end of the batch to avoid cascading updates
func Batch(fn func()) {
	// Get the current batch depth before incrementing
	wasBatchingBefore := atomic.LoadInt32(&batchDepth) > 0

	// Increment batch depth
	atomic.AddInt32(&batchDepth, 1)

	// For debugging
	// currentDepth := atomic.LoadInt32(&batchDepth)
	// fmt.Printf("Batch start - depth: %d\n", currentDepth)

	// Run the function
	fn()

	// Decrement batch depth
	newDepth := atomic.AddInt32(&batchDepth, -1)

	// For debugging
	// fmt.Printf("Batch end - depth: %d\n", newDepth)

	// Only process pending effects if this was the outermost batch
	if newDepth == 0 && !wasBatchingBefore {
		// Process pending effects
		globalMutex.Lock()
		effectsToRun := make(map[string]bool)
		for effectID := range pendingEffects {
			effectsToRun[effectID] = true
		}
		pendingEffects = make(map[string]bool)
		globalMutex.Unlock()

		// For debugging
		// fmt.Printf("Running %d pending effects\n", len(effectsToRun))

		// Run all pending effects
		for effectID := range effectsToRun {
			globalMutex.RLock()
			effect, exists := effectsRegistry[effectID]
			globalMutex.RUnlock()

			if exists {
				effect.Notify()
			}
		}
	}
}

// SetErrorHandler sets a global error handler for effects
func SetErrorHandler(handler func(error)) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	errorHandler = handler
}

// Notify implements the Dependency interface for Effect
func (e *Effect) Notify() {
	safelyRunEffect(e.fn)
}

// Notify implements the Dependency interface for AsyncEffect
func (a *AsyncEffect) Notify() {
	safelyRunEffect(a.fn)
}

// safelyRunEffect runs an effect and catches any panics
func safelyRunEffect(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			globalMutex.RLock()
			handler := errorHandler
			globalMutex.RUnlock()

			if handler != nil {
				var err error
				switch v := r.(type) {
				case error:
					err = v
				case string:
					err = fmt.Errorf("%s", v)
				default:
					err = fmt.Errorf("%v", v)
				}
				handler(err)
			}
		}
	}()

	fn()
}

// registerDependencyWithSignal registers a dependency with a signal by ID
func registerDependencyWithSignal(signalID, effectID string, dep Dependency) {
	// Find the signal by ID
	globalMutex.RLock()
	signalObj, exists := signalRegistry[signalID]
	globalMutex.RUnlock()

	if !exists {
		// Signal not found, nothing to do
		return
	}

	// Type assertion to check what type of signal we have
	// This is a bit of a hack due to Go's type system limitations with generics
	// In a real implementation, we might have a different approach

	// We use reflection to call the appropriate method on the signal
	// This is a simplification for demonstration purposes
	switch s := signalObj.(type) {
	case interface{ AddDependency(string, Dependency) }:
		s.AddDependency(effectID, dep)
	default:
		// Unknown signal type, can't register dependency
		fmt.Printf("Warning: Unknown signal type, can't register dependency: %T\n", s)
	}
}
