package testutil

import (
	"reflect"
)

// AssertEventFired asserts that an event with the given name was fired at least once.
// It reports an error via t.Errorf if the event was not fired.
//
// Parameters:
//   - name: The name of the event to check
//
// Example:
//
//	ct := harness.Mount(createComponent())
//	ct.component.Emit("click", nil)
//	ct.AssertEventFired("click")  // Passes
//	ct.AssertEventFired("submit") // Fails - event not fired
func (ct *ComponentTest) AssertEventFired(name string) {
	ct.harness.t.Helper()

	if !ct.events.tracker.WasFired(name) {
		ct.harness.t.Errorf("event %q was not fired", name)
	}
}

// AssertEventNotFired asserts that an event with the given name was NOT fired.
// It reports an error via t.Errorf if the event was fired.
//
// Parameters:
//   - name: The name of the event to check
//
// Example:
//
//	ct := harness.Mount(createComponent())
//	ct.AssertEventNotFired("click")  // Passes - no click event
//	ct.component.Emit("click", nil)
//	ct.AssertEventNotFired("click")  // Fails - event was fired
func (ct *ComponentTest) AssertEventNotFired(name string) {
	ct.harness.t.Helper()

	if ct.events.tracker.WasFired(name) {
		ct.harness.t.Errorf("event %q should not have fired", name)
	}
}

// AssertEventPayload asserts that the most recent event with the given name
// has a payload that matches the expected value using reflect.DeepEqual.
//
// If no events with the given name were fired, the assertion fails.
// If multiple events with the same name were fired, only the last one is checked.
//
// Parameters:
//   - name: The name of the event to check
//   - expected: The expected payload value
//
// Example:
//
//	ct := harness.Mount(createComponent())
//	ct.component.Emit("submit", "test-data")
//	ct.AssertEventPayload("submit", "test-data")  // Passes
//	ct.AssertEventPayload("submit", "other-data") // Fails - payload doesn't match
func (ct *ComponentTest) AssertEventPayload(name string, expected interface{}) {
	ct.harness.t.Helper()

	// Get all events with this name
	events := ct.events.tracker.GetEvents(name)

	// Check if any events were fired
	if len(events) == 0 {
		ct.harness.t.Errorf("event %q was not fired", name)
		return
	}

	// Get the last event's payload
	actual := events[len(events)-1].Payload

	// Compare payloads using reflect.DeepEqual
	if !reflect.DeepEqual(actual, expected) {
		ct.harness.t.Errorf("event %q payload: expected %v, got %v",
			name, formatValue(expected), formatValue(actual))
	}
}

// AssertEventCount asserts that an event with the given name was fired
// exactly the expected number of times.
//
// Parameters:
//   - name: The name of the event to check
//   - count: The expected number of times the event should have fired
//
// Example:
//
//	ct := harness.Mount(createComponent())
//	ct.component.Emit("click", nil)
//	ct.component.Emit("click", nil)
//	ct.AssertEventCount("click", 2)  // Passes
//	ct.AssertEventCount("click", 1)  // Fails - fired 2 times, not 1
//	ct.AssertEventCount("submit", 0) // Passes - never fired
func (ct *ComponentTest) AssertEventCount(name string, count int) {
	ct.harness.t.Helper()

	actual := ct.events.tracker.FiredCount(name)

	if actual != count {
		ct.harness.t.Errorf("event %q: expected %d occurrences, got %d",
			name, count, actual)
	}
}
