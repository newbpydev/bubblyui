package core

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"
)

// Signal implementation is in signal.go
// This file contains factory functions for the reactive state system

// SignalOptions defines the configuration options for creating a signal
type SignalOptions struct {
	// Custom equality function for comparing values
	Equals func(a, b any) bool

	// Debug information
	DebugName  string
	SourceFile string
	SourceLine int
}

// CreateSignal creates a reactive signal with the given initial value
// and optional configuration options
func CreateSignal[T any](initialValue T, options ...SignalOptions) *Signal[T] {
	var opts SignalOptions
	if len(options) > 0 {
		opts = options[0]
	} else {
		// Capture caller information for debugging by default
		_, file, line, _ := runtime.Caller(1)
		opts = SignalOptions{
			SourceFile: file,
			SourceLine: line,
		}
	}

	// Create a generic equals function wrapper if one was provided
	var equalsFn func(a, b T) bool
	if opts.Equals != nil {
		equalsFn = func(a, b T) bool {
			return opts.Equals(a, b)
		}
	}

	// Create the signal with the appropriate config
	signal := NewSignalWithEquals(initialValue, equalsFn)
	// Register the signal in the global registry with its id
	globalMutex.Lock()
	signalRegistry[signal.id] = signal
	globalMutex.Unlock()

	// Store debug information
	setSignalMetadata(signal, map[string]interface{}{
		"debugName":  opts.DebugName,
		"sourceFile": opts.SourceFile,
		"sourceLine": opts.SourceLine,
		"createdAt":  time.Now(),
	})

	return signal
}

// CreateComputed creates a computed signal that derives its value
// from other signals and automatically updates when dependencies change
func CreateComputed[T any](computeFn func() T, options ...SignalOptions) *Signal[T] {
	var opts SignalOptions
	if len(options) > 0 {
		opts = options[0]
	} else {
		// Capture caller information for debugging by default
		_, file, line, _ := runtime.Caller(1)
		opts = SignalOptions{
			SourceFile: file,
			SourceLine: line,
		}
	}

	// Create a generic equals function wrapper if one was provided
	var equalsFn func(a, b T) bool
	if opts.Equals != nil {
		equalsFn = func(a, b T) bool {
			return opts.Equals(a, b)
		}
	}

	// Track dependencies during the initial computation
	// Use the proper tracking mechanisms provided by the signal system
	startTrackingLocal()
	initialValue := computeFn() // Only run the compute function once initially
	deps := stopTrackingLocal()

	// Create the signal with the initial value and equality function
	signal := NewSignalWithEquals(initialValue, equalsFn)
	// Register the signal in the global registry with its id
	globalMutex.Lock()
	signalRegistry[signal.id] = signal
	globalMutex.Unlock()

	// Store debug information
	setSignalMetadata(signal, map[string]interface{}{
		"debugName":  opts.DebugName,
		"sourceFile": opts.SourceFile,
		"sourceLine": opts.SourceLine,
		"createdAt":  time.Now(),
		"isComputed": true,
	})

	// Create an effect function that will be called when dependencies change
	// Using a separate effect function allows us to apply custom equality logic
	effectFn := func() {
		// During effect execution, we don't want to track dependencies again
		// We use the fixed dependency list from the initial computation
		newValue := computeFn()

		// Get the current value for comparison without holding the mutex
		var currentValue T
		{
			signal.mutex.Lock()
			currentValue = signal.value
			signal.mutex.Unlock()
		}

		// Determine if we should update based on equality function
		shouldUpdate := true
		if equalsFn != nil {
			// If custom equality is provided, use it to determine if value changed
			// Note: we invert the result because equalsFn returns true when values are equal
			// but we want to update when they're different
			shouldUpdate = !equalsFn(currentValue, newValue)
		} else {
			// Default equality comparison for any type
			// Convert to string for non-comparable types
			valueStr := fmt.Sprintf("%v", currentValue)
			newValueStr := fmt.Sprintf("%v", newValue)
			shouldUpdate = (valueStr != newValueStr)
		}

		if shouldUpdate {
			// Update the signal value
			signal.mutex.Lock()
			signal.value = newValue
			signal.mutex.Unlock()
			signal.Set(newValue)
		}
	}

	// Register the effect to run when dependencies change
	// We use RegisterEffectWithoutInitialRun because we've already computed the initial value
	effectID := RegisterEffectWithoutInitialRun(effectFn, deps, fmt.Sprintf("computed_%p", signal))

	// Store the effect ID for cleanup
	signal.SetMetadata("effectID", effectID)

	return signal
}

// CreateComputedWithCleanup creates a computed signal with a cleanup function
// that is called when the signal is disposed. This is useful for cleaning up
// resources that were created during the computation.
func CreateComputedWithCleanup[T any](
	computeFn func() T,
	cleanupFn func() error,
	options ...SignalOptions,
) *Signal[T] {
	var opts SignalOptions
	if len(options) > 0 {
		opts = options[0]
	} else {
		// Capture caller information for debugging by default
		_, file, line, _ := runtime.Caller(1)
		opts = SignalOptions{
			SourceFile: file,
			SourceLine: line,
		}
	}

	// Create a generic equals function wrapper if one was provided
	var equalsFn func(a, b T) bool
	if opts.Equals != nil {
		equalsFn = func(a, b T) bool {
			return opts.Equals(a, b)
		}
	}

	// Create a state to hold the computed value - this ensures a stable reference
	// that we can update from the effect
	state := NewState[T](func() T {
		// Track dependencies during the initial computation
		startTrackingLocal()
		initialValue := computeFn() // Initial computation
		_ = stopTrackingLocal()

		return initialValue
	}())

	// Create the signal with the initial value and custom equality function
	signal := state.GetSignal()

	// If custom equality is provided, apply it to the signal
	if equalsFn != nil {
		// We need to reinstantiate the signal with the custom equality function
		// Since we can't modify an existing signal's equality function directly
		initialValue := signal.Value()
		signal = NewSignalWithEquals(initialValue, equalsFn)
	}

	// Store debug information
	setSignalMetadata(signal, map[string]interface{}{
		"debugName":  opts.DebugName,
		"sourceFile": opts.SourceFile,
		"sourceLine": opts.SourceLine,
		"createdAt":  time.Now(),
		"isComputed": true,
		"hasCleanup": cleanupFn != nil,
	})

	// We need a variable to store the effect ID that can be accessed from inner closures
	var effectIDVar string

	// Register the effect that will update the computed value when dependencies change
	effectHandler := func() {
		// When a dependency changes, recompute the value
		startTrackingLocal() // Start tracking to capture any new dependencies
		newValue := computeFn()
		newDeps := stopTrackingLocal()

		// Update dependencies if they've changed
		if !depsEqual(effectDeps[effectIDVar], newDeps) && len(newDeps) > 0 {
			// Re-register the effect with the new dependencies
			oldEffectID := effectIDVar
			RemoveEffect(oldEffectID)

			// Register a new effect with the updated dependencies
			effectIDVar = RegisterEffectWithoutInitialRun(func() {
				// Recompute and update
				newValue := computeFn()
				state.Set(newValue)
			}, newDeps)

			// Update deps
			effectDeps[effectIDVar] = newDeps
		}

		// Update the signal value
		state.Set(newValue)
	}

	// Register the effect
	effectIDVar = RegisterEffectWithoutInitialRun(effectHandler, effectDeps[effectIDVar], fmt.Sprintf("computed_cleanup_%p", signal))

	// Register the cleanup function for this effect
	if cleanupFn != nil {
		registerCleanupForEffect(effectIDVar, cleanupFn)
	}

	// Store the effect ID for cleanup
	signal.SetMetadata("effectID", effectIDVar)

	// Also store in debug metadata if present
	metadata, _ := signal.GetMetadata("debug")
	if metadata != nil {
		metadataMap := metadata.(map[string]interface{})
		metadataMap["effect_id"] = effectIDVar
		signal.SetMetadata("debug", metadataMap)
	}

	return signal
}

// setSignalMetadata sets metadata on a signal for debugging purposes
func setSignalMetadata(s interface{}, metadata map[string]interface{}) {
	// Try using direct type assertion for common types first
	switch signal := s.(type) {
	case *Signal[bool]:
		if signal.metadata == nil {
			signal.metadata = make(map[string]interface{})
		}
		for k, v := range metadata {
			signal.metadata[k] = v
		}
		return
	case *Signal[string]:
		if signal.metadata == nil {
			signal.metadata = make(map[string]interface{})
		}
		for k, v := range metadata {
			signal.metadata[k] = v
		}
		return
	case *Signal[int]:
		if signal.metadata == nil {
			signal.metadata = make(map[string]interface{})
		}
		for k, v := range metadata {
			signal.metadata[k] = v
		}
		return
	case *Signal[float64]:
		if signal.metadata == nil {
			signal.metadata = make(map[string]interface{})
		}
		for k, v := range metadata {
			signal.metadata[k] = v
		}
		return
	case *Signal[int64]:
		if signal.metadata == nil {
			signal.metadata = make(map[string]interface{})
		}
		for k, v := range metadata {
			signal.metadata[k] = v
		}
		return
	case *Signal[uint64]:
		if signal.metadata == nil {
			signal.metadata = make(map[string]interface{})
		}
		for k, v := range metadata {
			signal.metadata[k] = v
		}
		return
	}

	// If we can't directly assert the type, use reflection as a fallback
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
		if val.Kind() == reflect.Struct {
			// Try to find the metadata field
			field := val.FieldByName("metadata")

			// If we found the field and it's a map
			if field.IsValid() && field.CanInterface() {
				if metadataMap, ok := field.Interface().(map[string]interface{}); ok {
					// Copy the metadata
					for k, v := range metadata {
						metadataMap[k] = v
					}
					return
				}
			}
		}
	}

	// If we get here, we couldn't set the metadata
	if localDebugMode {
		fmt.Printf("[DEBUG] Could not set metadata on signal of type %T\n", s)
	}
}

// defaultEquals is a default equality function that works for comparable types
func defaultEquals[T comparable](a, b T) bool {
	return a == b
}

// EnableDebugMode enables debug logging for signals
func EnableDebugMode() {
	signalFactoryMutex.Lock()
	localDebugMode = true
	signalFactoryMutex.Unlock()
}

// DisableDebugMode disables debug logging for signals
func DisableDebugMode() {
	signalFactoryMutex.Lock()
	localDebugMode = false
	signalFactoryMutex.Unlock()
}

// Global cleanup registry (implementation uses a mutex to be thread-safe)
var (
	cleanupRegistryMutex sync.RWMutex
	cleanupRegistry      = make(map[string]func() error)
	// Factory specific mutex for thread safety
	signalFactoryMutex sync.RWMutex
	// Local error handler that integrates with the global error handler
	localErrorHandler func(error) = func(err error) {
		// Forward the error to the global error handler if it exists
		globalMutex.RLock()
		handler := errorHandler
		globalMutex.RUnlock()

		if handler != nil {
			handler(err)
		} else {
			// If no global handler, log the error (in a production system this might use a logger)
			fmt.Printf("Unhandled error in effect cleanup: %v\n", err)
		}
	}
	// Local debug mode flag for linting purposes
	localDebugMode bool
)

// registerCleanupForEffect registers a cleanup function for an effect ID
func registerCleanupForEffect(effectID string, cleanupFn func() error) {
	cleanupRegistryMutex.Lock()
	defer cleanupRegistryMutex.Unlock()
	cleanupRegistry[effectID] = cleanupFn
}

// safelyExecuteCleanup runs a cleanup function and catches any panics,
// converting them to errors that can be handled more gracefully
func safelyExecuteCleanup(cleanupFn func() error) (err error) {
	// Use defer to catch panics
	defer func() {
		if r := recover(); r != nil {
			// Convert panic to error
			switch v := r.(type) {
			case error:
				err = fmt.Errorf("panic during cleanup: %w", v)
			default:
				err = fmt.Errorf("panic during cleanup: %v", v)
			}
		}
	}()

	// Execute the cleanup function
	return cleanupFn()
}

// RegisterEffect registers an effect with its dependencies (exported for tests)
func RegisterEffect(fn func(), deps []string) string {
	return registerEffectLocal(fn, deps)
}

// registerEffectWithID registers an effect with a stable ID, removing any previous effect with that ID
func registerEffectWithID(fn func(), deps []string, effectID string) string {
	// Remove old effect with this ID if it exists
	globalMutex.Lock()
	if _, exists := effectsRegistry[effectID]; exists {
		globalMutex.Unlock()
		RemoveEffect(effectID)
		globalMutex.Lock()
	}
	// Create the effect object with correct fields
	effect := &Effect{
		fn:        fn,
		deps:      deps,
		debugInfo: effectID,
	}
	// Add to global registry
	effectsRegistry[effectID] = effect
	effectDeps[effectID] = deps
	globalMutex.Unlock()
	// Register this effect as a dependency of each signal it depends on
	for _, depID := range deps {
		addEffectToSignal(depID, effect)
	}
	if debugMode {
		fmt.Printf("[DEBUG] Registered effect %s with %d dependencies\n", effectID, len(deps))
	}
	return effectID
}

// Register an effect with its dependencies and add it to the effect registry
func registerEffectLocal(fn func(), deps []string) string {
	id := fmt.Sprintf("effect_%d", time.Now().UnixNano())
	return registerEffectWithID(fn, deps, id)
}

// RegisterEffectWithoutInitialRun registers an effect with its dependencies without running it initially
func RegisterEffectWithoutInitialRun(fn func(), deps []string, effectID ...string) string {
	// Debug logging
	if debugMode {
		fmt.Printf("[DEBUG] RegisterEffectWithoutInitialRun: deps = %v\n", deps)
		fmt.Printf("[DEBUG] signalRegistry keys at effect registration: ")
		for k, v := range signalRegistry {
			fmt.Printf("%s(%p) ", k, v)
		}
		fmt.Println()
	}
	globalMutex.Lock()
	if len(effectID) > 0 {
		if _, exists := effectsRegistry[effectID[0]]; exists {
			globalMutex.Unlock()
			RemoveEffect(effectID[0])
			globalMutex.Lock()
		}
	} else {
		id := fmt.Sprintf("effect_%d", time.Now().UnixNano())
		effectID = append(effectID, id)
	}
	// Create the effect object with correct fields
	effect := &Effect{
		fn:        fn,
		deps:      deps,
		debugInfo: effectID[0],
	}
	// Add to global registry
	effectsRegistry[effectID[0]] = effect
	effectDeps[effectID[0]] = deps
	globalMutex.Unlock()
	// Register this effect as a dependency of each signal it depends on
	for _, depID := range deps {
		addEffectToSignal(depID, effect)
	}
	if debugMode {
		fmt.Printf("[DEBUG] Registered effect %s with %d dependencies\n", effectID[0], len(deps))
	}
	return effectID[0]
}

// Simplified implementations of tracking functions for linting purposes
// These will be properly integrated with the full reactive system in the future

// Global dependency tracking system to monitor which signals are accessed during effect execution
var (
	trackingMutex        sync.RWMutex
	trackingActive       bool
	trackingDependencies = make(map[string]bool)
)

// StartTracking begins tracking signal dependencies (exported for tests)
func StartTracking() {
	startTrackingLocal()
}

// startTrackingLocal begins tracking signal dependencies locally
func startTrackingLocal() {
	// Acquire exclusive lock
	trackingMutex.Lock()
	defer trackingMutex.Unlock()

	// Clear previous tracking context
	trackingDependencies = make(map[string]bool)
	// Set tracking flag
	trackingActive = true

	if debugMode {
		fmt.Println("[DEBUG] Started tracking signal dependencies")
	}
}

// StopTracking stops tracking signal dependencies and returns the collected dependencies (exported for tests)
func StopTracking() []string {
	return stopTrackingLocal()
}

// depsEqual compares two slices of strings for set equality (order-insensitive, nil/empty safe)
func depsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	// Use a map to count occurrences
	count := make(map[string]int, len(a))
	for _, v := range a {
		count[v]++
	}
	for _, v := range b {
		if count[v] == 0 {
			return false
		}
		count[v]--
	}
	return true
}

// stopTrackingLocal stops tracking signal dependencies and returns the collected dependencies
func stopTrackingLocal() []string {
	// Acquire exclusive lock
	trackingMutex.Lock()
	defer trackingMutex.Unlock()

	// Stop tracking
	trackingActive = false

	// Convert map keys to slice
	deps := make([]string, 0, len(trackingDependencies))
	for dep := range trackingDependencies {
		deps = append(deps, dep)
	}

	// Clear dependencies
	trackingDependencies = make(map[string]bool)

	if debugMode {
		fmt.Printf("[DEBUG] Stopped tracking, collected %d dependencies\n", len(deps))
	}

	return deps
}

// AddSignalDependency registers the given signal as a dependency in the current tracking context
// This is exported so that signals can call it during Value() access
func AddSignalDependency(signalID string) {
	// Use a read lock for quick check if tracking is active
	trackingMutex.RLock()
	active := trackingActive
	trackingMutex.RUnlock()

	// If not tracking, return quickly
	if !active {
		return
	}

	// Full lock to update dependencies
	trackingMutex.Lock()
	defer trackingMutex.Unlock()

	// Double-check tracking is still active
	if trackingActive {
		trackingDependencies[signalID] = true

		if debugMode {
			fmt.Printf("[DEBUG] Added signal %s as dependency\n", signalID)
		}
	}
}

// addEffectToSignal finds a signal by ID and adds the effect as a dependent
func addEffectToSignal(signalID string, effect *Effect) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	if debugMode {
		fmt.Printf("[DEBUG] addEffectToSignal: searching for signalID=%s in keys: ", signalID)
		for k, v := range signalRegistry {
			fmt.Printf("%s(%p) ", k, v)
		}
		fmt.Printf(" | effect addr: %p\n", effect)
	}

	if signal, found := signalRegistry[signalID]; found {
		effectID := effect.debugInfo
		switch s := signal.(type) {
		case interface{ AddDependent(string, Dependency) }:
			s.AddDependent(effectID, effect)
			if debugMode {
				fmt.Printf("[DEBUG] Added effect as dependent to signal %s with effectID %s\n", signalID, effectID)
			}
			return
		default:
			if debugMode {
				fmt.Printf("[DEBUG] Could not add effect to signal %s: unsupported type %T\n", signalID, signal)
			}
		}
	} else if debugMode {
		fmt.Printf("[DEBUG] Signal with ID %s not found in registry\n", signalID)
	}
}

// Debug flag to help with testing
var (
	debugFlag = false
)

// CreateEffect creates an effect with automatic dependency detection.
// The effect function will run immediately and then again whenever
// any signal accessed during its execution changes.
func CreateEffect(effectFn func()) string {
	// Begin tracking dependencies
	startTrackingLocal()

	// Execute the effect to capture dependencies during first run
	// This both initializes the effect and collects its dependencies
	func() {
		defer func() {
			if r := recover(); r != nil {
			}
		}()
		effectFn()
	}()

	// Get the tracked dependencies
	deps := stopTrackingLocal()

	// Use stable effect ID for registration
	effectID := fmt.Sprintf("effect_%p", effectFn)
	return RegisterEffectWithoutInitialRun(effectFn, deps, effectID)
}

// CreateEffectWithDeps creates an effect with an explicit dependency list.
// The effect will only re-run when signals with IDs in the explicit deps list change.
func CreateEffectWithDeps(effectFn func(), explicitDepIDs []string) string {
	// Use stable effect ID for registration
	effectID := fmt.Sprintf("effect_explicit_%p", effectFn)
	id := RegisterEffectWithoutInitialRun(effectFn, explicitDepIDs, effectID)

	// Run the effect once to initialize
	func() {
		defer func() {
			if r := recover(); r != nil {
			}
		}()
		effectFn()
	}()

	return id
}
