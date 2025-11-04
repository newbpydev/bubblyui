package commands

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CommandRef is a reactive reference that automatically generates Bubbletea commands
// when its value changes. It wraps a standard Ref[T] and extends it with command
// generation capabilities for the automatic reactive bridge.
//
// When enabled, calling Set() will:
//  1. Update the underlying Ref value synchronously
//  2. Generate a tea.Cmd using the configured CommandGenerator
//  3. Enqueue the command in the component's command queue
//  4. The command will be returned from the component's Update() method
//  5. Bubbletea will execute the command, triggering a UI update cycle
//
// When disabled, Set() behaves like a normal Ref.Set() with no command generation.
//
// Thread Safety:
//
// CommandRef is thread-safe. Multiple goroutines can call Set() concurrently.
// The underlying Ref handles synchronization for the value, and the command queue
// handles synchronization for command enqueueing.
//
// Example:
//
//	// Create a command ref
//	ref := bubbly.NewRef(0)
//	queue := NewCommandQueue()
//	gen := &DefaultCommandGenerator{}
//
//	cmdRef := &CommandRef[int]{
//	    Ref:         ref,
//	    componentID: "counter-1",
//	    refID:       "count",
//	    commandGen:  gen,
//	    queue:       queue,
//	    enabled:     true,
//	}
//
//	// Set value - generates command automatically
//	cmdRef.Set(42)
//
//	// Command is now queued and will be returned from component's Update()
type CommandRef[T any] struct {
	// Ref is the underlying reactive reference that holds the value
	*bubbly.Ref[T]

	// componentID identifies the component owning this ref
	componentID string

	// refID uniquely identifies this ref within the component
	refID string

	// commandGen generates tea.Cmd from state changes
	commandGen CommandGenerator

	// queue stores pending commands for the component
	queue *CommandQueue

	// enabled controls whether commands are generated
	// When false, Set() behaves like normal Ref.Set()
	enabled bool
}

// Set updates the value and generates a command if enabled.
//
// This method overrides the underlying Ref.Set() to add automatic command
// generation. The behavior depends on the enabled flag:
//
// When enabled=true:
//  1. Captures the old value
//  2. Updates the underlying Ref value (synchronous)
//  3. Generates a tea.Cmd using the CommandGenerator
//  4. Enqueues the command in the component's command queue
//
// When enabled=false:
//   - Calls the underlying Ref.Set() directly
//   - No command generation occurs
//   - Behaves exactly like a normal Ref
//
// The state update is always synchronous - the new value is immediately
// visible via Get(). The command generation and UI update are asynchronous
// and happen on the next Bubbletea update cycle.
//
// Thread Safety:
//
// This method is thread-safe and can be called concurrently from multiple
// goroutines. The underlying Ref and CommandQueue handle synchronization.
//
// Example:
//
//	cmdRef.Set(42)  // Value updated immediately, command queued
//	value := cmdRef.Get()  // Returns 42 (synchronous)
//	// UI will update on next Bubbletea cycle (asynchronous)
func (cr *CommandRef[T]) Set(value T) {
	// If disabled, just use normal Ref.Set()
	if !cr.enabled {
		cr.Ref.Set(value)
		return
	}

	// Capture old value before update
	oldValue := cr.Ref.Get()

	// Update the underlying Ref (synchronous)
	cr.Ref.Set(value)

	// Generate command for the state change
	cmd := cr.commandGen.Generate(
		cr.componentID,
		cr.refID,
		oldValue,
		value,
	)

	// Enqueue command for component to return from Update()
	cr.queue.Enqueue(cmd)
}
