package devtools

import (
	"errors"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// EventReplayer replays captured events with speed control and pause/resume functionality.
//
// It provides event replay capabilities for debugging and testing, allowing developers
// to replay captured events at different speeds, pause/resume playback, and track progress.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	events := []EventRecord{...}
//	replayer := NewEventReplayer(events)
//	replayer.SetSpeed(2.0) // 2x speed
//	cmd := replayer.Replay()
//	// Use cmd in Bubbletea Update() method
type EventReplayer struct {
	// events is the list of events to replay
	events []EventRecord

	// speed is the playback speed multiplier (1.0 = normal, 2.0 = 2x, 0.5 = half)
	speed float64

	// paused indicates if replay is paused
	paused bool

	// replaying indicates if replay is currently active
	replaying bool

	// currentIndex is the current position in the event list
	currentIndex int

	// mu protects concurrent access to all fields
	mu sync.RWMutex
}

// ReplayEventMsg is the message sent when an event is replayed.
//
// This message is sent to the Bubbletea Update() method for each replayed event.
// Applications can distinguish replayed events from real events by checking the message type.
type ReplayEventMsg struct {
	// Event is the replayed event
	Event EventRecord

	// Index is the current index in the replay sequence
	Index int

	// Total is the total number of events in the replay
	Total int

	// NextCmd is the command to execute for the next event (or nil if done)
	NextCmd tea.Cmd
}

// ReplayPausedMsg is sent when replay is paused.
type ReplayPausedMsg struct {
	// Index is the current index when paused
	Index int

	// Total is the total number of events
	Total int
}

// ReplayCompletedMsg is sent when replay completes.
type ReplayCompletedMsg struct {
	// TotalEvents is the number of events replayed
	TotalEvents int
}

// NewEventReplayer creates a new event replayer with the given events.
//
// The replayer starts with speed=1.0, not paused, and at index 0.
// Events are copied to prevent external modification.
//
// Example:
//
//	events := []EventRecord{...}
//	replayer := NewEventReplayer(events)
//
// Parameters:
//   - events: The events to replay
//
// Returns:
//   - *EventReplayer: A new event replayer instance
func NewEventReplayer(events []EventRecord) *EventReplayer {
	// Copy events to prevent external modification
	eventsCopy := make([]EventRecord, len(events))
	copy(eventsCopy, events)

	return &EventReplayer{
		events:       eventsCopy,
		speed:        1.0,
		paused:       false,
		replaying:    false,
		currentIndex: 0,
	}
}

// Replay starts replaying events and returns a Bubbletea command.
//
// The command emits ReplayEventMsg for each event with appropriate timing
// based on the original event timestamps and the current speed setting.
//
// If the event list is empty, returns nil immediately.
// If already replaying, returns nil.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	cmd := replayer.Replay()
//	if cmd != nil {
//	    return m, cmd
//	}
//
// Returns:
//   - tea.Cmd: The command to start replay, or nil if no events
func (er *EventReplayer) Replay() tea.Cmd {
	er.mu.Lock()
	defer er.mu.Unlock()

	// Don't start if empty
	if len(er.events) == 0 {
		return nil
	}

	// Don't start if already replaying
	if er.replaying {
		return nil
	}

	// Reset to beginning
	er.currentIndex = 0
	er.replaying = true
	er.paused = false

	// Start replay from first event
	return er.replayNextEvent()
}

// replayNextEvent creates a command to replay the next event.
// Must be called with lock held.
func (er *EventReplayer) replayNextEvent() tea.Cmd {
	// Check if paused
	if er.paused {
		index := er.currentIndex
		total := len(er.events)
		return func() tea.Msg {
			return ReplayPausedMsg{
				Index: index,
				Total: total,
			}
		}
	}

	// Check if done
	if er.currentIndex >= len(er.events) {
		er.replaying = false
		totalEvents := len(er.events)
		return func() tea.Msg {
			return ReplayCompletedMsg{
				TotalEvents: totalEvents,
			}
		}
	}

	// Get current event
	event := er.events[er.currentIndex]
	index := er.currentIndex
	total := len(er.events)
	speed := er.speed

	// Calculate delay to next event
	var nextCmd tea.Cmd
	if er.currentIndex+1 < len(er.events) {
		nextEvent := er.events[er.currentIndex+1]
		delay := nextEvent.Timestamp.Sub(event.Timestamp)

		// Apply speed multiplier
		adjustedDelay := time.Duration(float64(delay) / speed)

		// Ensure minimum delay of 1ms for same timestamps
		if adjustedDelay < time.Millisecond {
			adjustedDelay = time.Millisecond
		}

		// Increment index for next event
		er.currentIndex++

		// Create command for next event with delay
		nextCmd = tea.Tick(adjustedDelay, func(t time.Time) tea.Msg {
			er.mu.Lock()
			defer er.mu.Unlock()
			cmd := er.replayNextEvent()
			if cmd != nil {
				return cmd()
			}
			return nil
		})
	} else {
		// Last event - increment index but keep replaying flag
		// Will be set to false when completion message is sent
		er.currentIndex++
		nextCmd = func() tea.Msg {
			// Mark as done when completion message is sent
			er.mu.Lock()
			er.replaying = false
			er.mu.Unlock()
			return ReplayCompletedMsg{
				TotalEvents: total,
			}
		}
	}

	// Return command that emits the current event
	return func() tea.Msg {
		return ReplayEventMsg{
			Event:   event,
			Index:   index,
			Total:   total,
			NextCmd: nextCmd,
		}
	}
}

// SetSpeed sets the playback speed multiplier.
//
// Speed must be greater than 0. Common values:
//   - 1.0: Normal speed (real-time)
//   - 2.0: Double speed (2x faster)
//   - 0.5: Half speed (slower)
//   - 10.0: Very fast
//   - 0.1: Very slow
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	err := replayer.SetSpeed(2.0) // 2x speed
//	if err != nil {
//	    log.Printf("Invalid speed: %v", err)
//	}
//
// Parameters:
//   - speed: The speed multiplier (must be > 0)
//
// Returns:
//   - error: Error if speed is invalid
func (er *EventReplayer) SetSpeed(speed float64) error {
	if speed <= 0 {
		return errors.New("speed must be greater than 0")
	}

	er.mu.Lock()
	defer er.mu.Unlock()
	er.speed = speed
	return nil
}

// GetSpeed returns the current playback speed.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - float64: The current speed multiplier
func (er *EventReplayer) GetSpeed() float64 {
	er.mu.RLock()
	defer er.mu.RUnlock()
	return er.speed
}

// Pause pauses the replay.
//
// While paused, no new events will be emitted. The current position
// is preserved and can be resumed later.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (er *EventReplayer) Pause() {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.paused = true
}

// Resume resumes the replay from where it was paused.
//
// If not currently paused, this is a no-op.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - tea.Cmd: Command to continue replay, or nil if not replaying
func (er *EventReplayer) Resume() tea.Cmd {
	er.mu.Lock()
	defer er.mu.Unlock()

	if !er.paused {
		return nil
	}

	er.paused = false

	// Continue from current position
	if er.replaying && er.currentIndex < len(er.events) {
		return er.replayNextEvent()
	}

	return nil
}

// IsPaused returns whether the replay is currently paused.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - bool: True if paused, false otherwise
func (er *EventReplayer) IsPaused() bool {
	er.mu.RLock()
	defer er.mu.RUnlock()
	return er.paused
}

// IsReplaying returns whether replay is currently active.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - bool: True if replaying, false otherwise
func (er *EventReplayer) IsReplaying() bool {
	er.mu.RLock()
	defer er.mu.RUnlock()
	return er.replaying
}

// GetProgress returns the current replay progress.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - current: Current event index
//   - total: Total number of events
func (er *EventReplayer) GetProgress() (current int, total int) {
	er.mu.RLock()
	defer er.mu.RUnlock()
	return er.currentIndex, len(er.events)
}

// Reset resets the replayer to the beginning.
//
// This stops any active replay and resets the position to index 0.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (er *EventReplayer) Reset() {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.currentIndex = 0
	er.replaying = false
	er.paused = false
}
