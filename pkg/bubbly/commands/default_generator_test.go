package commands

import (
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestDefaultCommandGenerator_Generate verifies command generation works
func TestDefaultCommandGenerator_Generate(t *testing.T) {
	tests := []struct {
		name        string
		componentID string
		refID       string
		oldValue    interface{}
		newValue    interface{}
	}{
		{
			name:        "integer value change",
			componentID: "counter-1",
			refID:       "count",
			oldValue:    0,
			newValue:    1,
		},
		{
			name:        "string value change",
			componentID: "form-1",
			refID:       "name",
			oldValue:    "Alice",
			newValue:    "Bob",
		},
		{
			name:        "boolean toggle",
			componentID: "toggle-1",
			refID:       "enabled",
			oldValue:    false,
			newValue:    true,
		},
		{
			name:        "nil to value",
			componentID: "optional-1",
			refID:       "data",
			oldValue:    nil,
			newValue:    "initialized",
		},
		{
			name:        "complex type",
			componentID: "list-1",
			refID:       "items",
			oldValue:    []string{"a"},
			newValue:    []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &DefaultCommandGenerator{}

			cmd := gen.Generate(tt.componentID, tt.refID, tt.oldValue, tt.newValue)

			// Verify command is not nil
			assert.NotNil(t, cmd, "Generate should return a command")

			// Execute command and verify message
			msg := cmd()
			assert.NotNil(t, msg, "Command should return a message")

			// Verify message type
			stateMsg, ok := msg.(StateChangedMsg)
			assert.True(t, ok, "Message should be StateChangedMsg")

			// Verify message fields
			assert.Equal(t, tt.componentID, stateMsg.ComponentID)
			assert.Equal(t, tt.refID, stateMsg.RefID)
			assert.Equal(t, tt.oldValue, stateMsg.OldValue)
			assert.Equal(t, tt.newValue, stateMsg.NewValue)
			assert.False(t, stateMsg.Timestamp.IsZero(), "Timestamp should be set")
		})
	}
}

// TestDefaultCommandGenerator_MessageReturnedCorrectly verifies message structure
func TestDefaultCommandGenerator_MessageReturnedCorrectly(t *testing.T) {
	gen := &DefaultCommandGenerator{}

	componentID := "test-component"
	refID := "test-ref"
	oldValue := "old"
	newValue := "new"

	cmd := gen.Generate(componentID, refID, oldValue, newValue)
	msg := cmd()

	stateMsg, ok := msg.(StateChangedMsg)
	assert.True(t, ok, "Message must be StateChangedMsg type")

	// Verify all fields are correctly set
	assert.Equal(t, componentID, stateMsg.ComponentID, "ComponentID mismatch")
	assert.Equal(t, refID, stateMsg.RefID, "RefID mismatch")
	assert.Equal(t, oldValue, stateMsg.OldValue, "OldValue mismatch")
	assert.Equal(t, newValue, stateMsg.NewValue, "NewValue mismatch")
	assert.NotZero(t, stateMsg.Timestamp, "Timestamp must be set")
}

// TestDefaultCommandGenerator_TimestampSet verifies timestamp is always set
func TestDefaultCommandGenerator_TimestampSet(t *testing.T) {
	gen := &DefaultCommandGenerator{}

	before := time.Now()
	cmd := gen.Generate("comp", "ref", 0, 1)
	msg := cmd()
	after := time.Now()

	stateMsg := msg.(StateChangedMsg)

	// Timestamp should be between before and after
	assert.True(t, stateMsg.Timestamp.After(before) || stateMsg.Timestamp.Equal(before))
	assert.True(t, stateMsg.Timestamp.Before(after) || stateMsg.Timestamp.Equal(after))
}

// TestDefaultCommandGenerator_ValuesCapturedCorrectly verifies values are preserved
func TestDefaultCommandGenerator_ValuesCapturedCorrectly(t *testing.T) {
	tests := []struct {
		name     string
		oldValue interface{}
		newValue interface{}
	}{
		{
			name:     "integers",
			oldValue: 42,
			newValue: 100,
		},
		{
			name:     "strings",
			oldValue: "hello",
			newValue: "world",
		},
		{
			name:     "nil values",
			oldValue: nil,
			newValue: nil,
		},
		{
			name:     "mixed types",
			oldValue: 123,
			newValue: "456",
		},
		{
			name:     "slices",
			oldValue: []int{1, 2, 3},
			newValue: []int{4, 5, 6},
		},
		{
			name:     "maps",
			oldValue: map[string]int{"a": 1},
			newValue: map[string]int{"b": 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &DefaultCommandGenerator{}

			cmd := gen.Generate("comp", "ref", tt.oldValue, tt.newValue)
			msg := cmd()

			stateMsg := msg.(StateChangedMsg)
			assert.Equal(t, tt.oldValue, stateMsg.OldValue)
			assert.Equal(t, tt.newValue, stateMsg.NewValue)
		})
	}
}

// TestDefaultCommandGenerator_ThreadSafe verifies concurrent command generation
func TestDefaultCommandGenerator_ThreadSafe(t *testing.T) {
	gen := &DefaultCommandGenerator{}

	const goroutines = 100
	const iterations = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Track all generated messages
	messages := make(chan StateChangedMsg, goroutines*iterations)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				componentID := "comp"
				refID := "ref"
				oldValue := j
				newValue := j + 1

				cmd := gen.Generate(componentID, refID, oldValue, newValue)
				msg := cmd()

				stateMsg, ok := msg.(StateChangedMsg)
				assert.True(t, ok, "Message must be StateChangedMsg")

				messages <- stateMsg
			}
		}()
	}

	wg.Wait()
	close(messages)

	// Verify all messages were generated correctly
	count := 0
	for msg := range messages {
		count++
		assert.Equal(t, "comp", msg.ComponentID)
		assert.Equal(t, "ref", msg.RefID)
		assert.NotZero(t, msg.Timestamp)
	}

	assert.Equal(t, goroutines*iterations, count, "All messages should be generated")
}

// TestDefaultCommandGenerator_ImplementsInterface verifies interface compliance
func TestDefaultCommandGenerator_ImplementsInterface(t *testing.T) {
	var _ CommandGenerator = &DefaultCommandGenerator{}

	gen := &DefaultCommandGenerator{}
	assert.NotNil(t, gen)
}

// TestDefaultCommandGenerator_MultipleGenerations verifies multiple calls work
func TestDefaultCommandGenerator_MultipleGenerations(t *testing.T) {
	gen := &DefaultCommandGenerator{}

	// Generate multiple commands
	cmd1 := gen.Generate("comp1", "ref1", 0, 1)
	cmd2 := gen.Generate("comp2", "ref2", "a", "b")
	cmd3 := gen.Generate("comp3", "ref3", false, true)

	// Execute all commands
	msg1 := cmd1()
	msg2 := cmd2()
	msg3 := cmd3()

	// Verify all messages are correct
	stateMsg1 := msg1.(StateChangedMsg)
	assert.Equal(t, "comp1", stateMsg1.ComponentID)
	assert.Equal(t, "ref1", stateMsg1.RefID)

	stateMsg2 := msg2.(StateChangedMsg)
	assert.Equal(t, "comp2", stateMsg2.ComponentID)
	assert.Equal(t, "ref2", stateMsg2.RefID)

	stateMsg3 := msg3.(StateChangedMsg)
	assert.Equal(t, "comp3", stateMsg3.ComponentID)
	assert.Equal(t, "ref3", stateMsg3.RefID)
}

// TestDefaultCommandGenerator_CommandIsTeaCmd verifies tea.Cmd compatibility
func TestDefaultCommandGenerator_CommandIsTeaCmd(t *testing.T) {
	gen := &DefaultCommandGenerator{}

	cmd := gen.Generate("comp", "ref", 0, 1)

	// Verify it's a valid tea.Cmd
	var _ tea.Cmd = cmd

	// Verify it can be executed
	msg := cmd()
	assert.NotNil(t, msg)
}
