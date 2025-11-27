package devtools

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCommandTimeline(t *testing.T) {
	timeline := NewCommandTimeline(100)

	assert.NotNil(t, timeline)
	assert.Equal(t, 100, timeline.maxSize)
	assert.False(t, timeline.paused)
	assert.Empty(t, timeline.commands)
}

func TestCommandTimeline_RecordCommand(t *testing.T) {
	tests := []struct {
		name          string
		maxSize       int
		recordCount   int
		expectedCount int
	}{
		{
			name:          "single command",
			maxSize:       10,
			recordCount:   1,
			expectedCount: 1,
		},
		{
			name:          "multiple commands",
			maxSize:       10,
			recordCount:   5,
			expectedCount: 5,
		},
		{
			name:          "circular buffer overflow",
			maxSize:       10,
			recordCount:   15,
			expectedCount: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeline := NewCommandTimeline(tt.maxSize)

			for i := 0; i < tt.recordCount; i++ {
				record := CommandRecord{
					ID:        "cmd-" + string(rune('0'+i)),
					Type:      "TestCommand",
					Source:    "test",
					Generated: time.Now(),
					Executed:  time.Now(),
					Duration:  time.Millisecond,
				}
				timeline.RecordCommand(record)
			}

			assert.Equal(t, tt.expectedCount, timeline.GetCommandCount())
		})
	}
}

func TestCommandTimeline_Pause(t *testing.T) {
	timeline := NewCommandTimeline(10)

	// Initially not paused
	assert.False(t, timeline.IsPaused())

	// Record a command
	timeline.RecordCommand(CommandRecord{
		ID:        "cmd-1",
		Type:      "BeforePause",
		Source:    "test",
		Generated: time.Now(),
		Executed:  time.Now(),
		Duration:  time.Millisecond,
	})
	assert.Equal(t, 1, timeline.GetCommandCount())

	// Pause
	timeline.Pause()
	assert.True(t, timeline.IsPaused())

	// Try to record while paused (should be ignored)
	timeline.RecordCommand(CommandRecord{
		ID:        "cmd-2",
		Type:      "WhilePaused",
		Source:    "test",
		Generated: time.Now(),
		Executed:  time.Now(),
		Duration:  time.Millisecond,
	})
	assert.Equal(t, 1, timeline.GetCommandCount(), "command should not be recorded while paused")
}

func TestCommandTimeline_Resume(t *testing.T) {
	timeline := NewCommandTimeline(10)

	// Pause first
	timeline.Pause()
	assert.True(t, timeline.IsPaused())

	// Resume
	timeline.Resume()
	assert.False(t, timeline.IsPaused())

	// Record after resume (should work)
	timeline.RecordCommand(CommandRecord{
		ID:        "cmd-1",
		Type:      "AfterResume",
		Source:    "test",
		Generated: time.Now(),
		Executed:  time.Now(),
		Duration:  time.Millisecond,
	})
	assert.Equal(t, 1, timeline.GetCommandCount())
}

func TestCommandTimeline_IsPaused(t *testing.T) {
	timeline := NewCommandTimeline(10)

	// Initially not paused
	assert.False(t, timeline.IsPaused())

	// After pause
	timeline.Pause()
	assert.True(t, timeline.IsPaused())

	// After resume
	timeline.Resume()
	assert.False(t, timeline.IsPaused())
}

func TestCommandTimeline_Render_Empty(t *testing.T) {
	timeline := NewCommandTimeline(10)

	output := timeline.Render(80)

	assert.Contains(t, output, "Command Timeline")
	assert.Contains(t, output, "No commands recorded")
}

func TestCommandTimeline_Render_SingleCommand(t *testing.T) {
	timeline := NewCommandTimeline(10)

	now := time.Now()
	timeline.RecordCommand(CommandRecord{
		ID:        "cmd-1",
		Type:      "TestCommand",
		Source:    "test",
		Generated: now,
		Executed:  now.Add(5 * time.Millisecond),
		Duration:  5 * time.Millisecond,
	})

	output := timeline.Render(80)

	assert.Contains(t, output, "Command Timeline")
	assert.Contains(t, output, "TestCommand")
	assert.Contains(t, output, "Time span:")
}

func TestCommandTimeline_Render_MultipleCommands(t *testing.T) {
	timeline := NewCommandTimeline(10)

	baseTime := time.Now()
	commands := []CommandRecord{
		{
			ID:        "cmd-1",
			Type:      "Command1",
			Source:    "source1",
			Generated: baseTime,
			Executed:  baseTime.Add(10 * time.Millisecond),
			Duration:  10 * time.Millisecond,
		},
		{
			ID:        "cmd-2",
			Type:      "Command2",
			Source:    "source2",
			Generated: baseTime.Add(50 * time.Millisecond),
			Executed:  baseTime.Add(60 * time.Millisecond),
			Duration:  10 * time.Millisecond,
		},
		{
			ID:        "cmd-3",
			Type:      "Command3",
			Source:    "source3",
			Generated: baseTime.Add(100 * time.Millisecond),
			Executed:  baseTime.Add(105 * time.Millisecond),
			Duration:  5 * time.Millisecond,
		},
	}

	for _, cmd := range commands {
		timeline.RecordCommand(cmd)
	}

	output := timeline.Render(80)

	assert.Contains(t, output, "Command Timeline")
	assert.Contains(t, output, "Command1")
	assert.Contains(t, output, "Command2")
	assert.Contains(t, output, "Command3")
	assert.Contains(t, output, "Time span:")
}

func TestCommandTimeline_Render_TimelineVisualization(t *testing.T) {
	timeline := NewCommandTimeline(10)

	baseTime := time.Now()
	timeline.RecordCommand(CommandRecord{
		ID:        "cmd-1",
		Type:      "FastCommand",
		Source:    "test",
		Generated: baseTime,
		Executed:  baseTime.Add(1 * time.Millisecond),
		Duration:  1 * time.Millisecond,
	})

	output := timeline.Render(80)

	// Should contain timeline bar character
	assert.True(t, strings.Contains(output, "▬") || strings.Contains(output, "█"),
		"output should contain timeline bar character")
}

func TestCommandTimeline_GetCommands(t *testing.T) {
	timeline := NewCommandTimeline(10)

	baseTime := time.Now()
	expected := []CommandRecord{
		{
			ID:        "cmd-1",
			Type:      "Command1",
			Source:    "test",
			Generated: baseTime,
			Executed:  baseTime.Add(5 * time.Millisecond),
			Duration:  5 * time.Millisecond,
		},
		{
			ID:        "cmd-2",
			Type:      "Command2",
			Source:    "test",
			Generated: baseTime.Add(10 * time.Millisecond),
			Executed:  baseTime.Add(15 * time.Millisecond),
			Duration:  5 * time.Millisecond,
		},
	}

	for _, cmd := range expected {
		timeline.RecordCommand(cmd)
	}

	commands := timeline.GetCommands()

	assert.Equal(t, len(expected), len(commands))
	for i, cmd := range commands {
		assert.Equal(t, expected[i].ID, cmd.ID)
		assert.Equal(t, expected[i].Type, cmd.Type)
	}
}

func TestCommandTimeline_Clear(t *testing.T) {
	timeline := NewCommandTimeline(10)

	// Add some commands
	for i := 0; i < 5; i++ {
		timeline.RecordCommand(CommandRecord{
			ID:        "cmd-" + string(rune('0'+i)),
			Type:      "TestCommand",
			Source:    "test",
			Generated: time.Now(),
			Executed:  time.Now(),
			Duration:  time.Millisecond,
		})
	}
	assert.Equal(t, 5, timeline.GetCommandCount())

	// Clear
	timeline.Clear()

	assert.Equal(t, 0, timeline.GetCommandCount())
	assert.Empty(t, timeline.GetCommands())
}

func TestCommandTimeline_Concurrent(t *testing.T) {
	timeline := NewCommandTimeline(1000)

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent writes
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			timeline.RecordCommand(CommandRecord{
				ID:        "cmd-" + string(rune('0'+id)),
				Type:      "ConcurrentCommand",
				Source:    "test",
				Generated: time.Now(),
				Executed:  time.Now(),
				Duration:  time.Millisecond,
			})
		}(i)
	}

	// Concurrent reads
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = timeline.GetCommandCount()
			_ = timeline.GetCommands()
			_ = timeline.IsPaused()
		}()
	}

	// Concurrent pause/resume
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				timeline.Pause()
			} else {
				timeline.Resume()
			}
		}(i)
	}

	wg.Wait()

	// Should have some commands recorded (not all due to pausing)
	count := timeline.GetCommandCount()
	assert.True(t, count > 0 && count <= numGoroutines,
		"should have recorded some commands")
}

func TestCommandTimeline_Render_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*CommandTimeline)
		width    int
		contains []string
	}{
		{
			name: "very small width",
			setup: func(ct *CommandTimeline) {
				ct.RecordCommand(CommandRecord{
					ID:        "cmd-1",
					Type:      "Test",
					Source:    "test",
					Generated: time.Now(),
					Executed:  time.Now().Add(5 * time.Millisecond),
					Duration:  5 * time.Millisecond,
				})
			},
			width:    20,
			contains: []string{"Command Timeline", "Test"},
		},
		{
			name: "zero duration command",
			setup: func(ct *CommandTimeline) {
				now := time.Now()
				ct.RecordCommand(CommandRecord{
					ID:        "cmd-1",
					Type:      "InstantCommand",
					Source:    "test",
					Generated: now,
					Executed:  now,
					Duration:  0,
				})
			},
			width:    80,
			contains: []string{"Command Timeline", "InstantCommand"},
		},
		{
			name: "offset at boundary",
			setup: func(ct *CommandTimeline) {
				baseTime := time.Now()
				ct.RecordCommand(CommandRecord{
					ID:        "cmd-1",
					Type:      "First",
					Source:    "test",
					Generated: baseTime,
					Executed:  baseTime.Add(1 * time.Millisecond),
					Duration:  1 * time.Millisecond,
				})
				ct.RecordCommand(CommandRecord{
					ID:        "cmd-2",
					Type:      "Last",
					Source:    "test",
					Generated: baseTime.Add(100 * time.Millisecond),
					Executed:  baseTime.Add(101 * time.Millisecond),
					Duration:  1 * time.Millisecond,
				})
			},
			width:    80,
			contains: []string{"First", "Last"},
		},
		{
			name: "very long command type",
			setup: func(ct *CommandTimeline) {
				ct.RecordCommand(CommandRecord{
					ID:        "cmd-1",
					Type:      "ThisIsAVeryLongCommandTypeThatShouldBeTruncated",
					Source:    "test",
					Generated: time.Now(),
					Executed:  time.Now().Add(5 * time.Millisecond),
					Duration:  5 * time.Millisecond,
				})
			},
			width:    80,
			contains: []string{"ThisIsAVeryLongCo..."},
		},
		{
			name: "negative offset edge case",
			setup: func(ct *CommandTimeline) {
				baseTime := time.Now()
				// Commands with same Generated time but different Executed
				ct.RecordCommand(CommandRecord{
					ID:        "cmd-1",
					Type:      "Cmd1",
					Source:    "test",
					Generated: baseTime,
					Executed:  baseTime.Add(10 * time.Millisecond),
					Duration:  10 * time.Millisecond,
				})
				ct.RecordCommand(CommandRecord{
					ID:        "cmd-2",
					Type:      "Cmd2",
					Source:    "test",
					Generated: baseTime,
					Executed:  baseTime.Add(5 * time.Millisecond),
					Duration:  5 * time.Millisecond,
				})
			},
			width:    80,
			contains: []string{"Cmd1", "Cmd2"},
		},
		{
			name: "duration exceeds remaining width",
			setup: func(ct *CommandTimeline) {
				baseTime := time.Now()
				ct.RecordCommand(CommandRecord{
					ID:        "cmd-1",
					Type:      "LongDuration",
					Source:    "test",
					Generated: baseTime.Add(90 * time.Millisecond),
					Executed:  baseTime.Add(200 * time.Millisecond),
					Duration:  110 * time.Millisecond,
				})
			},
			width:    80,
			contains: []string{"LongDuration"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeline := NewCommandTimeline(10)
			tt.setup(timeline)

			output := timeline.Render(tt.width)

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestFormatTimelineDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "nanoseconds",
			duration: 500 * time.Nanosecond,
			expected: "500ns",
		},
		{
			name:     "microseconds",
			duration: 250 * time.Microsecond,
			expected: "250.0µs",
		},
		{
			name:     "milliseconds",
			duration: 15 * time.Millisecond,
			expected: "15.0ms",
		},
		{
			name:     "seconds",
			duration: 2500 * time.Millisecond,
			expected: "2.50s",
		},
		{
			name:     "exactly 1 microsecond",
			duration: 1 * time.Microsecond,
			expected: "1.0µs",
		},
		{
			name:     "exactly 1 millisecond",
			duration: 1 * time.Millisecond,
			expected: "1.0ms",
		},
		{
			name:     "exactly 1 second",
			duration: 1 * time.Second,
			expected: "1.00s",
		},
		{
			name:     "fractional microseconds",
			duration: 1500 * time.Nanosecond,
			expected: "1.5µs",
		},
		{
			name:     "fractional milliseconds",
			duration: 1500 * time.Microsecond,
			expected: "1.5ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTimelineDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommandTimeline_Render_ConcurrentAccess(t *testing.T) {
	timeline := NewCommandTimeline(100)

	// Add some initial commands
	baseTime := time.Now()
	for i := 0; i < 10; i++ {
		timeline.RecordCommand(CommandRecord{
			ID:        fmt.Sprintf("cmd-%d", i),
			Type:      fmt.Sprintf("Command%d", i),
			Source:    "test",
			Generated: baseTime.Add(time.Duration(i*10) * time.Millisecond),
			Executed:  baseTime.Add(time.Duration(i*10+5) * time.Millisecond),
			Duration:  5 * time.Millisecond,
		})
	}

	var wg sync.WaitGroup

	// Concurrent renders
	wg.Add(50)
	for i := 0; i < 50; i++ {
		go func() {
			defer wg.Done()
			output := timeline.Render(80)
			assert.Contains(t, output, "Command Timeline")
		}()
	}

	// Concurrent modifications
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer wg.Done()
			timeline.RecordCommand(CommandRecord{
				ID:        fmt.Sprintf("concurrent-%d", id),
				Type:      "Concurrent",
				Source:    "test",
				Generated: time.Now(),
				Executed:  time.Now().Add(5 * time.Millisecond),
				Duration:  5 * time.Millisecond,
			})
		}(i)
	}

	wg.Wait()
}

// TestCommandTimeline_Append tests the Append method which is an alias for RecordCommand
func TestCommandTimeline_Append(t *testing.T) {
	tests := []struct {
		name          string
		maxSize       int
		appendCount   int
		expectedCount int
	}{
		{
			name:          "single append",
			maxSize:       10,
			appendCount:   1,
			expectedCount: 1,
		},
		{
			name:          "multiple appends",
			maxSize:       10,
			appendCount:   5,
			expectedCount: 5,
		},
		{
			name:          "append respects max size",
			maxSize:       5,
			appendCount:   10,
			expectedCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeline := NewCommandTimeline(tt.maxSize)

			for i := 0; i < tt.appendCount; i++ {
				record := CommandRecord{
					ID:        fmt.Sprintf("append-%d", i),
					Type:      "TestCommand",
					Source:    "test",
					Generated: time.Now(),
					Executed:  time.Now(),
					Duration:  time.Millisecond,
				}
				timeline.Append(record)
			}

			assert.Equal(t, tt.expectedCount, timeline.GetCommandCount())
		})
	}
}

// TestCommandTimeline_Append_IsAliasForRecordCommand tests that Append behaves identically to RecordCommand
func TestCommandTimeline_Append_IsAliasForRecordCommand(t *testing.T) {
	timeline1 := NewCommandTimeline(10)
	timeline2 := NewCommandTimeline(10)

	record := CommandRecord{
		ID:        "test-cmd",
		Type:      "TestCommand",
		Source:    "test",
		Generated: time.Now(),
		Executed:  time.Now(),
		Duration:  time.Millisecond,
	}

	// Use RecordCommand on timeline1
	timeline1.RecordCommand(record)

	// Use Append on timeline2
	timeline2.Append(record)

	// Both should have same count
	assert.Equal(t, timeline1.GetCommandCount(), timeline2.GetCommandCount())

	// Both should have the same command
	cmds1 := timeline1.GetCommands()
	cmds2 := timeline2.GetCommands()

	assert.Equal(t, len(cmds1), len(cmds2))
	assert.Equal(t, cmds1[0].ID, cmds2[0].ID)
	assert.Equal(t, cmds1[0].Type, cmds2[0].Type)
}
