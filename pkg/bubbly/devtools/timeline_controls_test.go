package devtools

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTimelineControls(t *testing.T) {
	timeline := NewCommandTimeline(100)
	controls := NewTimelineControls(timeline)

	assert.NotNil(t, controls)
	assert.Equal(t, timeline, controls.timeline)
	assert.Equal(t, 0, controls.position)
	assert.Equal(t, 1.0, controls.speed)
	assert.False(t, controls.replaying)
	assert.False(t, controls.paused)
}

func TestTimelineControls_Scrub(t *testing.T) {
	timeline := NewCommandTimeline(10)
	baseTime := time.Now()

	// Add some commands
	for i := 0; i < 5; i++ {
		timeline.RecordCommand(CommandRecord{
			ID:        "cmd-" + string(rune('0'+i)),
			Type:      "TestCommand",
			Source:    "test",
			Generated: baseTime.Add(time.Duration(i) * time.Second),
			Executed:  baseTime.Add(time.Duration(i) * time.Second),
			Duration:  time.Millisecond,
		})
	}

	controls := NewTimelineControls(timeline)

	tests := []struct {
		name     string
		position int
		expected int
	}{
		{
			name:     "scrub to middle",
			position: 2,
			expected: 2,
		},
		{
			name:     "scrub to start",
			position: 0,
			expected: 0,
		},
		{
			name:     "scrub to end",
			position: 4,
			expected: 4,
		},
		{
			name:     "scrub beyond end (clamps)",
			position: 10,
			expected: 4,
		},
		{
			name:     "scrub negative (clamps to 0)",
			position: -1,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controls.Scrub(tt.position)
			assert.Equal(t, tt.expected, controls.GetPosition())
		})
	}
}

func TestTimelineControls_ScrubForward(t *testing.T) {
	timeline := NewCommandTimeline(10)
	baseTime := time.Now()

	// Add 3 commands
	for i := 0; i < 3; i++ {
		timeline.RecordCommand(CommandRecord{
			ID:        "cmd-" + string(rune('0'+i)),
			Type:      "TestCommand",
			Source:    "test",
			Generated: baseTime.Add(time.Duration(i) * time.Second),
			Executed:  baseTime.Add(time.Duration(i) * time.Second),
			Duration:  time.Millisecond,
		})
	}

	controls := NewTimelineControls(timeline)

	// Start at 0
	assert.Equal(t, 0, controls.GetPosition())

	// Forward to 1
	controls.ScrubForward()
	assert.Equal(t, 1, controls.GetPosition())

	// Forward to 2
	controls.ScrubForward()
	assert.Equal(t, 2, controls.GetPosition())

	// Forward at end (stays at 2)
	controls.ScrubForward()
	assert.Equal(t, 2, controls.GetPosition())
}

func TestTimelineControls_ScrubBackward(t *testing.T) {
	timeline := NewCommandTimeline(10)
	baseTime := time.Now()

	// Add 3 commands
	for i := 0; i < 3; i++ {
		timeline.RecordCommand(CommandRecord{
			ID:        "cmd-" + string(rune('0'+i)),
			Type:      "TestCommand",
			Source:    "test",
			Generated: baseTime.Add(time.Duration(i) * time.Second),
			Executed:  baseTime.Add(time.Duration(i) * time.Second),
			Duration:  time.Millisecond,
		})
	}

	controls := NewTimelineControls(timeline)

	// Start at end
	controls.Scrub(2)
	assert.Equal(t, 2, controls.GetPosition())

	// Backward to 1
	controls.ScrubBackward()
	assert.Equal(t, 1, controls.GetPosition())

	// Backward to 0
	controls.ScrubBackward()
	assert.Equal(t, 0, controls.GetPosition())

	// Backward at start (stays at 0)
	controls.ScrubBackward()
	assert.Equal(t, 0, controls.GetPosition())
}

func TestTimelineControls_SetSpeed(t *testing.T) {
	timeline := NewCommandTimeline(10)
	controls := NewTimelineControls(timeline)

	tests := []struct {
		name      string
		speed     float64
		expectErr bool
	}{
		{
			name:      "normal speed",
			speed:     1.0,
			expectErr: false,
		},
		{
			name:      "2x speed",
			speed:     2.0,
			expectErr: false,
		},
		{
			name:      "half speed",
			speed:     0.5,
			expectErr: false,
		},
		{
			name:      "minimum speed",
			speed:     0.1,
			expectErr: false,
		},
		{
			name:      "maximum speed",
			speed:     10.0,
			expectErr: false,
		},
		{
			name:      "zero speed (invalid)",
			speed:     0.0,
			expectErr: true,
		},
		{
			name:      "negative speed (invalid)",
			speed:     -1.0,
			expectErr: true,
		},
		{
			name:      "too fast (invalid)",
			speed:     11.0,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := controls.SetSpeed(tt.speed)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.speed, controls.GetSpeed())
			}
		})
	}
}

func TestTimelineControls_Replay_Empty(t *testing.T) {
	timeline := NewCommandTimeline(10)
	controls := NewTimelineControls(timeline)

	cmd := controls.Replay()
	assert.Nil(t, cmd, "replay on empty timeline should return nil")
	assert.False(t, controls.IsReplaying())
}

func TestTimelineControls_Replay_SingleCommand(t *testing.T) {
	timeline := NewCommandTimeline(10)
	baseTime := time.Now()

	timeline.RecordCommand(CommandRecord{
		ID:        "cmd-1",
		Type:      "TestCommand",
		Source:    "test",
		Generated: baseTime,
		Executed:  baseTime.Add(5 * time.Millisecond),
		Duration:  5 * time.Millisecond,
	})

	controls := NewTimelineControls(timeline)

	cmd := controls.Replay()
	assert.NotNil(t, cmd)
	assert.True(t, controls.IsReplaying())

	// Execute command
	msg := cmd()
	replayMsg, ok := msg.(ReplayCommandMsg)
	assert.True(t, ok)
	assert.Equal(t, "cmd-1", replayMsg.Command.ID)
	assert.Equal(t, 0, replayMsg.Index)
	assert.Equal(t, 1, replayMsg.Total)
	assert.NotNil(t, replayMsg.NextCmd)

	// Execute next command (should be completion)
	msg = replayMsg.NextCmd()
	completedMsg, ok := msg.(ReplayCompletedMsg)
	assert.True(t, ok)
	assert.Equal(t, 1, completedMsg.TotalEvents)
	assert.False(t, controls.IsReplaying())
}

func TestTimelineControls_Pause(t *testing.T) {
	timeline := NewCommandTimeline(10)
	controls := NewTimelineControls(timeline)

	// Initially not paused
	assert.False(t, controls.IsPaused())

	// Pause
	controls.Pause()
	assert.True(t, controls.IsPaused())

	// Pause again (idempotent)
	controls.Pause()
	assert.True(t, controls.IsPaused())
}

func TestTimelineControls_Resume(t *testing.T) {
	timeline := NewCommandTimeline(10)
	controls := NewTimelineControls(timeline)

	// Pause first
	controls.Pause()
	assert.True(t, controls.IsPaused())

	// Resume
	controls.Resume()
	assert.False(t, controls.IsPaused())

	// Resume again (idempotent)
	controls.Resume()
	assert.False(t, controls.IsPaused())
}

func TestTimelineControls_Render_Empty(t *testing.T) {
	timeline := NewCommandTimeline(10)
	controls := NewTimelineControls(timeline)

	output := controls.Render(80)

	assert.Contains(t, output, "Timeline Controls")
	assert.Contains(t, output, "No commands")
}

func TestTimelineControls_Render_WithCommands(t *testing.T) {
	timeline := NewCommandTimeline(10)
	baseTime := time.Now()

	// Add commands
	for i := 0; i < 3; i++ {
		timeline.RecordCommand(CommandRecord{
			ID:        "cmd-" + string(rune('0'+i)),
			Type:      "Command" + string(rune('0'+i)),
			Source:    "test",
			Generated: baseTime.Add(time.Duration(i*10) * time.Millisecond),
			Executed:  baseTime.Add(time.Duration(i*10+5) * time.Millisecond),
			Duration:  5 * time.Millisecond,
		})
	}

	controls := NewTimelineControls(timeline)
	controls.Scrub(1) // Position at command 1

	output := controls.Render(80)

	assert.Contains(t, output, "Timeline Controls")
	assert.Contains(t, output, "Position: 2/3") // 1-indexed for display
	assert.Contains(t, output, "Speed: 1.0x")
	assert.Contains(t, output, "►") // Position indicator
}

func TestTimelineControls_Render_Replaying(t *testing.T) {
	timeline := NewCommandTimeline(10)
	baseTime := time.Now()

	timeline.RecordCommand(CommandRecord{
		ID:        "cmd-1",
		Type:      "TestCommand",
		Source:    "test",
		Generated: baseTime,
		Executed:  baseTime.Add(5 * time.Millisecond),
		Duration:  5 * time.Millisecond,
	})

	controls := NewTimelineControls(timeline)
	controls.Replay()

	output := controls.Render(80)

	assert.Contains(t, output, "Timeline Controls")
	assert.Contains(t, output, "Replaying")
}

func TestTimelineControls_Concurrent(t *testing.T) {
	timeline := NewCommandTimeline(100)
	baseTime := time.Now()

	// Add commands
	for i := 0; i < 10; i++ {
		timeline.RecordCommand(CommandRecord{
			ID:        "cmd-" + string(rune('0'+i)),
			Type:      "TestCommand",
			Source:    "test",
			Generated: baseTime.Add(time.Duration(i) * time.Millisecond),
			Executed:  baseTime.Add(time.Duration(i+1) * time.Millisecond),
			Duration:  time.Millisecond,
		})
	}

	controls := NewTimelineControls(timeline)

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent scrubbing
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(pos int) {
			defer wg.Done()
			controls.Scrub(pos % 10)
		}(i)
	}

	// Concurrent reads
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = controls.GetPosition()
			_ = controls.GetSpeed()
			_ = controls.IsReplaying()
			_ = controls.IsPaused()
		}()
	}

	// Concurrent pause/resume
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				controls.Pause()
			} else {
				controls.Resume()
			}
		}(i)
	}

	// Concurrent speed changes
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer wg.Done()
			speed := 0.5 + float64(id%10)*0.5
			_ = controls.SetSpeed(speed)
		}(i)
	}

	wg.Wait()

	// Should still be in valid state
	pos := controls.GetPosition()
	assert.True(t, pos >= 0 && pos < 10)
}

func TestTimelineControls_ReplayWithPause(t *testing.T) {
	timeline := NewCommandTimeline(10)
	baseTime := time.Now()

	// Add 2 commands
	for i := 0; i < 2; i++ {
		timeline.RecordCommand(CommandRecord{
			ID:        "cmd-" + string(rune('0'+i)),
			Type:      "TestCommand",
			Source:    "test",
			Generated: baseTime.Add(time.Duration(i*100) * time.Millisecond),
			Executed:  baseTime.Add(time.Duration(i*100+10) * time.Millisecond),
			Duration:  10 * time.Millisecond,
		})
	}

	controls := NewTimelineControls(timeline)

	// Start replay
	cmd := controls.Replay()
	assert.NotNil(t, cmd)
	assert.True(t, controls.IsReplaying())

	// Execute first command
	msg := cmd()
	replayMsg, ok := msg.(ReplayCommandMsg)
	assert.True(t, ok)
	assert.Equal(t, 0, replayMsg.Index)
	assert.NotNil(t, replayMsg.NextCmd)

	// Pause before next command
	controls.Pause()
	assert.True(t, controls.IsPaused())
	assert.True(t, controls.IsReplaying()) // Still replaying, just paused

	// Resume
	controls.Resume()
	assert.False(t, controls.IsPaused())
	assert.True(t, controls.IsReplaying())
}

func TestTimelineControls_GetMethods(t *testing.T) {
	timeline := NewCommandTimeline(10)
	baseTime := time.Now()

	// Add commands so we can scrub to position 5
	for i := 0; i < 10; i++ {
		timeline.RecordCommand(CommandRecord{
			ID:        "cmd-" + string(rune('0'+i)),
			Type:      "TestCommand",
			Source:    "test",
			Generated: baseTime.Add(time.Duration(i) * time.Millisecond),
			Executed:  baseTime.Add(time.Duration(i+1) * time.Millisecond),
			Duration:  time.Millisecond,
		})
	}

	controls := NewTimelineControls(timeline)

	// Test initial values
	assert.Equal(t, 0, controls.GetPosition())
	assert.Equal(t, 1.0, controls.GetSpeed())
	assert.False(t, controls.IsReplaying())
	assert.False(t, controls.IsPaused())

	// Change values
	controls.Scrub(5)
	_ = controls.SetSpeed(2.0)
	controls.Pause()

	// Test updated values
	assert.Equal(t, 5, controls.GetPosition())
	assert.Equal(t, 2.0, controls.GetSpeed())
	assert.True(t, controls.IsPaused())
}

// Edge case: Scrub on empty timeline
func TestTimelineControls_Scrub_EmptyTimeline(t *testing.T) {
	timeline := NewCommandTimeline(10)
	controls := NewTimelineControls(timeline)

	// Scrub on empty timeline should set position to 0
	controls.Scrub(5)
	assert.Equal(t, 0, controls.GetPosition())

	controls.Scrub(-1)
	assert.Equal(t, 0, controls.GetPosition())

	controls.Scrub(100)
	assert.Equal(t, 0, controls.GetPosition())
}

// Edge case: ScrubForward on empty timeline
func TestTimelineControls_ScrubForward_EmptyTimeline(t *testing.T) {
	timeline := NewCommandTimeline(10)
	controls := NewTimelineControls(timeline)

	// ScrubForward on empty timeline should not panic
	controls.ScrubForward()
	assert.Equal(t, 0, controls.GetPosition())
}

// Edge case: ScrubBackward on empty timeline
func TestTimelineControls_ScrubBackward_EmptyTimeline(t *testing.T) {
	timeline := NewCommandTimeline(10)
	controls := NewTimelineControls(timeline)

	// ScrubBackward on empty timeline should not panic
	controls.ScrubBackward()
	assert.Equal(t, 0, controls.GetPosition())
}

// Edge case: Multiple commands with same timestamp
func TestTimelineControls_Replay_SameTimestamps(t *testing.T) {
	timeline := NewCommandTimeline(10)
	baseTime := time.Now()

	// Add commands with same timestamp
	for i := 0; i < 3; i++ {
		timeline.RecordCommand(CommandRecord{
			ID:        "cmd-" + string(rune('0'+i)),
			Type:      "TestCommand",
			Source:    "test",
			Generated: baseTime, // Same timestamp
			Executed:  baseTime.Add(time.Millisecond),
			Duration:  time.Millisecond,
		})
	}

	controls := NewTimelineControls(timeline)
	cmd := controls.Replay()
	assert.NotNil(t, cmd)

	// Execute first command
	msg := cmd()
	replayMsg, ok := msg.(ReplayCommandMsg)
	assert.True(t, ok)
	assert.Equal(t, 0, replayMsg.Index)
	assert.NotNil(t, replayMsg.NextCmd)
}

// Edge case: Replay already replaying
func TestTimelineControls_Replay_AlreadyReplaying(t *testing.T) {
	timeline := NewCommandTimeline(10)
	baseTime := time.Now()

	timeline.RecordCommand(CommandRecord{
		ID:        "cmd-1",
		Type:      "TestCommand",
		Source:    "test",
		Generated: baseTime,
		Executed:  baseTime.Add(time.Millisecond),
		Duration:  time.Millisecond,
	})

	controls := NewTimelineControls(timeline)

	// Start first replay
	cmd1 := controls.Replay()
	assert.NotNil(t, cmd1)
	assert.True(t, controls.IsReplaying())

	// Try to start second replay (should return nil)
	cmd2 := controls.Replay()
	assert.Nil(t, cmd2)
}

// Edge case: Speed boundary values
func TestTimelineControls_SetSpeed_BoundaryValues(t *testing.T) {
	timeline := NewCommandTimeline(10)
	controls := NewTimelineControls(timeline)

	tests := []struct {
		name      string
		speed     float64
		expectErr bool
	}{
		{"exactly 0.1", 0.1, false},
		{"exactly 10.0", 10.0, false},
		{"just below 0.1", 0.09, true},
		{"just above 10.0", 10.01, true},
		{"very small positive", 0.001, true},
		{"very large", 100.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := controls.SetSpeed(tt.speed)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Edge case: Render with different widths
func TestTimelineControls_Render_VariousWidths(t *testing.T) {
	timeline := NewCommandTimeline(10)
	baseTime := time.Now()

	for i := 0; i < 3; i++ {
		timeline.RecordCommand(CommandRecord{
			ID:        "cmd-" + string(rune('0'+i)),
			Type:      "TestCommand",
			Source:    "test",
			Generated: baseTime.Add(time.Duration(i*10) * time.Millisecond),
			Executed:  baseTime.Add(time.Duration(i*10+5) * time.Millisecond),
			Duration:  5 * time.Millisecond,
		})
	}

	controls := NewTimelineControls(timeline)

	tests := []struct {
		name  string
		width int
	}{
		{"very narrow", 10},
		{"narrow", 40},
		{"normal", 80},
		{"wide", 120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := controls.Render(tt.width)
			assert.Contains(t, output, "Timeline Controls")
			assert.Contains(t, output, "Position:")
		})
	}
}

// Edge case: Render with paused state
func TestTimelineControls_Render_PausedState(t *testing.T) {
	timeline := NewCommandTimeline(10)
	baseTime := time.Now()

	timeline.RecordCommand(CommandRecord{
		ID:        "cmd-1",
		Type:      "TestCommand",
		Source:    "test",
		Generated: baseTime,
		Executed:  baseTime.Add(time.Millisecond),
		Duration:  time.Millisecond,
	})

	controls := NewTimelineControls(timeline)
	controls.Replay()
	controls.Pause()

	output := controls.Render(80)
	assert.Contains(t, output, "Timeline Controls")
	assert.Contains(t, output, "Paused")
}

// Edge case: Render stopped state
func TestTimelineControls_Render_StoppedState(t *testing.T) {
	timeline := NewCommandTimeline(10)
	baseTime := time.Now()

	timeline.RecordCommand(CommandRecord{
		ID:        "cmd-1",
		Type:      "TestCommand",
		Source:    "test",
		Generated: baseTime,
		Executed:  baseTime.Add(time.Millisecond),
		Duration:  time.Millisecond,
	})

	controls := NewTimelineControls(timeline)

	output := controls.Render(80)
	assert.Contains(t, output, "Timeline Controls")
	assert.Contains(t, output, "Stopped")
}

// Edge case: Render with very long command type names
func TestTimelineControls_Render_LongCommandNames(t *testing.T) {
	timeline := NewCommandTimeline(10)
	baseTime := time.Now()

	timeline.RecordCommand(CommandRecord{
		ID:        "cmd-1",
		Type:      "VeryLongCommandTypeNameThatExceedsMaxLength",
		Source:    "test",
		Generated: baseTime,
		Executed:  baseTime.Add(time.Millisecond),
		Duration:  time.Millisecond,
	})

	controls := NewTimelineControls(timeline)

	output := controls.Render(80)
	assert.Contains(t, output, "Timeline Controls")
	// Should be truncated
	assert.NotContains(t, output, "VeryLongCommandTypeNameThatExceedsMaxLength")
}

// Edge case: Position at various locations
func TestTimelineControls_Render_PositionIndicator(t *testing.T) {
	timeline := NewCommandTimeline(10)
	baseTime := time.Now()

	for i := 0; i < 5; i++ {
		timeline.RecordCommand(CommandRecord{
			ID:        "cmd-" + string(rune('0'+i)),
			Type:      "TestCommand",
			Source:    "test",
			Generated: baseTime.Add(time.Duration(i*10) * time.Millisecond),
			Executed:  baseTime.Add(time.Duration(i*10+5) * time.Millisecond),
			Duration:  5 * time.Millisecond,
		})
	}

	controls := NewTimelineControls(timeline)

	tests := []struct {
		name     string
		position int
	}{
		{"at start", 0},
		{"in middle", 2},
		{"at end", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controls.Scrub(tt.position)
			output := controls.Render(80)
			assert.Contains(t, output, "►") // Position indicator
		})
	}
}

// Edge case: Multiple speed changes
func TestTimelineControls_MultipleSpeedChanges(t *testing.T) {
	timeline := NewCommandTimeline(10)
	controls := NewTimelineControls(timeline)

	speeds := []float64{0.5, 1.0, 2.0, 5.0, 10.0, 0.1}
	for _, speed := range speeds {
		err := controls.SetSpeed(speed)
		assert.NoError(t, err)
		assert.Equal(t, speed, controls.GetSpeed())
	}
}

// Edge case: Pause/Resume multiple times
func TestTimelineControls_MultiplePauseResume(t *testing.T) {
	timeline := NewCommandTimeline(10)
	controls := NewTimelineControls(timeline)

	// Multiple pause calls
	controls.Pause()
	assert.True(t, controls.IsPaused())
	controls.Pause()
	assert.True(t, controls.IsPaused())

	// Multiple resume calls
	controls.Resume()
	assert.False(t, controls.IsPaused())
	controls.Resume()
	assert.False(t, controls.IsPaused())
}

// Edge case: Scrub while replaying
func TestTimelineControls_ScrubWhileReplaying(t *testing.T) {
	timeline := NewCommandTimeline(10)
	baseTime := time.Now()

	for i := 0; i < 5; i++ {
		timeline.RecordCommand(CommandRecord{
			ID:        "cmd-" + string(rune('0'+i)),
			Type:      "TestCommand",
			Source:    "test",
			Generated: baseTime.Add(time.Duration(i*10) * time.Millisecond),
			Executed:  baseTime.Add(time.Duration(i*10+5) * time.Millisecond),
			Duration:  5 * time.Millisecond,
		})
	}

	controls := NewTimelineControls(timeline)

	// Start replay
	cmd := controls.Replay()
	assert.NotNil(t, cmd)
	assert.True(t, controls.IsReplaying())

	// Scrub while replaying (should work)
	controls.Scrub(3)
	assert.Equal(t, 3, controls.GetPosition())
	assert.True(t, controls.IsReplaying())
}
