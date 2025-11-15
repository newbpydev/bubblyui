package testutil

import (
	"fmt"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// MockCommand is a mock Bubbletea command for testing.
//
// It tracks whether the command was executed and can return
// either a predefined message or an error message.
//
// Thread-safe for concurrent access.
//
// Example:
//
//	mock, cmd := testutil.NewMockCommand(myMsg)
//	result := cmd()  // Execute the command
//	assert.True(t, mock.Executed())
//	assert.Equal(t, myMsg, result)
type MockCommand struct {
	mu       sync.RWMutex
	executed bool
	message  tea.Msg
	error    error
}

// MockErrorMsg is a message type that wraps an error.
//
// This is returned by commands created with NewMockCommandWithError.
// It implements the error interface for convenience.
//
// Example:
//
//	mock, cmd := testutil.NewMockCommandWithError(myErr)
//	result := cmd()
//	errMsg, ok := result.(testutil.MockErrorMsg)
//	assert.True(t, ok)
//	assert.Equal(t, myErr, errMsg.Err)
type MockErrorMsg struct {
	Err error
}

// Error implements the error interface for MockErrorMsg.
//
// Returns the error string, or "<nil>" if the error is nil.
func (m MockErrorMsg) Error() string {
	if m.Err == nil {
		return "<nil>"
	}
	return m.Err.Error()
}

// NewMockCommand creates a new mock command that returns the specified message.
//
// The command tracks whether it was executed and returns the provided message
// when called. The message can be nil.
//
// Returns:
//   - *MockCommand: The mock command object for assertions
//   - tea.Cmd: The command function that can be executed
//
// Example:
//
//	type myMsg struct{ value string }
//	mock, cmd := testutil.NewMockCommand(myMsg{value: "test"})
//
//	// Execute the command
//	result := cmd()
//
//	// Assert execution
//	assert.True(t, mock.Executed())
//	assert.Equal(t, myMsg{value: "test"}, result)
func NewMockCommand(msg tea.Msg) (*MockCommand, tea.Cmd) {
	mock := &MockCommand{
		message: msg,
	}

	cmd := func() tea.Msg {
		mock.mu.Lock()
		mock.executed = true
		mock.mu.Unlock()
		return mock.message
	}

	return mock, cmd
}

// NewMockCommandWithError creates a new mock command that returns an error message.
//
// The command tracks whether it was executed and returns a MockErrorMsg
// containing the provided error when called.
//
// Parameters:
//   - err: The error to wrap in the message (can be nil)
//
// Returns:
//   - *MockCommand: The mock command object for assertions
//   - tea.Cmd: The command function that can be executed
//
// Example:
//
//	mock, cmd := testutil.NewMockCommandWithError(errors.New("test error"))
//
//	// Execute the command
//	result := cmd()
//
//	// Assert execution and error
//	assert.True(t, mock.Executed())
//	errMsg, ok := result.(testutil.MockErrorMsg)
//	assert.True(t, ok)
//	assert.Equal(t, "test error", errMsg.Error())
func NewMockCommandWithError(err error) (*MockCommand, tea.Cmd) {
	mock := &MockCommand{
		error: err,
	}

	cmd := func() tea.Msg {
		mock.mu.Lock()
		mock.executed = true
		mock.mu.Unlock()
		return MockErrorMsg{Err: mock.error}
	}

	return mock, cmd
}

// Executed returns whether the command has been executed.
//
// Thread-safe for concurrent access.
//
// Example:
//
//	mock, cmd := testutil.NewMockCommand(myMsg)
//	assert.False(t, mock.Executed())  // Not executed yet
//	cmd()
//	assert.True(t, mock.Executed())   // Now executed
func (mc *MockCommand) Executed() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.executed
}

// Message returns the message that will be returned by the command.
//
// This is the message provided to NewMockCommand, or nil for
// commands created with NewMockCommandWithError.
//
// Thread-safe for concurrent access.
//
// Example:
//
//	mock, _ := testutil.NewMockCommand(myMsg)
//	assert.Equal(t, myMsg, mock.Message())
func (mc *MockCommand) Message() tea.Msg {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.message
}

// Error returns the error that will be wrapped in the error message.
//
// This is the error provided to NewMockCommandWithError, or nil for
// commands created with NewMockCommand.
//
// Thread-safe for concurrent access.
//
// Example:
//
//	mock, _ := testutil.NewMockCommandWithError(myErr)
//	assert.Equal(t, myErr, mock.Error())
func (mc *MockCommand) Error() error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.error
}

// Reset clears the executed flag, allowing the command to be tested again.
//
// The message and error fields are not modified.
//
// Thread-safe for concurrent access.
//
// Example:
//
//	mock, cmd := testutil.NewMockCommand(myMsg)
//	cmd()
//	assert.True(t, mock.Executed())
//
//	mock.Reset()
//	assert.False(t, mock.Executed())
func (mc *MockCommand) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.executed = false
}

// AssertExecuted asserts that the command was executed.
//
// This is a convenience method for testing. It uses t.Helper() to
// report the error at the correct line in the test.
//
// Parameters:
//   - t: The testing interface (typically *testing.T)
//
// Example:
//
//	mock, cmd := testutil.NewMockCommand(myMsg)
//	cmd()
//	mock.AssertExecuted(t)  // Passes
func (mc *MockCommand) AssertExecuted(t testingT) {
	t.Helper()
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if !mc.executed {
		t.Errorf("command was not executed")
	}
}

// AssertNotExecuted asserts that the command was not executed.
//
// This is a convenience method for testing. It uses t.Helper() to
// report the error at the correct line in the test.
//
// Parameters:
//   - t: The testing interface (typically *testing.T)
//
// Example:
//
//	mock, _ := testutil.NewMockCommand(myMsg)
//	mock.AssertNotExecuted(t)  // Passes (not executed yet)
func (mc *MockCommand) AssertNotExecuted(t testingT) {
	t.Helper()
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.executed {
		t.Errorf("command was executed but should not have been")
	}
}

// String returns a string representation of the MockCommand for debugging.
//
// Example output: "MockCommand{executed=true, hasMessage=true, hasError=false}"
func (mc *MockCommand) String() string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return fmt.Sprintf("MockCommand{executed=%v, hasMessage=%v, hasError=%v}",
		mc.executed,
		mc.message != nil,
		mc.error != nil,
	)
}
