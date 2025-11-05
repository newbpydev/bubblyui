package bubbly

import (
	"fmt"
	"io"
	"log"
	"sync"
)

// maxCommandsPerRef is the maximum number of commands that can be generated
// for a single ref within one update cycle before a loop is detected.
//
// This constant matches the value in pkg/bubbly/commands/loop_detection.go
// and lifecycle.go's maxUpdateDepth for consistency across the framework.
const maxCommandsPerRef = 100

// loopDetector tracks command generation per component:ref pair to detect
// infinite loops. This is the internal implementation used by componentImpl.
//
// Note: This is a package-private type to avoid import cycles. The public
// API is exposed through pkg/bubbly/commands/loop_detection.go.
type loopDetector struct {
	commandCounts map[string]int
	mu            sync.RWMutex
}

// newLoopDetector creates a new loop detector with empty command counts.
func newLoopDetector() *loopDetector {
	return &loopDetector{
		commandCounts: make(map[string]int),
	}
}

// checkLoop increments the command count for the given component:ref pair
// and returns an error if the count exceeds the maximum allowed.
func (ld *loopDetector) checkLoop(componentID, refID string) error {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	key := componentID + ":" + refID
	ld.commandCounts[key]++

	if ld.commandCounts[key] > maxCommandsPerRef {
		return &commandLoopError{
			ComponentID:  componentID,
			RefID:        refID,
			CommandCount: ld.commandCounts[key],
			MaxCommands:  maxCommandsPerRef,
		}
	}

	return nil
}

// reset clears all command counts.
func (ld *loopDetector) reset() {
	ld.mu.Lock()
	defer ld.mu.Unlock()
	ld.commandCounts = make(map[string]int)
}

// commandLoopError indicates that a command generation loop was detected.
// This is the internal error type used by the bubbly package.
type commandLoopError struct {
	ComponentID  string
	RefID        string
	CommandCount int
	MaxCommands  int
}

// Error returns a clear, actionable error message for developers.
func (e *commandLoopError) Error() string {
	return fmt.Sprintf(
		"command generation loop detected for component '%s' ref '%s': "+
			"generated %d commands (max %d). "+
			"Check for recursive state updates in event handlers or lifecycle hooks.",
		e.ComponentID,
		e.RefID,
		e.CommandCount,
		e.MaxCommands,
	)
}

// CommandLogger is the package-private interface for logging command generation
// events within the bubbly package.
//
// This interface is defined internally to avoid import cycles between the
// main bubbly package and the commands package. The public API and full
// documentation are available in pkg/bubbly/commands/debug.go.
//
// Architecture Pattern:
//   - Public API: pkg/bubbly/commands.CommandLogger (exported, documented)
//   - Internal API: pkg/bubbly.CommandLogger (package-private, used by components)
//   - Implementations: Both packages provide compatible implementations
//
// Integration:
//
//	This interface is used by componentImpl when WithCommandDebug(true) is
//	specified in the component builder. The logging happens automatically
//	in the Context.Ref() setHook when reactive state changes trigger command
//	generation.
//
// Usage (Internal):
//
//	// In componentImpl
//	component.commandLogger.LogCommand(componentName, componentID, refID, oldValue, newValue)
//
// Performance:
//   - NopLogger: ~0.25 ns/op, 0 allocs/op (zero overhead)
//   - CommandLogger: ~2700 ns/op, 4 allocs/op (enabled logging)
//   - ~10,000x faster when disabled
type CommandLogger interface {
	// LogCommand logs a command generation event with component and ref details.
	//
	// This method is called automatically when a reactive state change triggers
	// command generation in the Automatic Reactive Bridge system.
	//
	// Parameters:
	//   - componentName: Human-readable name of the component (e.g., "Counter")
	//   - componentID: Unique identifier of the component (e.g., "component-42")
	//   - refID: Unique identifier of the ref (e.g., "ref-5")
	//   - oldValue: Value before the change (can be nil)
	//   - newValue: Value after the change (any type)
	//
	// Thread Safety:
	//   This method must be safe for concurrent calls from multiple goroutines.
	//   All provided implementations are thread-safe.
	//
	// Performance:
	//   - NopLogger: ~0.25 ns/op (inlined empty method)
	//   - CommandLogger: ~2700 ns/op (includes formatting and I/O)
	//
	LogCommand(componentName, componentID, refID string, oldValue, newValue interface{})
}

// commandLoggerImpl is the internal implementation that logs command generation
// events to an io.Writer with structured formatting.
//
// This implementation mirrors the public commandLogger in pkg/bubbly/commands
// but is defined locally to avoid import cycles. It provides the same
// functionality and performance characteristics.
//
// Features:
//   - Thread-safe logging (uses Go's standard log package)
//   - Timestamp prefix (standard log format)
//   - Structured format for easy parsing
//   - Flexible output destination (any io.Writer)
//
// Log Format:
//
//	[timestamp] [DEBUG] Command Generated | Component: <name> (<id>) | Ref: <refID> | <old> → <new>
//
// Performance:
//   - Overhead: ~2700 ns/op (includes formatting and I/O)
//   - Allocations: 4 allocs/op (string formatting)
//   - Thread Safety: Yes (log.Logger is thread-safe)
type commandLoggerImpl struct {
	logger *log.Logger // Standard Go logger with built-in thread safety
}

// newCommandLogger creates a new command logger for internal use within the
// bubbly package.
//
// This function creates a CommandLogger implementation that formats and writes
// command generation events to the specified writer. It's used internally by
// the component builder when WithCommandDebug(true) is specified.
//
// Parameters:
//   - writer: Destination for log output (e.g., os.Stdout, file, buffer)
//     If nil, logs are discarded (behaves like NopLogger)
//
// Returns:
//   - CommandLogger: Ready to use for logging command generation events
//
// Usage:
//
//	// Create logger for component
//	logger := newCommandLogger(os.Stdout)
//	logger.LogCommand("Counter", "component-1", "ref-5", 0, 1)
//
//	// Create logger that discards output
//	logger := newCommandLogger(nil) // Equivalent to NopLogger
func newCommandLogger(writer io.Writer) CommandLogger {
	// Handle nil writer gracefully - discard logs like NopLogger
	if writer == nil {
		writer = io.Discard
	}

	// Create standard Go logger with timestamp flags
	// log.LstdFlags provides: YYYY/MM/DD HH:MM:SS timestamp format
	return &commandLoggerImpl{
		logger: log.New(writer, "", log.LstdFlags),
	}
}

// LogCommand logs the command generation event with structured formatting.
//
// This method formats the command generation details into a consistent log message
// that includes all essential debugging information. The format matches the public
// implementation for consistency across the system.
//
// Log Format:
//
//	[timestamp] [DEBUG] Command Generated | Component: <name> (<id>) | Ref: <refID> | <old> → <new>
//
// Format Elements:
//   - timestamp: Added automatically by Go's log package (YYYY/MM/DD HH:MM:SS)
//   - [DEBUG]: Tag for filtering log output
//   - Component: Component name and unique ID for identification
//   - Ref: Ref ID for tracking specific state variables
//   - →: Unicode arrow (U+2192) showing state transition direction
//   - <old>, <new>: String representation of state values
//
// Examples:
//
//	// Integer counter
//	cl.LogCommand("Counter", "component-1", "ref-5", 0, 1)
//	// Output: 2025/11/05 10:30:45 [DEBUG] Command Generated | Component: Counter (component-1) | Ref: ref-5 | 0 → 1
//
//	// String field
//	cl.LogCommand("Form", "component-2", "ref-10", "", "john@example.com")
//	// Output: 2025/11/05 10:30:46 [DEBUG] Command Generated | Component: Form (component-2) | Ref: ref-10 |  → john@example.com
//
// Thread Safety:
//
//	This method is thread-safe and can be called concurrently from multiple
//	goroutines. The underlying log.Logger provides synchronization.
//
// Performance:
//   - Time complexity: O(n) where n is the length of formatted strings
//   - Space complexity: O(n) for the formatted log message
//   - Typical overhead: ~2700 ns/op
//   - Allocations: ~4 allocs/op for string formatting
func (cl *commandLoggerImpl) LogCommand(componentName, componentID, refID string, oldValue, newValue interface{}) {
	cl.logger.Printf(
		"[DEBUG] Command Generated | Component: %s (%s) | Ref: %s | %v → %v",
		componentName,
		componentID,
		refID,
		oldValue,
		newValue,
	)
}

// nopCommandLogger is a no-operation logger that provides zero overhead when
// debug logging is disabled.
//
// This implementation mirrors the public NopLogger in pkg/bubbly/commands but
// is defined locally to avoid import cycles. It provides identical performance
// characteristics and behavior.
//
// Performance Characteristics:
//   - LogCommand overhead: ~0.25 ns/op (compiler inlines empty method)
//   - Allocations: 0 per call
//   - Memory usage: Zero (no fields, no allocations)
//   - Thread safety: Trivial (no shared state)
//
// Use Cases:
//   - Default logger when debug is disabled
//   - Production environments where debugging is not needed
//   - Performance-critical applications
//   - Components without WithCommandDebug(true)
//
// Comparison:
//   - NopLogger: ~0.25 ns/op, 0 allocs/op
//   - CommandLogger: ~2700 ns/op, 4 allocs/op
//   - Performance ratio: ~10,000x faster when disabled
type nopCommandLogger struct{}

// newNopCommandLogger creates a no-operation logger for internal use.
//
// This function returns a CommandLogger implementation that does nothing.
// It's used when debug logging is disabled to ensure zero performance overhead.
// This is the default logger for components unless WithCommandDebug(true) is
// specified in the component builder.
//
// Returns:
//   - CommandLogger: A logger that discards all log calls
//
// Usage:
//
//	// Create no-op logger (default)
//	logger := newNopCommandLogger()
//	logger.LogCommand("Counter", "comp-1", "ref-5", 0, 1) // No-op
//
//	// Used internally by component builder when debug is disabled
//	component := newComponentImpl("Test")
//	// component.commandLogger = newNopCommandLogger() // Default
//
// Performance:
//   - Creation cost: Minimal (just returns a singleton)
//   - LogCommand cost: ~0.25 ns/op (inlined empty method)
//   - Memory overhead: Zero
//   - Thread safety: No synchronization needed (no shared state)
func newNopCommandLogger() CommandLogger {
	return &nopCommandLogger{}
}

// LogCommand does nothing and provides zero overhead.
//
// This method is empty and gets inlined away by the Go compiler,
// resulting in essentially no runtime cost. It's designed for maximum
// performance when debug logging is disabled.
//
// Parameters:
//   - componentName: Ignored
//   - componentID: Ignored
//   - refID: Ignored
//   - oldValue: Ignored
//   - newValue: Ignored
//
// Performance:
//   - Time complexity: O(1) (compiler inlines to nothing)
//   - Space complexity: O(0)
//   - Overhead: ~0.25 ns/op (measurement artifact)
//   - Allocations: 0
func (nl *nopCommandLogger) LogCommand(componentName, componentID, refID string, oldValue, newValue interface{}) {
	// No-op: zero overhead when debugging is disabled
	// This method is inlined by the Go compiler for maximum performance
}
