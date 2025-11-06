package commands

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestStateChangedMsg_Structure verifies the StateChangedMsg structure
func TestStateChangedMsg_Structure(t *testing.T) {
	tests := []struct {
		name        string
		componentID string
		refID       string
		oldValue    interface{}
		newValue    interface{}
	}{
		{
			name:        "string value change",
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
			name:        "boolean value change",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := StateChangedMsg{
				ComponentID: tt.componentID,
				RefID:       tt.refID,
				OldValue:    tt.oldValue,
				NewValue:    tt.newValue,
				Timestamp:   time.Now(),
			}

			assert.Equal(t, tt.componentID, msg.ComponentID)
			assert.Equal(t, tt.refID, msg.RefID)
			assert.Equal(t, tt.oldValue, msg.OldValue)
			assert.Equal(t, tt.newValue, msg.NewValue)
			assert.False(t, msg.Timestamp.IsZero())
		})
	}
}

// TestCommandGenerator_Interface verifies the CommandGenerator interface
func TestCommandGenerator_Interface(t *testing.T) {
	tests := []struct {
		name        string
		componentID string
		refID       string
		oldValue    interface{}
		newValue    interface{}
		wantCmd     bool
	}{
		{
			name:        "generates command for value change",
			componentID: "test-1",
			refID:       "value",
			oldValue:    0,
			newValue:    1,
			wantCmd:     true,
		},
		{
			name:        "generates command for nil to value",
			componentID: "test-2",
			refID:       "optional",
			oldValue:    nil,
			newValue:    "data",
			wantCmd:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test generator
			var gen CommandGenerator = &testGenerator{}

			cmd := gen.Generate(tt.componentID, tt.refID, tt.oldValue, tt.newValue)

			if tt.wantCmd {
				assert.NotNil(t, cmd, "expected command to be generated")

				// Execute command and verify message
				msg := cmd()
				assert.NotNil(t, msg, "expected message from command")

				// Verify it's a StateChangedMsg
				stateMsg, ok := msg.(StateChangedMsg)
				assert.True(t, ok, "expected StateChangedMsg type")
				assert.Equal(t, tt.componentID, stateMsg.ComponentID)
				assert.Equal(t, tt.refID, stateMsg.RefID)
			} else {
				assert.Nil(t, cmd, "expected no command")
			}
		})
	}
}

// TestStateChangedMsg_Timestamp verifies timestamp is set
func TestStateChangedMsg_Timestamp(t *testing.T) {
	before := time.Now()
	time.Sleep(1 * time.Millisecond) // Ensure time passes

	msg := StateChangedMsg{
		ComponentID: "test",
		RefID:       "value",
		OldValue:    0,
		NewValue:    1,
		Timestamp:   time.Now(),
	}

	time.Sleep(1 * time.Millisecond)
	after := time.Now()

	assert.True(t, msg.Timestamp.After(before) || msg.Timestamp.Equal(before))
	assert.True(t, msg.Timestamp.Before(after) || msg.Timestamp.Equal(after))
}

// testGenerator is a test implementation of CommandGenerator
type testGenerator struct{}

func (g *testGenerator) Generate(componentID, refID string, oldValue, newValue interface{}) tea.Cmd {
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
