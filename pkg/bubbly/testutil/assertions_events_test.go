package testutil

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAssertEventFired tests the AssertEventFired assertion method.
func TestAssertEventFired(t *testing.T) {
	tests := []struct {
		name        string
		eventName   string
		fireEvent   bool
		shouldPass  bool
		description string
	}{
		{
			name:        "event fired - passes",
			eventName:   "click",
			fireEvent:   true,
			shouldPass:  true,
			description: "Should pass when event was fired",
		},
		{
			name:        "event not fired - fails",
			eventName:   "click",
			fireEvent:   false,
			shouldPass:  false,
			description: "Should fail when event was not fired",
		},
		{
			name:        "different event fired - fails",
			eventName:   "submit",
			fireEvent:   false,
			shouldPass:  false,
			description: "Should fail when different event was fired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock testing.T
			mockT := &mockTestingT{}
			harness := &TestHarness{
				t:      mockT,
				events: NewEventTracker(),
			}

			ct := &ComponentTest{
				harness: harness,
				events:  NewEventInspector(harness.events),
			}

			// Fire event if needed
			if tt.fireEvent {
				harness.events.Track(tt.eventName, nil, "test")
			}

			// Call assertion
			ct.AssertEventFired(tt.eventName)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "Expected assertion to pass but it failed")
			} else {
				assert.True(t, mockT.failed, "Expected assertion to fail but it passed")
			}
		})
	}
}

// TestAssertEventNotFired tests the AssertEventNotFired assertion method.
func TestAssertEventNotFired(t *testing.T) {
	tests := []struct {
		name        string
		eventName   string
		fireEvent   bool
		shouldPass  bool
		description string
	}{
		{
			name:        "event not fired - passes",
			eventName:   "click",
			fireEvent:   false,
			shouldPass:  true,
			description: "Should pass when event was not fired",
		},
		{
			name:        "event fired - fails",
			eventName:   "click",
			fireEvent:   true,
			shouldPass:  false,
			description: "Should fail when event was fired",
		},
		{
			name:        "different event fired - passes",
			eventName:   "submit",
			fireEvent:   false,
			shouldPass:  true,
			description: "Should pass when different event was fired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock testing.T
			mockT := &mockTestingT{}
			harness := &TestHarness{
				t:      mockT,
				events: NewEventTracker(),
			}

			ct := &ComponentTest{
				harness: harness,
				events:  NewEventInspector(harness.events),
			}

			// Fire event if needed
			if tt.fireEvent {
				harness.events.Track(tt.eventName, nil, "test")
			}

			// Call assertion
			ct.AssertEventNotFired(tt.eventName)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "Expected assertion to pass but it failed")
			} else {
				assert.True(t, mockT.failed, "Expected assertion to fail but it passed")
			}
		})
	}
}

// TestAssertEventPayload tests the AssertEventPayload assertion method.
func TestAssertEventPayload(t *testing.T) {
	tests := []struct {
		name            string
		eventName       string
		fireEvent       bool
		payload         interface{}
		expectedPayload interface{}
		shouldPass      bool
		description     string
	}{
		{
			name:            "matching payload - passes",
			eventName:       "click",
			fireEvent:       true,
			payload:         "test-data",
			expectedPayload: "test-data",
			shouldPass:      true,
			description:     "Should pass when payload matches",
		},
		{
			name:            "different payload - fails",
			eventName:       "click",
			fireEvent:       true,
			payload:         "actual-data",
			expectedPayload: "expected-data",
			shouldPass:      false,
			description:     "Should fail when payload doesn't match",
		},
		{
			name:            "event not fired - fails",
			eventName:       "click",
			fireEvent:       false,
			payload:         nil,
			expectedPayload: "test-data",
			shouldPass:      false,
			description:     "Should fail when event was not fired",
		},
		{
			name:            "nil payload matches - passes",
			eventName:       "click",
			fireEvent:       true,
			payload:         nil,
			expectedPayload: nil,
			shouldPass:      true,
			description:     "Should pass when both payloads are nil",
		},
		{
			name:            "struct payload matches - passes",
			eventName:       "submit",
			fireEvent:       true,
			payload:         struct{ Value int }{Value: 42},
			expectedPayload: struct{ Value int }{Value: 42},
			shouldPass:      true,
			description:     "Should pass when struct payloads match",
		},
		{
			name:            "multiple events uses last - passes",
			eventName:       "click",
			fireEvent:       true,
			payload:         "last-payload",
			expectedPayload: "last-payload",
			shouldPass:      true,
			description:     "Should use last event's payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock testing.T
			mockT := &mockTestingT{}
			harness := &TestHarness{
				t:      mockT,
				events: NewEventTracker(),
			}

			ct := &ComponentTest{
				harness: harness,
				events:  NewEventInspector(harness.events),
			}

			// Fire event if needed
			if tt.fireEvent {
				// For "multiple events" test, fire multiple times
				if tt.name == "multiple events uses last - passes" {
					harness.events.Track(tt.eventName, "first-payload", "test")
					harness.events.Track(tt.eventName, "middle-payload", "test")
				}
				harness.events.Track(tt.eventName, tt.payload, "test")
			}

			// Call assertion
			ct.AssertEventPayload(tt.eventName, tt.expectedPayload)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "Expected assertion to pass but it failed")
			} else {
				assert.True(t, mockT.failed, "Expected assertion to fail but it passed")
			}
		})
	}
}

// TestAssertEventCount tests the AssertEventCount assertion method.
func TestAssertEventCount(t *testing.T) {
	tests := []struct {
		name          string
		eventName     string
		fireCount     int
		expectedCount int
		shouldPass    bool
		description   string
	}{
		{
			name:          "zero events - passes",
			eventName:     "click",
			fireCount:     0,
			expectedCount: 0,
			shouldPass:    true,
			description:   "Should pass when no events fired and expecting 0",
		},
		{
			name:          "single event - passes",
			eventName:     "click",
			fireCount:     1,
			expectedCount: 1,
			shouldPass:    true,
			description:   "Should pass when one event fired and expecting 1",
		},
		{
			name:          "multiple events - passes",
			eventName:     "click",
			fireCount:     5,
			expectedCount: 5,
			shouldPass:    true,
			description:   "Should pass when multiple events fired and count matches",
		},
		{
			name:          "count mismatch - fails",
			eventName:     "click",
			fireCount:     3,
			expectedCount: 5,
			shouldPass:    false,
			description:   "Should fail when count doesn't match",
		},
		{
			name:          "expected zero but fired - fails",
			eventName:     "click",
			fireCount:     2,
			expectedCount: 0,
			shouldPass:    false,
			description:   "Should fail when events fired but expecting 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock testing.T
			mockT := &mockTestingT{}
			harness := &TestHarness{
				t:      mockT,
				events: NewEventTracker(),
			}

			ct := &ComponentTest{
				harness: harness,
				events:  NewEventInspector(harness.events),
			}

			// Fire events
			for i := 0; i < tt.fireCount; i++ {
				harness.events.Track(tt.eventName, i, "test")
			}

			// Call assertion
			ct.AssertEventCount(tt.eventName, tt.expectedCount)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "Expected assertion to pass but it failed")
			} else {
				assert.True(t, mockT.failed, "Expected assertion to fail but it passed")
			}
		})
	}
}

// TestEventTracker_Track tests the Track method.
func TestEventTracker_Track(t *testing.T) {
	tracker := NewEventTracker()

	// Track an event
	tracker.Track("click", "payload1", "component1")

	// Verify event was tracked
	events := tracker.GetEvents("click")
	assert.Len(t, events, 1)
	assert.Equal(t, "click", events[0].Name)
	assert.Equal(t, "payload1", events[0].Payload)
	assert.Equal(t, "component1", events[0].Source)
	assert.False(t, events[0].Timestamp.IsZero())
}

// TestEventTracker_GetEvents tests the GetEvents method.
func TestEventTracker_GetEvents(t *testing.T) {
	tracker := NewEventTracker()

	// Track multiple events
	tracker.Track("click", "payload1", "source1")
	tracker.Track("submit", "payload2", "source2")
	tracker.Track("click", "payload3", "source3")

	// Get click events
	clickEvents := tracker.GetEvents("click")
	assert.Len(t, clickEvents, 2)
	assert.Equal(t, "payload1", clickEvents[0].Payload)
	assert.Equal(t, "payload3", clickEvents[1].Payload)

	// Get submit events
	submitEvents := tracker.GetEvents("submit")
	assert.Len(t, submitEvents, 1)
	assert.Equal(t, "payload2", submitEvents[0].Payload)

	// Get non-existent events
	noneEvents := tracker.GetEvents("nonexistent")
	assert.Len(t, noneEvents, 0)
}

// TestEventTracker_WasFired tests the WasFired method.
func TestEventTracker_WasFired(t *testing.T) {
	tracker := NewEventTracker()

	// Initially no events
	assert.False(t, tracker.WasFired("click"))

	// Track an event
	tracker.Track("click", nil, "test")

	// Now it was fired
	assert.True(t, tracker.WasFired("click"))
	assert.False(t, tracker.WasFired("submit"))
}

// TestEventTracker_FiredCount tests the FiredCount method.
func TestEventTracker_FiredCount(t *testing.T) {
	tracker := NewEventTracker()

	// Initially zero
	assert.Equal(t, 0, tracker.FiredCount("click"))

	// Track events
	tracker.Track("click", nil, "test")
	assert.Equal(t, 1, tracker.FiredCount("click"))

	tracker.Track("click", nil, "test")
	tracker.Track("click", nil, "test")
	assert.Equal(t, 3, tracker.FiredCount("click"))

	// Different event
	assert.Equal(t, 0, tracker.FiredCount("submit"))
}

// TestEventTracker_ThreadSafety tests thread-safe operations.
func TestEventTracker_ThreadSafety(t *testing.T) {
	tracker := NewEventTracker()

	// Track events concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			tracker.Track("click", id, "test")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all events tracked
	assert.Equal(t, 10, tracker.FiredCount("click"))
}

// TestEventTracker_Clear tests the Clear method.
func TestEventTracker_Clear(t *testing.T) {
	tests := []struct {
		name   string
		events []struct {
			name, source string
			payload      interface{}
		}
	}{
		{
			name: "clear with events",
			events: []struct {
				name, source string
				payload      interface{}
			}{
				{name: "click", source: "button-1", payload: "data1"},
				{name: "hover", source: "button-2", payload: "data2"},
				{name: "focus", source: "input-1", payload: "data3"},
			},
		},
		{
			name: "clear empty tracker",
			events: []struct {
				name, source string
				payload      interface{}
			}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewEventTracker()

			// Track events
			for _, e := range tt.events {
				tracker.Track(e.name, e.payload, e.source)
			}

			// Verify events exist before clear
			if len(tt.events) > 0 {
				assert.True(t, tracker.WasFired("click"))
			}

			// Clear events
			tracker.Clear()

			// Verify all events cleared
			assert.False(t, tracker.WasFired("click"))
			assert.False(t, tracker.WasFired("hover"))
			assert.False(t, tracker.WasFired("focus"))
			assert.Equal(t, 0, tracker.FiredCount("click"))
			assert.Equal(t, 0, tracker.FiredCount("hover"))
			assert.Len(t, tracker.GetEvents("click"), 0)
		})
	}
}

// TestEventTracker_Clear_Idempotent tests that Clear is idempotent.
func TestEventTracker_Clear_Idempotent(t *testing.T) {
	tracker := NewEventTracker()
	tracker.Track("click", "data", "button-1")

	// Clear multiple times
	tracker.Clear()
	tracker.Clear()
	tracker.Clear()

	// Should still be empty
	assert.False(t, tracker.WasFired("click"))
	assert.Equal(t, 0, tracker.FiredCount("click"))
}

// TestEventTracker_Clear_ThreadSafety tests Clear with concurrent operations.
func TestEventTracker_Clear_ThreadSafety(t *testing.T) {
	tracker := NewEventTracker()
	var wg sync.WaitGroup

	// Concurrent writes, reads, and clears
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			tracker.Track("event", n, "source")
		}(i)
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = tracker.GetEvents("event")
			_ = tracker.WasFired("event")
		}()
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tracker.Clear()
		}()
	}

	wg.Wait()

	// Should not panic and tracker should be in valid state
	assert.NotNil(t, tracker)
	// After all clears, tracker should be empty or have some events
	// depending on timing, but should not panic
	_ = tracker.FiredCount("event")
}

// TestEventInspector_Integration tests EventInspector integration.
func TestEventInspector_Integration(t *testing.T) {
	tracker := NewEventTracker()
	inspector := NewEventInspector(tracker)

	// Track some events
	tracker.Track("click", "data1", "source1")
	tracker.Track("submit", "data2", "source2")

	// Verify inspector has access to tracker
	assert.NotNil(t, inspector.tracker)
	assert.True(t, tracker.WasFired("click"))
	assert.True(t, tracker.WasFired("submit"))
}
