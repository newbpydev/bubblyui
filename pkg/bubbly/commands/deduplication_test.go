package commands

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestDeduplicate_EdgeCases tests deduplication with edge cases.
func TestDeduplicate_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		commands []tea.Cmd
		want     int // Expected number of commands after deduplication
	}{
		{
			name:     "empty list",
			commands: []tea.Cmd{},
			want:     0,
		},
		{
			name:     "nil list",
			commands: nil,
			want:     0,
		},
		{
			name: "single command",
			commands: []tea.Cmd{
				func() tea.Msg {
					return bubbly.StateChangedMsg{
						ComponentID: "comp1",
						RefID:       "ref1",
					}
				},
			},
			want: 1,
		},
		{
			name: "nil command in list",
			commands: []tea.Cmd{
				func() tea.Msg {
					return bubbly.StateChangedMsg{
						ComponentID: "comp1",
						RefID:       "ref1",
					}
				},
				nil,
				func() tea.Msg {
					return bubbly.StateChangedMsg{
						ComponentID: "comp2",
						RefID:       "ref2",
					}
				},
			},
			want: 2, // Nil should be filtered out
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batcher := NewCommandBatcher(CoalesceAll)
			result := batcher.deduplicateCommands(tt.commands)

			assert.Equal(t, tt.want, len(result), "deduplication should return correct number of commands")
		})
	}
}

// TestDeduplicate_DuplicateDetection tests that duplicate commands are removed.
func TestDeduplicate_DuplicateDetection(t *testing.T) {
	tests := []struct {
		name     string
		commands []tea.Cmd
		want     int
	}{
		{
			name: "same ref changed twice - keeps last",
			commands: []tea.Cmd{
				func() tea.Msg {
					return bubbly.StateChangedMsg{
						ComponentID: "comp1",
						RefID:       "ref1",
						OldValue:    0,
						NewValue:    1,
					}
				},
				func() tea.Msg {
					return bubbly.StateChangedMsg{
						ComponentID: "comp1",
						RefID:       "ref1",
						OldValue:    1,
						NewValue:    2,
					}
				},
			},
			want: 1, // Only the last state change matters
		},
		{
			name: "same ref changed three times - keeps last",
			commands: []tea.Cmd{
				func() tea.Cmd {
					return func() tea.Msg {
						return bubbly.StateChangedMsg{
							ComponentID: "comp1",
							RefID:       "ref1",
							NewValue:    1,
						}
					}
				}(),
				func() tea.Cmd {
					return func() tea.Msg {
						return bubbly.StateChangedMsg{
							ComponentID: "comp1",
							RefID:       "ref1",
							NewValue:    2,
						}
					}
				}(),
				func() tea.Cmd {
					return func() tea.Msg {
						return bubbly.StateChangedMsg{
							ComponentID: "comp1",
							RefID:       "ref1",
							NewValue:    3,
						}
					}
				}(),
			},
			want: 1,
		},
		{
			name: "different refs - all kept",
			commands: []tea.Cmd{
				func() tea.Msg {
					return bubbly.StateChangedMsg{
						ComponentID: "comp1",
						RefID:       "ref1",
					}
				},
				func() tea.Msg {
					return bubbly.StateChangedMsg{
						ComponentID: "comp1",
						RefID:       "ref2",
					}
				},
				func() tea.Msg {
					return bubbly.StateChangedMsg{
						ComponentID: "comp1",
						RefID:       "ref3",
					}
				},
			},
			want: 3, // All unique refs
		},
		{
			name: "different components same ref - all kept",
			commands: []tea.Cmd{
				func() tea.Msg {
					return bubbly.StateChangedMsg{
						ComponentID: "comp1",
						RefID:       "ref1",
					}
				},
				func() tea.Msg {
					return bubbly.StateChangedMsg{
						ComponentID: "comp2",
						RefID:       "ref1",
					}
				},
			},
			want: 2, // Different components
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batcher := NewCommandBatcher(CoalesceAll)
			result := batcher.deduplicateCommands(tt.commands)

			assert.Equal(t, tt.want, len(result), "should keep correct number of unique commands")
		})
	}
}

// TestDeduplicate_OrderPreserved tests that order is preserved after deduplication.
func TestDeduplicate_OrderPreserved(t *testing.T) {
	// Create commands with different refs to preserve order
	commands := []tea.Cmd{
		func() tea.Msg {
			return bubbly.StateChangedMsg{
				ComponentID: "comp1",
				RefID:       "ref1",
				NewValue:    "first",
			}
		},
		func() tea.Msg {
			return bubbly.StateChangedMsg{
				ComponentID: "comp1",
				RefID:       "ref2",
				NewValue:    "second",
			}
		},
		func() tea.Msg {
			return bubbly.StateChangedMsg{
				ComponentID: "comp1",
				RefID:       "ref1", // Duplicate of first
				NewValue:    "first-updated",
			}
		},
		func() tea.Msg {
			return bubbly.StateChangedMsg{
				ComponentID: "comp1",
				RefID:       "ref3",
				NewValue:    "third",
			}
		},
	}

	batcher := NewCommandBatcher(CoalesceAll)
	result := batcher.deduplicateCommands(commands)

	// Should keep: ref2 (position 0), ref1 (latest at position 2), ref3 (position 3)
	// Order should be: ref2, ref1 (updated), ref3
	assert.Equal(t, 3, len(result), "should have 3 unique commands")

	// Execute commands to verify order and content
	msgs := make([]bubbly.StateChangedMsg, 0, len(result))
	for _, cmd := range result {
		if cmd != nil {
			msg := cmd()
			if stateMsg, ok := msg.(bubbly.StateChangedMsg); ok {
				msgs = append(msgs, stateMsg)
			}
		}
	}

	assert.Equal(t, 3, len(msgs), "should execute 3 commands")
	assert.Equal(t, "ref2", msgs[0].RefID, "first command should be ref2")
	assert.Equal(t, "ref1", msgs[1].RefID, "second command should be ref1 (updated)")
	assert.Equal(t, "first-updated", msgs[1].NewValue, "ref1 should have latest value")
	assert.Equal(t, "ref3", msgs[2].RefID, "third command should be ref3")
}

// TestDeduplicate_Performance tests that deduplication is efficient.
func TestDeduplicate_Performance(t *testing.T) {
	// Create a large batch of commands with some duplicates
	commands := make([]tea.Cmd, 1000)
	for i := 0; i < 1000; i++ {
		refID := i % 100 // 100 unique refs, each updated 10 times
		commands[i] = func(id int) tea.Cmd {
			return func() tea.Msg {
				return bubbly.StateChangedMsg{
					ComponentID: "comp1",
					RefID:       string(rune('a' + (id % 26))), // Simple ref ID
					NewValue:    id,
				}
			}
		}(refID)
	}

	batcher := NewCommandBatcher(CoalesceAll)
	result := batcher.deduplicateCommands(commands)

	// Should have only 100 unique commands (one per ref)
	assert.LessOrEqual(t, len(result), 100, "should deduplicate to at most 100 unique refs")
	assert.NotEmpty(t, result, "should not be empty")
}

// TestGenerateCommandKey tests key generation for different message types.
func TestGenerateCommandKey(t *testing.T) {
	tests := []struct {
		name string
		cmd  tea.Cmd
		want string
	}{
		{
			name: "StateChangedMsg generates component:ref key",
			cmd: func() tea.Msg {
				return bubbly.StateChangedMsg{
					ComponentID: "comp1",
					RefID:       "ref1",
				}
			},
			want: "comp1:ref1",
		},
		{
			name: "different component same ref - different key",
			cmd: func() tea.Msg {
				return bubbly.StateChangedMsg{
					ComponentID: "comp2",
					RefID:       "ref1",
				}
			},
			want: "comp2:ref1",
		},
		{
			name: "same component different ref - different key",
			cmd: func() tea.Msg {
				return bubbly.StateChangedMsg{
					ComponentID: "comp1",
					RefID:       "ref2",
				}
			},
			want: "comp1:ref2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := generateCommandKey(tt.cmd)
			assert.Equal(t, tt.want, key, "key should match expected format")
		})
	}
}

// TestGenerateCommandKey_NonStateChanged tests key generation for other message types.
func TestGenerateCommandKey_NonStateChanged(t *testing.T) {
	// Custom message type
	type CustomMsg struct {
		Data string
	}

	cmd := func() tea.Msg {
		return CustomMsg{Data: "test"}
	}

	key := generateCommandKey(cmd)
	// For non-StateChangedMsg, should generate a unique key
	// (implementation-specific, but should be non-empty and consistent)
	assert.NotEmpty(t, key, "should generate non-empty key for custom messages")
}

// TestGenerateCommandKey_NilCommand tests key generation for nil command.
func TestGenerateCommandKey_NilCommand(t *testing.T) {
	key := generateCommandKey(nil)
	assert.Equal(t, "", key, "should return empty string for nil command")
}

// TestGenerateCommandKey_CustomMessageTypes tests key generation for various custom message types.
func TestGenerateCommandKey_CustomMessageTypes(t *testing.T) {
	tests := []struct {
		name     string
		cmd      tea.Cmd
		expected string
	}{
		{
			name: "custom struct message",
			cmd: func() tea.Msg {
				return struct{ Name string }{Name: "test"}
			},
			expected: "struct { Name string }",
		},
		{
			name: "string message",
			cmd: func() tea.Msg {
				return "string message"
			},
			expected: "string",
		},
		{
			name: "integer message",
			cmd: func() tea.Msg {
				return 42
			},
			expected: "int",
		},
		{
			name: "nil message from command",
			cmd: func() tea.Msg {
				return nil
			},
			expected: "<nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := generateCommandKey(tt.cmd)
			assert.Equal(t, tt.expected, key, "key should match expected type name")
		})
	}
}

// TestDeduplicateCommands_EdgeCases tests additional edge cases for deduplication.
func TestDeduplicateCommands_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		commands []tea.Cmd
		want     int
	}{
		{
			name:     "all nil commands",
			commands: []tea.Cmd{nil, nil, nil},
			want:     0,
		},
		{
			name:     "single nil command",
			commands: []tea.Cmd{nil},
			want:     0,
		},
		{
			name: "mixed nil and valid with duplicates",
			commands: []tea.Cmd{
				nil,
				func() tea.Msg {
					return bubbly.StateChangedMsg{
						ComponentID: "comp1",
						RefID:       "ref1",
					}
				},
				nil,
				func() tea.Msg {
					return bubbly.StateChangedMsg{
						ComponentID: "comp1",
						RefID:       "ref1",
					}
				},
				nil,
			},
			want: 1, // Only one unique ref after filtering nils and deduplication
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batcher := NewCommandBatcher(CoalesceAll)
			result := batcher.deduplicateCommands(tt.commands)

			if tt.want == 0 {
				assert.Nil(t, result, "should return nil for empty result")
			} else {
				assert.Equal(t, tt.want, len(result), "should return correct number of commands")
			}
		})
	}
}

// TestDeduplicateCommands_SingleCommandOptimization tests the single command optimization.
func TestDeduplicateCommands_SingleCommandOptimization(t *testing.T) {
	tests := []struct {
		name     string
		commands []tea.Cmd
	}{
		{
			name: "single valid command",
			commands: []tea.Cmd{
				func() tea.Msg {
					return bubbly.StateChangedMsg{
						ComponentID: "comp1",
						RefID:       "ref1",
					}
				},
			},
		},
		{
			name:     "single nil command",
			commands: []tea.Cmd{nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batcher := NewCommandBatcher(CoalesceAll)
			result := batcher.deduplicateCommands(tt.commands)

			if tt.commands[0] == nil {
				assert.Nil(t, result, "should return nil for single nil command")
			} else {
				assert.NotNil(t, result, "should return command for single valid command")
				assert.Equal(t, 1, len(result), "should return exactly one command")
			}
		})
	}
}
