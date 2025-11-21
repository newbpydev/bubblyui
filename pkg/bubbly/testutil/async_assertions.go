package testutil

import (
	"reflect"
	"time"
)

// WaitOptions configures the behavior of WaitFor.
type WaitOptions struct {
	// Timeout is the maximum duration to wait for the condition.
	// Default: 5 seconds
	Timeout time.Duration

	// Interval is the polling interval between condition checks.
	// Default: 10 milliseconds
	Interval time.Duration

	// Message is a custom error message to display on timeout.
	// If empty, a default message is used.
	Message string
}

// WaitFor polls a condition function until it returns true or the timeout is reached.
// It uses the provided WaitOptions to configure timeout, polling interval, and error message.
//
// Example:
//
//	WaitFor(t, func() bool {
//	    return counter.Get() == 5
//	}, WaitOptions{
//	    Timeout:  2 * time.Second,
//	    Interval: 50 * time.Millisecond,
//	    Message:  "counter never reached 5",
//	})
func WaitFor(t testingT, condition func() bool, opts WaitOptions) {
	t.Helper()

	// Set defaults
	if opts.Timeout == 0 {
		opts.Timeout = 5 * time.Second
	}
	if opts.Interval == 0 {
		opts.Interval = 10 * time.Millisecond
	}

	deadline := time.Now().Add(opts.Timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(opts.Interval)
	}

	// Timeout reached
	msg := "timeout waiting for condition"
	if opts.Message != "" {
		msg = opts.Message
	}
	t.Errorf(msg)
}

// WaitForRef waits for a ref value to match the expected value.
// It polls the ref until the value matches or the timeout is reached.
//
// Example:
//
//	ct.WaitForRef("loading", false, 2*time.Second)
func (ct *ComponentTest) WaitForRef(name string, expected interface{}, timeout time.Duration) {
	ct.harness.t.Helper()

	// Create a testingT wrapper that calls t.Fatal
	deadline := time.Now().Add(timeout)
	interval := 10 * time.Millisecond

	for time.Now().Before(deadline) {
		actual := ct.state.GetRefValue(name)
		if reflect.DeepEqual(actual, expected) {
			return
		}
		time.Sleep(interval)
	}

	// Timeout reached
	ct.harness.t.Errorf("timeout waiting for ref %q to equal %v", name, expected)
}

// WaitForEvent waits for an event to be fired.
// It polls the event tracker until the event is detected or the timeout is reached.
//
// Example:
//
//	ct.WaitForEvent("data-loaded", 2*time.Second)
func (ct *ComponentTest) WaitForEvent(name string, timeout time.Duration) {
	ct.harness.t.Helper()

	deadline := time.Now().Add(timeout)
	interval := 10 * time.Millisecond

	for time.Now().Before(deadline) {
		if ct.events.tracker.WasFired(name) {
			return
		}
		time.Sleep(interval)
	}

	// Timeout reached
	ct.harness.t.Errorf("timeout waiting for event %q to be fired", name)
}
