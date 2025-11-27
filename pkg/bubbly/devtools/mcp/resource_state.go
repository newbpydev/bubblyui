package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// StateResource represents the reactive state resource.
//
// This structure is returned by the bubblyui://state/refs resource
// and provides a snapshot of all reactive state in the application.
//
// JSON Schema:
//
//	{
//	  "refs": [RefInfo],
//	  "computed": [ComputedInfo],
//	  "timestamp": string (ISO 8601)
//	}
type StateResource struct {
	// Refs contains all reactive references across all components
	Refs []*RefInfo `json:"refs"`

	// Computed contains all computed values (future enhancement)
	Computed []*ComputedInfo `json:"computed"`

	// Timestamp indicates when this snapshot was captured
	Timestamp time.Time `json:"timestamp"`
}

// RefInfo provides detailed information about a reactive reference.
//
// This structure includes the ref's identity, value, type, ownership,
// and watcher information for debugging and analysis.
type RefInfo struct {
	// ID is the unique identifier of the ref
	ID string `json:"id"`

	// Name is the variable name of the ref
	Name string `json:"name"`

	// Type is the Go type of the ref's value
	Type string `json:"type"`

	// Value is the current value of the ref
	Value interface{} `json:"value"`

	// OwnerID is the ID of the component that owns this ref
	OwnerID string `json:"owner_id"`

	// OwnerName is the name of the component that owns this ref
	OwnerName string `json:"owner_name,omitempty"`

	// Watchers is the number of active watchers on this ref
	Watchers int `json:"watchers"`
}

// ComputedInfo provides information about computed values (future enhancement).
//
// This structure will be populated in future versions when computed value
// tracking is fully implemented in the devtools system.
type ComputedInfo struct {
	// ID is the unique identifier of the computed value
	ID string `json:"id"`

	// Name is the variable name
	Name string `json:"name"`

	// Type is the Go type
	Type string `json:"type"`

	// Value is the current computed value
	Value interface{} `json:"value"`

	// Dependencies lists the ref IDs this computed value depends on
	Dependencies []string `json:"dependencies,omitempty"`
}

// RegisterStateResource registers the state resources.
//
// This method registers two resources:
//   - bubblyui://state/refs - All reactive references
//   - bubblyui://state/history - State change history
//
// Thread Safety:
//
//	Safe to call concurrently. Resource reads use Store's thread-safe methods.
//
// Example Response (bubblyui://state/refs):
//
//	{
//	  "refs": [
//	    {
//	      "id": "ref-1",
//	      "name": "count",
//	      "type": "int",
//	      "value": 42,
//	      "owner_id": "comp-1",
//	      "owner_name": "Counter",
//	      "watchers": 2
//	    }
//	  ],
//	  "computed": [],
//	  "timestamp": "2025-01-13T14:30:00Z"
//	}
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
func (s *Server) RegisterStateResource() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Register refs resource
	s.registerResource(
		"bubblyui://state/refs",
		"state-refs",
		"All reactive references across components",
		func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return s.readStateRefsResource(ctx, req)
		},
	)

	// Register history resource
	s.registerResource(
		"bubblyui://state/history",
		"state-history",
		"State change history",
		func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return s.readStateHistoryResource(ctx, req)
		},
	)

	return nil
}

// readStateRefsResource handles reading the refs resource.
//
// This handler collects all refs from all components and returns them
// as a structured JSON response.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses Store's thread-safe methods.
func (s *Server) readStateRefsResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Collect all refs from all components
	refs := s.collectAllRefs()

	// Build resource response
	resource := StateResource{
		Refs:      refs,
		Computed:  []*ComputedInfo{}, // Empty for now, future enhancement
		Timestamp: time.Now(),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(resource, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state resource: %w", err)
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

// readStateHistoryResource handles reading the state history resource.
//
// This handler retrieves all state changes from the Store
// and returns them as a structured JSON response.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses Store's thread-safe methods.
func (s *Server) readStateHistoryResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Get state history from store
	history := s.store.GetStateHistory().GetAll()

	// Build response
	response := map[string]interface{}{
		"changes":   history,
		"count":     len(history),
		"timestamp": time.Now(),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state history: %w", err)
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

// collectAllRefs collects all refs from all components.
//
// This function iterates through all components in the Store
// and extracts their refs, building a complete list of all reactive
// references in the application.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses Store's thread-safe methods.
//
// Returns:
//   - []*RefInfo: List of all refs with their details
func (s *Server) collectAllRefs() []*RefInfo {
	// Get all components from store
	components := s.store.GetAllComponents()

	// Collect refs from all components
	refs := make([]*RefInfo, 0)

	for _, comp := range components {
		if comp.Refs == nil {
			continue
		}

		for _, refSnapshot := range comp.Refs {
			refInfo := &RefInfo{
				ID:        refSnapshot.ID,
				Name:      refSnapshot.Name,
				Type:      refSnapshot.Type,
				Value:     refSnapshot.Value,
				OwnerID:   comp.ID,
				OwnerName: comp.Name,
				Watchers:  refSnapshot.Watchers,
			}
			refs = append(refs, refInfo)
		}
	}

	return refs
}
