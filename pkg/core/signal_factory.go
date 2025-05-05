package core

import (
	"fmt"
	"reflect"
	"runtime"
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
