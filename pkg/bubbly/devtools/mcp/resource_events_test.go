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

// TestRegisterEventsResource_Success tests successful registration of events resources.
func TestRegisterEventsResource_Success(t *testing.T) {
	// Create MCP server with devtools
	dt := devtools.Enable()
	defer devtools.Disable()

	config := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            100,
	}

	server, err := NewMCPServer(config, dt)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Register events resource
	err = server.RegisterEventsResource()
	assert.NoError(t, err, "RegisterEventsResource should succeed")
}

// TestEventsResource_Empty tests reading events when no events exist.
func TestEventsResource_Empty(t *testing.T) {
	// Create MCP server with devtools
	dt := devtools.Enable()
	defer devtools.Disable()

	// Clear any residual state
	dt.GetStore().GetEventLog().Clear()

	config := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            100,
	}

	server, err := NewMCPServer(config, dt)
	require.NoError(t, err)

	// Register events resource
	err = server.RegisterEventsResource()
	require.NoError(t, err)

	// Read events log resource
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://events/log",
		},
	}

	result, err := server.readEventsLogResource(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Contents, 1)

	// Parse JSON response
	var resource EventsResource
	err = json.Unmarshal([]byte(result.Contents[0].Text), &resource)
	require.NoError(t, err)

	// Verify empty events
	assert.Empty(t, resource.Events, "Should have no events initially")
	assert.Equal(t, 0, resource.TotalCount, "Total count should be 0")
	assert.NotZero(t, resource.Timestamp, "Should have timestamp")
}

// TestEventsResource_WithData tests reading events with actual data.
func TestEventsResource_WithData(t *testing.T) {
	// Create MCP server with devtools
	dt := devtools.Enable()
	defer devtools.Disable()

	store := dt.GetStore()
	eventLog := store.GetEventLog()

	// Add events
	eventLog.Append(devtools.EventRecord{
		ID:        "event-1",
		Name:      "click",
		SourceID:  "comp-1",
		TargetID:  "comp-2",
		Payload:   map[string]interface{}{"button": "submit"},
		Timestamp: time.Now(),
		Duration:  5 * time.Millisecond,
	})

	eventLog.Append(devtools.EventRecord{
		ID:        "event-2",
		Name:      "change",
		SourceID:  "comp-3",
		TargetID:  "comp-3",
		Payload:   map[string]interface{}{"value": "test"},
		Timestamp: time.Now(),
		Duration:  2 * time.Millisecond,
	})

	config := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            100,
	}

	server, err := NewMCPServer(config, dt)
	require.NoError(t, err)

	// Register events resource
	err = server.RegisterEventsResource()
	require.NoError(t, err)

	// Read events log resource
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://events/log",
		},
	}

	result, err := server.readEventsLogResource(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Parse JSON response
	var resource EventsResource
	err = json.Unmarshal([]byte(result.Contents[0].Text), &resource)
	require.NoError(t, err)

	// Verify events
	assert.Len(t, resource.Events, 2, "Should have 2 events")
	assert.Equal(t, 2, resource.TotalCount, "Total count should be 2")
	assert.NotZero(t, resource.Timestamp, "Should have timestamp")

	// Verify event details
	eventMap := make(map[string]devtools.EventRecord)
	for _, event := range resource.Events {
		eventMap[event.ID] = event
	}

	assert.Contains(t, eventMap, "event-1")
	assert.Equal(t, "click", eventMap["event-1"].Name)
	assert.Equal(t, "comp-1", eventMap["event-1"].SourceID)
	assert.Equal(t, "comp-2", eventMap["event-1"].TargetID)

	assert.Contains(t, eventMap, "event-2")
	assert.Equal(t, "change", eventMap["event-2"].Name)
	assert.Equal(t, "comp-3", eventMap["event-2"].SourceID)
}

// TestEventResource_ByID tests reading a single event by ID.
func TestEventResource_ByID(t *testing.T) {
	// Create MCP server with devtools
	dt := devtools.Enable()
	defer devtools.Disable()

	store := dt.GetStore()
	eventLog := store.GetEventLog()

	// Add event
	eventLog.Append(devtools.EventRecord{
		ID:        "event-123",
		Name:      "submit",
		SourceID:  "comp-1",
		TargetID:  "comp-1",
		Payload:   map[string]interface{}{"form": "login"},
		Timestamp: time.Now(),
		Duration:  10 * time.Millisecond,
	})

	config := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            100,
	}

	server, err := NewMCPServer(config, dt)
	require.NoError(t, err)

	// Register events resource
	err = server.RegisterEventsResource()
	require.NoError(t, err)

	// Read individual event resource
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://events/event-123",
		},
	}

	result, err := server.readEventResource(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Parse JSON response
	var event devtools.EventRecord
	err = json.Unmarshal([]byte(result.Contents[0].Text), &event)
	require.NoError(t, err)

	// Verify event
	assert.Equal(t, "event-123", event.ID)
	assert.Equal(t, "submit", event.Name)
	assert.Equal(t, "comp-1", event.SourceID)
	assert.Equal(t, "comp-1", event.TargetID)
}

// TestEventResource_NotFound tests reading non-existent event.
func TestEventResource_NotFound(t *testing.T) {
	// Create MCP server with devtools
	dt := devtools.Enable()
	defer devtools.Disable()

	config := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            100,
	}

	server, err := NewMCPServer(config, dt)
	require.NoError(t, err)

	// Register events resource
	err = server.RegisterEventsResource()
	require.NoError(t, err)

	// Try to read non-existent event
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://events/non-existent",
		},
	}

	result, err := server.readEventResource(context.Background(), req)
	assert.Error(t, err, "Should return error for non-existent event")
	assert.Nil(t, result)
}

// TestEventsResource_JSONSchema tests that response matches expected schema.
func TestEventsResource_JSONSchema(t *testing.T) {
	// Create MCP server with devtools
	dt := devtools.Enable()
	defer devtools.Disable()

	store := dt.GetStore()
	eventLog := store.GetEventLog()

	// Add event
	eventLog.Append(devtools.EventRecord{
		ID:        "event-1",
		Name:      "test",
		SourceID:  "comp-1",
		TargetID:  "comp-1",
		Payload:   nil,
		Timestamp: time.Now(),
		Duration:  1 * time.Millisecond,
	})

	config := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            100,
	}

	server, err := NewMCPServer(config, dt)
	require.NoError(t, err)

	// Register events resource
	err = server.RegisterEventsResource()
	require.NoError(t, err)

	// Read events log resource
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://events/log",
		},
	}

	result, err := server.readEventsLogResource(context.Background(), req)
	require.NoError(t, err)

	// Verify JSON structure
	var resource EventsResource
	err = json.Unmarshal([]byte(result.Contents[0].Text), &resource)
	require.NoError(t, err)

	// Verify schema fields exist
	assert.NotNil(t, resource.Events)
	assert.GreaterOrEqual(t, resource.TotalCount, 0)
	assert.NotZero(t, resource.Timestamp)

	// Verify event schema
	assert.NotEmpty(t, resource.Events[0].ID)
	assert.NotEmpty(t, resource.Events[0].Name)
	assert.NotEmpty(t, resource.Events[0].SourceID)
	assert.NotZero(t, resource.Events[0].Timestamp)
}

// TestEventsResource_ThreadSafe tests concurrent access to events resources.
func TestEventsResource_ThreadSafe(t *testing.T) {
	// Create MCP server with devtools
	dt := devtools.Enable()
	defer devtools.Disable()

	store := dt.GetStore()
	eventLog := store.GetEventLog()

	// Add event
	eventLog.Append(devtools.EventRecord{
		ID:        "event-1",
		Name:      "test",
		SourceID:  "comp-1",
		TargetID:  "comp-1",
		Payload:   nil,
		Timestamp: time.Now(),
		Duration:  1 * time.Millisecond,
	})

	config := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            100,
	}

	server, err := NewMCPServer(config, dt)
	require.NoError(t, err)

	// Register events resource
	err = server.RegisterEventsResource()
	require.NoError(t, err)

	// Concurrent reads
	const numReaders = 10
	done := make(chan bool, numReaders)

	for i := 0; i < numReaders; i++ {
		go func() {
			defer func() { done <- true }()

			req := &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{
					URI: "bubblyui://events/log",
				},
			}

			result, err := server.readEventsLogResource(context.Background(), req)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		}()
	}

	// Wait for all readers
	for i := 0; i < numReaders; i++ {
		<-done
	}
}

// TestEventsResource_LargeLog tests handling of large event logs.
func TestEventsResource_LargeLog(t *testing.T) {
	// Create MCP server with devtools
	dt := devtools.Enable()
	defer devtools.Disable()

	store := dt.GetStore()
	eventLog := store.GetEventLog()

	// Add many events
	const numEvents = 1000
	for i := 0; i < numEvents; i++ {
		eventLog.Append(devtools.EventRecord{
			ID:        "event-" + string(rune(i)),
			Name:      "test",
			SourceID:  "comp-1",
			TargetID:  "comp-1",
			Payload:   nil,
			Timestamp: time.Now(),
			Duration:  1 * time.Millisecond,
		})
	}

	config := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            100,
	}

	server, err := NewMCPServer(config, dt)
	require.NoError(t, err)

	// Register events resource
	err = server.RegisterEventsResource()
	require.NoError(t, err)

	// Read events log resource
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://events/log",
		},
	}

	result, err := server.readEventsLogResource(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Parse JSON response
	var resource EventsResource
	err = json.Unmarshal([]byte(result.Contents[0].Text), &resource)
	require.NoError(t, err)

	// Verify all events present
	assert.Len(t, resource.Events, numEvents, "Should have all events")
	assert.Equal(t, numEvents, resource.TotalCount, "Total count should match")
}

// TestExtractEventID tests the event ID extraction helper.
func TestExtractEventID(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected string
	}{
		{
			name:     "simple ID",
			uri:      "bubblyui://events/event-123",
			expected: "event-123",
		},
		{
			name:     "ID with special chars",
			uri:      "bubblyui://events/event-0x456abc",
			expected: "event-0x456abc",
		},
		{
			name:     "empty after prefix",
			uri:      "bubblyui://events/",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractEventID(tt.uri)
			assert.Equal(t, tt.expected, result)
		})
	}
}
