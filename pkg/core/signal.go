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
	mutex       sync.Mutex            // Mutex for thread-safe access
	deps        map[string]Dependency  // Dependencies that need to be notified on updates
	equalsFn    func(a, b T) bool      // Custom equality function for value comparison
	id          string                 // Unique identifier for this signal
	lastAccess  int64                  // Timestamp of last access for tracking
	accessCount uint64                 // Number of times this signal has been accessed
	writes      uint64                 // Number of write operations performed on this signal
	createdAt   time.Time              // Time when the signal was created
	metadata    map[string]interface{} // Metadata for debugging and tracing purposes
}

// SetMetadata sets metadata for the signal
func (s *Signal[T]) SetMetadata(key string, value interface{}) {
	s.mutex.Lock()
	if s.metadata == nil {
		s.metadata = make(map[string]interface{})
	}
	s.metadata[key] = value
	s.mutex.Unlock()
}

// GetMetadata retrieves metadata for the signal
func (s *Signal[T]) GetMetadata(key string) (interface{}, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.metadata == nil {
		return nil, false
	}
	value, exists := s.metadata[key]
	return value, exists
}

// GetStats returns statistics about the signal's usage
func (s *Signal[T]) GetStats() map[string]interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return map[string]interface{}{
		"version":      s.version,
		"accessCount":  s.accessCount,
		"writes":       s.writes,
		"createdAt":    s.createdAt,
		"lastAccess":   s.lastAccess,
		"dependencies": len(s.deps),
	}
}

// AddDependency adds a dependency to this signal
func (s *Signal[T]) AddDependency(id string, dep Dependency) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.deps[id] = dep
}

// RemoveDependency removes a dependency from this signal by ID
func (s *Signal[T]) RemoveDependency(effectID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.deps != nil {
		delete(s.deps, effectID)
	}
}

// AddDependent adds an effect or other dependency to this signal
// This will be called by the dependency tracking system when an effect
// accesses this signal's value
func (s *Signal[T]) AddDependent(effectID string, dep Dependency) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.deps == nil {
		s.deps = make(map[string]Dependency)
	}

	s.deps[effectID] = dep

	if debugMode {
		fmt.Printf("[DEBUG] Signal %s added dependent %s\n", s.id, effectID)
	}
}

// Dependency represents an entity that depends on a signal
type Dependency interface {
	Notify()
}

// Effect represents a function that will be called when its dependencies change
type Effect struct {
	fn        func()
	deps      []string
	debugInfo string //nolint:unused // Additional debugging information - will be used in future development
}

// Notify implements the Dependency interface
func (e *Effect) Notify() {
	safelyRunEffect(e.fn)
}

// DebugInfo returns debugging information about this effect
func (e *Effect) DebugInfo() string {
	return e.debugInfo
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
	effectInfos     = make(map[string]EffectInfo) //nolint:unused // Will be used for effect scheduling in future implementation
	effectQueue     = make([]string, 0)           //nolint:unused // Will be used for effect scheduling in future implementation
	processingQueue bool                          //nolint:unused // Will be used for effect scheduling in future implementation
	idCounter       uint64
	errorHandler    func(error)
	globalMutex     sync.RWMutex
	debugMode       bool // Flag to enable/disable debug logging
	effectDeps      = make(map[string][]string) // Track dependencies for each effect
)

// NewSignal creates a new signal with the given initial value
func NewSignal[T any](initialValue T) *Signal[T] {
	return NewSignalWithEquals(initialValue, nil)
}

// NewSignalWithEquals creates a new signal with a custom equality function
func NewSignalWithEquals[T any](initialValue T, equalsFn func(a, b T) bool) *Signal[T] {
	globalMutex.Lock()
	s := &Signal[T]{
		value:       initialValue,
		version:     1,
		deps:        make(map[string]Dependency),
		equalsFn:    equalsFn,
		id:          fmt.Sprintf("signal_%d", atomic.AddUint64(&idCounter, 1)),
		createdAt:   time.Now(),
		lastAccess:  time.Now().UnixNano(),
		metadata:    make(map[string]interface{}),
	}
	globalMutex.Unlock()

	// Register the signal in the global registry
	globalMutex.Lock()
	signalRegistry[s.id] = s
	if debugMode {
		fmt.Printf("[DEBUG] Registered signal %s. Registry keys: ", s.id)
		for k := range signalRegistry {
			fmt.Printf("%s ", k)
		}
		fmt.Println()
	}
	globalMutex.Unlock()

	// Set the signal id as top-level metadata
	s.SetMetadata("id", s.id)

	return s
}

// Value returns the current value of the signal and registers it as a dependency
// in the current tracking context if one exists
func (s *Signal[T]) Value() T {
	// Update access statistics with atomic operations
	atomic.AddUint64(&s.accessCount, 1)
	atomic.StoreInt64(&s.lastAccess, time.Now().UnixNano())

	// Get the value first without holding the mutex
	var value T
	{
		s.mutex.Lock()
		value = s.value
		s.mutex.Unlock()
	}

	// Register this signal as a dependency if we're in a tracking context
	// This uses the AddSignalDependency function from signal_factory.go
	AddSignalDependency(s.id)

	// If we're tracking dependencies, record this signal
	if len(trackingStack) > 0 {
		// Use a separate goroutine to handle the dependency registration
		// to prevent potential deadlocks
		go func() {
			globalMutex.Lock()
			trackingContext := trackingStack[len(trackingStack)-1]
			trackingContext[s.id] = true
			globalMutex.Unlock()
		}()
	}

	return value
}

// tryLock attempts to acquire the mutex lock with a timeout
// Returns true if the lock was acquired, false otherwise
//
//nolint:unused // Will be used for advanced synchronization in future implementation
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
		s.mutex.Unlock()
		return
	}

	// Debug logging
	if debugMode {
		fmt.Printf("[DEBUG] Signal %s changing from %v to %v\n", s.id, s.value, newValue)
	}

	// Update the value and increment version atomically
	s.value = newValue
	atomic.AddUint64(&s.version, 1)

	// Copy dependencies to notify, then unlock
	depsCopy := make([]Dependency, 0, len(s.deps))
	for _, dep := range s.deps {
		depsCopy = append(depsCopy, dep)
	}
	s.mutex.Unlock()

	// Notify dependencies synchronously (no goroutine)
	for _, dep := range depsCopy {
		dep.Notify()
	}
}

// notifyDependents notifies all dependencies that this signal has changed
func (s *Signal[T]) notifyDependents() {
	// Get a copy of all dependencies to notify (copy both IDs and Dependency values)
	depsCopy := make([]Dependency, 0, len(s.deps))
	{
		s.mutex.Lock()
		for effectID, dep := range s.deps {
			_ = effectID // effectID is available for future use
			depsCopy = append(depsCopy, dep)
		}
		s.mutex.Unlock()
	}

	// Debug logging
	if debugMode {
		fmt.Printf("[DEBUG] Signal %s notifying %d dependencies\n", s.id, len(depsCopy))
	}

	// Check if we're currently in a batch update
	globalMutex.RLock()
	inBatch := atomic.LoadInt32(&batchDepth) > 0
	globalMutex.RUnlock()

	// If we're in batch mode, queue the signal for batch processing
	if inBatch {
		globalMutex.Lock()
		batchedSignals[s.id] = s

		// Queue all effects for later processing
		for effectID, dep := range s.deps {
			if _, ok := dep.(*Effect); ok {
				pendingEffects[effectID] = true
			}
		}

		// Log summary
		if debugMode {
			fmt.Printf("[DEBUG] Signal %s queued for batch processing\n", s.id)
		}
		globalMutex.Unlock()
	} else {
		// Otherwise notify immediately
		if debugMode {
			fmt.Printf("[DEBUG] Not in batch mode, notifying all %d dependencies immediately\n", len(depsCopy))
		}

		// Notify dependencies synchronously (no goroutine)
		for _, dep := range depsCopy {
			dep.Notify()
		}
	}
}


// ... (rest of the code remains the same)
// but does not run the effect immediately. This is useful for computed signals that have already
// calculated their initial value.
// func RegisterEffectWithoutInitialRun(fn func(), deps []string, effectID ...string) string {
// Implementation moved to signal_factory.go
// }

// RegisterAsyncEffect registers an effect that performs asynchronous operations
func RegisterAsyncEffect(fn func(), deps []string, effectID ...string) string {
	id := ""
	if len(effectID) > 0 && effectID[0] != "" {
		id = effectID[0]
	} else {
		id = fmt.Sprintf("async_effect_%d", atomic.AddUint64(&idCounter, 1))
	}
	registerEffectWithID(fn, deps, id)
	// Run the effect once to initialize
	safelyRunEffect(fn)
	return id
}

// Batch executes a function in a batched context, deferring signal notifications
// until the end of the batch to avoid cascading updates
func Batch(fn func()) {
	// Increment batch depth
	newDepth := atomic.AddInt32(&batchDepth, 1)

	// Debug logging
	if debugMode {
		fmt.Printf("[DEBUG] Batch start - depth: %d\n", newDepth)
	}

	// Run the function
	fn()

	// Decrement batch depth
	newDepth = atomic.AddInt32(&batchDepth, -1)

	// Debug logging
	if debugMode {
		fmt.Printf("[DEBUG] Batch end - depth: %d\n", newDepth)
	}

	// Process pending effects if this was the outermost batch
	if newDepth == 0 {
		// Process batched signals and effects atomically
		globalMutex.Lock()
		defer globalMutex.Unlock()

		// Reset batch mode
		batchMode = false

		// Process batched signals - only trigger one update per signal
		processedSignals := make(map[string]bool)
		for id := range batchedSignals {
			if !processedSignals[id] {
				processedSignals[id] = true
			}
		}
		batchedSignals = make(map[string]any)

		// Process pending effects - only run one update per effect
		effectsToRun := make(map[string]bool)
		for effectID := range pendingEffects {
			if !effectsToRun[effectID] {
				effectsToRun[effectID] = true
			}
		}
		pendingEffects = make(map[string]bool)

		// Debug logging
		if debugMode {
			fmt.Printf("[DEBUG] Running %d pending effects\n", len(effectsToRun))
		}

		// Run all pending effects while holding the lock
		for effectID := range effectsToRun {
			effect, exists := effectsRegistry[effectID]
			if exists {
				effect.Notify()
			}
		}

		// Flush batched state notifications
		batchedStateChangesMu.Lock()
		for statePtr, v := range batchedStateChanges {
			switch s := statePtr.(type) {
			case *State[any]:
				callbacks := make([]func(old, new any), len(s.onChange))
				copy(callbacks, s.onChange)
				for _, cb := range callbacks {
					cb(v.old, v.new)
				}
			default:
				// Try to call onChange for generic State[T] via reflection
				// This is a fallback for generic support
			}
		}
		batchedStateChanges = make(map[any]struct{old, new any})
		batchedStateChangesMu.Unlock()
	}
}

// RemoveEffectFromSignals removes an effect from all signals in the given dependency list.
func RemoveEffectFromSignals(effectID string, deps []string) {
	for _, depID := range deps {
		if signal, ok := signalRegistry[depID]; ok {
			// Use type assertion to call RemoveDependency if available
			switch s := signal.(type) {
			case interface{ RemoveDependency(string) }:
				s.RemoveDependency(effectID)
			}
		}
	}
}

// RemoveEffect removes an effect from the registry and performs cleanup
func RemoveEffect(effectID interface{}) {
	id, ok := effectID.(string)
	if !ok || id == "" {
		// Ignore nil or non-string effectIDs
		return
	}
	globalMutex.Lock()
	deps, hasDeps := effectDeps[id]
	if hasDeps {
		globalMutex.Unlock() // Unlock before calling RemoveEffectFromSignals (may lock internally)
		RemoveEffectFromSignals(id, deps)
		globalMutex.Lock()
		delete(effectDeps, id)
	}
	if _, exists := effectsRegistry[id]; exists {
		delete(effectsRegistry, id)
		// Additional cleanup logic can be added here if needed
	}
	globalMutex.Unlock()
}

// SetErrorHandler sets a global error handler for effects
func SetErrorHandler(handler func(error)) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	errorHandler = handler
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
