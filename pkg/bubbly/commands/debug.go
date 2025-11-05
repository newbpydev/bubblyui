package commands

import (
	"fmt"
	"io"
	"log"
)

// CommandLogger is the interface for logging command generation events.
//
// Implementations can log to different outputs (stdout, files, custom writers)
// or skip logging entirely for zero overhead when debugging is disabled.
//
// Usage:
//
//	// Enable logging
//	logger := NewCommandLogger(os.Stdout)
//	logger.LogCommand("Counter", "comp-1", "ref-5", 0, 1)
//
//	// Disable logging (zero overhead)
//	logger := NewNopLogger()
//	logger.LogCommand("Counter", "comp-1", "ref-5", 0, 1) // No-op
type CommandLogger interface {
	// LogCommand logs a command generation event with component and ref details.
	//
	// Parameters:
	//   - componentName: Human-readable name of the component (e.g., "Counter")
	//   - componentID: Unique identifier of the component (e.g., "component-42")
	//   - refID: Unique identifier of the ref (e.g., "ref-5")
	//   - oldValue: Value before the change
	//   - newValue: Value after the change
	LogCommand(componentName, componentID, refID string, oldValue, newValue interface{})
}

// commandLogger is the standard implementation that logs to an io.Writer.
//
// It formats log messages with:
//   - Timestamp (from Go's standard log package)
//   - [DEBUG] prefix for visibility
//   - Component identification (name and ID)
//   - Ref identification
//   - State transition (old → new)
//   - Action description
//
// Example output:
//
//	2024/01/15 10:30:45 [DEBUG] Command Generated | Component: Counter (component-1) | Ref: ref-5 | 0 → 1
type commandLogger struct {
	logger *log.Logger
}

// NewCommandLogger creates a new command logger that writes to the given io.Writer.
//
// The writer is typically os.Stdout, os.Stderr, or a custom buffer/file for testing.
// If writer is nil, logs are discarded (similar to NopLogger).
//
// Example:
//
//	logger := NewCommandLogger(os.Stdout)
func NewCommandLogger(writer io.Writer) CommandLogger {
	if writer == nil {
		writer = io.Discard
	}

	return &commandLogger{
		logger: log.New(writer, "", log.LstdFlags), // Standard date/time flags
	}
}

// LogCommand logs the command generation with a clear, structured format.
//
// Format:
//
//	[DEBUG] Command Generated | Component: <name> (<id>) | Ref: <refID> | <old> → <new>
//
// The arrow (→) provides a clear visual indicator of the state transition.
func (cl *commandLogger) LogCommand(componentName, componentID, refID string, oldValue, newValue interface{}) {
	cl.logger.Printf(
		"[DEBUG] Command Generated | Component: %s (%s) | Ref: %s | %v → %v",
		componentName,
		componentID,
		refID,
		oldValue,
		newValue,
	)
}

// nopLogger is a no-operation logger that does nothing.
//
// This implementation provides zero overhead when debug logging is disabled.
// All method calls are inlined away by the Go compiler, resulting in no
// runtime cost.
//
// This is the default logger used when WithCommandDebug(false) or when
// command debugging is not explicitly enabled.
type nopLogger struct{}

// NewNopLogger creates a no-operation logger that discards all log calls.
//
// This is used when debug logging is disabled to ensure zero overhead.
//
// Example:
//
//	logger := NewNopLogger()
//	logger.LogCommand(...) // No-op, zero cost
func NewNopLogger() CommandLogger {
	return &nopLogger{}
}

// LogCommand does nothing. This method is inlined by the compiler for zero overhead.
func (nl *nopLogger) LogCommand(componentName, componentID, refID string, oldValue, newValue interface{}) {
	// No-op: zero overhead when debugging is disabled
}

// defaultLogger is the package-level default logger (disabled by default).
//
// Components use this logger unless they explicitly enable debugging via
// WithCommandDebug(true), which sets a custom logger.
var defaultLogger CommandLogger = NewNopLogger()

// SetDefaultLogger sets the package-level default logger.
//
// This allows applications to enable debug logging globally without
// modifying individual component creation calls.
//
// Example:
//
//	// Enable debug logging for all components
//	commands.SetDefaultLogger(commands.NewCommandLogger(os.Stdout))
//
// Note: Most applications should use WithCommandDebug(true) on specific
// components rather than setting a global logger.
func SetDefaultLogger(logger CommandLogger) {
	if logger == nil {
		logger = NewNopLogger()
	}
	defaultLogger = logger
}

// GetDefaultLogger returns the package-level default logger.
//
// This is used by components that don't have command debugging explicitly
// enabled or disabled.
func GetDefaultLogger() CommandLogger {
	return defaultLogger
}

// FormatValue formats a value for logging, handling special cases.
//
// This is a helper function that can be used by custom CommandLogger
// implementations to format values consistently.
//
// Special handling:
//   - nil: returns "<nil>"
//   - strings: quoted
//   - other: uses fmt.Sprintf with %v
//
// Example:
//
//	formatted := FormatValue(nil)        // "<nil>"
//	formatted = FormatValue("hello")     // "\"hello\""
//	formatted = FormatValue(42)          // "42"
func FormatValue(value interface{}) string {
	if value == nil {
		return "<nil>"
	}

	// Use %v for simple formatting (same as in LogCommand)
	return fmt.Sprintf("%v", value)
}
