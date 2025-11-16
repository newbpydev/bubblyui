package testutil

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// GenerateArgs captures the arguments passed to MockCommandGenerator.Generate().
//
// This struct stores all parameters from a Generate() call for later inspection
// and verification in tests. It's useful for asserting that the correct arguments
// were passed when testing command generation.
//
// Example usage:
//
//	mock := NewMockCommandGenerator(someCmd)
//	mock.Generate("comp-1", "count", 0, 1)
//
//	args := mock.GetCapturedArgs()
//	assert.Equal(t, "comp-1", args[0].ComponentID)
//	assert.Equal(t, "count", args[0].RefID)
type GenerateArgs struct {
	ComponentID string
	RefID       string
	OldValue    interface{}
	NewValue    interface{}
}

// MockCommandGenerator is a mock implementation of bubbly.CommandGenerator for testing.
//
// This mock allows you to:
//   - Configure what command to return from Generate()
//   - Track how many times Generate() was called
//   - Capture all arguments passed to Generate()
//   - Assert on call counts and arguments
//
// Thread Safety:
//
// MockCommandGenerator is thread-safe. All methods use a mutex to protect internal
// state, allowing concurrent calls from multiple goroutines during testing.
//
// Example usage:
//
//	// Create mock that returns a specific command
//	mock := NewMockCommandGenerator(func() tea.Msg {
//	    return bubbly.StateChangedMsg{ComponentID: "test"}
//	})
//
//	// Use in tests
//	cmd := mock.Generate("comp-1", "count", 0, 1)
//	assert.NotNil(t, cmd)
//
//	// Verify it was called
//	mock.AssertCalled(t, 1)
//
//	// Inspect captured arguments
//	args := mock.GetCapturedArgs()
//	assert.Equal(t, "comp-1", args[0].ComponentID)
type MockCommandGenerator struct {
	mu             sync.Mutex
	generateCalled int
	returnCmd      tea.Cmd
	capturedArgs   []GenerateArgs
}

// NewMockCommandGenerator creates a new mock command generator.
//
// The mock will return the provided command from all Generate() calls.
// If returnCmd is nil, Generate() will return nil, but the call will still
// be tracked for verification.
//
// Parameters:
//   - returnCmd: The command to return from Generate() (can be nil)
//
// Returns:
//   - *MockCommandGenerator: Ready to use for testing
//
// Example:
//
//	// Mock that returns a StateChangedMsg
//	mock := NewMockCommandGenerator(func() tea.Msg {
//	    return bubbly.StateChangedMsg{
//	        ComponentID: "test-comp",
//	        RefID:       "test-ref",
//	    }
//	})
//
//	// Mock that returns nil
//	nilMock := NewMockCommandGenerator(nil)
func NewMockCommandGenerator(returnCmd tea.Cmd) *MockCommandGenerator {
	return &MockCommandGenerator{
		returnCmd:    returnCmd,
		capturedArgs: []GenerateArgs{},
	}
}

// Generate implements bubbly.CommandGenerator.Generate().
//
// This method:
//  1. Increments the call counter (thread-safe)
//  2. Captures the arguments for later inspection
//  3. Returns the configured command (or nil)
//
// The method is thread-safe and can be called concurrently from multiple
// goroutines. All state updates are protected by a mutex.
//
// Parameters:
//   - componentID: Component identifier (captured for inspection)
//   - refID: Ref identifier (captured for inspection)
//   - oldValue: Previous value (captured for inspection)
//   - newValue: New value (captured for inspection)
//
// Returns:
//   - tea.Cmd: The configured return command (or nil)
//
// Example:
//
//	mock := NewMockCommandGenerator(func() tea.Msg { return "msg" })
//
//	// Generate command
//	cmd := mock.Generate("comp-1", "count", 0, 1)
//	assert.NotNil(t, cmd)
//
//	// Verify call was tracked
//	assert.Equal(t, 1, mock.generateCalled)
//
//	// Inspect arguments
//	args := mock.GetCapturedArgs()
//	assert.Equal(t, "comp-1", args[0].ComponentID)
func (mcg *MockCommandGenerator) Generate(
	componentID, refID string,
	oldValue, newValue interface{},
) tea.Cmd {
	mcg.mu.Lock()
	defer mcg.mu.Unlock()

	// Track call count
	mcg.generateCalled++

	// Capture arguments
	mcg.capturedArgs = append(mcg.capturedArgs, GenerateArgs{
		ComponentID: componentID,
		RefID:       refID,
		OldValue:    oldValue,
		NewValue:    newValue,
	})

	// Return configured command
	return mcg.returnCmd
}

// AssertCalled asserts that Generate() was called the expected number of times.
//
// This is a convenience assertion method that checks the call count and reports
// a clear error message if it doesn't match. It uses t.Helper() to ensure the
// error is reported at the correct line in the test.
//
// Parameters:
//   - t: The testing.T instance (or testingT interface)
//   - times: The expected number of calls
//
// Example:
//
//	mock := NewMockCommandGenerator(someCmd)
//	mock.Generate("comp-1", "ref-1", 0, 1)
//	mock.Generate("comp-2", "ref-2", "a", "b")
//
//	mock.AssertCalled(t, 2) // Passes
//	mock.AssertCalled(t, 3) // Fails with clear error
func (mcg *MockCommandGenerator) AssertCalled(t testingT, times int) {
	t.Helper()

	mcg.mu.Lock()
	actual := mcg.generateCalled
	mcg.mu.Unlock()

	if actual != times {
		t.Errorf("mock command generator: expected Generate() called %d times, got %d", times, actual)
	}
}

// GetCapturedArgs returns all captured Generate() arguments.
//
// This method returns a copy of the captured arguments slice, so modifications
// to the returned slice do not affect the mock's internal state.
//
// Returns:
//   - []GenerateArgs: All captured arguments from Generate() calls
//
// Example:
//
//	mock := NewMockCommandGenerator(someCmd)
//	mock.Generate("comp-1", "ref-1", 0, 1)
//	mock.Generate("comp-2", "ref-2", "a", "b")
//
//	args := mock.GetCapturedArgs()
//	assert.Len(t, args, 2)
//	assert.Equal(t, "comp-1", args[0].ComponentID)
//	assert.Equal(t, "comp-2", args[1].ComponentID)
func (mcg *MockCommandGenerator) GetCapturedArgs() []GenerateArgs {
	mcg.mu.Lock()
	defer mcg.mu.Unlock()

	// Return a copy to prevent external modification
	result := make([]GenerateArgs, len(mcg.capturedArgs))
	copy(result, mcg.capturedArgs)
	return result
}

// Clear resets all tracking state.
//
// This method clears the call count and captured arguments. It's useful for
// resetting state between test cases or cleaning up after testing. Safe to
// call multiple times.
//
// The configured return command is NOT cleared, so the mock can be reused
// with the same behavior after clearing.
//
// Example:
//
//	mock := NewMockCommandGenerator(someCmd)
//	mock.Generate("comp-1", "ref-1", 0, 1)
//	assert.Equal(t, 1, mock.generateCalled)
//
//	mock.Clear()
//	assert.Equal(t, 0, mock.generateCalled)
//	assert.Empty(t, mock.GetCapturedArgs())
//
//	// Can be used again
//	mock.Generate("comp-2", "ref-2", 1, 2)
//	assert.Equal(t, 1, mock.generateCalled)
func (mcg *MockCommandGenerator) Clear() {
	mcg.mu.Lock()
	defer mcg.mu.Unlock()

	mcg.generateCalled = 0
	mcg.capturedArgs = []GenerateArgs{}
}
