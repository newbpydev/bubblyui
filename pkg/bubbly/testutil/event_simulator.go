package testutil

import (
	"time"
)

// Event represents an event to be emitted with its name and payload.
// Used by EmitMultiple to emit multiple events in sequence.
type Event struct {
	// Name is the event name to emit
	Name string
	
	// Payload is the data to pass with the event
	Payload interface{}
}

// Emit simulates event emission on the component.
// It calls the component's Emit method and allows a small delay
// for event handlers to execute.
//
// This is the primary method for testing event-driven behavior.
// The event is emitted synchronously and handlers execute immediately.
//
// Example:
//
//	ct := harness.Mount(createComponent())
//	ct.Emit("submit", FormData{Username: "test"})
//	ct.AssertEventFired("submit")
func (ct *ComponentTest) Emit(name string, payload interface{}) {
	ct.harness.t.Helper()
	
	// Emit event on the component
	ct.component.Emit(name, payload)
	
	// Give a tiny delay for event handlers to execute
	// This ensures handlers run before assertions
	time.Sleep(1 * time.Millisecond)
}

// EmitAndWait emits an event and waits for it to be processed.
// It polls the event tracker until the event is detected or the timeout is reached.
//
// This is useful for testing async event handling where you need to ensure
// the event has been fully processed before making assertions.
//
// Example:
//
//	ct := harness.Mount(createComponent())
//	ct.EmitAndWait("fetch-data", nil, 2*time.Second)
//	ct.AssertRefEquals("loading", false)
func (ct *ComponentTest) EmitAndWait(name string, payload interface{}, timeout time.Duration) {
	ct.harness.t.Helper()
	
	// Emit the event
	ct.component.Emit(name, payload)
	
	// Wait for event to be tracked
	deadline := time.Now().Add(timeout)
	interval := 10 * time.Millisecond
	
	for time.Now().Before(deadline) {
		if ct.events.tracker.WasFired(name) {
			// Event was tracked, give handlers time to complete
			time.Sleep(10 * time.Millisecond)
			return
		}
		time.Sleep(interval)
	}
	
	// Timeout reached - event may not have been tracked
	// Don't fail here, just return (caller can assert if needed)
}

// EmitMultiple emits multiple events in sequence.
// Each event is emitted with a small delay between them to ensure
// handlers execute in order.
//
// This is useful for testing sequences of events or complex workflows
// that involve multiple event emissions.
//
// Example:
//
//	events := []Event{
//	    {Name: "start", Payload: nil},
//	    {Name: "process", Payload: data},
//	    {Name: "complete", Payload: result},
//	}
//	ct.EmitMultiple(events)
func (ct *ComponentTest) EmitMultiple(events []Event) {
	ct.harness.t.Helper()
	
	for _, event := range events {
		ct.Emit(event.Name, event.Payload)
	}
}
