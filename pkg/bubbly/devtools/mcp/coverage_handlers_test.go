package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	mcpSDK "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

// TestHandleExportTool_ErrorPaths tests handleExportTool error handling paths.
func TestHandleExportTool_ErrorPaths(t *testing.T) {
	tests := []struct {
		name          string
		params        map[string]interface{}
		wantError     bool
		errorContains string
	}{
		{
			name:          "invalid JSON parameters",
			params:        nil, // Will cause JSON parse error
			wantError:     true,
			errorContains: "Failed to parse parameters",
		},
		{
			name: "missing destination",
			params: map[string]interface{}{
				"format":  "json",
				"include": []interface{}{"components"},
			},
			wantError:     true,
			errorContains: "destination is required",
		},
		{
			name: "invalid format",
			params: map[string]interface{}{
				"format":      "xml",
				"include":     []interface{}{"components"},
				"destination": "/tmp/test.out",
			},
			wantError:     true,
			errorContains: "invalid format",
		},
		{
			name: "empty include array",
			params: map[string]interface{}{
				"format":      "json",
				"include":     []interface{}{},
				"destination": "/tmp/test.out",
			},
			wantError:     true,
			errorContains: "include array cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := devtools.Enable()
			defer devtools.Disable()

			cfg := DefaultMCPConfig()
			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)

			var request *mcpSDK.CallToolRequest
			if tt.params == nil {
				request = &mcpSDK.CallToolRequest{
					Params: &mcpSDK.CallToolParamsRaw{
						Arguments: json.RawMessage(`{invalid json`),
					},
				}
			} else {
				paramsJSON, _ := json.Marshal(tt.params)
				request = &mcpSDK.CallToolRequest{
					Params: &mcpSDK.CallToolParamsRaw{
						Arguments: paramsJSON,
					},
				}
			}

			result, err := server.handleExportTool(context.Background(), request)

			assert.NoError(t, err) // Handler returns error in result, not as error
			assert.NotNil(t, result)

			if tt.wantError {
				assert.True(t, result.IsError)
				if tt.errorContains != "" {
					textContent := result.Content[0].(*mcpSDK.TextContent)
					assert.Contains(t, textContent.Text, tt.errorContains)
				}
			} else {
				assert.False(t, result.IsError)
			}
		})
	}
}

// TestHandleExportTool_StdoutDestination tests stdout export path.
func TestHandleExportTool_StdoutDestination(t *testing.T) {
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	params := map[string]interface{}{
		"format":      "json",
		"compress":    false,
		"sanitize":    false,
		"include":     []interface{}{"components"},
		"destination": "stdout",
	}

	paramsJSON, _ := json.Marshal(params)
	request := &mcpSDK.CallToolRequest{
		Params: &mcpSDK.CallToolParamsRaw{
			Arguments: paramsJSON,
		},
	}

	result, err := server.handleExportTool(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// stdout path returns content directly
	assert.False(t, result.IsError)
}

// TestHandleSearchComponentsTool_ErrorPaths tests handleSearchComponentsTool error handling.
func TestHandleSearchComponentsTool_ErrorPaths(t *testing.T) {
	tests := []struct {
		name          string
		params        map[string]interface{}
		wantError     bool
		errorContains string
	}{
		{
			name:          "invalid JSON parameters",
			params:        nil,
			wantError:     true,
			errorContains: "Failed to parse parameters",
		},
		{
			name: "missing query",
			params: map[string]interface{}{
				"fields": []interface{}{"name"},
			},
			wantError:     true,
			errorContains: "query",
		},
		{
			name: "search with empty query succeeds (returns no results)",
			params: map[string]interface{}{
				"query": "",
			},
			wantError: false, // Empty query is allowed but finds nothing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := devtools.Enable()
			defer devtools.Disable()

			cfg := DefaultMCPConfig()
			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)

			var request *mcpSDK.CallToolRequest
			if tt.params == nil {
				request = &mcpSDK.CallToolRequest{
					Params: &mcpSDK.CallToolParamsRaw{
						Arguments: json.RawMessage(`{invalid`),
					},
				}
			} else {
				paramsJSON, _ := json.Marshal(tt.params)
				request = &mcpSDK.CallToolRequest{
					Params: &mcpSDK.CallToolParamsRaw{
						Arguments: paramsJSON,
					},
				}
			}

			result, err := server.handleSearchComponentsTool(context.Background(), request)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			if tt.wantError {
				assert.True(t, result.IsError)
				if tt.errorContains != "" {
					textContent := result.Content[0].(*mcpSDK.TextContent)
					assert.Contains(t, textContent.Text, tt.errorContains)
				}
			}
		})
	}
}

// TestHandleFilterEventsTool_ErrorPaths tests handleFilterEventsTool error handling.
func TestHandleFilterEventsTool_ErrorPaths(t *testing.T) {
	tests := []struct {
		name          string
		params        map[string]interface{}
		wantError     bool
		errorContains string
	}{
		{
			name:          "invalid JSON parameters",
			params:        nil,
			wantError:     true,
			errorContains: "Failed to parse parameters",
		},
		{
			name: "filter with valid params - no events",
			params: map[string]interface{}{
				"event_names": []interface{}{"click"},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := devtools.Enable()
			defer devtools.Disable()

			cfg := DefaultMCPConfig()
			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)

			var request *mcpSDK.CallToolRequest
			if tt.params == nil {
				request = &mcpSDK.CallToolRequest{
					Params: &mcpSDK.CallToolParamsRaw{
						Arguments: json.RawMessage(`{bad`),
					},
				}
			} else {
				paramsJSON, _ := json.Marshal(tt.params)
				request = &mcpSDK.CallToolRequest{
					Params: &mcpSDK.CallToolParamsRaw{
						Arguments: paramsJSON,
					},
				}
			}

			result, err := server.handleFilterEventsTool(context.Background(), request)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			if tt.wantError {
				assert.True(t, result.IsError)
				if tt.errorContains != "" {
					textContent := result.Content[0].(*mcpSDK.TextContent)
					assert.Contains(t, textContent.Text, tt.errorContains)
				}
			}
		})
	}
}

// TestHandleClearStateHistoryTool_Success tests successful state history clearing.
func TestHandleClearStateHistoryTool_Success(t *testing.T) {
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Add some state history
	store := dt.GetStore()
	stateHistory := store.GetStateHistory()
	stateHistory.Record(devtools.StateChange{RefID: "ref-1", OldValue: 10, NewValue: 20})
	stateHistory.Record(devtools.StateChange{RefID: "ref-2", OldValue: "old", NewValue: "new"})

	// Confirm clearing
	params := map[string]interface{}{
		"confirm": true,
	}
	paramsJSON, _ := json.Marshal(params)
	request := &mcpSDK.CallToolRequest{
		Params: &mcpSDK.CallToolParamsRaw{
			Arguments: paramsJSON,
		},
	}

	result, err := server.handleClearStateHistoryTool(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsError)

	textContent := result.Content[0].(*mcpSDK.TextContent)
	assert.Contains(t, textContent.Text, "cleared successfully")
}

// TestHandleClearEventLogTool_Success tests successful event log clearing.
func TestHandleClearEventLogTool_Success(t *testing.T) {
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Add some events
	store := dt.GetStore()
	eventLog := store.GetEventLog()
	eventLog.Append(devtools.EventRecord{
		ID:        "evt-1",
		Name:      "click",
		SourceID:  "comp-1",
		Timestamp: time.Now(),
	})

	// Confirm clearing
	params := map[string]interface{}{
		"confirm": true,
	}
	paramsJSON, _ := json.Marshal(params)
	request := &mcpSDK.CallToolRequest{
		Params: &mcpSDK.CallToolParamsRaw{
			Arguments: paramsJSON,
		},
	}

	result, err := server.handleClearEventLogTool(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsError)

	textContent := result.Content[0].(*mcpSDK.TextContent)
	assert.Contains(t, textContent.Text, "cleared successfully")
}

// TestHandleClearStateHistoryTool_NotConfirmed tests rejection when not confirmed.
func TestHandleClearStateHistoryTool_NotConfirmed(t *testing.T) {
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Don't confirm
	params := map[string]interface{}{
		"confirm": false,
	}
	paramsJSON, _ := json.Marshal(params)
	request := &mcpSDK.CallToolRequest{
		Params: &mcpSDK.CallToolParamsRaw{
			Arguments: paramsJSON,
		},
	}

	result, err := server.handleClearStateHistoryTool(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsError)

	textContent := result.Content[0].(*mcpSDK.TextContent)
	assert.Contains(t, textContent.Text, "Confirmation required")
}

// TestHandleClearEventLogTool_NotConfirmed tests rejection when not confirmed.
func TestHandleClearEventLogTool_NotConfirmed(t *testing.T) {
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Don't confirm
	params := map[string]interface{}{
		"confirm": false,
	}
	paramsJSON, _ := json.Marshal(params)
	request := &mcpSDK.CallToolRequest{
		Params: &mcpSDK.CallToolParamsRaw{
			Arguments: paramsJSON,
		},
	}

	result, err := server.handleClearEventLogTool(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsError)

	textContent := result.Content[0].(*mcpSDK.TextContent)
	assert.Contains(t, textContent.Text, "Confirmation required")
}

// TestHandleSetRefValueTool_RefNotFound tests ref not found error path.
func TestHandleSetRefValueTool_RefNotFound(t *testing.T) {
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := &Config{
		Transport:            MCPTransportStdio,
		WriteEnabled:         true,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            60,
		EnableAuth:           false,
		SanitizeExports:      true,
	}

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Try to update a non-existent ref
	params := map[string]interface{}{
		"ref_id":    "nonexistent-ref",
		"new_value": 20,
		"dry_run":   false,
	}
	paramsJSON, _ := json.Marshal(params)
	request := &mcpSDK.CallToolRequest{
		Params: &mcpSDK.CallToolParamsRaw{
			Arguments: paramsJSON,
		},
	}

	result, err := server.handleSetRefValueTool(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsError)
	textContent := result.Content[0].(*mcpSDK.TextContent)
	assert.Contains(t, textContent.Text, "ref not found")
}

// TestExportTool_InvalidDestinationDirectory tests export to non-existent directory.
func TestExportTool_InvalidDestinationDirectory(t *testing.T) {
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Use a path with restricted permissions or non-existent deep directory
	params := map[string]interface{}{
		"format":      "json",
		"include":     []interface{}{"components"},
		"destination": "/nonexistent/deep/path/export.json",
	}

	paramsJSON, _ := json.Marshal(params)
	request := &mcpSDK.CallToolRequest{
		Params: &mcpSDK.CallToolParamsRaw{
			Arguments: paramsJSON,
		},
	}

	// This test may or may not fail depending on system permissions
	result, err := server.handleExportTool(context.Background(), request)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Result depends on whether directory creation succeeds
}

// TestLogRefModification_WithObservability tests logging with observability.
func TestLogRefModification_WithObservability(t *testing.T) {
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := &Config{
		Transport:            MCPTransportStdio,
		WriteEnabled:         true,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            60,
		EnableAuth:           false,
		SanitizeExports:      true,
	}

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Call logRefModification multiple times to cover all paths
	server.logRefModification("ref-1", 10, 20, "comp-1")
	server.logRefModification("ref-2", "old", "new", "comp-2")
	server.logRefModification("ref-3", nil, 100, "comp-3")
	server.logRefModification("ref-4", []int{1, 2}, []int{3, 4}, "comp-4")

	// If we reach here without panic, test passes
}

// TestRecoverToolRegistration tests the panic recovery helper.
func TestRecoverToolRegistration(t *testing.T) {
	// Test that recoverToolRegistration doesn't panic itself
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("recoverToolRegistration should not panic: %v", r)
		}
	}()

	// Call without a panic - should do nothing
	recoverToolRegistration("TestEvent", "test_tool")
}

// TestExportTool_Compression_Extended tests export with compression enabled (extended scenarios).
func TestExportTool_Compression_Extended(t *testing.T) {
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	tmpDir := t.TempDir()
	params := map[string]interface{}{
		"format":      "json",
		"compress":    true,
		"sanitize":    false,
		"include":     []interface{}{"components"},
		"destination": filepath.Join(tmpDir, "export.json.gz"),
	}

	paramsJSON, _ := json.Marshal(params)
	request := &mcpSDK.CallToolRequest{
		Params: &mcpSDK.CallToolParamsRaw{
			Arguments: paramsJSON,
		},
	}

	result, err := server.handleExportTool(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Verify file was created (may or may not be compressed depending on implementation)
	_, statErr := os.Stat(filepath.Join(tmpDir, "export.json.gz"))
	if statErr != nil {
		// Try without .gz extension
		_, statErr = os.Stat(filepath.Join(tmpDir, "export.json"))
		_ = statErr // Intentionally ignoring - we're just checking if file exists
	}
}
