package mcp

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	mcpSDK "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

// TestMCPServer_GetSDKServer tests the GetSDKServer method.
func TestMCPServer_GetSDKServer(t *testing.T) {
	dt := devtools.Enable()
	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Get SDK server
	sdkServer := server.GetSDKServer()

	// Verify it returns the underlying MCP SDK server
	assert.NotNil(t, sdkServer, "GetSDKServer should return non-nil MCP SDK server")
}

// TestMCPServer_GetSDKServer_Concurrent tests concurrent access to GetSDKServer.
func TestMCPServer_GetSDKServer_Concurrent(t *testing.T) {
	dt := devtools.Enable()
	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	const goroutines = 10
	const iterations = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				sdkServer := server.GetSDKServer()
				assert.NotNil(t, sdkServer)
			}
		}()
	}

	wg.Wait()
}

// TestStateChangeDetector_SetNotifier tests the SetNotifier method.
func TestStateChangeDetector_SetNotifier(t *testing.T) {
	sm := NewSubscriptionManager(50)
	detector := NewStateChangeDetector(sm)

	// Create a mock notifier
	mockNotifier := &mockNotificationSender{
		notifications: make([]mockNotification, 0),
	}

	// Set the notifier
	detector.SetNotifier(mockNotifier)

	// Verify notifier was set (indirectly through HandleRefChange)
	detector.subscriptions["client-1"] = []*Subscription{
		{
			ID:          "sub-1",
			ClientID:    "client-1",
			ResourceURI: "bubblyui://state/refs",
			Filters:     nil,
			CreatedAt:   time.Now(),
		},
	}

	// Trigger a ref change
	detector.HandleRefChange("ref-1", 10, 20)

	// Verify notification was queued
	assert.Len(t, mockNotifier.notifications, 1)
	assert.Equal(t, "client-1", mockNotifier.notifications[0].clientID)
	assert.Equal(t, "bubblyui://state/refs", mockNotifier.notifications[0].uri)
}

// TestStateChangeDetector_SetNotifier_ThreadSafe tests SetNotifier is thread-safe.
func TestStateChangeDetector_SetNotifier_ThreadSafe(t *testing.T) {
	sm := NewSubscriptionManager(50)
	detector := NewStateChangeDetector(sm)

	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer wg.Done()
			mockNotifier := &mockNotificationSender{
				notifications: make([]mockNotification, 0),
			}
			detector.SetNotifier(mockNotifier)
		}(i)
	}

	wg.Wait()
}

// TestStateChangeDetector_getAllSubscriptions tests subscription retrieval.
func TestStateChangeDetector_getAllSubscriptions(t *testing.T) {
	tests := []struct {
		name               string
		setupDetector      func() *StateChangeDetector
		expectedCount      int
		useSubscriptionMgr bool
	}{
		{
			name: "empty subscriptions",
			setupDetector: func() *StateChangeDetector {
				sm := NewSubscriptionManager(50)
				return NewStateChangeDetector(sm)
			},
			expectedCount: 0,
		},
		{
			name: "subscriptions from detector directly",
			setupDetector: func() *StateChangeDetector {
				sm := NewSubscriptionManager(50)
				detector := NewStateChangeDetector(sm)
				detector.subscriptions["client-1"] = []*Subscription{
					{ID: "sub-1", ClientID: "client-1"},
					{ID: "sub-2", ClientID: "client-1"},
				}
				detector.subscriptions["client-2"] = []*Subscription{
					{ID: "sub-3", ClientID: "client-2"},
				}
				return detector
			},
			expectedCount: 3,
		},
		{
			name: "subscriptions from manager",
			setupDetector: func() *StateChangeDetector {
				sm := NewSubscriptionManager(50)
				// Add subscriptions through manager
				_ = sm.Subscribe("client-1", "bubblyui://state/refs", nil)
				_ = sm.Subscribe("client-2", "bubblyui://components", nil)
				detector := NewStateChangeDetector(sm)
				return detector
			},
			expectedCount:      2,
			useSubscriptionMgr: true,
		},
		{
			name: "nil subscription manager",
			setupDetector: func() *StateChangeDetector {
				detector := NewStateChangeDetector(nil)
				return detector
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := tt.setupDetector()
			subs := detector.getAllSubscriptions()

			assert.Len(t, subs, tt.expectedCount)
		})
	}
}

// TestCheckTypeCompatibility_Extended tests additional type compatibility cases.
func TestCheckTypeCompatibility_Extended(t *testing.T) {
	tests := []struct {
		name       string
		oldValue   interface{}
		newValue   interface{}
		compatible bool
	}{
		{
			name:       "slice same type",
			oldValue:   []int{1, 2, 3},
			newValue:   []int{4, 5, 6},
			compatible: true,
		},
		{
			name:       "map same type",
			oldValue:   map[string]int{"a": 1},
			newValue:   map[string]int{"b": 2},
			compatible: true,
		},
		{
			name:       "struct same type",
			oldValue:   struct{ Name string }{Name: "old"},
			newValue:   struct{ Name string }{Name: "new"},
			compatible: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkTypeCompatibility(tt.oldValue, tt.newValue)
			assert.Equal(t, tt.compatible, result)
		})
	}
}

// TestLogRefModification tests ref modification logging.
func TestLogRefModification(t *testing.T) {
	dt := devtools.Enable()
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

	// Call logRefModification - should not panic even without error reporter
	server.logRefModification("ref-123", 10, 20, "comp-1")

	// If we reach here without panic, test passes
}

// TestHandleClearStateHistoryTool_InvalidJSON tests handling of invalid JSON.
func TestHandleClearStateHistoryTool_InvalidJSON(t *testing.T) {
	dt := devtools.Enable()
	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Create request with invalid JSON arguments
	request := &mcpSDK.CallToolRequest{
		Params: &mcpSDK.CallToolParamsRaw{
			Arguments: json.RawMessage(`{invalid json}`),
		},
	}

	result, err := server.handleClearStateHistoryTool(context.Background(), request)

	assert.NoError(t, err) // Handler returns error in result, not as error
	assert.NotNil(t, result)
	assert.True(t, result.IsError)
	assert.Len(t, result.Content, 1)
	textContent := result.Content[0].(*mcpSDK.TextContent)
	assert.Contains(t, textContent.Text, "Failed to parse parameters")
}

// TestHandleClearStateHistoryTool_MissingConfirm tests handling of missing confirm param.
func TestHandleClearStateHistoryTool_MissingConfirm(t *testing.T) {
	dt := devtools.Enable()
	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Create request without confirm param
	argsJSON, _ := json.Marshal(map[string]interface{}{})
	request := &mcpSDK.CallToolRequest{
		Params: &mcpSDK.CallToolParamsRaw{
			Arguments: argsJSON,
		},
	}

	result, err := server.handleClearStateHistoryTool(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsError)
	textContent := result.Content[0].(*mcpSDK.TextContent)
	assert.Contains(t, textContent.Text, "Invalid parameters")
}

// TestHandleClearEventLogTool_InvalidJSON tests handling of invalid JSON.
func TestHandleClearEventLogTool_InvalidJSON(t *testing.T) {
	dt := devtools.Enable()
	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Create request with invalid JSON arguments
	request := &mcpSDK.CallToolRequest{
		Params: &mcpSDK.CallToolParamsRaw{
			Arguments: json.RawMessage(`{invalid json}`),
		},
	}

	result, err := server.handleClearEventLogTool(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsError)
	textContent := result.Content[0].(*mcpSDK.TextContent)
	assert.Contains(t, textContent.Text, "Failed to parse parameters")
}

// TestHandleClearEventLogTool_MissingConfirm tests handling of missing confirm param.
func TestHandleClearEventLogTool_MissingConfirm(t *testing.T) {
	dt := devtools.Enable()
	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Create request without confirm param
	argsJSON, _ := json.Marshal(map[string]interface{}{})
	request := &mcpSDK.CallToolRequest{
		Params: &mcpSDK.CallToolParamsRaw{
			Arguments: argsJSON,
		},
	}

	result, err := server.handleClearEventLogTool(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsError)
	textContent := result.Content[0].(*mcpSDK.TextContent)
	assert.Contains(t, textContent.Text, "Invalid parameters")
}

// TestHandleSetRefValueTool_InvalidJSON tests handling of invalid JSON.
func TestHandleSetRefValueTool_InvalidJSON(t *testing.T) {
	dt := devtools.Enable()
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

	// Create request with invalid JSON arguments
	request := &mcpSDK.CallToolRequest{
		Params: &mcpSDK.CallToolParamsRaw{
			Arguments: json.RawMessage(`{not valid json`),
		},
	}

	result, err := server.handleSetRefValueTool(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsError)
	textContent := result.Content[0].(*mcpSDK.TextContent)
	assert.Contains(t, textContent.Text, "Failed to parse parameters")
}

// TestParseSetRefValueParams tests parameter parsing for set_ref_value.
func TestParseSetRefValueParams(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		wantError bool
		errMsg    string
	}{
		{
			name: "valid params",
			args: map[string]interface{}{
				"ref_id":    "ref-123",
				"new_value": 42,
				"dry_run":   true,
			},
			wantError: false,
		},
		{
			name: "missing ref_id",
			args: map[string]interface{}{
				"new_value": 42,
			},
			wantError: true,
			errMsg:    "ref_id parameter is required",
		},
		{
			name: "empty ref_id",
			args: map[string]interface{}{
				"ref_id":    "",
				"new_value": 42,
			},
			wantError: true,
			errMsg:    "ref_id cannot be empty",
		},
		{
			name: "missing new_value",
			args: map[string]interface{}{
				"ref_id": "ref-123",
			},
			wantError: true,
			errMsg:    "new_value parameter is required",
		},
		{
			name: "dry_run defaults to false",
			args: map[string]interface{}{
				"ref_id":    "ref-123",
				"new_value": 42,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := parseSetRefValueParams(tt.args)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, params)
			}
		})
	}
}

// TestParseClearStateHistoryParams tests parameter parsing.
func TestParseClearStateHistoryParams(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid confirm true",
			args:      map[string]interface{}{"confirm": true},
			wantError: false,
		},
		{
			name:      "valid confirm false",
			args:      map[string]interface{}{"confirm": false},
			wantError: false,
		},
		{
			name:      "missing confirm",
			args:      map[string]interface{}{},
			wantError: true,
			errMsg:    "confirm parameter is required",
		},
		{
			name:      "confirm wrong type",
			args:      map[string]interface{}{"confirm": "yes"},
			wantError: true,
			errMsg:    "confirm parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := parseClearStateHistoryParams(tt.args)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, params)
			}
		})
	}
}

// TestParseClearEventLogParams tests parameter parsing.
func TestParseClearEventLogParams(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid confirm true",
			args:      map[string]interface{}{"confirm": true},
			wantError: false,
		},
		{
			name:      "valid confirm false",
			args:      map[string]interface{}{"confirm": false},
			wantError: false,
		},
		{
			name:      "missing confirm",
			args:      map[string]interface{}{},
			wantError: true,
			errMsg:    "confirm parameter is required",
		},
		{
			name:      "confirm wrong type",
			args:      map[string]interface{}{"confirm": 1},
			wantError: true,
			errMsg:    "confirm parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := parseClearEventLogParams(tt.args)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, params)
			}
		})
	}
}

// TestParseExportParams tests export parameter parsing.
func TestParseExportParams(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		wantError bool
		errMsg    string
		checkFunc func(t *testing.T, params *ExportParams)
	}{
		{
			name: "all params provided",
			args: map[string]interface{}{
				"format":      "yaml",
				"compress":    true,
				"sanitize":    true,
				"include":     []interface{}{"components", "state"},
				"destination": "/tmp/export.yaml",
			},
			wantError: false,
			checkFunc: func(t *testing.T, params *ExportParams) {
				assert.Equal(t, "yaml", params.Format)
				assert.True(t, params.Compress)
				assert.True(t, params.Sanitize)
				assert.Equal(t, []string{"components", "state"}, params.Include)
				assert.Equal(t, "/tmp/export.yaml", params.Destination)
			},
		},
		{
			name: "defaults applied",
			args: map[string]interface{}{
				"destination": "/tmp/export.json",
			},
			wantError: false,
			checkFunc: func(t *testing.T, params *ExportParams) {
				assert.Equal(t, "json", params.Format)
				assert.False(t, params.Compress)
				assert.False(t, params.Sanitize)
				assert.Equal(t, []string{"components", "state", "events", "performance"}, params.Include)
			},
		},
		{
			name:      "missing destination",
			args:      map[string]interface{}{},
			wantError: true,
			errMsg:    "destination is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := parseExportParams(tt.args)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, params)
				if tt.checkFunc != nil {
					tt.checkFunc(t, params)
				}
			}
		})
	}
}

// TestValidateExportParams tests export parameter validation.
func TestValidateExportParams(t *testing.T) {
	tests := []struct {
		name      string
		params    *ExportParams
		wantError bool
		errMsg    string
	}{
		{
			name: "valid params",
			params: &ExportParams{
				Format:      "json",
				Destination: "/tmp/export.json",
				Include:     []string{"components"},
			},
			wantError: false,
		},
		{
			name: "invalid format",
			params: &ExportParams{
				Format:      "xml",
				Destination: "/tmp/export.xml",
				Include:     []string{"components"},
			},
			wantError: true,
			errMsg:    "invalid format",
		},
		{
			name: "empty destination",
			params: &ExportParams{
				Format:      "json",
				Destination: "",
				Include:     []string{"components"},
			},
			wantError: true,
			errMsg:    "destination cannot be empty",
		},
		{
			name: "empty include",
			params: &ExportParams{
				Format:      "json",
				Destination: "/tmp/export.json",
				Include:     []string{},
			},
			wantError: true,
			errMsg:    "include array cannot be empty",
		},
		{
			name: "invalid section in include",
			params: &ExportParams{
				Format:      "json",
				Destination: "/tmp/export.json",
				Include:     []string{"components", "invalid_section"},
			},
			wantError: true,
			errMsg:    "invalid section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExportParams(tt.params)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestContains tests the contains helper function.
func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "item exists",
			slice:    []string{"a", "b", "c"},
			item:     "b",
			expected: true,
		},
		{
			name:     "item not exists",
			slice:    []string{"a", "b", "c"},
			item:     "d",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			item:     "a",
			expected: false,
		},
		{
			name:     "empty item",
			slice:    []string{"a", "b", ""},
			item:     "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatSetRefResult tests result formatting.
func TestFormatSetRefResult(t *testing.T) {
	tests := []struct {
		name     string
		result   SetRefResult
		contains []string
	}{
		{
			name: "dry run result",
			result: SetRefResult{
				RefID:     "ref-123",
				OldValue:  10,
				NewValue:  20,
				OwnerID:   "comp-1",
				Timestamp: time.Now(),
				DryRun:    true,
				TypeMatch: true,
			},
			contains: []string{"Dry-run", "ref-123", "Current value: 10", "New value: 20", "No changes were applied"},
		},
		{
			name: "actual update result",
			result: SetRefResult{
				RefID:     "ref-456",
				OldValue:  "hello",
				NewValue:  "world",
				OwnerID:   "comp-2",
				Timestamp: time.Now(),
				DryRun:    false,
				TypeMatch: true,
			},
			contains: []string{"Successfully updated", "ref-456", "Old value: hello", "New value: world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := formatSetRefResult(tt.result)

			for _, expected := range tt.contains {
				assert.Contains(t, formatted, expected)
			}
		})
	}
}

// TestMatchesFilter_Extended tests additional filter matching scenarios.
func TestMatchesFilter_Extended(t *testing.T) {
	tests := []struct {
		name     string
		filters  map[string]interface{}
		data     map[string]interface{}
		expected bool
	}{
		{
			name:     "nil filters matches everything",
			filters:  nil,
			data:     map[string]interface{}{"key": "value"},
			expected: true,
		},
		{
			name:     "empty filters matches everything",
			filters:  map[string]interface{}{},
			data:     map[string]interface{}{"key": "value"},
			expected: true,
		},
		{
			name:     "exact match",
			filters:  map[string]interface{}{"ref_id": "ref-1"},
			data:     map[string]interface{}{"ref_id": "ref-1", "value": 42},
			expected: true,
		},
		{
			name:     "no match - different value",
			filters:  map[string]interface{}{"ref_id": "ref-1"},
			data:     map[string]interface{}{"ref_id": "ref-2", "value": 42},
			expected: false,
		},
		{
			name:     "no match - key not in data",
			filters:  map[string]interface{}{"ref_id": "ref-1"},
			data:     map[string]interface{}{"other_key": "value"},
			expected: false,
		},
		{
			name:     "multiple filters all match",
			filters:  map[string]interface{}{"ref_id": "ref-1", "component_id": "comp-1"},
			data:     map[string]interface{}{"ref_id": "ref-1", "component_id": "comp-1"},
			expected: true,
		},
		{
			name:     "multiple filters one fails",
			filters:  map[string]interface{}{"ref_id": "ref-1", "component_id": "comp-1"},
			data:     map[string]interface{}{"ref_id": "ref-1", "component_id": "comp-2"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesFilter(tt.filters, tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}
