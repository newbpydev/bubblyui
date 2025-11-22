package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

// TestRegisterComponentsResource_Success tests successful registration
func TestRegisterComponentsResource_Success(t *testing.T) {
	// Create test setup
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Register resource
	err = server.RegisterComponentsResource()
	assert.NoError(t, err, "RegisterComponentsResource should succeed")
}

// TestComponentsResource_FullTree tests retrieving full component tree
func TestComponentsResource_FullTree(t *testing.T) {
	// Create test setup with components
	dt := devtools.Enable()
	defer devtools.Disable()

	store := dt.GetStore()

	// Add test components
	comp1 := &devtools.ComponentSnapshot{
		ID:        "comp-1",
		Name:      "App",
		Type:      "App",
		Timestamp: time.Now(),
		State:     make(map[string]interface{}),
		Props:     make(map[string]interface{}),
	}
	comp2 := &devtools.ComponentSnapshot{
		ID:        "comp-2",
		Name:      "Counter",
		Type:      "Counter",
		Timestamp: time.Now(),
		State:     make(map[string]interface{}),
		Props:     make(map[string]interface{}),
	}

	store.AddComponent(comp1)
	store.AddComponent(comp2)

	// Create MCP server
	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterComponentsResource()
	require.NoError(t, err)

	// Simulate resource read
	ctx := context.Background()
	result, err := server.readComponentsResource(ctx, &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://components",
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Contents, 1)

	// Parse JSON response
	var resource ComponentsResource
	err = json.Unmarshal([]byte(result.Contents[0].Text), &resource)
	require.NoError(t, err)

	// Verify structure
	assert.Equal(t, 2, resource.TotalCount, "Should have 2 components")
	assert.Len(t, resource.Roots, 2, "Should have 2 root components")
	assert.NotZero(t, resource.Timestamp, "Timestamp should be set")
}

// TestComponentResource_ByID tests retrieving individual component
func TestComponentResource_ByID(t *testing.T) {
	// Create test setup
	dt := devtools.Enable()
	defer devtools.Disable()

	store := dt.GetStore()

	// Add test component
	comp := &devtools.ComponentSnapshot{
		ID:        "comp-123",
		Name:      "TestComponent",
		Type:      "Test",
		Timestamp: time.Now(),
		State:     make(map[string]interface{}),
		Props:     make(map[string]interface{}),
	}
	store.AddComponent(comp)

	// Create MCP server
	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterComponentResource()
	require.NoError(t, err)

	// Simulate resource read
	ctx := context.Background()
	result, err := server.readComponentResource(ctx, &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://components/comp-123",
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Contents, 1)

	// Parse JSON response
	var snapshot devtools.ComponentSnapshot
	err = json.Unmarshal([]byte(result.Contents[0].Text), &snapshot)
	require.NoError(t, err)

	// Verify component
	assert.Equal(t, "comp-123", snapshot.ID)
	assert.Equal(t, "TestComponent", snapshot.Name)
	assert.Equal(t, "Test", snapshot.Type)
}

// TestComponentResource_NotFound tests missing component error
func TestComponentResource_NotFound(t *testing.T) {
	// Create test setup
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterComponentResource()
	require.NoError(t, err)

	// Try to read non-existent component
	ctx := context.Background()
	result, err := server.readComponentResource(ctx, &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://components/non-existent",
		},
	})

	// Should return error
	assert.Error(t, err, "Should return error for missing component")
	assert.Nil(t, result, "Result should be nil")
	assert.Contains(t, err.Error(), "not found", "Error should mention 'not found'")
}

// TestComponentsResource_ThreadSafe tests concurrent access
func TestComponentsResource_ThreadSafe(t *testing.T) {
	// Create test setup
	dt := devtools.Enable()
	defer devtools.Disable()

	store := dt.GetStore()

	// Add test components
	for i := 0; i < 10; i++ {
		comp := &devtools.ComponentSnapshot{
			ID:        fmt.Sprintf("comp-%d", i),
			Name:      "Component",
			Type:      "Test",
			Timestamp: time.Now(),
			State:     make(map[string]interface{}),
			Props:     make(map[string]interface{}),
		}
		store.AddComponent(comp)
	}

	// Create MCP server
	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterComponentsResource()
	require.NoError(t, err)

	// Concurrent reads
	ctx := context.Background()
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := server.readComponentsResource(ctx, &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{
					URI: "bubblyui://components",
				},
			})
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent read failed: %v", err)
	}
}

// TestComponentsResource_LargeTree tests handling large component trees
func TestComponentsResource_LargeTree(t *testing.T) {
	// Create test setup
	dt := devtools.Enable()
	defer devtools.Disable()

	store := dt.GetStore()

	// Add 1000 components (scaled down from 10,000 for test speed)
	for i := 0; i < 1000; i++ {
		comp := &devtools.ComponentSnapshot{
			ID:        fmt.Sprintf("comp-%d", i),
			Name:      "Component",
			Type:      "Test",
			Timestamp: time.Now(),
			State:     make(map[string]interface{}),
			Props:     make(map[string]interface{}),
		}
		store.AddComponent(comp)
	}

	// Create MCP server
	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterComponentsResource()
	require.NoError(t, err)

	// Read large tree
	ctx := context.Background()
	result, err := server.readComponentsResource(ctx, &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://components",
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)

	// Parse and verify
	var resource ComponentsResource
	err = json.Unmarshal([]byte(result.Contents[0].Text), &resource)
	require.NoError(t, err)

	assert.Equal(t, 1000, resource.TotalCount, "Should have 1000 components")
}

// TestExtractComponentID tests URI parsing
func TestExtractComponentID(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected string
	}{
		{
			name:     "simple ID",
			uri:      "bubblyui://components/comp-123",
			expected: "comp-123",
		},
		{
			name:     "ID with special chars",
			uri:      "bubblyui://components/comp-0x123abc",
			expected: "comp-0x123abc",
		},
		{
			name:     "empty after prefix",
			uri:      "bubblyui://components/",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractComponentID(tt.uri)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestComponentsResource_JSONSchema tests JSON structure
func TestComponentsResource_JSONSchema(t *testing.T) {
	// Create test setup
	dt := devtools.Enable()
	defer devtools.Disable()

	store := dt.GetStore()
	store.Clear() // Clear any previous test data

	// Add component with full data
	comp := &devtools.ComponentSnapshot{
		ID:        "comp-1",
		Name:      "TestComponent",
		Type:      "Test",
		Timestamp: time.Now(),
		State: map[string]interface{}{
			"count": 42,
		},
		Props: map[string]interface{}{
			"title": "Test",
		},
		Refs: []*devtools.RefSnapshot{
			{
				ID:    "ref-1",
				Name:  "count",
				Value: 42,
				Type:  "int",
			},
		},
	}
	store.AddComponent(comp)

	// Create MCP server
	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterComponentsResource()
	require.NoError(t, err)

	// Read resource
	ctx := context.Background()
	result, err := server.readComponentsResource(ctx, &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://components",
		},
	})

	require.NoError(t, err)

	// Verify JSON structure
	var resource ComponentsResource
	err = json.Unmarshal([]byte(result.Contents[0].Text), &resource)
	require.NoError(t, err)

	// Verify all fields present
	assert.NotNil(t, resource.Roots)
	assert.NotZero(t, resource.TotalCount)
	assert.NotZero(t, resource.Timestamp)

	// Verify component structure
	assert.Len(t, resource.Roots, 1)
	root := resource.Roots[0]
	assert.Equal(t, "comp-1", root.ID)
	assert.Equal(t, "TestComponent", root.Name)
	assert.NotNil(t, root.State)
	assert.NotNil(t, root.Props)
	assert.NotNil(t, root.Refs)
	assert.Len(t, root.Refs, 1)
}
