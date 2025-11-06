package commands

import (
	"fmt"
	"sync"
)

// maxCommandsPerRef is the maximum number of commands that can be generated
// for a single ref within one update cycle before a loop is detected.
//
// This matches the maxUpdateDepth constant used in lifecycle.go for consistency.
// A value of 100 is high enough for legitimate rapid updates (like animations
// or batch processing) but low enough to catch infinite loops quickly.
const maxCommandsPerRef = 100

// LoopDetector tracks command generation per component:ref pair to detect
// infinite loops where state changes trigger more state changes recursively.
//
// Thread Safety:
//
// All methods are thread-safe and can be called concurrently from multiple
// goroutines. A RWMutex protects the internal command count map.
//
// Usage Pattern:
//
// The detector is typically integrated into the component runtime and reset
// after each Update() cycle (similar to lifecycle's updateCount reset):
//
//	detector := NewLoopDetector()
//
//	// In command generation (setHook):
//	if err := detector.CheckLoop(componentID, refID); err != nil {
//	    // Report error to observability
//	    return
//	}
//
//	// In component Update() after processing:
//	detector.Reset()
//
// Design Note:
//
// The detector uses a map key format of "componentID:refID" to track each
// unique ref independently. This allows multiple refs in the same component
// to each generate up to maxCommandsPerRef commands without false positives.
type LoopDetector struct {
	commandCounts map[string]int
	mu            sync.RWMutex
}

// NewLoopDetector creates a new loop detector with empty command counts.
//
// Example:
//
//	detector := NewLoopDetector()
func NewLoopDetector() *LoopDetector {
	return &LoopDetector{
		commandCounts: make(map[string]int),
	}
}

// CheckLoop increments the command count for the given component:ref pair
// and returns an error if the count exceeds the maximum allowed.
//
// This method is thread-safe and can be called concurrently.
//
// Parameters:
//   - componentID: Unique identifier of the component
//   - refID: Unique identifier of the ref (e.g., "ref-42")
//
// Returns:
//   - nil if the command is allowed (count <= maxCommandsPerRef)
//   - ErrCommandLoop if the count exceeds the limit
//
// Example:
//
//	if err := detector.CheckLoop("counter-1", "ref-5"); err != nil {
//	    // Handle loop detection error
//	    return err
//	}
func (ld *LoopDetector) CheckLoop(componentID, refID string) error {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	key := componentID + ":" + refID
	ld.commandCounts[key]++

	if ld.commandCounts[key] > maxCommandsPerRef {
		return &CommandLoopError{
			ComponentID:  componentID,
			RefID:        refID,
			CommandCount: ld.commandCounts[key],
			MaxCommands:  maxCommandsPerRef,
		}
	}

	return nil
}

// Reset clears all command counts, typically called after each component
// Update() cycle completes.
//
// This method is thread-safe and can be called concurrently.
//
// Example:
//
//	// After component Update() cycle
//	detector.Reset()
func (ld *LoopDetector) Reset() {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	// Clear the map by creating a new one
	// This is more efficient than iterating and deleting keys
	ld.commandCounts = make(map[string]int)
}

// CommandLoopError indicates that a command generation loop was detected.
//
// This error is returned by CheckLoop when a ref generates more than
// maxCommandsPerRef commands within a single update cycle, which typically
// indicates an infinite loop where state changes trigger more state changes.
//
// The error message includes the component ID, ref ID, and command count
// to help developers identify the problematic code.
//
// Example error message:
//
//	command generation loop detected for component 'counter' ref 'count':
//	generated 101 commands (max 100). Check for recursive state updates in
//	event handlers or lifecycle hooks.
type CommandLoopError struct {
	ComponentID  string
	RefID        string
	CommandCount int
	MaxCommands  int
}

// Error returns a clear, actionable error message for developers.
//
// The message includes:
//   - Component and ref identification
//   - Actual vs maximum command counts
//   - Guidance on where to look for the problem
func (e *CommandLoopError) Error() string {
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
