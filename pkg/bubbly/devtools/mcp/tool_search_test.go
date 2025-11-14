package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
