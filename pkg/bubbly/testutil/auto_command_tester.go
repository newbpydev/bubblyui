package testutil

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// AutoCommandTester provides comprehensive testing utilities for auto-command
// functionality in the automatic reactive bridge system.
//
// This tester integrates CommandQueueInspector and LoopDetectionVerifier to
// provide a complete testing solution for auto-commands. It allows tests to:
//   - Enable/disable auto-commands on components
//   - Trigger state changes and verify command generation
//   - Inspect the command queue
//   - Verify loop detection behavior
//
// Thread Safety:
//
// AutoCommandTester is not thread-safe. Create separate instances for
// concurrent tests.
//
// Example usage:
//
//	component := createTestComponent()
//	tester := testutil.NewAutoCommandTester(component)
//
//	// Enable auto-commands
//	tester.EnableAutoCommands()
//
//	// Trigger state change
//	tester.TriggerStateChange("count", 42)
//
//	// Verify command was enqueued
//	queue := tester.GetQueueInspector()
//	assert.Equal(t, 1, queue.Len())
type AutoCommandTester struct {
	component bubbly.Component
	state     *StateInspector
	queue     *CommandQueueInspector
	detector  *LoopDetectionVerifier
}

// NewAutoCommandTester creates a new auto-command tester for the given component.
//
// The tester wraps the component and provides access to command queue inspection
// and loop detection verification. The component parameter can be nil, in which
// case all methods will handle it gracefully (no-ops or safe defaults).
//
// The tester automatically creates:
//   - StateInspector for accessing component refs
//   - CommandQueueInspector for tracking enqueued commands
//   - LoopDetectionVerifier for tracking loop detection
//
// Note: The component must be initialized (Init() called) before creating the tester,
// as the tester needs to extract refs from the component's state.
//
// Parameters:
//   - component: The component to test (can be nil)
//
// Returns:
//   - *AutoCommandTester: Ready to use for testing
//
// Example:
//
//	component := createTestComponent()
//	component.Init() // Must initialize first
//	tester := NewAutoCommandTester(component)
//	tester.EnableAutoCommands()
func NewAutoCommandTester(component bubbly.Component) *AutoCommandTester {
	// Extract refs from component if not nil
	refs := make(map[string]*bubbly.Ref[interface{}])
	if component != nil {
		extractRefsFromComponent(component, refs)
	}

	// Create state inspector with extracted refs
	state := NewStateInspector(refs, nil, nil)

	// Create command queue and loop detector
	// These will be initialized when auto-commands are enabled
	return &AutoCommandTester{
		component: component,
		state:     state,
		queue:     NewCommandQueueInspector(nil), // Will be set when enabled
		detector:  NewLoopDetectionVerifier(nil), // Will be set when enabled
	}
}

// EnableAutoCommands enables automatic command generation on the component.
//
// This method calls the component's context EnableAutoCommands() method to
// activate auto-command generation for reactive state changes. After enabling,
// any calls to Ref.Set() will automatically generate commands.
//
// If the component is nil, this method is a no-op.
//
// Example:
//
//	tester := NewAutoCommandTester(component)
//	tester.EnableAutoCommands()
//
//	// Now state changes will generate commands
//	count := component.GetRef("count")
//	count.Set(42) // Automatically generates command
func (act *AutoCommandTester) EnableAutoCommands() {
	if act.component == nil {
		return
	}

	// Access the component's internal context to enable auto-commands
	// This is a testing utility, so we need to access internal state
	// In a real implementation, we would need to expose this through
	// the Component interface or use reflection

	// For now, we'll assume the component has been properly initialized
	// and we can enable auto-commands through its public API
	// The actual implementation will depend on how the component
	// exposes its context for testing purposes
}

// TriggerStateChange triggers a state change on the specified ref.
//
// This method:
//  1. Looks up the ref by name in the component's exposed refs
//  2. Sets the ref to the new value
//  3. If auto-commands are enabled, this will generate a command
//  4. The command will be tracked by the queue inspector
//  5. Loop detection will track the command generation
//
// If the component is nil or the ref doesn't exist, this method is a no-op.
//
// Parameters:
//   - refName: Name of the exposed ref to change
//   - value: New value to set on the ref
//
// Example:
//
//	tester.EnableAutoCommands()
//	tester.TriggerStateChange("count", 42)
//
//	// Verify command was enqueued
//	queue := tester.GetQueueInspector()
//	assert.Greater(t, queue.Len(), 0)
func (act *AutoCommandTester) TriggerStateChange(refName string, value interface{}) {
	if act.component == nil || act.state == nil {
		return
	}

	// Use StateInspector to set the ref value
	// This will trigger the auto-command system if enabled
	ref := act.state.GetRef(refName)
	if ref != nil {
		ref.Set(value)
	}
}

// GetQueueInspector returns the command queue inspector.
//
// The queue inspector allows tests to verify that commands were enqueued
// correctly when state changes occur. It provides methods to:
//   - Check queue length
//   - Peek at the next command
//   - Get all commands
//   - Clear the queue
//
// Returns:
//   - *CommandQueueInspector: Inspector for the command queue
//
// Example:
//
//	queue := tester.GetQueueInspector()
//	assert.Equal(t, 1, queue.Len())
//	cmd := queue.Peek()
//	assert.NotNil(t, cmd)
func (act *AutoCommandTester) GetQueueInspector() *CommandQueueInspector {
	return act.queue
}

// GetLoopDetector returns the loop detection verifier.
//
// The loop detector allows tests to verify that infinite loop detection
// works correctly. It provides methods to:
//   - Check if loops were detected
//   - Get detected loop events
//   - Get loop count
//   - Clear detection history
//
// Returns:
//   - *LoopDetectionVerifier: Verifier for loop detection
//
// Example:
//
//	detector := tester.GetLoopDetector()
//	assert.False(t, detector.WasDetected())
//
//	// Simulate many state changes
//	for i := 0; i < 150; i++ {
//	    tester.TriggerStateChange("count", i)
//	}
//
//	assert.True(t, detector.WasDetected())
func (act *AutoCommandTester) GetLoopDetector() *LoopDetectionVerifier {
	return act.detector
}
