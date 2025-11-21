package testutil

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestTimeSimulator_Creation tests creating a new TimeSimulator
func TestTimeSimulator_Creation(t *testing.T) {
	ts := NewTimeSimulator()

	assert.NotNil(t, ts, "TimeSimulator should not be nil")
	assert.NotZero(t, ts.Now(), "TimeSimulator should have a current time")
}

// TestTimeSimulator_Now tests getting the current simulated time
func TestTimeSimulator_Now(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"returns current time"},
		{"time does not advance automatically"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTimeSimulator()

			now1 := ts.Now()
			time.Sleep(10 * time.Millisecond)
			now2 := ts.Now()

			// Simulated time should not advance automatically
			assert.Equal(t, now1, now2, "simulated time should not advance without Advance()")
		})
	}
}

// TestTimeSimulator_Advance tests advancing simulated time
func TestTimeSimulator_Advance(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
	}{
		{"advance 1 second", 1 * time.Second},
		{"advance 100 milliseconds", 100 * time.Millisecond},
		{"advance 1 hour", 1 * time.Hour},
		{"advance 0 duration", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTimeSimulator()

			before := ts.Now()
			ts.Advance(tt.duration)
			after := ts.Now()

			expected := before.Add(tt.duration)
			assert.Equal(t, expected, after, "time should advance by duration")
		})
	}
}

// TestTimeSimulator_After tests creating simulated timers
func TestTimeSimulator_After(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
	}{
		{"100ms timer", 100 * time.Millisecond},
		{"1s timer", 1 * time.Second},
		{"500ms timer", 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTimeSimulator()

			ch := ts.After(tt.duration)
			assert.NotNil(t, ch, "After should return a channel")

			// Timer should not fire immediately
			select {
			case <-ch:
				t.Error("timer fired immediately, should wait for Advance()")
			default:
				// Expected - timer hasn't fired yet
			}
		})
	}
}

// TestTimeSimulator_TimerFiring tests that timers fire when time advances
func TestTimeSimulator_TimerFiring(t *testing.T) {
	tests := []struct {
		name       string
		timerDelay time.Duration
		advanceBy  time.Duration
		shouldFire bool
	}{
		{"timer fires when time reached", 100 * time.Millisecond, 100 * time.Millisecond, true},
		{"timer fires when time exceeded", 100 * time.Millisecond, 200 * time.Millisecond, true},
		{"timer does not fire when time not reached", 100 * time.Millisecond, 50 * time.Millisecond, false},
		{"zero duration timer fires immediately", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTimeSimulator()

			ch := ts.After(tt.timerDelay)
			ts.Advance(tt.advanceBy)

			if tt.shouldFire {
				select {
				case receivedTime := <-ch:
					expectedTime := ts.Now()
					assert.Equal(t, expectedTime, receivedTime, "timer should send current time")
				case <-time.After(100 * time.Millisecond):
					t.Error("timer should have fired but didn't")
				}
			} else {
				select {
				case <-ch:
					t.Error("timer fired when it shouldn't have")
				default:
					// Expected - timer hasn't fired
				}
			}
		})
	}
}

// TestTimeSimulator_MultipleTimers tests multiple timers firing at different times
func TestTimeSimulator_MultipleTimers(t *testing.T) {
	ts := NewTimeSimulator()

	timer1 := ts.After(100 * time.Millisecond)
	timer2 := ts.After(200 * time.Millisecond)
	timer3 := ts.After(300 * time.Millisecond)

	// Advance to 150ms - only timer1 should fire
	ts.Advance(150 * time.Millisecond)

	select {
	case <-timer1:
		// Expected
	default:
		t.Error("timer1 should have fired at 150ms")
	}

	select {
	case <-timer2:
		t.Error("timer2 should not have fired at 150ms")
	default:
		// Expected
	}

	select {
	case <-timer3:
		t.Error("timer3 should not have fired at 150ms")
	default:
		// Expected
	}

	// Advance to 250ms total - timer2 should now fire
	ts.Advance(100 * time.Millisecond)

	select {
	case <-timer2:
		// Expected
	default:
		t.Error("timer2 should have fired at 250ms")
	}

	select {
	case <-timer3:
		t.Error("timer3 should not have fired at 250ms")
	default:
		// Expected
	}

	// Advance to 350ms total - timer3 should now fire
	ts.Advance(100 * time.Millisecond)

	select {
	case <-timer3:
		// Expected
	default:
		t.Error("timer3 should have fired at 350ms")
	}
}

// TestTimeSimulator_ThreadSafe tests concurrent access to TimeSimulator
func TestTimeSimulator_ThreadSafe(t *testing.T) {
	ts := NewTimeSimulator()

	var wg sync.WaitGroup
	iterations := 100

	// Multiple goroutines calling Now()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = ts.Now()
			}
		}()
	}

	// Multiple goroutines calling Advance()
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				ts.Advance(1 * time.Millisecond)
			}
		}()
	}

	// Multiple goroutines creating timers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = ts.After(10 * time.Millisecond)
			}
		}()
	}

	wg.Wait()

	// If we get here without race detector errors, we're good
	assert.NotNil(t, ts)
}

// TestTimeSimulator_FastForward tests fast-forwarding time
func TestTimeSimulator_FastForward(t *testing.T) {
	ts := NewTimeSimulator()

	timer1 := ts.After(1 * time.Second)
	timer2 := ts.After(2 * time.Second)
	timer3 := ts.After(3 * time.Second)

	// Fast-forward past all timers
	ts.Advance(5 * time.Second)

	// All timers should have fired
	select {
	case <-timer1:
		// Expected
	default:
		t.Error("timer1 should have fired")
	}

	select {
	case <-timer2:
		// Expected
	default:
		t.Error("timer2 should have fired")
	}

	select {
	case <-timer3:
		// Expected
	default:
		t.Error("timer3 should have fired")
	}
}

// TestTimeSimulator_TimerOrder tests that timers fire when expected
func TestTimeSimulator_TimerOrder(t *testing.T) {
	ts := NewTimeSimulator()

	var wg sync.WaitGroup
	var fired []int
	var mu sync.Mutex

	// Create timers BEFORE starting goroutines
	timer1 := ts.After(100 * time.Millisecond)
	timer2 := ts.After(200 * time.Millisecond)
	timer3 := ts.After(300 * time.Millisecond)

	// Start goroutines to wait on timers
	wg.Add(3)

	go func() {
		defer wg.Done()
		<-timer1
		mu.Lock()
		fired = append(fired, 1)
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		<-timer2
		mu.Lock()
		fired = append(fired, 2)
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		<-timer3
		mu.Lock()
		fired = append(fired, 3)
		mu.Unlock()
	}()

	// Advance time to fire all timers
	ts.Advance(400 * time.Millisecond)

	// Wait for all goroutines to complete
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	assert.Equal(t, 3, len(fired), "all 3 timers should have fired")
	// Note: We can't guarantee order due to goroutine scheduling,
	// but all should have fired
}
