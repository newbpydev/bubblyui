package testutil

import (
	"sync"
	"time"
)

// SimulatedTimer represents a timer in the simulated time system.
// It fires when the simulated time reaches or exceeds the target time.
type SimulatedTimer struct {
	targetTime time.Time
	ch         chan time.Time
	fired      bool
}

// TimeSimulator simulates the passage of time for testing time-dependent code
// without actual delays. This is essential for testing composables like useDebounce
// and useThrottle deterministically and quickly.
//
// Example:
//
//	ts := NewTimeSimulator()
//	timer := ts.After(100 * time.Millisecond)
//	ts.Advance(100 * time.Millisecond)
//	<-timer // Fires immediately
type TimeSimulator struct {
	mu          sync.Mutex
	currentTime time.Time
	timers      []*SimulatedTimer
}

// NewTimeSimulator creates a new TimeSimulator with the current time as the starting point.
// The simulated time will not advance automatically - you must call Advance() to move time forward.
//
// Example:
//
//	ts := NewTimeSimulator()
//	fmt.Println(ts.Now()) // Current time
func NewTimeSimulator() *TimeSimulator {
	return &TimeSimulator{
		currentTime: time.Now(),
		timers:      make([]*SimulatedTimer, 0),
	}
}

// Now returns the current simulated time.
// Unlike time.Now(), this value does not change unless Advance() is called.
//
// Example:
//
//	ts := NewTimeSimulator()
//	t1 := ts.Now()
//	time.Sleep(1 * time.Second)
//	t2 := ts.Now()
//	// t1 == t2 (simulated time doesn't advance automatically)
func (ts *TimeSimulator) Now() time.Time {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.currentTime
}

// Advance advances the simulated time by the given duration and fires any timers
// that have reached or exceeded their target time.
//
// Example:
//
//	ts := NewTimeSimulator()
//	timer := ts.After(100 * time.Millisecond)
//	ts.Advance(100 * time.Millisecond) // Timer fires
//	<-timer // Receives immediately
func (ts *TimeSimulator) Advance(d time.Duration) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Advance the current time
	ts.currentTime = ts.currentTime.Add(d)

	// Fire timers that have reached their target time
	for _, timer := range ts.timers {
		if !timer.fired && !ts.currentTime.Before(timer.targetTime) {
			timer.fired = true
			// Send current time to the channel (non-blocking)
			select {
			case timer.ch <- ts.currentTime:
			default:
				// Channel already has a value or is closed
			}
		}
	}
}

// After creates a simulated timer that will fire after the specified duration
// from the current simulated time. The timer fires when Advance() moves the
// simulated time to or past the target time.
//
// Example:
//
//	ts := NewTimeSimulator()
//	timer := ts.After(100 * time.Millisecond)
//	ts.Advance(50 * time.Millisecond) // Timer doesn't fire yet
//	ts.Advance(50 * time.Millisecond) // Timer fires now
//	<-timer // Receives the simulated time
func (ts *TimeSimulator) After(d time.Duration) <-chan time.Time {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	targetTime := ts.currentTime.Add(d)
	ch := make(chan time.Time, 1) // Buffered to prevent blocking

	timer := &SimulatedTimer{
		targetTime: targetTime,
		ch:         ch,
		fired:      false,
	}

	// If duration is 0 or negative, fire immediately
	if d <= 0 {
		timer.fired = true
		ch <- ts.currentTime
	} else {
		// Add to timers list for future firing
		ts.timers = append(ts.timers, timer)
	}

	return ch
}
