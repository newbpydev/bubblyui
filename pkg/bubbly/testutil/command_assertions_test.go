package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAssertCommandEnqueued tests the AssertCommandEnqueued helper
func TestAssertCommandEnqueued(t *testing.T) {
	tests := []struct {
		name          string
		setupCommands int // Number of commands to enqueue before assertion
		expectedCount int // Expected count to assert
		shouldFail    bool
	}{
		{
			name:          "zero commands expected and found",
			setupCommands: 0,
			expectedCount: 0,
			shouldFail:    false,
		},
		{
			name:          "one command expected and found",
			setupCommands: 1,
			expectedCount: 1,
			shouldFail:    false,
		},
		{
			name:          "multiple commands expected and found",
			setupCommands: 5,
			expectedCount: 5,
			shouldFail:    false,
		},
		{
			name:          "expected more commands than enqueued",
			setupCommands: 2,
			expectedCount: 5,
			shouldFail:    true,
		},
		{
			name:          "expected fewer commands than enqueued",
			setupCommands: 5,
			expectedCount: 2,
			shouldFail:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock testing.T to capture failures
			mockT := &mockTestingT{}

			// Create test harness with component
			component := createAutoTestComponentWithRef("count")
			tester := NewAutoCommandTester(component)
			tester.EnableAutoCommands()

			// Enqueue the specified number of commands
			for i := 0; i < tt.setupCommands; i++ {
				tester.TriggerStateChange("count", i)
			}

			// Get the queue inspector
			queue := tester.GetQueueInspector()

			// Call the assertion helper
			AssertCommandEnqueued(mockT, queue, tt.expectedCount)

			// Verify the assertion result
			if tt.shouldFail {
				assert.True(t, mockT.failed, "assertion should have failed")
				assert.NotEmpty(t, mockT.errors, "should have error message")
			} else {
				assert.False(t, mockT.failed, "assertion should have passed")
			}
		})
	}
}

// TestAssertCommandEnqueued_NilQueue tests behavior with nil queue
func TestAssertCommandEnqueued_NilQueue(t *testing.T) {
	mockT := &mockTestingT{}

	// Call with nil queue
	AssertCommandEnqueued(mockT, nil, 0)

	// Should fail with clear error
	assert.True(t, mockT.failed, "should fail with nil queue")
	assert.NotEmpty(t, mockT.errors, "should have error message")
	assert.Contains(t, mockT.errors[0], "queue inspector is nil", "should mention nil queue")
}

// TestAssertNoCommandLoop tests the AssertNoCommandLoop helper
func TestAssertNoCommandLoop(t *testing.T) {
	tests := []struct {
		name       string
		setupLoop  bool // Whether to simulate a loop
		shouldFail bool
	}{
		{
			name:       "no loop detected",
			setupLoop:  false,
			shouldFail: false,
		},
		{
			name:       "loop detected",
			setupLoop:  true,
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock testing.T to capture failures
			mockT := &mockTestingT{}

			// Create a fresh loop detector for testing
			// (not from component, to ensure clean state)
			detector := NewLoopDetectionVerifier(nil)

			// Simulate loop if needed
			if tt.setupLoop {
				// Manually set detected loops to simulate detection
				// This mimics what would happen if a loop was actually detected
				detector.detected = []LoopEvent{
					{
						ComponentID:  "test-component",
						RefID:        "count",
						CommandCount: 150,
					},
				}
			}

			// Call the assertion helper
			AssertNoCommandLoop(mockT, detector)

			// Verify the assertion result
			if tt.shouldFail {
				assert.True(t, mockT.failed, "assertion should have failed")
				assert.NotEmpty(t, mockT.errors, "should have error message")
				assert.Contains(t, mockT.errors[0], "loop", "error should mention loop")
			} else {
				assert.False(t, mockT.failed, "assertion should have passed")
			}
		})
	}
}

// TestAssertNoCommandLoop_NilDetector tests behavior with nil detector
func TestAssertNoCommandLoop_NilDetector(t *testing.T) {
	mockT := &mockTestingT{}

	// Call with nil detector
	AssertNoCommandLoop(mockT, nil)

	// Should fail with clear error
	assert.True(t, mockT.failed, "should fail with nil detector")
	assert.NotEmpty(t, mockT.errors, "should have error message")
	assert.Contains(t, mockT.errors[0], "detector is nil", "should mention nil detector")
}

// TestAssertCommandEnqueued_ErrorMessages tests error message quality
func TestAssertCommandEnqueued_ErrorMessages(t *testing.T) {
	mockT := &mockTestingT{}

	// Create queue with 3 commands
	component := createAutoTestComponentWithRef("count")
	tester := NewAutoCommandTester(component)
	tester.EnableAutoCommands()

	for i := 0; i < 3; i++ {
		tester.TriggerStateChange("count", i)
	}

	queue := tester.GetQueueInspector()

	// Assert expecting 5 commands (should fail)
	AssertCommandEnqueued(mockT, queue, 5)

	// Verify error message contains useful information
	assert.NotEmpty(t, mockT.errors, "should have error message")
	assert.Contains(t, mockT.errors[0], "expected 5", "should mention expected count")
	assert.Contains(t, mockT.errors[0], "got 3", "should mention actual count")
}

// TestAssertNoCommandLoop_ErrorMessages tests error message quality
func TestAssertNoCommandLoop_ErrorMessages(t *testing.T) {
	mockT := &mockTestingT{}

	// Create detector with simulated loop
	detector := NewLoopDetectionVerifier(nil)
	detector.detected = []LoopEvent{
		{
			ComponentID:  "test-component",
			RefID:        "count",
			CommandCount: 150,
		},
	}

	// Assert no loop (should fail)
	AssertNoCommandLoop(mockT, detector)

	// Verify error message contains useful information
	assert.NotEmpty(t, mockT.errors, "should have error message")
	assert.Contains(t, mockT.errors[0], "loop detected", "should mention loop detection")
	assert.Contains(t, mockT.errors[0], "1", "should mention loop count")
}
