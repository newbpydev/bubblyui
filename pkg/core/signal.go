package core

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Signal represents a reactive value container that tracks dependencies
// and notifies subscribers when the value changes.
type Signal[T any] struct {
	value       T                      // The current value of the signal
	version     uint64                 // Version counter for tracking updates
	mutex       sync.RWMutex           // Mutex for thread-safe access
	deps        map[string]Dependency  // Dependencies that need to be notified on updates
	equalsFn    func(a, b T) bool      // Custom equality function for value comparison
	id          string                 // Unique identifier for this signal
	lastAccess  int64                  // Timestamp of last access for tracking purposes
	accessCount uint64                 // Number of times this signal has been accessed
	writes      uint64                 // Number of write operations performed on this signal
	createdAt   time.Time              // Time when the signal was created
	metadata    map[string]interface{} // Metadata for debugging and tracing purposes
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
	fn        func()
	deps      []string
	debugInfo string // Additional debugging information
}

// AsyncEffect represents an effect that handles asynchronous operations
type AsyncEffect struct {
	fn   func()
	deps []string
}

// EffectPriority represents the priority level of an effect
type EffectPriority int

// Priority constants for effect scheduling
const (
	PriorityLow    EffectPriority = 0
	PriorityNormal EffectPriority = 50
	PriorityHigh   EffectPriority = 100
)

// EffectInfo contains metadata about an effect
type EffectInfo struct {
	Priority    EffectPriority
	IsCancelled bool
	IsDeferred  bool
	BatchID     string      // Used to group related effects
	Metadata    interface{} // Any additional metadata
}

var (
	// Global tracking context
	trackingStack   = []map[string]bool{}
	effectsRegistry = map[string]Dependency{}
	signalRegistry  = make(map[string]any)
	batchDepth      int32
	batchMode       bool
	batchedSignals  = make(map[string]any)
	pendingEffects  = make(map[string]bool)
	// Effect scheduling information
	effectInfos     = make(map[string]EffectInfo)
	effectQueue     = make([]string, 0)
	processingQueue bool
	idCounter       uint64
	errorHandler    func(error)
	globalMutex     sync.RWMutex
	debugMode       bool // Flag to enable/disable debug logging
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
		value:       initialValue,
		version:     0,
		deps:        make(map[string]Dependency),
		equalsFn:    equalsFn,
		id:          fmt.Sprintf("signal_%d", atomic.AddUint64(&idCounter, 1)),
		createdAt:   time.Now(),
		writes:      0,
		accessCount: 0,
		lastAccess:  time.Now().UnixNano(),
		metadata:    make(map[string]interface{}),
	}

	// Register the signal in the global registry
	signalRegistry[signal.id] = signal

	return signal
}

// Value gets the current value of the signal and records the dependency
// in the current tracking context if one exists
func (s *Signal[T]) Value() T {
	// Update access statistics with atomic operations
	atomic.AddUint64(&s.accessCount, 1)
	atomic.StoreInt64(&s.lastAccess, time.Now().UnixNano())

	// Fast path: If no active tracking context, we can do a simple read
	if len(trackingStack) == 0 {
		s.mutex.RLock()
		value := s.value
		s.mutex.RUnlock()
		return value
	}

	// When tracking dependencies, we need to register this signal
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Record dependency if we're currently tracking
	globalMutex.Lock()
	if len(trackingStack) > 0 { // Double-check under lock
		trackingContext := trackingStack[len(trackingStack)-1]
		trackingContext[s.id] = true
	}
	globalMutex.Unlock()

	return s.value
}

// tryLock attempts to acquire the mutex lock with a timeout
// Returns true if the lock was acquired, false otherwise
func (s *Signal[T]) tryLock(timeout time.Duration) bool {
	// Create a buffered channel to avoid goroutine leak
	done := make(chan struct{}, 1)

	go func() {
		s.mutex.Lock()
		done <- struct{}{} // Signal that we got the lock
	}()

	select {
	case <-done:
		return true
	case <-time.After(timeout):
		// The lock attempt timed out
		return false
	}
}

// Set updates the signal value if it differs from the current value
// and notifies all dependencies
func (s *Signal[T]) Set(newValue T) {
	// Record write operation atomically
	atomic.AddUint64(&s.writes, 1)

	// Use standard mutex locking for better reliability
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if value is actually changing
	isEqual := false
	if s.equalsFn != nil {
		// Use custom equality function if provided
		isEqual = s.equalsFn(s.value, newValue)
	} else {
		// Safe basic equality check with panic recovery
		func() {
			// Use a defer-recover to handle potential panics during comparison
			defer func() {
				if r := recover(); r != nil {
					// If we panic during comparison, values are not equal
					isEqual = false
				}
			}()

			// Simple string-based comparison as a fallback
			valueStr := fmt.Sprintf("%v", s.value)
			newValueStr := fmt.Sprintf("%v", newValue)
			isEqual = (valueStr == newValueStr)
		}()
	}

	// If values are equal, no need to update
	if isEqual {
		return
	}

	// Debug logging
	if debugMode {
		fmt.Printf("[DEBUG] Signal %s changing from %v to %v\n", s.id, s.value, newValue)
	}

	// Update the value and increment version atomically
	s.value = newValue
	atomic.AddUint64(&s.version, 1)

	// Copy dependencies to notify to avoid any potential deadlocks
	// from holding the lock during notification
	deps := make([]Dependency, 0, len(s.deps))
	for _, dep := range s.deps {
		deps = append(deps, dep)
	}

	// Check if we're in batch mode
	globalMutex.RLock()
	inBatch := batchMode
	globalMutex.RUnlock()

	// Release the lock before notifying dependencies to prevent potential deadlocks
	s.mutex.Unlock()

	// If in batch mode, queue for later processing
	if inBatch {
		globalMutex.Lock()
		batchedSignals[s.id] = s
		for _, dep := range deps {
			if effect, ok := dep.(*Effect); ok {
				// Find the effect ID
				for id, registered := range effectsRegistry {
					if registered == effect {
						pendingEffects[id] = true
						break
					}
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

	// Re-acquire the lock to maintain the deferred unlock semantics
	// This ensures our defer s.mutex.Unlock() doesn't cause issues
	s.mutex.Lock()
}

// notifyDependents notifies all dependencies that this signal has changed
func (s *Signal[T]) notifyDependents() {
	// Copy dependencies to avoid deadlocks
	s.mutex.RLock()
	deps := make([]Dependency, 0, len(s.deps))
	for _, dep := range s.deps {
		deps = append(deps, dep)
	}
	s.mutex.RUnlock()

	// Debug logging
	if debugMode {
		fmt.Printf("[DEBUG] Signal %s notifying %d dependencies\n", s.id, len(deps))
	}

	// Check if we're currently in a batch update
	globalMutex.RLock()
	inBatch := atomic.LoadInt32(&batchDepth) > 0
	globalMutex.RUnlock()

	// If in batch mode, queue notifications for later
	if inBatch {
		globalMutex.Lock()
		// Add this signal to the batched signals map
		batchedSignals[s.id] = s

		// Register effects for later execution
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
	if debugMode {
		fmt.Printf("[DEBUG] Started tracking, depth: %d\n", len(trackingStack))
	}
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

	// Debug logging
	if debugMode {
		fmt.Printf("[DEBUG] Stopped tracking, collected %d dependencies\n", len(result))
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

// RegisterEffectWithoutInitialRun registers a function to be called when any of its dependencies change,
// but does not run the effect immediately. This is useful for computed signals that have already
// calculated their initial value.
func RegisterEffectWithoutInitialRun(fn func(), deps []string) string {
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

	// Do not run the effect immediately
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

	// Set batch mode flag
	globalMutex.Lock()
	batchMode = true
	globalMutex.Unlock()

	// Debug logging
	if debugMode {
		currentDepth := atomic.LoadInt32(&batchDepth)
		fmt.Printf("[DEBUG] Batch start - depth: %d\n", currentDepth)
	}

	// Run the function
	fn()

	// Decrement batch depth
	newDepth := atomic.AddInt32(&batchDepth, -1)

	// Debug logging
	if debugMode {
		fmt.Printf("[DEBUG] Batch end - depth: %d\n", newDepth)
	}

	// Only process pending effects if this was the outermost batch
	if newDepth == 0 && !wasBatchingBefore {
		// Reset batch mode
		globalMutex.Lock()
		batchMode = false

		// Process batched signals
		processedSignals := make(map[string]bool)
		for id := range batchedSignals {
			processedSignals[id] = true
		}
		batchedSignals = make(map[string]any)

		// Process pending effects
		effectsToRun := make(map[string]bool)
		for effectID := range pendingEffects {
			effectsToRun[effectID] = true
		}
		pendingEffects = make(map[string]bool)
		globalMutex.Unlock()

		// Debug logging
		if debugMode {
			fmt.Printf("[DEBUG] Running %d pending effects\n", len(effectsToRun))
		}

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
				default:
					err = fmt.Errorf("%v", v)
				}
				handler(err)
			} else {
				// If no custom handler, print to stderr
				fmt.Fprintf(os.Stderr, "Unhandled panic in effect: %v\n", r)
			}
		}
	}()

	fn()
}

// GetStats returns statistics about the signal
func (s *Signal[T]) GetStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stats := map[string]interface{}{
		"id":              s.id,
		"version":         s.version,
		"accessCount":     s.accessCount,
		"writeCount":      s.writes,
		"createdAt":       s.createdAt,
		"age":             time.Since(s.createdAt).String(),
		"dependencyCount": len(s.deps),
	}

	// Add any custom metadata
	for k, v := range s.metadata {
		stats[k] = v
	}

	return stats
}

// GetMetadata returns the metadata associated with this signal
func (s *Signal[T]) GetMetadata(key string) (interface{}, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, exists := s.metadata[key]
	return value, exists
}

// SetMetadata sets a metadata value on this signal
func (s *Signal[T]) SetMetadata(key string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.metadata[key] = value
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
