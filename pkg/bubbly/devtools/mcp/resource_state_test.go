package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

// TestRegisterStateResource_Success tests successful registration of state resources.
func TestRegisterStateResource_Success(t *testing.T) {
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

	// Register state resource
	err = server.RegisterStateResource()
	assert.NoError(t, err, "RegisterStateResource should succeed")
}

// TestStateResource_RefsEmpty tests reading refs when no refs exist.
func TestStateResource_RefsEmpty(t *testing.T) {
	// Create MCP server with devtools
	dt := devtools.Enable()
	defer devtools.Disable()

	// Clear any residual state
	dt.GetStore().Clear()

	config := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            100,
	}

	server, err := NewMCPServer(config, dt)
	require.NoError(t, err)

	// Register state resource
	err = server.RegisterStateResource()
	require.NoError(t, err)

	// Read refs resource
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://state/refs",
		},
	}

	result, err := server.readStateRefsResource(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Contents, 1)

	// Parse JSON response
	var resource StateResource
	err = json.Unmarshal([]byte(result.Contents[0].Text), &resource)
	require.NoError(t, err)

	// Verify empty refs
	assert.Empty(t, resource.Refs, "Should have no refs initially")
	assert.Empty(t, resource.Computed, "Should have no computed values initially")
	assert.NotZero(t, resource.Timestamp, "Should have timestamp")
}

// TestStateResource_RefsWithData tests reading refs with actual data.
func TestStateResource_RefsWithData(t *testing.T) {
	// Create MCP server with devtools
	dt := devtools.Enable()
	defer devtools.Disable()

	store := dt.GetStore()

	// Add component with refs
	snapshot := &devtools.ComponentSnapshot{
		ID:   "comp-1",
		Name: "Counter",
		Type: "Counter",
		Refs: []*devtools.RefSnapshot{
			{
				ID:       "ref-1",
				Name:     "count",
				Type:     "int",
				Value:    42,
				Watchers: 2,
			},
			{
				ID:       "ref-2",
				Name:     "name",
				Type:     "string",
				Value:    "test",
				Watchers: 0,
			},
		},
		Timestamp: time.Now(),
	}
	store.AddComponent(snapshot)

	// Register ref ownership
	store.RegisterRefOwner("comp-1", "ref-1")
	store.RegisterRefOwner("comp-1", "ref-2")

	config := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            100,
	}

	server, err := NewMCPServer(config, dt)
	require.NoError(t, err)

	// Register state resource
	err = server.RegisterStateResource()
	require.NoError(t, err)

	// Read refs resource
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://state/refs",
		},
	}

	result, err := server.readStateRefsResource(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Parse JSON response
	var resource StateResource
	err = json.Unmarshal([]byte(result.Contents[0].Text), &resource)
	require.NoError(t, err)

	// Verify refs
	assert.Len(t, resource.Refs, 2, "Should have 2 refs")
	assert.NotZero(t, resource.Timestamp, "Should have timestamp")

	// Verify ref details
	refMap := make(map[string]*RefInfo)
	for _, ref := range resource.Refs {
		refMap[ref.ID] = ref
	}

	assert.Contains(t, refMap, "ref-1")
	assert.Equal(t, "count", refMap["ref-1"].Name)
	assert.Equal(t, "int", refMap["ref-1"].Type)
	assert.Equal(t, float64(42), refMap["ref-1"].Value) // JSON unmarshals numbers as float64
	assert.Equal(t, "comp-1", refMap["ref-1"].OwnerID)
	assert.Equal(t, 2, refMap["ref-1"].Watchers)

	assert.Contains(t, refMap, "ref-2")
	assert.Equal(t, "name", refMap["ref-2"].Name)
	assert.Equal(t, "string", refMap["ref-2"].Type)
	assert.Equal(t, "test", refMap["ref-2"].Value)
	assert.Equal(t, "comp-1", refMap["ref-2"].OwnerID)
	assert.Equal(t, 0, refMap["ref-2"].Watchers)
}

// TestStateHistoryResource_Empty tests reading history when empty.
func TestStateHistoryResource_Empty(t *testing.T) {
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

	// Register state resource
	err = server.RegisterStateResource()
	require.NoError(t, err)

	// Read history resource
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://state/history",
		},
	}

	result, err := server.readStateHistoryResource(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Parse JSON response
	var response map[string]interface{}
	err = json.Unmarshal([]byte(result.Contents[0].Text), &response)
	require.NoError(t, err)

	// Verify empty history
	changes := response["changes"].([]interface{})
	assert.Empty(t, changes, "Should have no changes initially")
	assert.NotZero(t, response["timestamp"], "Should have timestamp")
}

// TestStateHistoryResource_WithChanges tests reading history with data.
func TestStateHistoryResource_WithChanges(t *testing.T) {
	// Create MCP server with devtools
	dt := devtools.Enable()
	defer devtools.Disable()

	store := dt.GetStore()

	// Add state changes
	store.GetStateHistory().Record(devtools.StateChange{
		RefID:     "ref-1",
		RefName:   "count",
		OldValue:  41,
		NewValue:  42,
		Timestamp: time.Now(),
		Source:    "increment",
	})

	store.GetStateHistory().Record(devtools.StateChange{
		RefID:     "ref-1",
		RefName:   "count",
		OldValue:  42,
		NewValue:  43,
		Timestamp: time.Now(),
		Source:    "increment",
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

	// Register state resource
	err = server.RegisterStateResource()
	require.NoError(t, err)

	// Read history resource
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://state/history",
		},
	}

	result, err := server.readStateHistoryResource(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Parse JSON response
	var response map[string]interface{}
	err = json.Unmarshal([]byte(result.Contents[0].Text), &response)
	require.NoError(t, err)

	// Verify history
	changes := response["changes"].([]interface{})
	assert.Len(t, changes, 2, "Should have 2 state changes")
	assert.Equal(t, float64(2), response["count"], "Count should be 2")
	assert.NotZero(t, response["timestamp"], "Should have timestamp")
}

// TestStateResource_JSONSchema tests that response matches expected schema.
func TestStateResource_JSONSchema(t *testing.T) {
	// Create MCP server with devtools
	dt := devtools.Enable()
	defer devtools.Disable()

	store := dt.GetStore()

	// Add component with ref
	snapshot := &devtools.ComponentSnapshot{
		ID:   "comp-1",
		Name: "Counter",
		Refs: []*devtools.RefSnapshot{
			{
				ID:       "ref-1",
				Name:     "count",
				Type:     "int",
				Value:    42,
				Watchers: 1,
			},
		},
		Timestamp: time.Now(),
	}
	store.AddComponent(snapshot)
	store.RegisterRefOwner("comp-1", "ref-1")

	config := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            100,
	}

	server, err := NewMCPServer(config, dt)
	require.NoError(t, err)

	// Register state resource
	err = server.RegisterStateResource()
	require.NoError(t, err)

	// Read refs resource
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://state/refs",
		},
	}

	result, err := server.readStateRefsResource(context.Background(), req)
	require.NoError(t, err)

	// Verify JSON structure
	var resource StateResource
	err = json.Unmarshal([]byte(result.Contents[0].Text), &resource)
	require.NoError(t, err)

	// Verify schema fields exist
	assert.NotNil(t, resource.Refs)
	assert.NotNil(t, resource.Computed)
	assert.NotZero(t, resource.Timestamp)

	// Verify ref schema
	assert.NotEmpty(t, resource.Refs[0].ID)
	assert.NotEmpty(t, resource.Refs[0].Name)
	assert.NotEmpty(t, resource.Refs[0].Type)
	assert.NotNil(t, resource.Refs[0].Value)
	assert.NotEmpty(t, resource.Refs[0].OwnerID)
	assert.GreaterOrEqual(t, resource.Refs[0].Watchers, 0)
}

// TestStateResource_ThreadSafe tests concurrent access to state resources.
func TestStateResource_ThreadSafe(t *testing.T) {
	// Create MCP server with devtools
	dt := devtools.Enable()
	defer devtools.Disable()

	store := dt.GetStore()

	// Add component with ref
	snapshot := &devtools.ComponentSnapshot{
		ID:   "comp-1",
		Name: "Counter",
		Refs: []*devtools.RefSnapshot{
			{
				ID:    "ref-1",
				Name:  "count",
				Type:  "int",
				Value: 0,
			},
		},
		Timestamp: time.Now(),
	}
	store.AddComponent(snapshot)
	store.RegisterRefOwner("comp-1", "ref-1")

	config := &MCPConfig{
		Transport:            MCPTransportStdio,
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            100,
	}

	server, err := NewMCPServer(config, dt)
	require.NoError(t, err)

	// Register state resource
	err = server.RegisterStateResource()
	require.NoError(t, err)

	// Concurrent reads
	const numReaders = 10
	done := make(chan bool, numReaders)

	for i := 0; i < numReaders; i++ {
		go func() {
			defer func() { done <- true }()

			req := &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{
					URI: "bubblyui://state/refs",
				},
			}

			result, err := server.readStateRefsResource(context.Background(), req)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		}()
	}

	// Wait for all readers
	for i := 0; i < numReaders; i++ {
		<-done
	}
}

// TestStateResource_LargeHistory tests handling of large history.
func TestStateResource_LargeHistory(t *testing.T) {
	// Create MCP server with devtools
	dt := devtools.Enable()
	defer devtools.Disable()

	store := dt.GetStore()

	// Add many state changes
	const numChanges = 1000
	for i := 0; i < numChanges; i++ {
		store.GetStateHistory().Record(devtools.StateChange{
			RefID:     "ref-1",
			RefName:   "count",
			OldValue:  i,
			NewValue:  i + 1,
			Timestamp: time.Now(),
			Source:    "increment",
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

	// Register state resource
	err = server.RegisterStateResource()
	require.NoError(t, err)

	// Read history resource
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://state/history",
		},
	}

	result, err := server.readStateHistoryResource(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Parse JSON response
	var response map[string]interface{}
	err = json.Unmarshal([]byte(result.Contents[0].Text), &response)
	require.NoError(t, err)

	// Verify all changes present
	changes := response["changes"].([]interface{})
	assert.Len(t, changes, numChanges, "Should have all state changes")
}
