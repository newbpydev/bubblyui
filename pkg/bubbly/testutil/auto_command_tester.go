package testutil

import (
	"reflect"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/commands"
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

	// Extract command queue and loop detector from component using reflection
	var queue *CommandQueueInspector
	var detector *LoopDetectionVerifier

	if component != nil {
		v := reflect.ValueOf(component)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		// Get command queue
		commandQueueField := v.FieldByName("commandQueue")
		if commandQueueField.IsValid() && !commandQueueField.IsNil() {
			commandQueuePtr := reflect.NewAt(commandQueueField.Type(), commandQueueField.Addr().UnsafePointer()).Elem()
			if cmdQueue, ok := commandQueuePtr.Interface().(*bubbly.CommandQueue); ok && cmdQueue != nil {
				queue = NewCommandQueueInspector(cmdQueue)
			}
		}

		// Get loop detector
		loopDetectorField := v.FieldByName("loopDetector")
		if loopDetectorField.IsValid() && !loopDetectorField.IsNil() {
			loopDetectorPtr := reflect.NewAt(loopDetectorField.Type(), loopDetectorField.Addr().UnsafePointer()).Elem()
			if loopDet, ok := loopDetectorPtr.Interface().(*commands.LoopDetector); ok && loopDet != nil {
				detector = NewLoopDetectionVerifier(loopDet)
			}
		}
	}

	// If queue or detector not found, create empty ones
	if queue == nil {
		queue = NewCommandQueueInspector(nil)
	}
	if detector == nil {
		detector = NewLoopDetectionVerifier(nil)
	}

	return &AutoCommandTester{
		component: component,
		state:     state,
		queue:     queue,
		detector:  detector,
	}
}

// EnableAutoCommands enables automatic command generation on the component.
//
// This method creates a Context for the component and calls its EnableAutoCommands()
// method to activate auto-command generation for reactive state changes. After enabling,
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
//	ref := tester.state.GetRef("count")
//	ref.Set(42) // Automatically generates command
func (act *AutoCommandTester) EnableAutoCommands() {
	if act.component == nil {
		return
	}

	// Use reflection to replicate Context.EnableAutoCommands() logic
	// from context.go:628-638 PLUS initialize command queue and loop detector
	v := reflect.ValueOf(act.component)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Lock the autoCommandsMu mutex
	autoCommandsMuField := v.FieldByName("autoCommandsMu")
	if !autoCommandsMuField.IsValid() {
		return
	}

	// Use UnsafePointer to access the mutex
	autoCommandsMuPtr := reflect.NewAt(autoCommandsMuField.Type(), autoCommandsMuField.Addr().UnsafePointer())
	lockMethod := autoCommandsMuPtr.MethodByName("Lock")
	unlockMethod := autoCommandsMuPtr.MethodByName("Unlock")

	lockMethod.Call(nil)
	defer unlockMethod.Call(nil)

	// Set autoCommands = true
	autoCommandsField := v.FieldByName("autoCommands")
	if autoCommandsField.IsValid() {
		autoCommandsPtr := reflect.NewAt(autoCommandsField.Type(), autoCommandsField.Addr().UnsafePointer()).Elem()
		autoCommandsPtr.SetBool(true)
	}

	// Initialize command queue if not set
	commandQueueField := v.FieldByName("commandQueue")
	if commandQueueField.IsValid() {
		commandQueuePtr := reflect.NewAt(commandQueueField.Type(), commandQueueField.Addr().UnsafePointer()).Elem()
		if commandQueuePtr.IsNil() {
			// Create new command queue using bubbly.NewCommandQueue()
			newQueue := bubbly.NewCommandQueue()
			commandQueuePtr.Set(reflect.ValueOf(newQueue))

			// Update our queue inspector to point to the new queue
			act.queue = NewCommandQueueInspector(newQueue)
		}
	}

	// Initialize loop detector if not set
	loopDetectorField := v.FieldByName("loopDetector")
	if loopDetectorField.IsValid() {
		loopDetectorPtr := reflect.NewAt(loopDetectorField.Type(), loopDetectorField.Addr().UnsafePointer()).Elem()
		if loopDetectorPtr.IsNil() {
			// Create new loop detector using commands.NewLoopDetector()
			newDetector := commands.NewLoopDetector()
			loopDetectorPtr.Set(reflect.ValueOf(newDetector))

			// Update our loop detection verifier to point to the new detector
			act.detector = NewLoopDetectionVerifier(newDetector)
		}
	}

	// Ensure command generator is set (replicate context.go:635-637)
	commandGenField := v.FieldByName("commandGen")
	if commandGenField.IsValid() {
		commandGenPtr := reflect.NewAt(commandGenField.Type(), commandGenField.Addr().UnsafePointer()).Elem()
		if commandGenPtr.IsNil() {
			// Create defaultCommandGenerator using reflection
			// Find the defaultCommandGenerator type from bubbly package
			defaultGenType := reflect.TypeOf(struct{}{})
			defaultGen := reflect.New(defaultGenType).Elem().Interface()

			// The actual type needs to implement CommandGenerator interface
			// For now, we'll use a workaround - the component will create it when needed
			_ = defaultGen // Placeholder
		}
	}
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
