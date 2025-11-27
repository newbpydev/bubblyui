package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly/commands"
)

// TestNewLoopDetectionVerifier tests the constructor with various inputs.
func TestNewLoopDetectionVerifier(t *testing.T) {
	tests := []struct {
		name     string
		detector *commands.LoopDetector
		wantNil  bool
	}{
		{
			name:     "with valid detector",
			detector: commands.NewLoopDetector(),
			wantNil:  false,
		},
		{
			name:     "with nil detector",
			detector: nil,
			wantNil:  false, // Verifier itself should not be nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier := NewLoopDetectionVerifier(tt.detector)

			if tt.wantNil {
				assert.Nil(t, verifier)
			} else {
				assert.NotNil(t, verifier)
				assert.Equal(t, 0, verifier.GetLoopCount())
				assert.False(t, verifier.WasDetected())
			}
		})
	}
}

// TestSimulateLoop_NoLoop tests that no loop is detected below threshold.
func TestSimulateLoop_NoLoop(t *testing.T) {
	detector := commands.NewLoopDetector()
	verifier := NewLoopDetectionVerifier(detector)

	// Simulate 50 iterations (below 100 threshold)
	verifier.SimulateLoop("test-component", "test-ref", 50)

	// Should not detect loop
	assert.False(t, verifier.WasDetected())
	assert.Equal(t, 0, verifier.GetLoopCount())
	assert.Empty(t, verifier.GetDetectedLoops())
}

// TestSimulateLoop_DetectsLoop tests that loop is detected above threshold.
func TestSimulateLoop_DetectsLoop(t *testing.T) {
	detector := commands.NewLoopDetector()
	verifier := NewLoopDetectionVerifier(detector)

	// Simulate 150 iterations (above 100 threshold)
	verifier.SimulateLoop("test-component", "test-ref", 150)

	// Should detect loop
	assert.True(t, verifier.WasDetected())
	assert.Equal(t, 1, verifier.GetLoopCount())

	loops := verifier.GetDetectedLoops()
	assert.Len(t, loops, 1)
	assert.Equal(t, "test-component", loops[0].ComponentID)
	assert.Equal(t, "test-ref", loops[0].RefID)
	assert.Equal(t, 101, loops[0].CommandCount) // 101 triggers the error
}

// TestSimulateLoop_MultipleRefs tests that different refs are tracked independently.
func TestSimulateLoop_MultipleRefs(t *testing.T) {
	detector := commands.NewLoopDetector()
	verifier := NewLoopDetectionVerifier(detector)

	// Simulate 50 iterations for ref1 (no loop)
	verifier.SimulateLoop("component", "ref1", 50)
	assert.False(t, verifier.WasDetected())

	// Simulate 50 iterations for ref2 (no loop)
	verifier.SimulateLoop("component", "ref2", 50)
	assert.False(t, verifier.WasDetected())

	// Simulate 150 iterations for ref3 (loop detected)
	verifier.SimulateLoop("component", "ref3", 150)
	assert.True(t, verifier.WasDetected())
	assert.Equal(t, 1, verifier.GetLoopCount())

	loops := verifier.GetDetectedLoops()
	assert.Len(t, loops, 1)
	assert.Equal(t, "ref3", loops[0].RefID)
}

// TestSimulateLoop_NilDetector tests behavior with nil detector.
func TestSimulateLoop_NilDetector(t *testing.T) {
	verifier := NewLoopDetectionVerifier(nil)

	// Should not panic with nil detector
	assert.NotPanics(t, func() {
		verifier.SimulateLoop("test-component", "test-ref", 150)
	})

	// Should not detect any loops (no detector to detect with)
	assert.False(t, verifier.WasDetected())
	assert.Equal(t, 0, verifier.GetLoopCount())
}

// TestAssertLoopDetected tests the assertion method for detected loops.
func TestAssertLoopDetected(t *testing.T) {
	t.Run("passes when loop detected", func(t *testing.T) {
		detector := commands.NewLoopDetector()
		verifier := NewLoopDetectionVerifier(detector)

		// Trigger loop
		verifier.SimulateLoop("component", "ref", 150)

		// Should pass
		mockT := &mockTestingT{}
		verifier.AssertLoopDetected(mockT)
		assert.False(t, mockT.failed, "assertion should pass when loop detected")
	})

	t.Run("fails when no loop detected", func(t *testing.T) {
		detector := commands.NewLoopDetector()
		verifier := NewLoopDetectionVerifier(detector)

		// No loop triggered
		verifier.SimulateLoop("component", "ref", 50)

		// Should fail
		mockT := &mockTestingT{}
		verifier.AssertLoopDetected(mockT)
		assert.True(t, mockT.failed, "assertion should fail when no loop detected")
		assert.NotEmpty(t, mockT.errors)
		assert.Contains(t, mockT.errors[0], "expected loop to be detected")
	})
}

// TestAssertNoLoop tests the assertion method for no loops.
func TestAssertNoLoop(t *testing.T) {
	t.Run("passes when no loop detected", func(t *testing.T) {
		detector := commands.NewLoopDetector()
		verifier := NewLoopDetectionVerifier(detector)

		// No loop triggered
		verifier.SimulateLoop("component", "ref", 50)

		// Should pass
		mockT := &mockTestingT{}
		verifier.AssertNoLoop(mockT)
		assert.False(t, mockT.failed, "assertion should pass when no loop detected")
	})

	t.Run("fails when loop detected", func(t *testing.T) {
		detector := commands.NewLoopDetector()
		verifier := NewLoopDetectionVerifier(detector)

		// Trigger loop
		verifier.SimulateLoop("component", "ref", 150)

		// Should fail
		mockT := &mockTestingT{}
		verifier.AssertNoLoop(mockT)
		assert.True(t, mockT.failed, "assertion should fail when loop detected")
		assert.NotEmpty(t, mockT.errors)
		assert.Contains(t, mockT.errors[0], "expected no loop")
	})
}

// TestGetDetectedLoops tests retrieving detected loop events.
func TestGetDetectedLoops(t *testing.T) {
	detector := commands.NewLoopDetector()
	verifier := NewLoopDetectionVerifier(detector)

	// No loops initially
	loops := verifier.GetDetectedLoops()
	assert.Empty(t, loops)

	// Trigger first loop
	verifier.SimulateLoop("component1", "ref1", 150)
	loops = verifier.GetDetectedLoops()
	assert.Len(t, loops, 1)

	// Reset detector and trigger second loop
	detector.Reset()
	verifier.SimulateLoop("component2", "ref2", 150)
	loops = verifier.GetDetectedLoops()
	assert.Len(t, loops, 2)

	// Verify loop details
	assert.Equal(t, "component1", loops[0].ComponentID)
	assert.Equal(t, "ref1", loops[0].RefID)
	assert.Equal(t, "component2", loops[1].ComponentID)
	assert.Equal(t, "ref2", loops[1].RefID)

	// Verify returned slice is a copy (modifications don't affect internal state)
	loops[0].ComponentID = "modified"
	freshLoops := verifier.GetDetectedLoops()
	assert.Equal(t, "component1", freshLoops[0].ComponentID, "internal state should not be modified")
}

// TestClear tests resetting the verifier state.
func TestClear(t *testing.T) {
	detector := commands.NewLoopDetector()
	verifier := NewLoopDetectionVerifier(detector)

	// Trigger loop
	verifier.SimulateLoop("component", "ref", 150)
	assert.True(t, verifier.WasDetected())
	assert.Equal(t, 1, verifier.GetLoopCount())

	// Clear state
	verifier.Clear()
	assert.False(t, verifier.WasDetected())
	assert.Equal(t, 0, verifier.GetLoopCount())
	assert.Empty(t, verifier.GetDetectedLoops())
}

// TestWasDetected tests the boolean check for loop detection.
func TestWasDetected(t *testing.T) {
	tests := []struct {
		name       string
		iterations int
		want       bool
	}{
		{
			name:       "no loop - 50 iterations",
			iterations: 50,
			want:       false,
		},
		{
			name:       "no loop - 100 iterations (at threshold)",
			iterations: 100,
			want:       false,
		},
		{
			name:       "loop detected - 101 iterations",
			iterations: 101,
			want:       true,
		},
		{
			name:       "loop detected - 150 iterations",
			iterations: 150,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := commands.NewLoopDetector()
			verifier := NewLoopDetectionVerifier(detector)

			verifier.SimulateLoop("component", "ref", tt.iterations)

			assert.Equal(t, tt.want, verifier.WasDetected())
		})
	}
}

// TestGetLoopCount tests counting detected loops.
func TestGetLoopCount(t *testing.T) {
	detector := commands.NewLoopDetector()
	verifier := NewLoopDetectionVerifier(detector)

	// Initially zero
	assert.Equal(t, 0, verifier.GetLoopCount())

	// Trigger first loop
	verifier.SimulateLoop("component1", "ref1", 150)
	assert.Equal(t, 1, verifier.GetLoopCount())

	// Reset detector and trigger second loop
	detector.Reset()
	verifier.SimulateLoop("component2", "ref2", 150)
	assert.Equal(t, 2, verifier.GetLoopCount())

	// Clear and count should be zero
	verifier.Clear()
	assert.Equal(t, 0, verifier.GetLoopCount())
}
