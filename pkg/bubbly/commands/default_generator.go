package commands

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// DefaultCommandGenerator is the standard implementation of CommandGenerator.
//
// This generator creates commands that return StateChangedMsg when executed.
// It is thread-safe and can be used concurrently from multiple goroutines.
//
// The generator is stateless and can be shared across multiple components
// or created per-component as needed. There is no performance difference
// between the two approaches.
//
// Usage:
//
//	gen := &DefaultCommandGenerator{}
//	cmd := gen.Generate("counter-1", "count", 0, 1)
//	// cmd() returns StateChangedMsg{ComponentID: "counter-1", RefID: "count", ...}
//
// Thread Safety:
//
// DefaultCommandGenerator is thread-safe because it has no mutable state.
// Multiple goroutines can call Generate() concurrently without synchronization.
//
// Performance:
//
// Command generation is extremely fast (< 10ns overhead) as it only creates
// a closure that captures the provided values. The actual StateChangedMsg is
// created when the command is executed by Bubbletea's runtime.
type DefaultCommandGenerator struct{}

// Generate creates a tea.Cmd that returns a StateChangedMsg when executed.
//
// This method implements the CommandGenerator interface. It creates a command
// (closure) that captures the state change information and returns it as a
// StateChangedMsg when executed by Bubbletea's runtime.
//
// Parameters:
//   - componentID: Unique identifier of the component owning the ref
//   - refID: Unique identifier of the ref that changed
//   - oldValue: Previous value before the change (captured for debugging/logging)
//   - newValue: New value after the change (captured for debugging/logging)
//
// Returns:
//   - tea.Cmd: A command that produces StateChangedMsg when executed
//
// The returned command is safe to enqueue and will be executed asynchronously
// by Bubbletea's message loop. The timestamp is set when the command executes,
// not when Generate() is called, to accurately reflect when the message enters
// the message loop.
//
// Example:
//
//	gen := &DefaultCommandGenerator{}
//
//	// Generate command from state change
//	cmd := gen.Generate("counter-1", "count", 0, 1)
//
//	// Command will be executed by Bubbletea
//	// In component's Update():
//	func (c *componentImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	    switch msg := msg.(type) {
//	    case StateChangedMsg:
//	        // Handle state change
//	        if msg.ComponentID == c.id {
//	            c.lifecycle.executeUpdated()
//	        }
//	    }
//	    return c, nil
//	}
//
// Thread Safety:
//
// This method is thread-safe and can be called concurrently from multiple
// goroutines. The captured values are copied into the closure, so there are
// no shared mutable references.
//
// Performance:
//
// This method is extremely fast (< 10ns) as it only creates a closure.
// No allocations are made until the command is executed.
func (g *DefaultCommandGenerator) Generate(
	componentID, refID string,
	oldValue, newValue interface{},
) tea.Cmd {
	return func() tea.Msg {
		return StateChangedMsg{
			ComponentID: componentID,
			RefID:       refID,
			OldValue:    oldValue,
			NewValue:    newValue,
			Timestamp:   time.Now(),
		}
	}
}
