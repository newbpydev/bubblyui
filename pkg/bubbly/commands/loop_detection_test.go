package commands

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoopDetector_NormalOperation tests that legitimate rapid updates are allowed.
func TestLoopDetector_NormalOperation(t *testing.T) {
	tests := []struct {
		name          string
		updates       int
		componentID   string
		refID         string
		shouldError   bool
		errorContains string
	}{
		{
			name:        "single update",
			updates:     1,
			componentID: "comp-1",
			refID:       "ref-1",
			shouldError: false,
		},
		{
			name:        "50 updates (under limit)",
			updates:     50,
			componentID: "comp-1",
			refID:       "ref-1",
			shouldError: false,
		},
		{
			name:        "exactly at limit (100)",
			updates:     100,
			componentID: "comp-1",
			refID:       "ref-1",
			shouldError: false,
		},
		{
			name:          "exceeds limit (101)",
			updates:       101,
			componentID:   "comp-1",
			refID:         "ref-1",
			shouldError:   true,
			errorContains: "command generation loop detected",
		},
		{
			name:          "far exceeds limit (200)",
			updates:       200,
			componentID:   "comp-1",
			refID:         "ref-1",
			shouldError:   true,
			errorContains: "command generation loop detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewLoopDetector()

			var err error
			for i := 0; i < tt.updates; i++ {
				err = detector.CheckLoop(tt.componentID, tt.refID)
				if err != nil {
					break
				}
			}

			if tt.shouldError {
				require.Error(t, err, "expected error for %d updates", tt.updates)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err, "should allow %d updates", tt.updates)
			}
		})
	}
}

// TestLoopDetector_Reset tests that reset clears command counts.
func TestLoopDetector_Reset(t *testing.T) {
	detector := NewLoopDetector()

	// Generate 50 commands
	for i := 0; i < 50; i++ {
		err := detector.CheckLoop("comp-1", "ref-1")
		require.NoError(t, err)
	}

	// Reset detector
	detector.Reset()

	// Should be able to generate another 100 commands
	for i := 0; i < 100; i++ {
		err := detector.CheckLoop("comp-1", "ref-1")
		assert.NoError(t, err, "should allow 100 more commands after reset")
	}

	// 101st should fail
	err := detector.CheckLoop("comp-1", "ref-1")
	assert.Error(t, err, "should detect loop after 100 commands")
}

// TestLoopDetector_MultipleRefs tests tracking multiple refs independently.
func TestLoopDetector_MultipleRefs(t *testing.T) {
	detector := NewLoopDetector()

	// Each ref should have independent counter
	for i := 0; i < 100; i++ {
		err1 := detector.CheckLoop("comp-1", "ref-1")
		err2 := detector.CheckLoop("comp-1", "ref-2")
		err3 := detector.CheckLoop("comp-2", "ref-1")

		assert.NoError(t, err1, "ref-1 should allow 100 commands")
		assert.NoError(t, err2, "ref-2 should allow 100 commands")
		assert.NoError(t, err3, "comp-2/ref-1 should allow 100 commands")
	}

	// Each should fail independently
	err1 := detector.CheckLoop("comp-1", "ref-1")
	assert.Error(t, err1, "comp-1/ref-1 should detect loop")

	err2 := detector.CheckLoop("comp-1", "ref-2")
	assert.Error(t, err2, "comp-1/ref-2 should detect loop")

	err3 := detector.CheckLoop("comp-2", "ref-1")
	assert.Error(t, err3, "comp-2/ref-1 should detect loop")
}

// TestLoopDetector_ErrorMessage tests that error messages are clear and helpful.
func TestLoopDetector_ErrorMessage(t *testing.T) {
	detector := NewLoopDetector()

	// Generate 101 commands to trigger error
	var err error
	for i := 0; i < 101; i++ {
		err = detector.CheckLoop("counter-comp", "count-ref")
	}

	require.Error(t, err)

	// Error should contain helpful information
	errMsg := err.Error()
	assert.Contains(t, errMsg, "command generation loop detected", "should mention loop detection")
	assert.Contains(t, errMsg, "counter-comp", "should include component ID")
	assert.Contains(t, errMsg, "count-ref", "should include ref ID")
	assert.Contains(t, errMsg, "100", "should mention the limit")
}

// TestLoopDetector_ConcurrentAccess tests thread-safety with concurrent updates.
func TestLoopDetector_ConcurrentAccess(t *testing.T) {
	detector := NewLoopDetector()

	const goroutines = 10
	const updatesPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()

			componentID := fmt.Sprintf("comp-%d", id)
			refID := fmt.Sprintf("ref-%d", id)

			for j := 0; j < updatesPerGoroutine; j++ {
				err := detector.CheckLoop(componentID, refID)
				assert.NoError(t, err, "concurrent access should work")
			}
		}(i)
	}

	wg.Wait()
}

// TestLoopDetector_ResetConcurrentAccess tests thread-safety of reset.
func TestLoopDetector_ResetConcurrentAccess(t *testing.T) {
	detector := NewLoopDetector()

	const goroutines = 5
	const cycles = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()

			componentID := fmt.Sprintf("comp-%d", id)
			refID := fmt.Sprintf("ref-%d", id)

			for cycle := 0; cycle < cycles; cycle++ {
				// Generate some commands
				for j := 0; j < 10; j++ {
					detector.CheckLoop(componentID, refID)
				}

				// Reset
				detector.Reset()
			}
		}(i)
	}

	wg.Wait()
}

// TestLoopDetector_LegitimateRapidUpdates tests that rapid but non-looping updates work.
func TestLoopDetector_LegitimateRapidUpdates(t *testing.T) {
	tests := []struct {
		name        string
		scenario    string
		setupFunc   func(*LoopDetector) error
		shouldError bool
	}{
		{
			name:     "batch processing with reset between batches",
			scenario: "Process 100 items, reset, process 100 more",
			setupFunc: func(ld *LoopDetector) error {
				// First batch
				for i := 0; i < 100; i++ {
					if err := ld.CheckLoop("batch-processor", "items"); err != nil {
						return err
					}
				}

				// Reset between batches
				ld.Reset()

				// Second batch
				for i := 0; i < 100; i++ {
					if err := ld.CheckLoop("batch-processor", "items"); err != nil {
						return err
					}
				}

				return nil
			},
			shouldError: false,
		},
		{
			name:     "animation with 60 updates per second",
			scenario: "60 frames/second animation should work within limit",
			setupFunc: func(ld *LoopDetector) error {
				// 60 updates (typical for one second of animation)
				for i := 0; i < 60; i++ {
					if err := ld.CheckLoop("animation", "position"); err != nil {
						return err
					}
				}
				return nil
			},
			shouldError: false,
		},
		{
			name:     "form with multiple rapid field changes",
			scenario: "User types quickly in form fields",
			setupFunc: func(ld *LoopDetector) error {
				// Simulate rapid typing (20 characters)
				for i := 0; i < 20; i++ {
					if err := ld.CheckLoop("form", "name"); err != nil {
						return err
					}
				}

				// Simulate email field (15 characters)
				for i := 0; i < 15; i++ {
					if err := ld.CheckLoop("form", "email"); err != nil {
						return err
					}
				}

				return nil
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewLoopDetector()

			err := tt.setupFunc(detector)

			if tt.shouldError {
				assert.Error(t, err, "scenario %q should fail", tt.scenario)
			} else {
				assert.NoError(t, err, "scenario %q should succeed: %s", tt.name, tt.scenario)
			}
		})
	}
}

// TestLoopDetector_CountAccuracy tests that counts are tracked accurately.
func TestLoopDetector_CountAccuracy(t *testing.T) {
	detector := NewLoopDetector()

	// Generate exactly 99 commands
	for i := 0; i < 99; i++ {
		err := detector.CheckLoop("comp-1", "ref-1")
		require.NoError(t, err, "should allow 99 commands")
	}

	// 100th should still succeed
	err := detector.CheckLoop("comp-1", "ref-1")
	assert.NoError(t, err, "should allow 100th command")

	// 101st should fail
	err = detector.CheckLoop("comp-1", "ref-1")
	assert.Error(t, err, "should detect loop on 101st command")
}

// TestLoopDetector_NoFalsePositives tests that detector doesn't trigger incorrectly.
func TestLoopDetector_NoFalsePositives(t *testing.T) {
	detector := NewLoopDetector()

	// Scenario: Component with many refs, each used within limit
	for refNum := 0; refNum < 20; refNum++ {
		refID := fmt.Sprintf("ref-%d", refNum)

		// Each ref gets 50 updates (well under limit)
		for i := 0; i < 50; i++ {
			err := detector.CheckLoop("multi-ref-comp", refID)
			assert.NoError(t, err, "ref %s update %d should succeed", refID, i)
		}
	}

	// All refs should still work after reset
	detector.Reset()

	for refNum := 0; refNum < 20; refNum++ {
		refID := fmt.Sprintf("ref-%d", refNum)
		err := detector.CheckLoop("multi-ref-comp", refID)
		assert.NoError(t, err, "ref %s should work after reset", refID)
	}
}
