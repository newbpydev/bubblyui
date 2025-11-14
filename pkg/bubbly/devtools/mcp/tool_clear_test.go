package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createCallToolRequest creates a properly formatted CallToolRequest for testing
func createCallToolRequest(toolName string, params map[string]interface{}) (*mcp.CallToolRequest, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      toolName,
			Arguments: paramsJSON,
		},
	}, nil
}

// TestRegisterClearStateHistoryTool tests the registration of the clear_state_history tool.
func TestRegisterClearStateHistoryTool(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "successful registration",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			cfg := DefaultMCPConfig()
			dt := devtools.Enable()
			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)

			// Register tool
			err = server.RegisterClearStateHistoryTool()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRegisterClearEventLogTool tests the registration of the clear_event_log tool.
func TestRegisterClearEventLogTool(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "successful registration",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			cfg := DefaultMCPConfig()
			dt := devtools.Enable()
			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)

			// Register tool
			err = server.RegisterClearEventLogTool()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestHandleClearStateHistoryTool tests the clear_state_history tool handler.
func TestHandleClearStateHistoryTool(t *testing.T) {
	tests := []struct {
		name          string
		setupHistory  func(*devtools.DevToolsStore)
		params        map[string]interface{}
		wantError     bool
		wantCleared   int
		errorContains string
	}{
		{
			name: "clear empty history with confirmation",
			setupHistory: func(store *devtools.DevToolsStore) {
				// No history to add
			},
			params: map[string]interface{}{
				"confirm": true,
			},
			wantError:   false,
			wantCleared: 0,
		},
		{
			name: "clear history with items",
			setupHistory: func(store *devtools.DevToolsStore) {
				history := store.GetStateHistory()
				for i := 0; i < 5; i++ {
					history.Record(devtools.StateChange{
						RefID:     "ref-1",
						RefName:   "counter",
						OldValue:  i,
						NewValue:  i + 1,
						Timestamp: time.Now(),
						Source:    "test",
					})
				}
			},
			params: map[string]interface{}{
				"confirm": true,
			},
			wantError:   false,
			wantCleared: 5,
		},
		{
			name: "clear without confirmation",
			setupHistory: func(store *devtools.DevToolsStore) {
				history := store.GetStateHistory()
				history.Record(devtools.StateChange{
					RefID:     "ref-1",
					RefName:   "counter",
					OldValue:  0,
					NewValue:  1,
					Timestamp: time.Now(),
					Source:    "test",
				})
			},
			params: map[string]interface{}{
				"confirm": false,
			},
			wantError:     true,
			errorContains: "Confirmation required",
		},
		{
			name: "missing confirm parameter",
			setupHistory: func(store *devtools.DevToolsStore) {
				// No history
			},
			params:        map[string]interface{}{},
			wantError:     true,
			errorContains: "confirm parameter is required",
		},
		{
			name: "invalid confirm type",
			setupHistory: func(store *devtools.DevToolsStore) {
				// No history
			},
			params: map[string]interface{}{
				"confirm": "true", // String instead of bool
			},
			wantError:     true,
			errorContains: "confirm parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			cfg := DefaultMCPConfig()
			dt := devtools.Enable()
			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)

			// Setup history
			tt.setupHistory(server.store)

			// Register tool
			err = server.RegisterClearStateHistoryTool()
			require.NoError(t, err)

			// Create request
			request, err := createCallToolRequest("clear_state_history", tt.params)
			require.NoError(t, err)

			// Call handler
			result, err := server.handleClearStateHistoryTool(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Check result
			if tt.wantError {
				assert.True(t, result.IsError, "Expected error result")
				if tt.errorContains != "" {
					assert.NotEmpty(t, result.Content)
					textContent := result.Content[0].(*mcp.TextContent)
					assert.Contains(t, textContent.Text, tt.errorContains)
				}
			} else {
				assert.False(t, result.IsError, "Expected success result")
				assert.NotEmpty(t, result.Content)
				textContent := result.Content[0].(*mcp.TextContent)
				assert.Contains(t, textContent.Text, "cleared successfully")

				// Verify history was actually cleared
				history := server.store.GetStateHistory()
				assert.Equal(t, 0, len(history.GetAll()), "History should be empty after clear")
			}
		})
	}
}

// TestHandleClearEventLogTool tests the clear_event_log tool handler.
func TestHandleClearEventLogTool(t *testing.T) {
	tests := []struct {
		name          string
		setupEvents   func(*devtools.DevToolsStore)
		params        map[string]interface{}
		wantError     bool
		wantCleared   int
		errorContains string
	}{
		{
			name: "clear empty event log with confirmation",
			setupEvents: func(store *devtools.DevToolsStore) {
				// No events to add
			},
			params: map[string]interface{}{
				"confirm": true,
			},
			wantError:   false,
			wantCleared: 0,
		},
		{
			name: "clear event log with items",
			setupEvents: func(store *devtools.DevToolsStore) {
				eventLog := store.GetEventLog()
				for i := 0; i < 10; i++ {
					eventLog.Append(devtools.EventRecord{
						ID:        "event-1",
						Name:      "click",
						SourceID:  "button-1",
						TargetID:  "handler-1",
						Payload:   nil,
						Timestamp: time.Now(),
						Duration:  time.Millisecond,
					})
				}
			},
			params: map[string]interface{}{
				"confirm": true,
			},
			wantError:   false,
			wantCleared: 10,
		},
		{
			name: "clear without confirmation",
			setupEvents: func(store *devtools.DevToolsStore) {
				eventLog := store.GetEventLog()
				eventLog.Append(devtools.EventRecord{
					ID:        "event-1",
					Name:      "click",
					SourceID:  "button-1",
					TargetID:  "handler-1",
					Payload:   nil,
					Timestamp: time.Now(),
					Duration:  time.Millisecond,
				})
			},
			params: map[string]interface{}{
				"confirm": false,
			},
			wantError:     true,
			errorContains: "Confirmation required",
		},
		{
			name: "missing confirm parameter",
			setupEvents: func(store *devtools.DevToolsStore) {
				// No events
			},
			params:        map[string]interface{}{},
			wantError:     true,
			errorContains: "confirm parameter is required",
		},
		{
			name: "invalid confirm type",
			setupEvents: func(store *devtools.DevToolsStore) {
				// No events
			},
			params: map[string]interface{}{
				"confirm": 1, // Number instead of bool
			},
			wantError:     true,
			errorContains: "confirm parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			cfg := DefaultMCPConfig()
			dt := devtools.Enable()
			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)

			// Setup events
			tt.setupEvents(server.store)

			// Register tool
			err = server.RegisterClearEventLogTool()
			require.NoError(t, err)

			// Create request
			request, err := createCallToolRequest("clear_event_log", tt.params)
			require.NoError(t, err)

			// Call handler
			result, err := server.handleClearEventLogTool(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Check result
			if tt.wantError {
				assert.True(t, result.IsError, "Expected error result")
				if tt.errorContains != "" {
					assert.NotEmpty(t, result.Content)
					textContent := result.Content[0].(*mcp.TextContent)
					assert.Contains(t, textContent.Text, tt.errorContains)
				}
			} else {
				assert.False(t, result.IsError, "Expected success result")
				assert.NotEmpty(t, result.Content)
				textContent := result.Content[0].(*mcp.TextContent)
				assert.Contains(t, textContent.Text, "cleared successfully")

				// Verify event log was actually cleared
				eventLog := server.store.GetEventLog()
				assert.Equal(t, 0, eventLog.Len(), "Event log should be empty after clear")
			}
		})
	}
}

// TestClearStateHistory_ThreadSafety tests concurrent access to clear state history.
func TestClearStateHistory_ThreadSafety(t *testing.T) {
	// Create test server
	cfg := DefaultMCPConfig()
	dt := devtools.Enable()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Register tool
	err = server.RegisterClearStateHistoryTool()
	require.NoError(t, err)

	// Add some history
	history := server.store.GetStateHistory()
	for i := 0; i < 100; i++ {
		history.Record(devtools.StateChange{
			RefID:     "ref-1",
			RefName:   "counter",
			OldValue:  i,
			NewValue:  i + 1,
			Timestamp: time.Now(),
			Source:    "test",
		})
	}

	// Run concurrent clear operations
	var wg sync.WaitGroup
	concurrency := 10

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			params := map[string]interface{}{
				"confirm": true,
			}
			request, _ := createCallToolRequest("clear_state_history", params)

			result, err := server.handleClearStateHistoryTool(context.Background(), request)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		}()
	}

	wg.Wait()

	// Verify history is cleared
	finalHistory := server.store.GetStateHistory()
	assert.Equal(t, 0, len(finalHistory.GetAll()), "History should be empty after concurrent clears")
}

// TestClearEventLog_ThreadSafety tests concurrent access to clear event log.
func TestClearEventLog_ThreadSafety(t *testing.T) {
	// Create test server
	cfg := DefaultMCPConfig()
	dt := devtools.Enable()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Register tool
	err = server.RegisterClearEventLogTool()
	require.NoError(t, err)

	// Add some events
	eventLog := server.store.GetEventLog()
	for i := 0; i < 100; i++ {
		eventLog.Append(devtools.EventRecord{
			ID:        "event-1",
			Name:      "click",
			SourceID:  "button-1",
			TargetID:  "handler-1",
			Payload:   nil,
			Timestamp: time.Now(),
			Duration:  time.Millisecond,
		})
	}

	// Run concurrent clear operations
	var wg sync.WaitGroup
	concurrency := 10

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			params := map[string]interface{}{
				"confirm": true,
			}
			request, _ := createCallToolRequest("clear_event_log", params)

			result, err := server.handleClearEventLogTool(context.Background(), request)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		}()
	}

	wg.Wait()

	// Verify event log is cleared
	finalEventLog := server.store.GetEventLog()
	assert.Equal(t, 0, finalEventLog.Len(), "Event log should be empty after concurrent clears")
}

// TestClearTools_InvalidJSON tests handling of malformed JSON parameters.
func TestClearTools_InvalidJSON(t *testing.T) {
	tests := []struct {
		name        string
		toolHandler func(*MCPServer, context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error)
		toolName    string
	}{
		{
			name:        "clear_state_history with invalid JSON",
			toolHandler: (*MCPServer).handleClearStateHistoryTool,
			toolName:    "clear_state_history",
		},
		{
			name:        "clear_event_log with invalid JSON",
			toolHandler: (*MCPServer).handleClearEventLogTool,
			toolName:    "clear_event_log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			cfg := DefaultMCPConfig()
			dt := devtools.Enable()
			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)

			// Create request with invalid JSON
			request := &mcp.CallToolRequest{
				Params: &mcp.CallToolParamsRaw{
					Name:      tt.toolName,
					Arguments: []byte("{invalid json}"),
				},
			}

			// Call handler
			result, err := tt.toolHandler(server, context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Should return error result
			assert.True(t, result.IsError)
			assert.NotEmpty(t, result.Content)
			textContent := result.Content[0].(*mcp.TextContent)
			assert.Contains(t, strings.ToLower(textContent.Text), "parse")
		})
	}
}

// TestClearTools_AtomicOperations tests that clear operations are atomic.
func TestClearTools_AtomicOperations(t *testing.T) {
	t.Run("state history clear is atomic", func(t *testing.T) {
		// Create test server
		cfg := DefaultMCPConfig()
		dt := devtools.Enable()
		server, err := NewMCPServer(cfg, dt)
		require.NoError(t, err)

		// Add history
		history := server.store.GetStateHistory()
		for i := 0; i < 50; i++ {
			history.Record(devtools.StateChange{
				RefID:     "ref-1",
				RefName:   "counter",
				OldValue:  i,
				NewValue:  i + 1,
				Timestamp: time.Now(),
				Source:    "test",
			})
		}

		// Clear should be atomic - either all cleared or none
		params := map[string]interface{}{
			"confirm": true,
		}
		request, _ := createCallToolRequest("clear_state_history", params)

		result, err := server.handleClearStateHistoryTool(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)

		// Verify all cleared
		finalHistory := server.store.GetStateHistory()
		assert.Equal(t, 0, len(finalHistory.GetAll()), "All history should be cleared atomically")
	})

	t.Run("event log clear is atomic", func(t *testing.T) {
		// Create test server
		cfg := DefaultMCPConfig()
		dt := devtools.Enable()
		server, err := NewMCPServer(cfg, dt)
		require.NoError(t, err)

		// Add events
		eventLog := server.store.GetEventLog()
		for i := 0; i < 50; i++ {
			eventLog.Append(devtools.EventRecord{
				ID:        "event-1",
				Name:      "click",
				SourceID:  "button-1",
				TargetID:  "handler-1",
				Payload:   nil,
				Timestamp: time.Now(),
				Duration:  time.Millisecond,
			})
		}

		// Clear should be atomic - either all cleared or none
		params := map[string]interface{}{
			"confirm": true,
		}
		request, _ := createCallToolRequest("clear_event_log", params)

		result, err := server.handleClearEventLogTool(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)

		// Verify all cleared
		finalEventLog := server.store.GetEventLog()
		assert.Equal(t, 0, finalEventLog.Len(), "All events should be cleared atomically")
	})
}
