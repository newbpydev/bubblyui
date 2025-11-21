package testutil

import (
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/commands"
)

// LoopEvent captures information about a detected command generation loop.
//
// This struct records when a loop was detected, which component and ref were
// involved, and how many commands were generated before the loop was caught.
//
// Example:
//
//	event := LoopEvent{
//	    ComponentID:  "counter-1",
//	    RefID:        "ref-5",
//	    CommandCount: 101,
//	    DetectedAt:   time.Now(),
//	}
type LoopEvent struct {
	ComponentID  string
	RefID        string
	CommandCount int
	DetectedAt   time.Time
}

// LoopDetectionVerifier provides testing utilities for command generation loop
// detection in the automatic reactive bridge system.
//
// This verifier wraps a LoopDetector and provides methods to simulate loops,
// track detected loops, and make assertions about loop detection behavior.
// It's designed to help verify that infinite loop detection works correctly
// when reactive state changes trigger recursive command generation.
//
// Thread Safety:
//
// LoopDetectionVerifier is not thread-safe. Create separate instances for
// concurrent tests.
//
// Example usage:
//
//	detector := commands.NewLoopDetector()
//	verifier := testutil.NewLoopDetectionVerifier(detector)
//
//	// Simulate loop (150 iterations exceeds 100 threshold)
//	verifier.SimulateLoop("counter", "count", 150)
//
//	// Verify loop was detected
//	verifier.AssertLoopDetected(t)
//	assert.Equal(t, 1, verifier.GetLoopCount())
type LoopDetectionVerifier struct {
	detector *commands.LoopDetector
	detected []LoopEvent
}

// NewLoopDetectionVerifier creates a new loop detection verifier.
//
// The verifier wraps the given LoopDetector and tracks all detected loops
// during simulation. The detector parameter can be nil, in which case
// SimulateLoop() will be a no-op (no loops can be detected without a detector).
//
// Parameters:
//   - detector: The LoopDetector to wrap (can be nil)
//
// Returns:
//   - *LoopDetectionVerifier: Ready to use for testing
//
// Example:
//
//	detector := commands.NewLoopDetector()
//	verifier := NewLoopDetectionVerifier(detector)
//	assert.Equal(t, 0, verifier.GetLoopCount())
func NewLoopDetectionVerifier(detector *commands.LoopDetector) *LoopDetectionVerifier {
	return &LoopDetectionVerifier{
		detector: detector,
		detected: []LoopEvent{},
	}
}

// SimulateLoop simulates command generation for a component:ref pair by
// calling the detector's CheckLoop() method repeatedly.
//
// This method:
//  1. Loops for the specified number of iterations
//  2. Calls detector.CheckLoop(componentID, refID) each iteration
//  3. If a loop is detected (error returned), captures it in detected []LoopEvent
//  4. Stops simulating after first loop detection (mimics real behavior)
//
// If the detector is nil, this method is a no-op.
//
// Parameters:
//   - componentID: Unique identifier of the component
//   - refID: Unique identifier of the ref (e.g., "ref-42")
//   - iterations: Number of times to call CheckLoop()
//
// Example:
//
//	// Simulate 50 iterations (no loop expected)
//	verifier.SimulateLoop("counter", "count", 50)
//	assert.False(t, verifier.WasDetected())
//
//	// Simulate 150 iterations (loop expected at 101)
//	verifier.SimulateLoop("counter", "count", 150)
//	assert.True(t, verifier.WasDetected())
func (ldv *LoopDetectionVerifier) SimulateLoop(componentID, refID string, iterations int) {
	if ldv.detector == nil {
		return
	}

	for i := 0; i < iterations; i++ {
		err := ldv.detector.CheckLoop(componentID, refID)
		if err != nil {
			// Loop detected - capture event
			if loopErr, ok := err.(*commands.CommandLoopError); ok {
				ldv.detected = append(ldv.detected, LoopEvent{
					ComponentID:  loopErr.ComponentID,
					RefID:        loopErr.RefID,
					CommandCount: loopErr.CommandCount,
					DetectedAt:   time.Now(),
				})
			}
			// Stop simulating after loop detected (mimics real behavior)
			return
		}
	}
}

// GetDetectedLoops returns all detected loop events.
//
// The returned slice is a copy, so modifications to it do not affect the
// verifier's internal state. This is useful for inspecting loop details
// or verifying that specific loops were detected.
//
// Returns:
//   - []LoopEvent: All detected loops (empty slice if none detected)
//
// Example:
//
//	verifier.SimulateLoop("counter", "count", 150)
//	loops := verifier.GetDetectedLoops()
//	assert.Len(t, loops, 1)
//	assert.Equal(t, "counter", loops[0].ComponentID)
//	assert.Equal(t, "count", loops[0].RefID)
//	assert.Equal(t, 101, loops[0].CommandCount)
func (ldv *LoopDetectionVerifier) GetDetectedLoops() []LoopEvent {
	// Return a copy to prevent external modification
	result := make([]LoopEvent, len(ldv.detected))
	copy(result, ldv.detected)
	return result
}

// GetLoopCount returns the number of loops that have been detected.
//
// This provides a quick way to verify how many loops occurred during
// simulation without retrieving all loop details.
//
// Returns:
//   - int: Number of detected loops
//
// Example:
//
//	verifier.SimulateLoop("counter", "count", 150)
//	assert.Equal(t, 1, verifier.GetLoopCount())
//
//	detector.Reset()
//	verifier.SimulateLoop("counter", "count", 150)
//	assert.Equal(t, 2, verifier.GetLoopCount())
func (ldv *LoopDetectionVerifier) GetLoopCount() int {
	return len(ldv.detected)
}

// WasDetected returns true if any loops have been detected.
//
// This is a convenience method for checking if loop detection occurred
// without needing to check the loop count or retrieve loop details.
//
// Returns:
//   - bool: True if at least one loop was detected
//
// Example:
//
//	verifier.SimulateLoop("counter", "count", 50)
//	assert.False(t, verifier.WasDetected())
//
//	verifier.SimulateLoop("counter", "count", 150)
//	assert.True(t, verifier.WasDetected())
func (ldv *LoopDetectionVerifier) WasDetected() bool {
	return len(ldv.detected) > 0
}

// Clear resets all tracking state.
//
// This method clears the detected loops history. It's useful for resetting
// state between test cases or cleaning up after testing. Safe to call
// multiple times.
//
// Note: This does NOT reset the underlying LoopDetector. If you need to
// reset the detector's command counts, call detector.Reset() separately.
//
// Example:
//
//	verifier.SimulateLoop("counter", "count", 150)
//	assert.True(t, verifier.WasDetected())
//
//	verifier.Clear()
//	assert.False(t, verifier.WasDetected())
//	assert.Equal(t, 0, verifier.GetLoopCount())
func (ldv *LoopDetectionVerifier) Clear() {
	ldv.detected = []LoopEvent{}
}

// AssertLoopDetected asserts that at least one loop was detected.
//
// This is a convenience assertion method that checks if any loops were
// detected and reports a clear error message if not. It uses t.Helper()
// to ensure the error is reported at the correct line in the test.
//
// Parameters:
//   - t: The testing.T instance (or testingT interface)
//
// Example:
//
//	verifier.SimulateLoop("counter", "count", 150)
//	verifier.AssertLoopDetected(t) // Passes
//
//	verifier.Clear()
//	verifier.SimulateLoop("counter", "count", 50)
//	verifier.AssertLoopDetected(t) // Fails with clear error
func (ldv *LoopDetectionVerifier) AssertLoopDetected(t testingT) {
	t.Helper()

	if !ldv.WasDetected() {
		t.Errorf("loop detection: expected loop to be detected, but none was detected")
	}
}

// AssertNoLoop asserts that no loops were detected.
//
// This is a convenience assertion method that checks that no loops were
// detected and reports a clear error message if any were found. It uses
// t.Helper() to ensure the error is reported at the correct line in the test.
//
// Parameters:
//   - t: The testing.T instance (or testingT interface)
//
// Example:
//
//	verifier.SimulateLoop("counter", "count", 50)
//	verifier.AssertNoLoop(t) // Passes
//
//	verifier.SimulateLoop("counter", "count", 150)
//	verifier.AssertNoLoop(t) // Fails with clear error
func (ldv *LoopDetectionVerifier) AssertNoLoop(t testingT) {
	t.Helper()

	if ldv.WasDetected() {
		t.Errorf("loop detection: expected no loop, but detected %d loop(s)", ldv.GetLoopCount())
	}
}
