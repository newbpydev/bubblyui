package devtools

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TimelineControls provides scrubbing and replay functionality for command timelines.
//
// It allows developers to navigate through command history, replay commands at different
// speeds, and pause/resume playback. This is useful for debugging asynchronous behavior
// and understanding command execution flow.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	timeline := NewCommandTimeline(1000)
//	controls := NewTimelineControls(timeline)
//	controls.SetSpeed(2.0) // 2x speed
//	cmd := controls.Replay()
//	// Use cmd in Bubbletea Update() method
type TimelineControls struct {
	// timeline is the command timeline to control
	timeline *CommandTimeline

	// position is the current position in the timeline (0-indexed)
	position int

	// speed is the playback speed multiplier (1.0 = normal, 2.0 = 2x, 0.5 = half)
	speed float64

	// replaying indicates if replay is currently active
	replaying bool

	// paused indicates if replay is paused
	paused bool

	// mu protects concurrent access to all fields
	mu sync.RWMutex
}

// ReplayCommandMsg is the message sent when a command is replayed.
//
// This message is sent to the Bubbletea Update() method for each replayed command.
// Applications can distinguish replayed commands from real commands by checking the message type.
type ReplayCommandMsg struct {
	// Command is the replayed command
	Command CommandRecord

	// Index is the current index in the replay sequence
	Index int

	// Total is the total number of commands in the replay
	Total int

	// NextCmd is the command to execute for the next command (or nil if done)
	NextCmd tea.Cmd
}

// Note: ReplayPausedMsg and ReplayCompletedMsg are defined in event_replay.go
// and are reused here for consistency across replay functionality.

// NewTimelineControls creates a new timeline controls instance.
//
// The controls start at position 0, speed 1.0, not replaying, and not paused.
//
// Example:
//
//	timeline := NewCommandTimeline(1000)
//	controls := NewTimelineControls(timeline)
//
// Parameters:
//   - timeline: The command timeline to control
//
// Returns:
//   - *TimelineControls: A new timeline controls instance
func NewTimelineControls(timeline *CommandTimeline) *TimelineControls {
	return &TimelineControls{
		timeline:  timeline,
		position:  0,
		speed:     1.0,
		replaying: false,
		paused:    false,
	}
}

// Scrub sets the timeline position to the specified index.
//
// The position is clamped to valid range [0, commandCount-1].
// If the timeline is empty, position is set to 0.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	controls.Scrub(5) // Jump to command 5
//
// Parameters:
//   - position: The target position (0-indexed)
func (tc *TimelineControls) Scrub(position int) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	commandCount := tc.timeline.GetCommandCount()
	if commandCount == 0 {
		tc.position = 0
		return
	}

	// Clamp to valid range
	if position < 0 {
		tc.position = 0
	} else if position >= commandCount {
		tc.position = commandCount - 1
	} else {
		tc.position = position
	}
}

// ScrubForward moves the position forward by one command.
//
// If already at the end, position remains unchanged.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	controls.ScrubForward() // Move to next command
func (tc *TimelineControls) ScrubForward() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	commandCount := tc.timeline.GetCommandCount()
	if commandCount == 0 {
		return
	}

	if tc.position < commandCount-1 {
		tc.position++
	}
}

// ScrubBackward moves the position backward by one command.
//
// If already at the start, position remains unchanged.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	controls.ScrubBackward() // Move to previous command
func (tc *TimelineControls) ScrubBackward() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.position > 0 {
		tc.position--
	}
}

// GetPosition returns the current timeline position.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - int: The current position (0-indexed)
func (tc *TimelineControls) GetPosition() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.position
}

// SetSpeed sets the replay speed multiplier.
//
// Valid range is 0.1 to 10.0. Values outside this range return an error.
// Speed of 1.0 is normal speed, 2.0 is double speed, 0.5 is half speed.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	err := controls.SetSpeed(2.0) // 2x speed
//	if err != nil {
//	    // Handle error
//	}
//
// Parameters:
//   - speed: The speed multiplier (0.1 to 10.0)
//
// Returns:
//   - error: Error if speed is out of valid range
func (tc *TimelineControls) SetSpeed(speed float64) error {
	if speed <= 0 || speed < 0.1 || speed > 10.0 {
		return errors.New("speed must be between 0.1 and 10.0")
	}

	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.speed = speed
	return nil
}

// GetSpeed returns the current replay speed multiplier.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - float64: The current speed multiplier
func (tc *TimelineControls) GetSpeed() float64 {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.speed
}

// Replay starts replaying commands and returns a Bubbletea command.
//
// The command emits ReplayCommandMsg for each command with appropriate timing
// based on the original command timestamps and the current speed setting.
//
// If the timeline is empty, returns nil immediately.
// If already replaying, returns nil.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	cmd := controls.Replay()
//	if cmd != nil {
//	    return m, cmd
//	}
//
// Returns:
//   - tea.Cmd: The command to start replay, or nil if no commands
func (tc *TimelineControls) Replay() tea.Cmd {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Don't start if empty
	commands := tc.timeline.GetCommands()
	if len(commands) == 0 {
		return nil
	}

	// Don't start if already replaying
	if tc.replaying {
		return nil
	}

	// Reset to beginning
	tc.position = 0
	tc.replaying = true
	tc.paused = false

	// Start replay from first command
	return tc.replayNextCommand(commands)
}

// replayNextCommand creates a command to replay the next command.
// Must be called with lock held.
func (tc *TimelineControls) replayNextCommand(commands []CommandRecord) tea.Cmd {
	// Check if paused
	if tc.paused {
		index := tc.position
		total := len(commands)
		return func() tea.Msg {
			return ReplayPausedMsg{
				Index: index,
				Total: total,
			}
		}
	}

	// Check if done
	if tc.position >= len(commands) {
		tc.replaying = false
		totalCommands := len(commands)
		return func() tea.Msg {
			return ReplayCompletedMsg{
				TotalEvents: totalCommands,
			}
		}
	}

	// Get current command
	command := commands[tc.position]
	index := tc.position
	total := len(commands)
	speed := tc.speed

	// Calculate delay to next command
	var nextCmd tea.Cmd
	if tc.position+1 < len(commands) {
		nextCommand := commands[tc.position+1]
		delay := nextCommand.Generated.Sub(command.Generated)

		// Apply speed multiplier
		adjustedDelay := time.Duration(float64(delay) / speed)

		// Ensure minimum delay of 1ms for same timestamps
		if adjustedDelay < time.Millisecond {
			adjustedDelay = time.Millisecond
		}

		// Increment position for next command
		tc.position++

		// Create next command with delay
		nextCmd = tea.Tick(adjustedDelay, func(t time.Time) tea.Msg {
			tc.mu.Lock()
			defer tc.mu.Unlock()

			// Get fresh commands list
			freshCommands := tc.timeline.GetCommands()
			return tc.replayNextCommand(freshCommands)
		})
	} else {
		// Last command - increment to trigger completion
		tc.position++

		// Create completion command
		nextCmd = func() tea.Msg {
			tc.mu.Lock()
			defer tc.mu.Unlock()

			tc.replaying = false
			return ReplayCompletedMsg{
				TotalEvents: total,
			}
		}
	}

	// Return message with current command
	return func() tea.Msg {
		return ReplayCommandMsg{
			Command: command,
			Index:   index,
			Total:   total,
			NextCmd: nextCmd,
		}
	}
}

// Pause pauses the replay.
//
// Commands already in flight will complete, but no new commands will be emitted
// until Resume is called.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	controls.Pause()
func (tc *TimelineControls) Pause() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.paused = true
}

// Resume resumes the replay.
//
// After calling Resume, replay will continue from the current position.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	controls.Resume()
func (tc *TimelineControls) Resume() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.paused = false
}

// IsPaused returns whether the replay is currently paused.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - bool: True if paused, false otherwise
func (tc *TimelineControls) IsPaused() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.paused
}

// IsReplaying returns whether replay is currently active.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - bool: True if replaying, false otherwise
func (tc *TimelineControls) IsReplaying() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.replaying
}

// Render generates a visual representation of the timeline with position indicator.
//
// The visualization shows the current position, replay status, speed, and a timeline
// with a marker at the current position.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	output := controls.Render(80) // 80 character width
//	fmt.Println(output)
//
// Parameters:
//   - width: Maximum width for the visualization
//
// Returns:
//   - string: The rendered timeline with Lipgloss styling
// renderStatusInfo returns status, position, and speed lines.
func (tc *TimelineControls) renderStatusInfo(numCommands int) []string {
	status := "Stopped"
	if tc.replaying {
		if tc.paused {
			status = "Paused"
		} else {
			status = "Replaying"
		}
	}

	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("35"))
	if tc.paused {
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229"))
	}

	positionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	speedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("35"))

	return []string{
		fmt.Sprintf("Status: %s", statusStyle.Render(status)),
		fmt.Sprintf("Position: %s", positionStyle.Render(fmt.Sprintf("%d/%d", tc.position+1, numCommands))),
		fmt.Sprintf("Speed: %s", speedStyle.Render(fmt.Sprintf("%.1fx", tc.speed))),
	}
}

// renderTimelineBar renders a single timeline bar for a command.
func (tc *TimelineControls) renderTimelineBar(cmd CommandRecord, index int, barWidth int, startTime time.Time, totalDuration time.Duration) string {
	offset := cmd.Generated.Sub(startTime)
	offsetChars := clampInt(int(float64(barWidth)*offset.Seconds()/totalDuration.Seconds()), 0, barWidth-1)

	durationChars := int(float64(barWidth) * cmd.Duration.Seconds() / totalDuration.Seconds())
	if durationChars < 1 {
		durationChars = 1
	}
	if offsetChars+durationChars > barWidth {
		durationChars = barWidth - offsetChars
	}

	indicator := " "
	if index == tc.position {
		indicator = selectionIndicator
	}

	bar := strings.Repeat(" ", offsetChars) + strings.Repeat("▬", durationChars)

	label := cmd.Type
	if len(label) > 12 {
		label = label[:9] + "..."
	}

	lineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("35"))
	if index == tc.position {
		lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true)
	}

	return indicator + " " + bar + " " + lineStyle.Render(label)
}

// clampInt clamps an integer value to the range [min, max].
func clampInt(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func (tc *TimelineControls) Render(width int) string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")).MarginBottom(1)
	commands := tc.timeline.GetCommands()

	if len(commands) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
		return headerStyle.Render("Timeline Controls") + "\n\n" + emptyStyle.Render("No commands in timeline")
	}

	lines := []string{headerStyle.Render("Timeline Controls")}
	lines = append(lines, tc.renderStatusInfo(len(commands))...)
	lines = append(lines, "")

	startTime := commands[0].Generated
	endTime := commands[len(commands)-1].Executed
	totalDuration := endTime.Sub(startTime)
	if totalDuration == 0 {
		totalDuration = 1 * time.Millisecond
	}

	barWidth := width - 30
	if barWidth < 10 {
		barWidth = 10
	}

	for i, cmd := range commands {
		lines = append(lines, tc.renderTimelineBar(cmd, i, barWidth, startTime, totalDuration))
	}

	return strings.Join(lines, "\n")
}
