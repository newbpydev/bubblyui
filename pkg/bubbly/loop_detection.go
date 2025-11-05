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

// CommandLogger is the package-private interface for logging command generation.
// This is defined here to avoid import cycles with pkg/bubbly/commands.
//
// The public API is in pkg/bubbly/commands/debug.go.
type CommandLogger interface {
	LogCommand(componentName, componentID, refID string, oldValue, newValue interface{})
}

// commandLoggerImpl is the internal implementation that logs to an io.Writer.
type commandLoggerImpl struct {
	logger *log.Logger
}

// newCommandLogger creates a new command logger for internal use.
func newCommandLogger(writer io.Writer) CommandLogger {
	if writer == nil {
		writer = io.Discard
	}

	return &commandLoggerImpl{
		logger: log.New(writer, "", log.LstdFlags),
	}
}

// LogCommand logs the command generation event.
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

// nopCommandLogger is a no-operation logger (zero overhead when disabled).
type nopCommandLogger struct{}

// newNopCommandLogger creates a no-operation logger.
func newNopCommandLogger() CommandLogger {
	return &nopCommandLogger{}
}

// LogCommand does nothing (inlined for zero overhead).
func (nl *nopCommandLogger) LogCommand(componentName, componentID, refID string, oldValue, newValue interface{}) {
	// No-op: zero overhead when debugging is disabled
}
