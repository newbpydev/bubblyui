package testutil

// AssertCommandEnqueued asserts that the command queue has exactly the expected number of commands.
//
// This is a high-level assertion helper that provides clear error messages when the
// command queue doesn't match expectations. It's designed to make command testing
// more readable and maintainable.
//
// The function checks:
//  1. That the queue inspector is not nil
//  2. That the queue length matches the expected count
//
// If the assertion fails, it provides a detailed error message including:
//   - The expected command count
//   - The actual command count
//   - Clear indication of what went wrong
//
// Parameters:
//   - t: The testing.T instance (or compatible interface)
//   - queue: The CommandQueueInspector to check
//   - count: The expected number of commands in the queue
//
// Example usage:
//
//	tester := NewAutoCommandTester(component)
//	tester.EnableAutoCommands()
//	tester.TriggerStateChange("count", 42)
//
//	queue := tester.GetQueueInspector()
//	AssertCommandEnqueued(t, queue, 1) // Assert exactly 1 command was enqueued
//
// Error messages:
//   - "command queue inspector is nil" - when queue parameter is nil
//   - "expected X commands enqueued, got Y" - when counts don't match
func AssertCommandEnqueued(t testingT, queue *CommandQueueInspector, count int) {
	t.Helper()

	// Check for nil queue
	if queue == nil {
		t.Errorf("command queue inspector is nil")
		return
	}

	// Get actual queue length
	actual := queue.Len()

	// Compare with expected
	if actual != count {
		t.Errorf("expected %d commands enqueued, got %d", count, actual)
	}
}

// AssertNoCommandLoop asserts that no command loop has been detected.
//
// This is a high-level assertion helper that verifies the loop detection system
// hasn't detected any infinite command generation loops. It's designed to make
// loop detection testing more readable and maintainable.
//
// The function checks:
//  1. That the loop detector is not nil
//  2. That no loops have been detected
//
// If the assertion fails, it provides a detailed error message including:
//   - That a loop was detected
//   - The number of iterations that triggered the detection
//   - Clear indication of what went wrong
//
// Parameters:
//   - t: The testing.T instance (or compatible interface)
//   - detector: The LoopDetectionVerifier to check
//
// Example usage:
//
//	tester := NewAutoCommandTester(component)
//	tester.EnableAutoCommands()
//
//	// Trigger some state changes
//	for i := 0; i < 10; i++ {
//	    tester.TriggerStateChange("count", i)
//	}
//
//	detector := tester.GetLoopDetector()
//	AssertNoCommandLoop(t, detector) // Assert no loop was detected
//
// Error messages:
//   - "loop detector is nil" - when detector parameter is nil
//   - "command loop detected: X iterations" - when a loop was detected
func AssertNoCommandLoop(t testingT, detector *LoopDetectionVerifier) {
	t.Helper()

	// Check for nil detector
	if detector == nil {
		t.Errorf("loop detector is nil")
		return
	}

	// Check if loop was detected
	if detector.WasDetected() {
		loopCount := detector.GetLoopCount()
		t.Errorf("command loop detected: %d iterations", loopCount)
	}
}
