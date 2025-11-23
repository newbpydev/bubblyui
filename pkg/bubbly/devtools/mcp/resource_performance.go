package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

// PerformanceResource represents the performance metrics resource.
//
// This structure is returned by the bubblyui://performance/metrics resource
// and provides a comprehensive view of component performance metrics.
//
// JSON Schema:
//
//	{
//	  "components": {
//	    "component-id": ComponentPerformance
//	  },
//	  "summary": PerformanceSummary,
//	  "timestamp": string (ISO 8601)
//	}
type PerformanceResource struct {
	// Components maps component ID to performance metrics
	Components map[string]*devtools.ComponentPerformance `json:"components"`

	// Summary provides aggregated performance statistics
	Summary *PerformanceSummary `json:"summary"`

	// Timestamp indicates when this snapshot was captured
	Timestamp time.Time `json:"timestamp"`
}

// PerformanceSummary provides aggregated performance statistics.
//
// This structure summarizes performance metrics across all components,
// identifying the slowest, fastest, and most rendered components.
type PerformanceSummary struct {
	// TotalComponents is the total number of components tracked
	TotalComponents int `json:"total_components"`

	// TotalRenders is the total number of renders across all components
	TotalRenders int64 `json:"total_renders"`

	// SlowestComponent is the ID of the component with the slowest max render time
	SlowestComponent string `json:"slowest_component,omitempty"`

	// SlowestRenderTime is the slowest max render time
	SlowestRenderTime time.Duration `json:"slowest_render_time,omitempty"`

	// FastestComponent is the ID of the component with the fastest min render time
	FastestComponent string `json:"fastest_component,omitempty"`

	// FastestRenderTime is the fastest min render time
	FastestRenderTime time.Duration `json:"fastest_render_time,omitempty"`

	// MostRenderedComponent is the ID of the component with the most renders
	MostRenderedComponent string `json:"most_rendered_component,omitempty"`

	// MostRenderedCount is the number of renders for the most rendered component
	MostRenderedCount int64 `json:"most_rendered_count,omitempty"`
}

// RegisterPerformanceResource registers the performance resources.
//
// This method registers the following resource:
//   - bubblyui://performance/metrics - All performance metrics with summary
//
// Thread Safety:
//
//	Safe to call concurrently. Resource reads use Store's thread-safe methods.
//
// Example Response (bubblyui://performance/metrics):
//
//	{
//	  "components": {
//	    "comp-1": {
//	      "component_id": "comp-1",
//	      "component_name": "Counter",
//	      "render_count": 42,
//	      "total_render_time": "210ms",
//	      "avg_render_time": "5ms",
//	      "max_render_time": "15ms",
//	      "min_render_time": "2ms",
//	      "memory_usage": 1024,
//	      "last_update": "2025-01-13T14:30:00Z"
//	    }
//	  },
//	  "summary": {
//	    "total_components": 3,
//	    "total_renders": 150,
//	    "slowest_component": "comp-2",
//	    "slowest_render_time": "50ms",
//	    "fastest_component": "comp-1",
//	    "fastest_render_time": "2ms",
//	    "most_rendered_component": "comp-1",
//	    "most_rendered_count": 42
//	  },
//	  "timestamp": "2025-01-13T14:30:00Z"
//	}
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
func (s *Server) RegisterPerformanceResource() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Register metrics resource
	s.server.AddResource(
		&mcp.Resource{
			URI:         "bubblyui://performance/metrics",
			Name:        "performance-metrics",
			Description: "Component performance metrics with summary statistics",
			MIMEType:    "application/json",
		},
		func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return s.readPerformanceMetricsResource(ctx, req)
		},
	)

	return nil
}

// readPerformanceMetricsResource handles reading the performance metrics resource.
//
// This handler retrieves all performance metrics from the Store,
// calculates summary statistics, and returns them as a structured JSON response.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses Store's thread-safe methods.
func (s *Server) readPerformanceMetricsResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Get all performance data from store
	components := s.store.GetPerformanceData().GetAll()

	// Calculate summary statistics
	summary := s.calculatePerformanceSummary(components)

	// Build resource response
	resource := PerformanceResource{
		Components: components,
		Summary:    summary,
		Timestamp:  time.Now(),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(resource, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal performance resource: %w", err)
	}

	// Return MCP resource result
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(data),
			},
		},
	}, nil
}

// calculatePerformanceSummary calculates aggregated performance statistics.
//
// This function analyzes all component performance metrics to identify:
//   - Total number of components and renders
//   - Slowest component (by max render time)
//   - Fastest component (by min render time)
//   - Most rendered component (by render count)
//
// Thread Safety:
//
//	Safe to call concurrently. Operates on a copy of the data.
//
// Parameters:
//   - components: Map of component ID to performance metrics
//
// Returns:
//   - *PerformanceSummary: Aggregated statistics
func (s *Server) calculatePerformanceSummary(components map[string]*devtools.ComponentPerformance) *PerformanceSummary {
	summary := &PerformanceSummary{
		TotalComponents: len(components),
	}

	// Handle empty case
	if len(components) == 0 {
		return summary
	}

	// Initialize tracking variables
	var (
		slowestTime  time.Duration
		fastestTime  = time.Duration(1<<63 - 1) // Max duration
		mostRendered int64
		totalRenders int64
	)

	// Analyze all components
	for id, comp := range components {
		// Accumulate total renders
		totalRenders += comp.RenderCount

		// Track slowest component
		if comp.MaxRenderTime > slowestTime {
			slowestTime = comp.MaxRenderTime
			summary.SlowestComponent = id
			summary.SlowestRenderTime = comp.MaxRenderTime
		}

		// Track fastest component
		if comp.MinRenderTime < fastestTime && comp.MinRenderTime > 0 {
			fastestTime = comp.MinRenderTime
			summary.FastestComponent = id
			summary.FastestRenderTime = comp.MinRenderTime
		}

		// Track most rendered component
		if comp.RenderCount > mostRendered {
			mostRendered = comp.RenderCount
			summary.MostRenderedComponent = id
			summary.MostRenderedCount = comp.RenderCount
		}
	}

	summary.TotalRenders = totalRenders

	return summary
}
