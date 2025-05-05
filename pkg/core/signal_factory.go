package core

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"
)

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
	return CreateComputedWithCleanup(computeFn, nil, options...)
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

	// Track dependencies during the initial computation
	StartTracking()
	initialValue := computeFn() // This is the only time we should run computeFn during initialization
	deps := StopTracking()

	// Create the signal with the initial value
	signal := NewSignalWithEquals(initialValue, equalsFn)

	// Store debug information
	setSignalMetadata(signal, map[string]interface{}{
		"debugName":  opts.DebugName,
		"sourceFile": opts.SourceFile,
		"sourceLine": opts.SourceLine,
		"createdAt":  time.Now(),
		"isComputed": true,
		"hasCleanup": cleanupFn != nil,
	})

	// Register an effect that will be triggered when dependencies change, but do NOT run it immediately
	// This ensures the computation only happens once during initialization
	effectID := RegisterEffectWithoutInitialRun(func() {
		// When a dependency changes, recompute and update the signal value
		// Use StartTracking/StopTracking to automatically capture any updated dependencies
		StartTracking()
		newValue := computeFn()
		_ = StopTracking() // Collect but don't use dependencies yet

		// Update the signal value
		signal.Set(newValue)
	}, deps)

	// If a cleanup function was provided, register it in our cleanup tracking system
	if cleanupFn != nil {
		registerCleanupForEffect(effectID, cleanupFn)
	}

	// Store the effect ID for cleanup
	signal.SetMetadata("effectID", effectID)

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
	if debugMode {
		fmt.Printf("[DEBUG] Could not set metadata on signal of type %T\n", s)
	}
}

// defaultEquals is a default equality function that works for comparable types
func defaultEquals[T comparable](a, b T) bool {
	return a == b
}

// EnableDebugMode enables debug logging for signals
func EnableDebugMode() {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	debugMode = true
}

// DisableDebugMode disables debug logging for signals
func DisableDebugMode() {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	debugMode = false
}

// Global cleanup registry (implementation uses a mutex to be thread-safe)
var (
	cleanupRegistryMutex sync.RWMutex
	cleanupRegistry      = make(map[string]func() error)
)

// RemoveEffect removes an effect and executes its cleanup function if provided
func RemoveEffect(effectID string) {
	// First check if there's a cleanup function for this effect
	var cleanupFn func() error

	// Get the cleanup function if it exists
	cleanupRegistryMutex.RLock()
	cleanupFn = cleanupRegistry[effectID]
	cleanupRegistryMutex.RUnlock()

	// Remove the effect from the global registry
	globalMutex.Lock()
	delete(effectsRegistry, effectID)
	globalMutex.Unlock()

	// Execute the cleanup function if it exists
	if cleanupFn != nil {
		// Remove from registry first to prevent multiple executions
		cleanupRegistryMutex.Lock()
		delete(cleanupRegistry, effectID)
		cleanupRegistryMutex.Unlock()

		// Execute the cleanup with error handling
		err := safelyExecuteCleanup(cleanupFn)
		if err != nil && errorHandler != nil {
			errorHandler(err)
		}
	}
}

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

// CreateEffect creates an effect with automatic dependency detection.
// The effect function will run immediately and then again whenever
// any signal accessed during its execution changes.
func CreateEffect(effectFn func()) string {
	// Store previous values of dependencies to check for actual changes
	depValues := make(map[string]uint64)

	// Create a wrapper function that will track dependencies each time
	// the effect runs, ensuring we always have the most up-to-date dependencies
	dynamicTrackingFn := func() {
		// Get previous version numbers of known dependencies
		valuesChanged := false

		// Check if any dependency values have actually changed
		if len(depValues) > 0 {
			for depID, prevVersion := range depValues {
				// Find the signal by ID
				globalMutex.RLock()
				signal, exists := signalRegistry[depID]
				globalMutex.RUnlock()

				if exists {
					// Check if version has changed
					// Use type assertion to get the underlying signal
					sig, ok := signal.(*Signal[any])
					if !ok {
						// If the type assertion fails, try to use reflection to get the version
						// This handles the case where the signal is of a different generic type
						val := reflect.ValueOf(signal).Elem()
						if val.Kind() == reflect.Struct {
							verField := val.FieldByName("version")
							if verField.IsValid() && verField.CanUint() {
								currentVersion := verField.Uint()
								if uint64(currentVersion) != prevVersion {
									valuesChanged = true
									break
								}
							}
						}
						continue
					}

					// We have a valid signal, check its version
					sig.mutex.RLock()
					currentVersion := sig.version
					sig.mutex.RUnlock()

					if currentVersion != prevVersion {
						valuesChanged = true
						break
					}
				}
			}

			// If no values changed, we can skip this execution
			if !valuesChanged && len(depValues) > 0 {
				if debugMode {
					fmt.Println("[DEBUG] Skipping effect execution - no dependency changes detected")
				}
				return
			}
		}

		// Start tracking dependencies
		StartTracking()

		// Run the effect function, which will access signals and register dependencies
		effectFn()

		// Get the dependencies that were accessed and update their version numbers
		deps := StopTracking()

		// Update our stored dependency versions for the next run
		depValues = make(map[string]uint64, len(deps))
		for _, depID := range deps {
			globalMutex.RLock()
			signal, exists := signalRegistry[depID]
			globalMutex.RUnlock()

			if exists {
				// Use reflection to safely get the version field regardless of generic type
				val := reflect.ValueOf(signal).Elem()
				if val.Kind() == reflect.Struct {
					verField := val.FieldByName("version")
					if verField.IsValid() && verField.CanUint() {
						depValues[depID] = uint64(verField.Uint())
					}
				}
			}
		}
	}

	// First execution with dependency tracking to establish initial dependencies
	StartTracking()
	effectFn() // Execute once to gather initial dependencies
	deps := StopTracking()

	// Initialize dependency version tracking
	depValues = make(map[string]uint64, len(deps))
	for _, depID := range deps {
		globalMutex.RLock()
		signal, exists := signalRegistry[depID]
		globalMutex.RUnlock()

		if exists {
			// Use reflection to safely get the version field regardless of generic type
			val := reflect.ValueOf(signal).Elem()
			if val.Kind() == reflect.Struct {
				verField := val.FieldByName("version")
				if verField.IsValid() && verField.CanUint() {
					depValues[depID] = uint64(verField.Uint())
				}
			}
		}
	}

	// Register the effect with dependencies - but we want the wrapper to be called
	// each time, so it can dynamically track dependencies
	effectID := RegisterEffect(dynamicTrackingFn, deps)

	// Use debug info for naming
	_, file, line, _ := runtime.Caller(1)
	effectDebugInfo := fmt.Sprintf("effect_%s:%d", file, line)

	// Store in global registry with debug info
	globalMutex.Lock()
	defer globalMutex.Unlock()

	// If we have an existing effect in the registry, add debug info
	if effect, ok := effectsRegistry[effectID]; ok {
		if e, ok := effect.(*Effect); ok {
			e.debugInfo = effectDebugInfo
		}
	}

	return effectID
}

// CreateEffectWithDeps creates an effect with an explicit dependency list.
// The effect will only re-run when signals with IDs in the explicit deps list change.
func CreateEffectWithDeps(effectFn func(), explicitDepIDs []string) string {
	// Register the effect with the explicit dependencies
	effectID := RegisterEffect(effectFn, explicitDepIDs)

	// Use debug info for naming
	_, file, line, _ := runtime.Caller(1)
	effectDebugInfo := fmt.Sprintf("effect_explicit_%s:%d", file, line)

	// Store in global registry with debug info
	globalMutex.Lock()
	if effect, ok := effectsRegistry[effectID].(*Effect); ok {
		effect.debugInfo = effectDebugInfo
	}
	globalMutex.Unlock()

	return effectID
}
