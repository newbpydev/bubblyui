package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

// ComponentsResource represents the full component tree resource.
//
// This structure is returned by the bubblyui://components resource
// and provides a snapshot of all root components in the application.
//
// JSON Schema:
//
//	{
//	  "roots": [ComponentSnapshot],
//	  "total_count": int,
//	  "timestamp": string (ISO 8601)
//	}
type ComponentsResource struct {
	// Roots contains all root-level components (components without parents)
	Roots []*devtools.ComponentSnapshot `json:"roots"`

	// TotalCount is the total number of components in the tree
	TotalCount int `json:"total_count"`

	// Timestamp indicates when this snapshot was captured
	Timestamp time.Time `json:"timestamp"`
}

// RegisterComponentsResource registers the full component tree resource.
//
// This resource provides access to all root components and their children
// via the URI: bubblyui://components
//
// The resource returns a ComponentsResource JSON structure containing:
//   - roots: Array of root component snapshots
//   - total_count: Total number of components tracked
//   - timestamp: When the snapshot was captured
//
// Thread Safety:
//
//	Safe to call concurrently. Resource reads use DevToolsStore's thread-safe methods.
//
// Example Response:
//
//	{
//	  "roots": [
//	    {
//	      "id": "comp-1",
//	      "name": "App",
//	      "type": "App",
//	      "state": {},
//	      "refs": [],
//	      "props": {},
//	      "children": [...]
//	    }
//	  ],
//	  "total_count": 5,
//	  "timestamp": "2025-01-13T14:30:00Z"
//	}
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
func (s *MCPServer) RegisterComponentsResource() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Register the resource with MCP SDK
	// Note: AddResource doesn't return an error in the MCP Go SDK
	s.server.AddResource(
		&mcp.Resource{
			URI:         "bubblyui://components",
			Name:        "components",
			Description: "Full component tree snapshot",
			MIMEType:    "application/json",
		},
		func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return s.readComponentsResource(ctx, req)
		},
	)

	return nil
}

// readComponentsResource handles reading the full component tree resource.
//
// This is the handler function called by the MCP SDK when a client
// requests the bubblyui://components resource.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses DevToolsStore's thread-safe methods.
func (s *MCPServer) readComponentsResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Get all components and roots from store
	allComponents := s.store.GetAllComponents()
	roots := s.store.GetRootComponents()

	// Build resource response
	resource := ComponentsResource{
		Roots:      roots,
		TotalCount: len(allComponents),
		Timestamp:  time.Now(),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(resource, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal components resource: %w", err)
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

// RegisterComponentResource registers the individual component resource template.
//
// This resource template provides access to individual components by ID
// via the URI pattern: bubblyui://components/{id}
//
// The resource returns a ComponentSnapshot JSON structure for the requested component.
//
// Thread Safety:
//
//	Safe to call concurrently. Resource reads use DevToolsStore's thread-safe methods.
//
// Example Request:
//
//	URI: bubblyui://components/comp-123
//
// Example Response:
//
//	{
//	  "id": "comp-123",
//	  "name": "Counter",
//	  "type": "Counter",
//	  "state": {"count": 42},
//	  "refs": [...],
//	  "props": {...},
//	  "children": [...]
//	}
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
func (s *MCPServer) RegisterComponentResource() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Register the resource template with MCP SDK
	// Note: AddResourceTemplate doesn't return an error in the MCP Go SDK
	s.server.AddResourceTemplate(
		&mcp.ResourceTemplate{
			URITemplate: "bubblyui://components/{id}",
			Name:        "component",
			Description: "Individual component details by ID",
			MIMEType:    "application/json",
		},
		func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return s.readComponentResource(ctx, req)
		},
	)

	return nil
}

// readComponentResource handles reading an individual component resource.
//
// This is the handler function called by the MCP SDK when a client
// requests a specific component via bubblyui://components/{id}.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses DevToolsStore's thread-safe methods.
func (s *MCPServer) readComponentResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Extract component ID from URI
	componentID := extractComponentID(req.Params.URI)
	if componentID == "" {
		return nil, fmt.Errorf("invalid component URI: missing component ID")
	}

	// Get component from store
	component := s.store.GetComponent(componentID)
	if component == nil {
		// Return MCP ResourceNotFoundError
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(component, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal component: %w", err)
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

// extractComponentID extracts the component ID from a resource URI.
//
// This function parses URIs of the form:
//   - bubblyui://components/comp-123 → "comp-123"
//   - bubblyui://components/comp-0x456 → "comp-0x456"
//   - bubblyui://components/ → ""
//
// Thread Safety:
//
//	Safe to call concurrently (pure function).
//
// Parameters:
//   - uri: The resource URI to parse
//
// Returns:
//   - string: The extracted component ID, or empty string if not found
func extractComponentID(uri string) string {
	// Expected format: "bubblyui://components/{id}"
	prefix := "bubblyui://components/"

	if !strings.HasPrefix(uri, prefix) {
		return ""
	}

	// Extract everything after the prefix
	id := strings.TrimPrefix(uri, prefix)
	return id
}
