package devtools

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// CommandTimeline tracks Bubbletea command execution over time.
//
// It maintains a circular buffer of command records with configurable maximum size
// and provides timeline visualization capabilities. Commands can be paused/resumed
// for analysis without losing historical data.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	timeline := NewCommandTimeline(1000)
//	timeline.RecordCommand(CommandRecord{
//	    ID:        "cmd-1",
//	    Type:      "fetchData",
//	    Source:    "DataLoader",
//	    Generated: time.Now(),
//	    Executed:  time.Now().Add(5 * time.Millisecond),
//	    Duration:  5 * time.Millisecond,
//	})
//	output := timeline.Render(80)
type CommandTimeline struct {
	// commands is the circular buffer of command records
	commands []CommandRecord

	// paused indicates whether recording is paused
	paused bool

	// maxSize is the maximum number of commands to keep
	maxSize int

	// mu protects concurrent access to all fields
	mu sync.RWMutex
}

// CommandRecord represents a single command execution.
type CommandRecord struct {
	// ID is the unique identifier for this command
	ID string

	// Type is the command type or name
	Type string

	// Source identifies what generated the command (component ID, function name, etc.)
	Source string

	// Generated is when the command was created
	Generated time.Time

	// Executed is when the command was executed
	Executed time.Time

	// Duration is how long the command took to execute
	Duration time.Duration
}

// NewCommandTimeline creates a new command timeline with the specified maximum size.
//
// The maxSize parameter determines how many command records to keep in memory.
// When the limit is reached, the oldest records are discarded.
//
// Example:
//
//	timeline := NewCommandTimeline(1000) // Keep last 1000 commands
//
// Parameters:
//   - maxSize: Maximum number of command records to keep
//
// Returns:
//   - *CommandTimeline: A new command timeline instance
func NewCommandTimeline(maxSize int) *CommandTimeline {
	return &CommandTimeline{
		commands: make([]CommandRecord, 0, maxSize),
		paused:   false,
		maxSize:  maxSize,
	}
}

// RecordCommand adds a command record to the timeline.
//
// If the timeline is paused, the command is not recorded.
// If the timeline is at maximum capacity, the oldest record is removed
// to make room for the new one.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	timeline.RecordCommand(CommandRecord{
//	    ID:        "cmd-1",
//	    Type:      "fetchData",
//	    Source:    "DataLoader",
//	    Generated: time.Now(),
//	    Executed:  time.Now().Add(5 * time.Millisecond),
//	    Duration:  5 * time.Millisecond,
//	})
//
// Parameters:
//   - record: The command record to add
func (ct *CommandTimeline) RecordCommand(record CommandRecord) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	// Don't record if paused
	if ct.paused {
		return
	}

	// Append the record
	ct.commands = append(ct.commands, record)

	// Enforce max size (circular buffer)
	if len(ct.commands) > ct.maxSize {
		ct.commands = ct.commands[len(ct.commands)-ct.maxSize:]
	}
}

// Pause stops recording new commands.
//
// Commands recorded while paused are ignored. Existing commands remain in the timeline.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	timeline.Pause()
//	// Commands recorded here will be ignored
//	timeline.Resume()
func (ct *CommandTimeline) Pause() {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.paused = true
}

// Resume resumes recording new commands.
//
// After calling Resume, new commands will be recorded normally.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	timeline.Resume()
//	// Commands recorded here will be captured
func (ct *CommandTimeline) Resume() {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.paused = false
}

// IsPaused returns whether the timeline is currently paused.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - bool: True if paused, false otherwise
func (ct *CommandTimeline) IsPaused() bool {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.paused
}

// GetCommandCount returns the number of commands in the timeline.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - int: The number of commands currently stored
func (ct *CommandTimeline) GetCommandCount() int {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return len(ct.commands)
}

// GetCommands returns a copy of all commands in the timeline.
//
// The returned slice is a copy and can be safely modified without affecting
// the internal state.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - []CommandRecord: A copy of all command records
func (ct *CommandTimeline) GetCommands() []CommandRecord {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	// Return a copy to prevent external modification
	commands := make([]CommandRecord, len(ct.commands))
	copy(commands, ct.commands)
	return commands
}

// Clear removes all commands from the timeline.
//
// The paused state is not affected.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	timeline.Clear()
func (ct *CommandTimeline) Clear() {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.commands = make([]CommandRecord, 0, ct.maxSize)
}

// Render generates a visual timeline representation of command execution.
//
// The timeline shows commands as horizontal bars with their duration and timing.
// The width parameter controls the maximum width of the visualization.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	output := timeline.Render(80) // 80 character width
//	fmt.Println(output)
//
// Parameters:
//   - width: Maximum width for the timeline visualization
//
// Returns:
//   - string: The rendered timeline with Lipgloss styling
func (ct *CommandTimeline) Render(width int) string {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	// Header style
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		MarginBottom(1)

	// Empty state
	if len(ct.commands) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)

		return headerStyle.Render("Command Timeline") + "\n\n" +
			emptyStyle.Render("No commands recorded")
	}

	// Calculate time span
	startTime := ct.commands[0].Generated
	endTime := ct.commands[len(ct.commands)-1].Executed
	totalDuration := endTime.Sub(startTime)

	// Handle single command or zero duration
	if totalDuration == 0 {
		totalDuration = 1 * time.Millisecond
	}

	// Build output
	var lines []string
	lines = append(lines, headerStyle.Render("Command Timeline"))
	lines = append(lines, fmt.Sprintf("Time span: %s", formatTimelineDuration(totalDuration)))
	lines = append(lines, "")

	// Render timeline bars
	barWidth := width - 25 // Reserve space for label
	if barWidth < 10 {
		barWidth = 10
	}

	for _, cmd := range ct.commands {
		// Calculate offset from start
		offset := cmd.Generated.Sub(startTime)
		offsetChars := int(float64(barWidth) * offset.Seconds() / totalDuration.Seconds())
		if offsetChars < 0 {
			offsetChars = 0
		}
		if offsetChars >= barWidth {
			offsetChars = barWidth - 1
		}

		// Calculate duration bar width
		durationChars := int(float64(barWidth) * cmd.Duration.Seconds() / totalDuration.Seconds())
		if durationChars < 1 {
			durationChars = 1
		}
		if offsetChars+durationChars > barWidth {
			durationChars = barWidth - offsetChars
		}

		// Build bar
		bar := strings.Repeat(" ", offsetChars) + strings.Repeat("▬", durationChars)

		// Truncate label if needed
		label := cmd.Type
		maxLabelLen := 20
		if len(label) > maxLabelLen {
			label = label[:maxLabelLen-3] + "..."
		}

		// Style the line
		lineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("35"))
		line := bar + " " + lineStyle.Render(label)

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// formatTimelineDuration formats a duration for timeline display.
func formatTimelineDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.1fµs", float64(d.Nanoseconds())/1000.0)
	} else if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000.0)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
