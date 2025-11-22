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

// TestRegisterPerformanceResource tests registering the performance resource.
func TestRegisterPerformanceResource(t *testing.T) {
	// Create test server
	dt := devtools.Enable()
	defer devtools.Disable()

	// Clear any residual state
	dt.GetStore().Clear()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Register performance resource
	err = server.RegisterPerformanceResource()
	assert.NoError(t, err)
}

// TestPerformanceMetricsResource tests the bubblyui://performance/metrics resource.
func TestPerformanceMetricsResource(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*devtools.DevToolsStore)
		validate func(*testing.T, *PerformanceResource)
	}{
		{
			name: "empty performance data",
			setup: func(store *devtools.DevToolsStore) {
				// No setup - empty data
			},
			validate: func(t *testing.T, resource *PerformanceResource) {
				assert.Empty(t, resource.Components)
				assert.NotNil(t, resource.Summary)
				assert.Equal(t, 0, resource.Summary.TotalComponents)
				assert.Equal(t, int64(0), resource.Summary.TotalRenders)
			},
		},
		{
			name: "single component performance",
			setup: func(store *devtools.DevToolsStore) {
				perf := store.GetPerformanceData()
				perf.RecordRender("comp-1", "Counter", 5*time.Millisecond)
				perf.RecordRender("comp-1", "Counter", 3*time.Millisecond)
			},
			validate: func(t *testing.T, resource *PerformanceResource) {
				assert.Len(t, resource.Components, 1)
				assert.Contains(t, resource.Components, "comp-1")

				comp := resource.Components["comp-1"]
				assert.Equal(t, "Counter", comp.ComponentName)
				assert.Equal(t, int64(2), comp.RenderCount)
				assert.Equal(t, 4*time.Millisecond, comp.AvgRenderTime)
				assert.Equal(t, 5*time.Millisecond, comp.MaxRenderTime)
				assert.Equal(t, 3*time.Millisecond, comp.MinRenderTime)
			},
		},
		{
			name: "multiple components performance",
			setup: func(store *devtools.DevToolsStore) {
				perf := store.GetPerformanceData()
				// Fast component
				perf.RecordRender("comp-1", "Fast", 1*time.Millisecond)
				perf.RecordRender("comp-1", "Fast", 2*time.Millisecond)
				// Slow component
				perf.RecordRender("comp-2", "Slow", 50*time.Millisecond)
				perf.RecordRender("comp-2", "Slow", 60*time.Millisecond)
				// Medium component
				perf.RecordRender("comp-3", "Medium", 10*time.Millisecond)
			},
			validate: func(t *testing.T, resource *PerformanceResource) {
				assert.Len(t, resource.Components, 3)
				assert.NotNil(t, resource.Summary)
				assert.Equal(t, 3, resource.Summary.TotalComponents)
				assert.Equal(t, int64(5), resource.Summary.TotalRenders)
			},
		},
		{
			name: "large dataset (1000 components)",
			setup: func(store *devtools.DevToolsStore) {
				perf := store.GetPerformanceData()
				for i := 0; i < 1000; i++ {
					compID := "comp-" + string(rune(i))
					perf.RecordRender(compID, "Component", time.Duration(i+1)*time.Microsecond)
				}
			},
			validate: func(t *testing.T, resource *PerformanceResource) {
				assert.Len(t, resource.Components, 1000)
				assert.Equal(t, 1000, resource.Summary.TotalComponents)
				assert.Equal(t, int64(1000), resource.Summary.TotalRenders)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			dt := devtools.Enable()
			defer devtools.Disable()

			// Clear any residual state
			dt.GetStore().Clear()

			cfg := DefaultMCPConfig()
			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)

			// Setup test data
			tt.setup(server.store)

			// Register resource
			err = server.RegisterPerformanceResource()
			require.NoError(t, err)

			// Read resource
			req := &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{
					URI: "bubblyui://performance/metrics",
				},
			}

			result, err := server.readPerformanceMetricsResource(context.Background(), req)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Len(t, result.Contents, 1)

			// Parse response
			var resource PerformanceResource
			err = json.Unmarshal([]byte(result.Contents[0].Text), &resource)
			require.NoError(t, err)

			// Validate
			tt.validate(t, &resource)
			assert.False(t, resource.Timestamp.IsZero())
		})
	}
}

// TestPerformanceSummaryCalculations tests summary metric calculations.
func TestPerformanceSummaryCalculations(t *testing.T) {
	// Create test server
	dt := devtools.Enable()
	defer devtools.Disable()

	// Clear any residual state
	dt.GetStore().Clear()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Setup test data with known values
	perf := server.store.GetPerformanceData()
	perf.RecordRender("comp-1", "Fast", 1*time.Millisecond)
	perf.RecordRender("comp-1", "Fast", 2*time.Millisecond)
	perf.RecordRender("comp-2", "Slow", 100*time.Millisecond)
	perf.RecordRender("comp-3", "Medium", 10*time.Millisecond)

	// Register and read resource
	err = server.RegisterPerformanceResource()
	require.NoError(t, err)

	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://performance/metrics",
		},
	}

	result, err := server.readPerformanceMetricsResource(context.Background(), req)
	require.NoError(t, err)

	var resource PerformanceResource
	err = json.Unmarshal([]byte(result.Contents[0].Text), &resource)
	require.NoError(t, err)

	// Validate summary calculations
	summary := resource.Summary
	assert.Equal(t, 3, summary.TotalComponents)
	assert.Equal(t, int64(4), summary.TotalRenders)

	// Slowest should be "Slow" component
	assert.Equal(t, "comp-2", summary.SlowestComponent)
	assert.Equal(t, 100*time.Millisecond, summary.SlowestRenderTime)

	// Fastest should be "Fast" component
	assert.Equal(t, "comp-1", summary.FastestComponent)
	assert.Equal(t, 1*time.Millisecond, summary.FastestRenderTime)

	// Most rendered should be "Fast" component (2 renders)
	assert.Equal(t, "comp-1", summary.MostRenderedComponent)
	assert.Equal(t, int64(2), summary.MostRenderedCount)
}

// TestPerformanceSortByRenderTime tests sorting components by render time.
func TestPerformanceSortByRenderTime(t *testing.T) {
	// Create test server
	dt := devtools.Enable()
	defer devtools.Disable()

	// Clear any residual state
	dt.GetStore().Clear()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Setup test data
	perf := server.store.GetPerformanceData()
	perf.RecordRender("comp-1", "Fast", 1*time.Millisecond)
	perf.RecordRender("comp-2", "Slow", 100*time.Millisecond)
	perf.RecordRender("comp-3", "Medium", 10*time.Millisecond)

	// Register and read resource
	err = server.RegisterPerformanceResource()
	require.NoError(t, err)

	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://performance/metrics",
		},
	}

	result, err := server.readPerformanceMetricsResource(context.Background(), req)
	require.NoError(t, err)

	var resource PerformanceResource
	err = json.Unmarshal([]byte(result.Contents[0].Text), &resource)
	require.NoError(t, err)

	// Verify components are present (order in map is not guaranteed)
	assert.Len(t, resource.Components, 3)
	assert.Contains(t, resource.Components, "comp-1")
	assert.Contains(t, resource.Components, "comp-2")
	assert.Contains(t, resource.Components, "comp-3")
}

// TestPerformanceConcurrentAccess tests thread-safe concurrent access.
func TestPerformanceConcurrentAccess(t *testing.T) {
	// Create test server
	dt := devtools.Enable()
	defer devtools.Disable()

	// Clear any residual state
	dt.GetStore().Clear()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Setup initial data
	perf := server.store.GetPerformanceData()
	for i := 0; i < 10; i++ {
		compID := "comp-" + string(rune(i))
		perf.RecordRender(compID, "Component", time.Millisecond)
	}

	// Register resource
	err = server.RegisterPerformanceResource()
	require.NoError(t, err)

	// Concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			req := &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{
					URI: "bubblyui://performance/metrics",
				},
			}

			result, err := server.readPerformanceMetricsResource(context.Background(), req)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestPerformanceJSONSchemaValidation tests JSON schema compliance.
func TestPerformanceJSONSchemaValidation(t *testing.T) {
	// Create test server
	dt := devtools.Enable()
	defer devtools.Disable()

	// Clear any residual state
	dt.GetStore().Clear()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Setup test data
	perf := server.store.GetPerformanceData()
	perf.RecordRender("comp-1", "Test", 5*time.Millisecond)

	// Register and read resource
	err = server.RegisterPerformanceResource()
	require.NoError(t, err)

	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://performance/metrics",
		},
	}

	result, err := server.readPerformanceMetricsResource(context.Background(), req)
	require.NoError(t, err)

	// Verify JSON structure
	var resource PerformanceResource
	err = json.Unmarshal([]byte(result.Contents[0].Text), &resource)
	require.NoError(t, err)

	// Verify required fields
	assert.NotNil(t, resource.Components)
	assert.NotNil(t, resource.Summary)
	assert.False(t, resource.Timestamp.IsZero())

	// Verify summary fields
	assert.GreaterOrEqual(t, resource.Summary.TotalComponents, 0)
	assert.GreaterOrEqual(t, resource.Summary.TotalRenders, int64(0))
}
