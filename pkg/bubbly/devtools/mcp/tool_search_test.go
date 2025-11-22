package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

// TestRegisterSearchComponentsTool tests the search_components tool registration.
func TestRegisterSearchComponentsTool(t *testing.T) {
	dt := devtools.Enable()
	cfg := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            60,
		EnableAuth:           false,
		SanitizeExports:      true,
	}

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)
	require.NotNil(t, server)

	err = server.RegisterSearchComponentsTool()
	assert.NoError(t, err)
}

// TestSearchComponents_ByName tests searching components by name.
func TestSearchComponents_ByName(t *testing.T) {
	dt := devtools.Enable()
	cfg := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            60,
		EnableAuth:           false,
		SanitizeExports:      true,
	}

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterSearchComponentsTool()
	require.NoError(t, err)

	// Add test components
	store := dt.GetStore()
	store.AddComponent(&devtools.ComponentSnapshot{
		ID:     "comp-1",
		Name:   "Counter",
		Type:   "Counter",
		Status: "mounted",
	})
	store.AddComponent(&devtools.ComponentSnapshot{
		ID:     "comp-2",
		Name:   "TodoList",
		Type:   "TodoList",
		Status: "mounted",
	})
	store.AddComponent(&devtools.ComponentSnapshot{
		ID:     "comp-3",
		Name:   "CounterButton",
		Type:   "Button",
		Status: "mounted",
	})

	// Create tool call request
	params := map[string]interface{}{
		"query":  "counter",
		"fields": []interface{}{"name"},
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	request := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "search_components",
			Arguments: paramsJSON,
		},
	}

	// Call the tool
	result, err := server.handleSearchComponentsTool(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	// Verify results
	resultText := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, resultText, "Counter")
	assert.Contains(t, resultText, "CounterButton")
	assert.NotContains(t, resultText, "TodoList")
}

// TestSearchComponents_NoMatches tests searching with no matches.
func TestSearchComponents_NoMatches(t *testing.T) {
	dt := devtools.Enable()
	cfg := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            60,
		EnableAuth:           false,
		SanitizeExports:      true,
	}

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterSearchComponentsTool()
	require.NoError(t, err)

	// Add test component
	store := dt.GetStore()
	store.AddComponent(&devtools.ComponentSnapshot{
		ID:     "comp-1",
		Name:   "Counter",
		Type:   "Counter",
		Status: "mounted",
	})

	// Create tool call request with non-matching query
	params := map[string]interface{}{
		"query": "nonexistent",
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	request := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "search_components",
			Arguments: paramsJSON,
		},
	}

	// Call the tool
	result, err := server.handleSearchComponentsTool(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	// Verify no matches message
	resultText := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, resultText, "No components found")
}

// TestRegisterFilterEventsTool tests the filter_events tool registration.
func TestRegisterFilterEventsTool(t *testing.T) {
	dt := devtools.Enable()
	cfg := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            60,
		EnableAuth:           false,
		SanitizeExports:      true,
	}

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)
	require.NotNil(t, server)

	err = server.RegisterFilterEventsTool()
	assert.NoError(t, err)
}

// TestFilterEvents_ByName tests filtering events by name.
func TestFilterEvents_ByName(t *testing.T) {
	dt := devtools.Enable()
	cfg := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            60,
		EnableAuth:           false,
		SanitizeExports:      true,
	}

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterFilterEventsTool()
	require.NoError(t, err)

	// Add test events
	store := dt.GetStore()
	eventLog := store.GetEventLog()
	eventLog.Append(devtools.EventRecord{
		ID:        "event-1",
		Name:      "click",
		SourceID:  "comp-1",
		Timestamp: time.Now(),
	})
	eventLog.Append(devtools.EventRecord{
		ID:        "event-2",
		Name:      "submit",
		SourceID:  "comp-2",
		Timestamp: time.Now(),
	})
	eventLog.Append(devtools.EventRecord{
		ID:        "event-3",
		Name:      "click",
		SourceID:  "comp-3",
		Timestamp: time.Now(),
	})

	// Create tool call request
	params := map[string]interface{}{
		"event_names": []interface{}{"click"},
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	request := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "filter_events",
			Arguments: paramsJSON,
		},
	}

	// Call the tool
	result, err := server.handleFilterEventsTool(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	// Verify results
	resultText := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, resultText, "Found 2 event(s)")
	assert.Contains(t, resultText, "click")
	assert.NotContains(t, resultText, "submit")
}

// TestFilterEvents_NoMatches tests filtering with no matches.
func TestFilterEvents_NoMatches(t *testing.T) {
	dt := devtools.Enable()
	cfg := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            60,
		EnableAuth:           false,
		SanitizeExports:      true,
	}

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterFilterEventsTool()
	require.NoError(t, err)

	// Add test event
	store := dt.GetStore()
	eventLog := store.GetEventLog()
	eventLog.Append(devtools.EventRecord{
		ID:        "event-1",
		Name:      "click",
		SourceID:  "comp-1",
		Timestamp: time.Now(),
	})

	// Create tool call request with non-matching filter
	params := map[string]interface{}{
		"event_names": []interface{}{"nonexistent"},
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	request := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "filter_events",
			Arguments: paramsJSON,
		},
	}

	// Call the tool
	result, err := server.handleFilterEventsTool(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	// Verify no matches message
	resultText := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, resultText, "No events found")
}

// TestSearchComponents_Comprehensive tests all search scenarios with table-driven tests.
func TestSearchComponents_Comprehensive(t *testing.T) {
	tests := []struct {
		name          string
		setupData     func(*devtools.DevToolsStore)
		params        map[string]interface{}
		wantError     bool
		wantContains  []string
		wantExcludes  []string
		errorContains string
	}{
		{
			name: "search by type field",
			setupData: func(store *devtools.DevToolsStore) {
				store.AddComponent(&devtools.ComponentSnapshot{
					ID: "comp-1", Name: "SubmitButton", Type: "Button", Status: "mounted",
				})
				store.AddComponent(&devtools.ComponentSnapshot{
					ID: "comp-2", Name: "CancelButton", Type: "Button", Status: "mounted",
				})
				store.AddComponent(&devtools.ComponentSnapshot{
					ID: "comp-3", Name: "TodoList", Type: "List", Status: "mounted",
				})
			},
			params: map[string]interface{}{
				"query":  "button",
				"fields": []interface{}{"type"},
			},
			wantContains: []string{"SubmitButton", "CancelButton"},
			wantExcludes: []string{"TodoList"},
		},
		{
			name: "search by ID field",
			setupData: func(store *devtools.DevToolsStore) {
				store.AddComponent(&devtools.ComponentSnapshot{
					ID: "test-comp-1", Name: "Counter", Type: "Counter", Status: "mounted",
				})
				store.AddComponent(&devtools.ComponentSnapshot{
					ID: "comp-2", Name: "TestComponent", Type: "Component", Status: "mounted",
				})
			},
			params: map[string]interface{}{
				"query":  "test",
				"fields": []interface{}{"id"},
			},
			wantContains: []string{"test-comp-1"},
			wantExcludes: []string{"TestComponent"},
		},
		{
			name: "search all fields (no fields specified)",
			setupData: func(store *devtools.DevToolsStore) {
				store.AddComponent(&devtools.ComponentSnapshot{
					ID: "comp-test-1", Name: "Counter", Type: "Counter", Status: "mounted",
				})
				store.AddComponent(&devtools.ComponentSnapshot{
					ID: "comp-2", Name: "TestComponent", Type: "Component", Status: "mounted",
				})
			},
			params: map[string]interface{}{
				"query": "test",
			},
			wantContains: []string{"comp-test-1", "TestComponent"},
		},
		{
			name: "max_results limit enforced",
			setupData: func(store *devtools.DevToolsStore) {
				for i := 0; i < 20; i++ {
					store.AddComponent(&devtools.ComponentSnapshot{
						ID: fmt.Sprintf("comp-%d", i), Name: "Component", Type: "Component", Status: "mounted",
					})
				}
			},
			params: map[string]interface{}{
				"query":       "component",
				"max_results": 5,
			},
			wantContains: []string{"Found 5 component(s)"},
		},
		{
			name: "exact match scores highest",
			setupData: func(store *devtools.DevToolsStore) {
				store.AddComponent(&devtools.ComponentSnapshot{
					ID: "comp-1", Name: "Counter", Type: "Counter", Status: "mounted",
				})
				store.AddComponent(&devtools.ComponentSnapshot{
					ID: "comp-2", Name: "CounterButton", Type: "Button", Status: "mounted",
				})
			},
			params: map[string]interface{}{
				"query": "counter",
			},
			wantContains: []string{"Counter", "CounterButton"},
		},
		{
			name: "case insensitive search",
			setupData: func(store *devtools.DevToolsStore) {
				store.AddComponent(&devtools.ComponentSnapshot{
					ID: "comp-1", Name: "UPPERCASE", Type: "Component", Status: "mounted",
				})
			},
			params: map[string]interface{}{
				"query": "uppercase",
			},
			wantContains: []string{"UPPERCASE"},
		},
		{
			name: "invalid JSON parameters",
			setupData: func(store *devtools.DevToolsStore) {
				// No setup needed
			},
			params:        nil, // Will cause JSON unmarshal error
			wantError:     true,
			errorContains: "Failed to parse parameters",
		},
		{
			name: "missing query parameter",
			setupData: func(store *devtools.DevToolsStore) {
				// No setup needed
			},
			params: map[string]interface{}{
				"fields": []interface{}{"name"},
			},
			wantError:     true,
			errorContains: "query parameter is required",
		},
		{
			name: "max_results too low",
			setupData: func(store *devtools.DevToolsStore) {
				// No setup needed
			},
			params: map[string]interface{}{
				"query":       "test",
				"max_results": 0,
			},
			wantError:     true,
			errorContains: "max_results must be between 1 and 1000",
		},
		{
			name: "max_results too high",
			setupData: func(store *devtools.DevToolsStore) {
				// No setup needed
			},
			params: map[string]interface{}{
				"query":       "test",
				"max_results": 2000,
			},
			wantError:     true,
			errorContains: "max_results must be between 1 and 1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := devtools.Enable()
			cfg := DefaultMCPConfig()
			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)

			err = server.RegisterSearchComponentsTool()
			require.NoError(t, err)

			// Setup test data
			if tt.setupData != nil {
				tt.setupData(dt.GetStore())
			}

			// Create request
			var request *mcp.CallToolRequest
			if tt.params == nil {
				// Invalid JSON case
				request = &mcp.CallToolRequest{
					Params: &mcp.CallToolParamsRaw{
						Name:      "search_components",
						Arguments: []byte("invalid json"),
					},
				}
			} else {
				paramsJSON, err := json.Marshal(tt.params)
				require.NoError(t, err)
				request = &mcp.CallToolRequest{
					Params: &mcp.CallToolParamsRaw{
						Name:      "search_components",
						Arguments: paramsJSON,
					},
				}
			}

			// Call tool
			result, err := server.handleSearchComponentsTool(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Verify results
			if tt.wantError {
				assert.True(t, result.IsError, "Expected error result")
				if tt.errorContains != "" {
					textContent := result.Content[0].(*mcp.TextContent)
					assert.Contains(t, textContent.Text, tt.errorContains)
				}
			} else {
				assert.False(t, result.IsError, "Expected success result")
				textContent := result.Content[0].(*mcp.TextContent)
				for _, want := range tt.wantContains {
					assert.Contains(t, textContent.Text, want)
				}
				for _, exclude := range tt.wantExcludes {
					assert.NotContains(t, textContent.Text, exclude)
				}
			}
		})
	}
}

// TestFilterEvents_Comprehensive tests all filter scenarios with table-driven tests.
func TestFilterEvents_Comprehensive(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	tests := []struct {
		name          string
		setupData     func(*devtools.DevToolsStore)
		params        map[string]interface{}
		wantError     bool
		wantCount     int
		wantContains  []string
		wantExcludes  []string
		errorContains string
	}{
		{
			name: "filter by source ID",
			setupData: func(store *devtools.DevToolsStore) {
				eventLog := store.GetEventLog()
				eventLog.Append(devtools.EventRecord{
					ID: "event-1", Name: "click", SourceID: "comp-1", Timestamp: now,
				})
				eventLog.Append(devtools.EventRecord{
					ID: "event-2", Name: "submit", SourceID: "comp-1", Timestamp: now,
				})
				eventLog.Append(devtools.EventRecord{
					ID: "event-3", Name: "click", SourceID: "comp-2", Timestamp: now,
				})
			},
			params: map[string]interface{}{
				"source_ids": []interface{}{"comp-1"},
			},
			wantCount:    2,
			wantContains: []string{"comp-1"},
			wantExcludes: []string{"comp-2"},
		},
		{
			name: "filter by time range",
			setupData: func(store *devtools.DevToolsStore) {
				eventLog := store.GetEventLog()
				eventLog.Append(devtools.EventRecord{
					ID: "event-1", Name: "old-event", SourceID: "comp-1", Timestamp: past,
				})
				eventLog.Append(devtools.EventRecord{
					ID: "event-2", Name: "current-event", SourceID: "comp-2", Timestamp: now,
				})
				eventLog.Append(devtools.EventRecord{
					ID: "event-3", Name: "future-event", SourceID: "comp-3", Timestamp: future,
				})
			},
			params: map[string]interface{}{
				"start_time": now.Add(-30 * time.Minute).Format(time.RFC3339),
			},
			wantCount:    2,
			wantContains: []string{"current-event", "future-event"},
			wantExcludes: []string{"old-event"},
		},
		{
			name: "filter by end time",
			setupData: func(store *devtools.DevToolsStore) {
				eventLog := store.GetEventLog()
				eventLog.Append(devtools.EventRecord{
					ID: "event-1", Name: "past", SourceID: "comp-1", Timestamp: past,
				})
				eventLog.Append(devtools.EventRecord{
					ID: "event-2", Name: "future", SourceID: "comp-2", Timestamp: future,
				})
			},
			params: map[string]interface{}{
				"end_time": now.Add(30 * time.Minute).Format(time.RFC3339),
			},
			wantCount:    1,
			wantContains: []string{"past"},
			wantExcludes: []string{"future"},
		},
		{
			name: "filter with limit",
			setupData: func(store *devtools.DevToolsStore) {
				eventLog := store.GetEventLog()
				for i := 0; i < 20; i++ {
					eventLog.Append(devtools.EventRecord{
						ID: fmt.Sprintf("event-%d", i), Name: "click", SourceID: "comp-1", Timestamp: now,
					})
				}
			},
			params: map[string]interface{}{
				"limit": 5,
			},
			wantCount:    5,
			wantContains: []string{"Found 5 event(s)"},
		},
		{
			name: "filter by multiple event names",
			setupData: func(store *devtools.DevToolsStore) {
				eventLog := store.GetEventLog()
				eventLog.Append(devtools.EventRecord{
					ID: "event-1", Name: "click", SourceID: "comp-1", Timestamp: now,
				})
				eventLog.Append(devtools.EventRecord{
					ID: "event-2", Name: "submit", SourceID: "comp-2", Timestamp: now,
				})
				eventLog.Append(devtools.EventRecord{
					ID: "event-3", Name: "change", SourceID: "comp-3", Timestamp: now,
				})
			},
			params: map[string]interface{}{
				"event_names": []interface{}{"click", "submit"},
			},
			wantCount:    2,
			wantContains: []string{"click", "submit"},
			wantExcludes: []string{"change"},
		},
		{
			name: "combined filters",
			setupData: func(store *devtools.DevToolsStore) {
				eventLog := store.GetEventLog()
				eventLog.Append(devtools.EventRecord{
					ID: "event-1", Name: "click", SourceID: "comp-1", Timestamp: now,
				})
				eventLog.Append(devtools.EventRecord{
					ID: "event-2", Name: "click", SourceID: "comp-2", Timestamp: now,
				})
				eventLog.Append(devtools.EventRecord{
					ID: "event-3", Name: "submit", SourceID: "comp-1", Timestamp: now,
				})
			},
			params: map[string]interface{}{
				"event_names": []interface{}{"click"},
				"source_ids":  []interface{}{"comp-1"},
			},
			wantCount:    1,
			wantContains: []string{"event-1"},
		},
		{
			name: "invalid JSON parameters",
			setupData: func(store *devtools.DevToolsStore) {
				// No setup needed
			},
			params:        nil, // Will cause JSON unmarshal error
			wantError:     true,
			errorContains: "Failed to parse parameters",
		},
		{
			name: "invalid limit too low",
			setupData: func(store *devtools.DevToolsStore) {
				// No setup needed
			},
			params: map[string]interface{}{
				"limit": 0,
			},
			wantError:     true,
			errorContains: "limit must be between 1 and 10000",
		},
		{
			name: "invalid limit too high",
			setupData: func(store *devtools.DevToolsStore) {
				// No setup needed
			},
			params: map[string]interface{}{
				"limit": 20000,
			},
			wantError:     true,
			errorContains: "limit must be between 1 and 10000",
		},
		{
			name: "invalid start_time format",
			setupData: func(store *devtools.DevToolsStore) {
				// No setup needed
			},
			params: map[string]interface{}{
				"start_time": "not-a-time",
			},
			wantError:     true,
			errorContains: "invalid start_time format",
		},
		{
			name: "invalid end_time format",
			setupData: func(store *devtools.DevToolsStore) {
				// No setup needed
			},
			params: map[string]interface{}{
				"end_time": "not-a-time",
			},
			wantError:     true,
			errorContains: "invalid end_time format",
		},
		{
			name: "event with target ID and duration",
			setupData: func(store *devtools.DevToolsStore) {
				eventLog := store.GetEventLog()
				eventLog.Append(devtools.EventRecord{
					ID:        "event-1",
					Name:      "click",
					SourceID:  "comp-1",
					TargetID:  "handler-1",
					Timestamp: now,
					Duration:  5 * time.Millisecond,
				})
			},
			params: map[string]interface{}{
				"event_names": []interface{}{"click"},
			},
			wantCount:    1,
			wantContains: []string{"Target: handler-1", "Duration: 5ms"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := devtools.Enable()
			cfg := DefaultMCPConfig()
			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)

			err = server.RegisterFilterEventsTool()
			require.NoError(t, err)

			// Setup test data
			if tt.setupData != nil {
				tt.setupData(dt.GetStore())
			}

			// Create request
			var request *mcp.CallToolRequest
			if tt.params == nil {
				// Invalid JSON case
				request = &mcp.CallToolRequest{
					Params: &mcp.CallToolParamsRaw{
						Name:      "filter_events",
						Arguments: []byte("invalid json"),
					},
				}
			} else {
				paramsJSON, err := json.Marshal(tt.params)
				require.NoError(t, err)
				request = &mcp.CallToolRequest{
					Params: &mcp.CallToolParamsRaw{
						Name:      "filter_events",
						Arguments: paramsJSON,
					},
				}
			}

			// Call tool
			result, err := server.handleFilterEventsTool(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Verify results
			if tt.wantError {
				assert.True(t, result.IsError, "Expected error result")
				if tt.errorContains != "" {
					textContent := result.Content[0].(*mcp.TextContent)
					assert.Contains(t, textContent.Text, tt.errorContains)
				}
			} else {
				assert.False(t, result.IsError, "Expected success result")
				textContent := result.Content[0].(*mcp.TextContent)
				for _, want := range tt.wantContains {
					assert.Contains(t, textContent.Text, want)
				}
				for _, exclude := range tt.wantExcludes {
					assert.NotContains(t, textContent.Text, exclude)
				}
			}
		})
	}
}

// TestCalculateMatchScore tests the match scoring algorithm.
func TestCalculateMatchScore(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		query     string
		wantScore float64
	}{
		{"exact match", "counter", "counter", 1.0},
		{"exact match case insensitive", "Counter", "counter", 1.0},
		{"starts with", "CounterButton", "counter", 0.9},
		{"contains at start", "counter-button", "counter", 0.9},
		{"contains in middle", "MyCounterComponent", "counter", 0.7}, // Approximate
		{"no match", "TodoList", "counter", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateMatchScore(tt.text, tt.query)
			if tt.wantScore == 0.0 {
				assert.Equal(t, 0.0, score)
			} else if tt.wantScore == 1.0 || tt.wantScore == 0.9 {
				assert.Equal(t, tt.wantScore, score)
			} else {
				// For contains matches, just verify it's in the expected range
				assert.Greater(t, score, 0.5)
				assert.Less(t, score, 0.9)
			}
		})
	}
}
