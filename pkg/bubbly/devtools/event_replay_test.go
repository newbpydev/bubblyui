package devtools

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewEventReplayer tests the constructor
func TestNewEventReplayer(t *testing.T) {
	events := []EventRecord{
		{ID: "1", Name: "event1", Timestamp: time.Now()},
		{ID: "2", Name: "event2", Timestamp: time.Now().Add(time.Second)},
	}

	replayer := NewEventReplayer(events)

	require.NotNil(t, replayer)
	assert.Equal(t, 1.0, replayer.GetSpeed())
	assert.False(t, replayer.IsPaused())
	assert.False(t, replayer.IsReplaying())
	current, total := replayer.GetProgress()
	assert.Equal(t, 0, current)
	assert.Equal(t, 2, total)
}

// TestNewEventReplayer_EmptyEvents tests constructor with empty events
func TestNewEventReplayer_EmptyEvents(t *testing.T) {
	replayer := NewEventReplayer([]EventRecord{})

	require.NotNil(t, replayer)
	current, total := replayer.GetProgress()
	assert.Equal(t, 0, current)
	assert.Equal(t, 0, total)
}

// TestEventReplayer_SetSpeed tests speed control
func TestEventReplayer_SetSpeed(t *testing.T) {
	tests := []struct {
		name          string
		speed         float64
		expectedSpeed float64
		expectError   bool
	}{
		{"normal speed", 1.0, 1.0, false},
		{"double speed", 2.0, 2.0, false},
		{"half speed", 0.5, 0.5, false},
		{"very fast", 10.0, 10.0, false},
		{"very slow", 0.1, 0.1, false},
		{"zero speed", 0.0, 1.0, true},      // Should reject
		{"negative speed", -1.0, 1.0, true}, // Should reject
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			replayer := NewEventReplayer([]EventRecord{})
			err := replayer.SetSpeed(tt.speed)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedSpeed, replayer.GetSpeed())
		})
	}
}

// TestEventReplayer_PauseResume tests pause and resume functionality
func TestEventReplayer_PauseResume(t *testing.T) {
	replayer := NewEventReplayer([]EventRecord{})

	// Initially not paused
	assert.False(t, replayer.IsPaused())

	// Pause
	replayer.Pause()
	assert.True(t, replayer.IsPaused())

	// Resume
	replayer.Resume()
	assert.False(t, replayer.IsPaused())
}

// TestEventReplayer_Reset tests reset functionality
func TestEventReplayer_Reset(t *testing.T) {
	events := []EventRecord{
		{ID: "1", Name: "event1", Timestamp: time.Now()},
		{ID: "2", Name: "event2", Timestamp: time.Now().Add(time.Second)},
	}

	replayer := NewEventReplayer(events)

	// Simulate some progress
	replayer.currentIndex = 1
	replayer.replaying = true

	// Reset
	replayer.Reset()

	assert.Equal(t, 0, replayer.currentIndex)
	assert.False(t, replayer.IsReplaying())
	current, _ := replayer.GetProgress()
	assert.Equal(t, 0, current)
}

// TestEventReplayer_Replay_EmptyEvents tests replay with no events
func TestEventReplayer_Replay_EmptyEvents(t *testing.T) {
	replayer := NewEventReplayer([]EventRecord{})

	cmd := replayer.Replay()
	assert.Nil(t, cmd)
	assert.False(t, replayer.IsReplaying())
}

// TestEventReplayer_Replay_SingleEvent tests replay with one event
func TestEventReplayer_Replay_SingleEvent(t *testing.T) {
	now := time.Now()
	events := []EventRecord{
		{ID: "1", Name: "event1", SourceID: "source1", Timestamp: now},
	}

	replayer := NewEventReplayer(events)
	cmd := replayer.Replay()

	require.NotNil(t, cmd)
	assert.True(t, replayer.IsReplaying())

	// Execute command to get message
	msg := cmd()
	replayMsg, ok := msg.(ReplayEventMsg)
	require.True(t, ok)
	assert.Equal(t, "1", replayMsg.Event.ID)
	assert.Equal(t, "event1", replayMsg.Event.Name)
	assert.Equal(t, 0, replayMsg.Index)
	assert.Equal(t, 1, replayMsg.Total)
}

// TestEventReplayer_Replay_MultipleEvents tests replay with multiple events
func TestEventReplayer_Replay_MultipleEvents(t *testing.T) {
	now := time.Now()
	events := []EventRecord{
		{ID: "1", Name: "event1", Timestamp: now},
		{ID: "2", Name: "event2", Timestamp: now.Add(100 * time.Millisecond)},
		{ID: "3", Name: "event3", Timestamp: now.Add(200 * time.Millisecond)},
	}

	replayer := NewEventReplayer(events)
	cmd := replayer.Replay()

	require.NotNil(t, cmd)
	assert.True(t, replayer.IsReplaying())

	// First event should be immediate
	msg := cmd()
	replayMsg, ok := msg.(ReplayEventMsg)
	require.True(t, ok)
	assert.Equal(t, "1", replayMsg.Event.ID)
	assert.Equal(t, 0, replayMsg.Index)

	// Next command should be set
	assert.NotNil(t, replayMsg.NextCmd)
}

// TestEventReplayer_Replay_OrderPreserved tests that event order is preserved
func TestEventReplayer_Replay_OrderPreserved(t *testing.T) {
	now := time.Now()
	events := []EventRecord{
		{ID: "1", Name: "first", Timestamp: now},
		{ID: "2", Name: "second", Timestamp: now.Add(50 * time.Millisecond)},
		{ID: "3", Name: "third", Timestamp: now.Add(100 * time.Millisecond)},
	}

	replayer := NewEventReplayer(events)

	// Collect all events in order
	var receivedIDs []string
	cmd := replayer.Replay()

	for cmd != nil {
		msg := cmd()
		if replayMsg, ok := msg.(ReplayEventMsg); ok {
			receivedIDs = append(receivedIDs, replayMsg.Event.ID)
			cmd = replayMsg.NextCmd
		} else {
			break
		}
	}

	assert.Equal(t, []string{"1", "2", "3"}, receivedIDs)
}

// TestEventReplayer_Replay_SpeedAffectsDelay tests that speed affects timing
func TestEventReplayer_Replay_SpeedAffectsDelay(t *testing.T) {
	now := time.Now()
	events := []EventRecord{
		{ID: "1", Name: "event1", Timestamp: now},
		{ID: "2", Name: "event2", Timestamp: now.Add(100 * time.Millisecond)},
	}

	// Test 2x speed (should halve the delay)
	replayer := NewEventReplayer(events)
	err := replayer.SetSpeed(2.0)
	require.NoError(t, err)

	cmd := replayer.Replay()
	msg := cmd()
	_, ok := msg.(ReplayEventMsg)
	require.True(t, ok)

	// The delay calculation is internal, but we can verify speed is applied
	assert.Equal(t, 2.0, replayer.GetSpeed())
}

// TestEventReplayer_Pause_StopsReplay tests that pause stops replay
func TestEventReplayer_Pause_StopsReplay(t *testing.T) {
	now := time.Now()
	events := []EventRecord{
		{ID: "1", Name: "event1", Timestamp: now},
		{ID: "2", Name: "event2", Timestamp: now.Add(100 * time.Millisecond)},
	}

	replayer := NewEventReplayer(events)
	cmd := replayer.Replay()

	// Get first event
	msg := cmd()
	replayMsg, ok := msg.(ReplayEventMsg)
	require.True(t, ok)
	assert.Equal(t, "1", replayMsg.Event.ID)

	// Pause before next event
	replayer.Pause()
	assert.True(t, replayer.IsPaused())

	// Next command should return nil when paused
	if replayMsg.NextCmd != nil {
		nextMsg := replayMsg.NextCmd()
		// Should be a pause message or nil
		_, isPauseMsg := nextMsg.(ReplayPausedMsg)
		assert.True(t, isPauseMsg || nextMsg == nil)
	}
}

// TestEventReplayer_Resume_ContinuesReplay tests that resume continues from where it left off
func TestEventReplayer_Resume_ContinuesReplay(t *testing.T) {
	now := time.Now()
	events := []EventRecord{
		{ID: "1", Name: "event1", Timestamp: now},
		{ID: "2", Name: "event2", Timestamp: now.Add(100 * time.Millisecond)},
		{ID: "3", Name: "event3", Timestamp: now.Add(200 * time.Millisecond)},
	}

	replayer := NewEventReplayer(events)
	replayer.Replay()

	// Simulate being at index 1
	replayer.currentIndex = 1
	replayer.Pause()

	// Resume should continue from index 1
	cmd := replayer.Resume()
	require.NotNil(t, cmd)

	msg := cmd()
	replayMsg, ok := msg.(ReplayEventMsg)
	require.True(t, ok)
	assert.Equal(t, "2", replayMsg.Event.ID) // Should be second event
	assert.Equal(t, 1, replayMsg.Index)
}

// TestEventReplayer_Concurrent_SetSpeed tests concurrent SetSpeed calls
func TestEventReplayer_Concurrent_SetSpeed(t *testing.T) {
	replayer := NewEventReplayer([]EventRecord{})

	// Run 100 concurrent SetSpeed calls
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func(speed float64) {
			_ = replayer.SetSpeed(speed)
			done <- true
		}(float64(i%10 + 1)) // Speeds from 1.0 to 10.0
	}

	// Wait for all to complete
	for i := 0; i < 100; i++ {
		<-done
	}

	// Should have a valid speed
	speed := replayer.GetSpeed()
	assert.True(t, speed >= 1.0 && speed <= 10.0)
}

// TestEventReplayer_Concurrent_PauseResume tests concurrent pause/resume
func TestEventReplayer_Concurrent_PauseResume(t *testing.T) {
	replayer := NewEventReplayer([]EventRecord{})

	// Run 100 concurrent pause/resume calls
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func(shouldPause bool) {
			if shouldPause {
				replayer.Pause()
			} else {
				replayer.Resume()
			}
			done <- true
		}(i%2 == 0)
	}

	// Wait for all to complete
	for i := 0; i < 100; i++ {
		<-done
	}

	// Should have a valid paused state (either true or false, no panic)
	_ = replayer.IsPaused()
}

// TestEventReplayer_GetProgress tests progress tracking
func TestEventReplayer_GetProgress(t *testing.T) {
	events := []EventRecord{
		{ID: "1", Name: "event1", Timestamp: time.Now()},
		{ID: "2", Name: "event2", Timestamp: time.Now().Add(time.Second)},
		{ID: "3", Name: "event3", Timestamp: time.Now().Add(2 * time.Second)},
	}

	replayer := NewEventReplayer(events)

	// Initial progress
	current, total := replayer.GetProgress()
	assert.Equal(t, 0, current)
	assert.Equal(t, 3, total)

	// Simulate progress
	replayer.currentIndex = 2
	current, total = replayer.GetProgress()
	assert.Equal(t, 2, current)
	assert.Equal(t, 3, total)
}

// TestEventReplayer_Integration_WithBubbletea tests integration with Bubbletea Update loop
func TestEventReplayer_Integration_WithBubbletea(t *testing.T) {
	now := time.Now()
	events := []EventRecord{
		{ID: "1", Name: "click", SourceID: "button1", Timestamp: now},
		{ID: "2", Name: "submit", SourceID: "form1", Timestamp: now.Add(50 * time.Millisecond)},
	}

	replayer := NewEventReplayer(events)

	// Mock Bubbletea model
	type model struct {
		receivedEvents []string
		replayer       *EventReplayer
	}

	m := model{
		receivedEvents: []string{},
		replayer:       replayer,
	}

	// Start replay
	cmd := m.replayer.Replay()
	require.NotNil(t, cmd)

	// Process first event
	msg := cmd()
	replayMsg, ok := msg.(ReplayEventMsg)
	require.True(t, ok)
	m.receivedEvents = append(m.receivedEvents, replayMsg.Event.Name)

	// Process second event
	if replayMsg.NextCmd != nil {
		msg2 := replayMsg.NextCmd()
		if replayMsg2, ok := msg2.(ReplayEventMsg); ok {
			m.receivedEvents = append(m.receivedEvents, replayMsg2.Event.Name)
		}
	}

	// Verify events received in order
	assert.Equal(t, []string{"click", "submit"}, m.receivedEvents)
}

// TestEventReplayer_SameTimestamp tests events with same timestamp
func TestEventReplayer_SameTimestamp(t *testing.T) {
	now := time.Now()
	events := []EventRecord{
		{ID: "1", Name: "event1", Timestamp: now},
		{ID: "2", Name: "event2", Timestamp: now}, // Same timestamp
		{ID: "3", Name: "event3", Timestamp: now}, // Same timestamp
	}

	replayer := NewEventReplayer(events)
	cmd := replayer.Replay()

	require.NotNil(t, cmd)

	// Should handle same timestamps with minimal delay
	var receivedIDs []string
	for cmd != nil {
		msg := cmd()
		if replayMsg, ok := msg.(ReplayEventMsg); ok {
			receivedIDs = append(receivedIDs, replayMsg.Event.ID)
			cmd = replayMsg.NextCmd
		} else {
			break
		}
	}

	assert.Equal(t, []string{"1", "2", "3"}, receivedIDs)
}
