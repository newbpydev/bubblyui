package testutil

import (
	"errors"
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// Test message types
type testMsg struct {
	value string
}

type anotherMsg struct {
	count int
}

// TestNewMockCommand tests creating a mock command with a message
func TestNewMockCommand(t *testing.T) {
	tests := []struct {
		name    string
		message tea.Msg
	}{
		{
			name:    "string message",
			message: testMsg{value: "test"},
		},
		{
			name:    "int message",
			message: anotherMsg{count: 42},
		},
		{
			name:    "nil message",
			message: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, cmd := NewMockCommand(tt.message)

			// Should not be executed initially
			assert.False(t, mock.Executed(), "command should not be executed initially")
			assert.Equal(t, tt.message, mock.Message(), "message should match")
			assert.Nil(t, mock.Error(), "error should be nil")

			// Command should not be nil
			assert.NotNil(t, cmd, "command should not be nil")
		})
	}
}

// TestNewMockCommandWithError tests creating a mock command with an error
func TestNewMockCommandWithError(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "standard error",
			err:  errors.New("test error"),
		},
		{
			name: "nil error",
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, cmd := NewMockCommandWithError(tt.err)

			// Should not be executed initially
			assert.False(t, mock.Executed(), "command should not be executed initially")
			assert.Equal(t, tt.err, mock.Error(), "error should match")
			assert.Nil(t, mock.Message(), "message should be nil for error commands")

			// Command should not be nil
			assert.NotNil(t, cmd, "command should not be nil")
		})
	}
}

// TestMockCommand_Execute tests executing a mock command
func TestMockCommand_Execute(t *testing.T) {
	t.Run("execute command with message", func(t *testing.T) {
		expectedMsg := testMsg{value: "hello"}
		mock, cmd := NewMockCommand(expectedMsg)

		// Execute the command
		result := cmd()

		// Should be marked as executed
		assert.True(t, mock.Executed(), "command should be executed")

		// Should return the message
		assert.Equal(t, expectedMsg, result, "should return the message")
	})

	t.Run("execute command with error", func(t *testing.T) {
		expectedErr := errors.New("test error")
		mock, cmd := NewMockCommandWithError(expectedErr)

		// Execute the command
		result := cmd()

		// Should be marked as executed
		assert.True(t, mock.Executed(), "command should be executed")

		// Should return error message
		errMsg, ok := result.(MockErrorMsg)
		assert.True(t, ok, "should return MockErrorMsg type")
		assert.Equal(t, expectedErr, errMsg.Err, "error should match")
	})

	t.Run("multiple executions", func(t *testing.T) {
		mock, cmd := NewMockCommand(testMsg{value: "test"})

		// Execute multiple times
		cmd()
		cmd()
		cmd()

		// Should still be marked as executed
		assert.True(t, mock.Executed(), "command should be executed")
	})
}

// TestMockCommand_AssertExecuted tests the assertion helper
func TestMockCommand_AssertExecuted(t *testing.T) {
	t.Run("assert executed - success", func(t *testing.T) {
		mock, cmd := NewMockCommand(testMsg{value: "test"})
		cmd() // Execute it

		// Should not fail
		mockT := &mockTestingT{}
		mock.AssertExecuted(mockT)
		assert.False(t, mockT.failed, "assertion should pass")
	})

	t.Run("assert executed - failure", func(t *testing.T) {
		mock, _ := NewMockCommand(testMsg{value: "test"})
		// Don't execute it

		// Should fail
		mockT := &mockTestingT{}
		mock.AssertExecuted(mockT)
		assert.True(t, mockT.failed, "assertion should fail")
		assert.NotEmpty(t, mockT.errors, "should have error messages")
	})
}

// TestMockCommand_AssertNotExecuted tests the negative assertion helper
func TestMockCommand_AssertNotExecuted(t *testing.T) {
	t.Run("assert not executed - success", func(t *testing.T) {
		mock, _ := NewMockCommand(testMsg{value: "test"})
		// Don't execute it

		// Should not fail
		mockT := &mockTestingT{}
		mock.AssertNotExecuted(mockT)
		assert.False(t, mockT.failed, "assertion should pass")
	})

	t.Run("assert not executed - failure", func(t *testing.T) {
		mock, cmd := NewMockCommand(testMsg{value: "test"})
		cmd() // Execute it

		// Should fail
		mockT := &mockTestingT{}
		mock.AssertNotExecuted(mockT)
		assert.True(t, mockT.failed, "assertion should fail")
		assert.NotEmpty(t, mockT.errors, "should have error messages")
	})
}

// TestMockCommand_Reset tests resetting the mock
func TestMockCommand_Reset(t *testing.T) {
	mock, cmd := NewMockCommand(testMsg{value: "test"})

	// Execute it
	cmd()
	assert.True(t, mock.Executed(), "should be executed")

	// Reset
	mock.Reset()
	assert.False(t, mock.Executed(), "should not be executed after reset")

	// Message and error should remain
	assert.NotNil(t, mock.Message(), "message should remain after reset")
	assert.Nil(t, mock.Error(), "error should remain nil after reset")
}

// TestMockCommand_ThreadSafety tests concurrent access
func TestMockCommand_ThreadSafety(t *testing.T) {
	mock, cmd := NewMockCommand(testMsg{value: "test"})

	// Execute concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd()
			_ = mock.Executed()
			_ = mock.Message()
			_ = mock.Error()
		}()
	}

	wg.Wait()

	// Should be executed
	assert.True(t, mock.Executed(), "command should be executed")
}

// TestMockCommand_NilMessage tests handling nil messages
func TestMockCommand_NilMessage(t *testing.T) {
	mock, cmd := NewMockCommand(nil)

	result := cmd()

	assert.True(t, mock.Executed(), "command should be executed")
	assert.Nil(t, result, "should return nil message")
	assert.Nil(t, mock.Message(), "message should be nil")
}

// TestMockCommand_GettersBeforeExecution tests accessing getters before execution
func TestMockCommand_GettersBeforeExecution(t *testing.T) {
	msg := testMsg{value: "test"}
	mock, _ := NewMockCommand(msg)

	// Should be able to access message before execution
	assert.Equal(t, msg, mock.Message(), "should return message before execution")
	assert.Nil(t, mock.Error(), "should return nil error before execution")
	assert.False(t, mock.Executed(), "should not be executed")
}

// TestMockErrorMsg_Error tests the error interface implementation
func TestMockErrorMsg_Error(t *testing.T) {
	err := errors.New("test error")
	msg := MockErrorMsg{Err: err}

	assert.Equal(t, "test error", msg.Error(), "should implement error interface")
}

// TestMockErrorMsg_NilError tests MockErrorMsg with nil error
func TestMockErrorMsg_NilError(t *testing.T) {
	msg := MockErrorMsg{Err: nil}

	assert.Equal(t, "<nil>", msg.Error(), "should handle nil error")
}

// TestMockCommand_String tests the String method for debugging
func TestMockCommand_String(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		mock, _ := NewMockCommand(testMsg{value: "test"})
		str := mock.String()
		assert.Contains(t, str, "MockCommand")
		assert.Contains(t, str, "executed=false")
		assert.Contains(t, str, "hasMessage=true")
		assert.Contains(t, str, "hasError=false")
	})

	t.Run("with error", func(t *testing.T) {
		mock, _ := NewMockCommandWithError(errors.New("test"))
		str := mock.String()
		assert.Contains(t, str, "MockCommand")
		assert.Contains(t, str, "hasMessage=false")
		assert.Contains(t, str, "hasError=true")
	})

	t.Run("after execution", func(t *testing.T) {
		mock, cmd := NewMockCommand(testMsg{value: "test"})
		cmd()
		str := mock.String()
		assert.Contains(t, str, "executed=true")
	})
}

// mockTestingT is defined in assertions_state_test.go and reused here
