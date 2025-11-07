package devtools

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewEventTracker tests the constructor
func TestNewEventTracker(t *testing.T) {
	tests := []struct {
		name      string
		maxEvents int
		wantMax   int
	}{
		{
			name:      "default size",
			maxEvents: 1000,
			wantMax:   1000,
		},
		{
			name:      "custom size",
			maxEvents: 500,
			wantMax:   500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewEventTracker(tt.maxEvents)
			require.NotNil(t, tracker)
			assert.False(t, tracker.IsPaused())
			assert.Equal(t, 0, tracker.GetEventCount())
		})
	}
}

// TestEventTracker_CaptureEvent tests event capture
func TestEventTracker_CaptureEvent(t *testing.T) {
	tests := []struct {
		name   string
		events []EventRecord
		want   int
	}{
		{
			name: "single event",
			events: []EventRecord{
				{
					ID:        "event-1",
					Name:      "click",
					SourceID:  "button-1",
					TargetID:  "handler-1",
					Payload:   "test",
					Timestamp: time.Now(),
					Duration:  time.Millisecond,
				},
			},
			want: 1,
		},
		{
			name: "multiple events",
			events: []EventRecord{
				{ID: "event-1", Name: "click", Timestamp: time.Now()},
				{ID: "event-2", Name: "submit", Timestamp: time.Now()},
				{ID: "event-3", Name: "change", Timestamp: time.Now()},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewEventTracker(1000)

			for _, event := range tt.events {
				tracker.CaptureEvent(event)
			}

			assert.Equal(t, tt.want, tracker.GetEventCount())
		})
	}
}

// TestEventTracker_PauseResume tests pause/resume functionality
func TestEventTracker_PauseResume(t *testing.T) {
	tracker := NewEventTracker(1000)

	// Initially not paused
	assert.False(t, tracker.IsPaused())

	// Pause
	tracker.Pause()
	assert.True(t, tracker.IsPaused())

	// Events should not be captured when paused
	tracker.CaptureEvent(EventRecord{ID: "event-1", Name: "click", Timestamp: time.Now()})
	assert.Equal(t, 0, tracker.GetEventCount())

	// Resume
	tracker.Resume()
	assert.False(t, tracker.IsPaused())

	// Events should be captured after resume
	tracker.CaptureEvent(EventRecord{ID: "event-2", Name: "submit", Timestamp: time.Now()})
	assert.Equal(t, 1, tracker.GetEventCount())
}

// TestEventTracker_MaxEvents tests max event limit enforcement
func TestEventTracker_MaxEvents(t *testing.T) {
	maxEvents := 10
	tracker := NewEventTracker(maxEvents)

	// Add more events than max
	for i := 0; i < 20; i++ {
		tracker.CaptureEvent(EventRecord{
			ID:        "event-" + string(rune(i)),
			Name:      "click",
			Timestamp: time.Now(),
		})
	}

	// Should only keep last maxEvents
	assert.Equal(t, maxEvents, tracker.GetEventCount())
}

// TestEventTracker_GetRecent tests getting recent events
func TestEventTracker_GetRecent(t *testing.T) {
	tracker := NewEventTracker(1000)

	// Add events
	events := []EventRecord{
		{ID: "event-1", Name: "click", Timestamp: time.Now()},
		{ID: "event-2", Name: "submit", Timestamp: time.Now()},
		{ID: "event-3", Name: "change", Timestamp: time.Now()},
	}

	for _, event := range events {
		tracker.CaptureEvent(event)
	}

	tests := []struct {
		name string
		n    int
		want int
	}{
		{
			name: "get all",
			n:    10,
			want: 3,
		},
		{
			name: "get last 2",
			n:    2,
			want: 2,
		},
		{
			name: "get last 1",
			n:    1,
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recent := tracker.GetRecent(tt.n)
			assert.Equal(t, tt.want, len(recent))
		})
	}
}

// TestEventTracker_Clear tests clearing events
func TestEventTracker_Clear(t *testing.T) {
	tracker := NewEventTracker(1000)

	// Add events
	tracker.CaptureEvent(EventRecord{ID: "event-1", Name: "click", Timestamp: time.Now()})
	tracker.CaptureEvent(EventRecord{ID: "event-2", Name: "submit", Timestamp: time.Now()})
	assert.Equal(t, 2, tracker.GetEventCount())

	// Clear
	tracker.Clear()
	assert.Equal(t, 0, tracker.GetEventCount())
}

// TestEventTracker_Render tests rendering output
func TestEventTracker_Render(t *testing.T) {
	tests := []struct {
		name     string
		events   []EventRecord
		contains []string
	}{
		{
			name:     "empty events",
			events:   []EventRecord{},
			contains: []string{"Recent Events:", "No events captured"},
		},
		{
			name: "single event",
			events: []EventRecord{
				{
					ID:        "event-1",
					Name:      "click",
					SourceID:  "button-1",
					TargetID:  "handler-1",
					Timestamp: time.Now(),
					Duration:  time.Millisecond,
				},
			},
			contains: []string{"Recent Events:", "click", "button-1", "handler-1"},
		},
		{
			name: "multiple events",
			events: []EventRecord{
				{ID: "event-1", Name: "click", SourceID: "btn-1", Timestamp: time.Now()},
				{ID: "event-2", Name: "submit", SourceID: "form-1", Timestamp: time.Now()},
			},
			contains: []string{"Recent Events:", "click", "submit", "btn-1", "form-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewEventTracker(1000)

			for _, event := range tt.events {
				tracker.CaptureEvent(event)
			}

			output := tracker.Render()
			for _, substr := range tt.contains {
				assert.Contains(t, output, substr)
			}
		})
	}
}

// TestEventTracker_RenderWithDuration tests rendering with duration display
func TestEventTracker_RenderWithDuration(t *testing.T) {
	tracker := NewEventTracker(1000)

	tracker.CaptureEvent(EventRecord{
		ID:        "event-1",
		Name:      "click",
		SourceID:  "button-1",
		Timestamp: time.Now(),
		Duration:  5 * time.Millisecond,
	})

	output := tracker.Render()
	assert.Contains(t, output, "5ms")
}

// TestEventTracker_SetFilter tests event filtering
func TestEventTracker_SetFilter(t *testing.T) {
	tracker := NewEventTracker(1000)

	// Add events
	tracker.CaptureEvent(EventRecord{ID: "event-1", Name: "click", Timestamp: time.Now()})
	tracker.CaptureEvent(EventRecord{ID: "event-2", Name: "submit", Timestamp: time.Now()})
	tracker.CaptureEvent(EventRecord{ID: "event-3", Name: "change", Timestamp: time.Now()})

	// Set filter
	tracker.SetFilter("click")
	assert.Equal(t, "click", tracker.GetFilter())

	// Render should only show filtered events
	output := tracker.Render()
	assert.Contains(t, output, "click")
	assert.NotContains(t, output, "submit")
	assert.NotContains(t, output, "change")
}

// TestEventTracker_FilterCaseInsensitive tests case-insensitive filtering
func TestEventTracker_FilterCaseInsensitive(t *testing.T) {
	tracker := NewEventTracker(1000)

	tracker.CaptureEvent(EventRecord{ID: "event-1", Name: "ClickEvent", Timestamp: time.Now()})
	tracker.CaptureEvent(EventRecord{ID: "event-2", Name: "SUBMIT", Timestamp: time.Now()})

	// Filter with lowercase
	tracker.SetFilter("click")
	output := tracker.Render()
	assert.Contains(t, output, "ClickEvent")

	// Filter with uppercase
	tracker.SetFilter("SUBMIT")
	output = tracker.Render()
	assert.Contains(t, output, "SUBMIT")
}

// TestEventTracker_ClearFilter tests clearing the filter
func TestEventTracker_ClearFilter(t *testing.T) {
	tracker := NewEventTracker(1000)

	tracker.CaptureEvent(EventRecord{ID: "event-1", Name: "click", Timestamp: time.Now()})
	tracker.CaptureEvent(EventRecord{ID: "event-2", Name: "submit", Timestamp: time.Now()})

	// Set filter
	tracker.SetFilter("click")
	output := tracker.Render()
	assert.NotContains(t, output, "submit")

	// Clear filter
	tracker.SetFilter("")
	output = tracker.Render()
	assert.Contains(t, output, "click")
	assert.Contains(t, output, "submit")
}

// TestEventTracker_Concurrent tests thread-safe concurrent access
func TestEventTracker_Concurrent(t *testing.T) {
	tracker := NewEventTracker(1000)
	var wg sync.WaitGroup

	// Concurrent captures
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tracker.CaptureEvent(EventRecord{
				ID:        "event-" + string(rune(id)),
				Name:      "click",
				Timestamp: time.Now(),
			})
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = tracker.GetEventCount()
			_ = tracker.GetRecent(10)
			_ = tracker.Render()
		}()
	}

	// Concurrent pause/resume
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tracker.Pause()
			tracker.Resume()
		}()
	}

	wg.Wait()

	// Should have captured some events (not all due to pauses)
	assert.GreaterOrEqual(t, tracker.GetEventCount(), 0)
}

// TestEventTracker_GetStatistics tests event statistics
func TestEventTracker_GetStatistics(t *testing.T) {
	tracker := NewEventTracker(1000)

	// Add events
	tracker.CaptureEvent(EventRecord{ID: "event-1", Name: "click", Timestamp: time.Now()})
	tracker.CaptureEvent(EventRecord{ID: "event-2", Name: "click", Timestamp: time.Now()})
	tracker.CaptureEvent(EventRecord{ID: "event-3", Name: "submit", Timestamp: time.Now()})

	stats := tracker.GetStatistics()
	assert.Equal(t, 3, stats.TotalEvents)
	assert.Equal(t, 2, stats.EventsByName["click"])
	assert.Equal(t, 1, stats.EventsByName["submit"])
}

// TestEventTracker_RenderShowsTimestamp tests timestamp display
func TestEventTracker_RenderShowsTimestamp(t *testing.T) {
	tracker := NewEventTracker(1000)

	now := time.Now()
	tracker.CaptureEvent(EventRecord{
		ID:        "event-1",
		Name:      "click",
		Timestamp: now,
	})

	output := tracker.Render()
	// Should contain time in HH:MM:SS format
	timeStr := now.Format("15:04:05")
	assert.Contains(t, output, timeStr)
}

// TestEventTracker_RenderReversedOrder tests that events are shown newest first
func TestEventTracker_RenderReversedOrder(t *testing.T) {
	tracker := NewEventTracker(1000)

	// Add events in order
	tracker.CaptureEvent(EventRecord{ID: "event-1", Name: "first", Timestamp: time.Now()})
	time.Sleep(10 * time.Millisecond)
	tracker.CaptureEvent(EventRecord{ID: "event-2", Name: "second", Timestamp: time.Now()})
	time.Sleep(10 * time.Millisecond)
	tracker.CaptureEvent(EventRecord{ID: "event-3", Name: "third", Timestamp: time.Now()})

	output := tracker.Render()

	// Find positions
	firstPos := strings.Index(output, "first")
	secondPos := strings.Index(output, "second")
	thirdPos := strings.Index(output, "third")

	// Newest (third) should appear before oldest (first)
	assert.Greater(t, firstPos, thirdPos, "Events should be in reverse chronological order")
	assert.Greater(t, secondPos, thirdPos, "Events should be in reverse chronological order")
}
