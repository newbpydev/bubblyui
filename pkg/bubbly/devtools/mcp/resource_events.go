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

// registerResource is a helper function to register a resource with the MCP server.
// It reduces code duplication in resource registration methods.
func (s *MCPServer) registerResource(uri, name, description string, handler func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error)) {
	s.server.AddResource(
		&mcp.Resource{
			URI:         uri,
			Name:        name,
			Description: description,
			MIMEType:    "application/json",
		},
		handler,
	)
}

// EventsResource represents the events log resource.
//
// This structure is returned by the bubblyui://events/log resource
// and provides a snapshot of all events that occurred in the application.
//
// JSON Schema:
//
//	{
//	  "events": [EventRecord],
//	  "total_count": int,
//	  "timestamp": string (ISO 8601)
//	}
type EventsResource struct {
	// Events contains all event records
	Events []devtools.EventRecord `json:"events"`

	// TotalCount is the total number of events
	TotalCount int `json:"total_count"`

	// Timestamp indicates when this snapshot was captured
	Timestamp time.Time `json:"timestamp"`
}

// RegisterEventsResource registers the events resources.
//
// This method registers two resources:
//   - bubblyui://events/log - All event records
//   - bubblyui://events/{id} - Individual event by ID
//
// Thread Safety:
//
//	Safe to call concurrently. Resource reads use DevToolsStore's thread-safe methods.
//
// Example Response (bubblyui://events/log):
//
//	{
//	  "events": [
//	    {
//	      "seq_id": 1,
//	      "id": "event-1",
//	      "name": "click",
//	      "source_id": "comp-1",
//	      "target_id": "comp-2",
//	      "payload": {"button": "submit"},
//	      "timestamp": "2025-01-13T14:30:00Z",
//	      "duration": 5000000
//	    }
//	  ],
//	  "total_count": 1,
//	  "timestamp": "2025-01-13T14:30:00Z"
//	}
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
func (s *MCPServer) RegisterEventsResource() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Register events log resource
	s.registerResource(
		"bubblyui://events/log",
		"events-log",
		"All event records",
		func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return s.readEventsLogResource(ctx, req)
		},
	)

	// Register individual event resource template
	s.server.AddResourceTemplate(
		&mcp.ResourceTemplate{
			URITemplate: "bubblyui://events/{id}",
			Name:        "event",
			Description: "Individual event details by ID",
			MIMEType:    "application/json",
		},
		func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return s.readEventResource(ctx, req)
		},
	)

	return nil
}

// readEventsLogResource handles reading the events log resource.
//
// This handler retrieves all events from the DevToolsStore
// and returns them as a structured JSON response.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses DevToolsStore's thread-safe methods.
func (s *MCPServer) readEventsLogResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Get event log from store
	eventLog := s.store.GetEventLog()
	events := eventLog.GetRecent(eventLog.Len())

	// Build resource response
	resource := EventsResource{
		Events:     events,
		TotalCount: len(events),
		Timestamp:  time.Now(),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(resource, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal events resource: %w", err)
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

// readEventResource handles reading an individual event resource.
//
// This handler retrieves a specific event by ID from the DevToolsStore
// and returns it as a structured JSON response.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses DevToolsStore's thread-safe methods.
func (s *MCPServer) readEventResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Extract event ID from URI
	eventID := extractEventID(req.Params.URI)
	if eventID == "" {
		return nil, fmt.Errorf("invalid event URI: missing event ID")
	}

	// Get all events and find the one with matching ID
	eventLog := s.store.GetEventLog()
	events := eventLog.GetRecent(eventLog.Len())

	var foundEvent *devtools.EventRecord
	for i := range events {
		if events[i].ID == eventID {
			foundEvent = &events[i]
			break
		}
	}

	if foundEvent == nil {
		// Return MCP ResourceNotFoundError
		return nil, mcp.ResourceNotFoundError(req.Params.URI)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(foundEvent, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
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

// extractEventID extracts the event ID from a resource URI.
//
// This function parses URIs of the form:
//   - bubblyui://events/event-123 → "event-123"
//   - bubblyui://events/event-0x456 → "event-0x456"
//   - bubblyui://events/ → ""
//
// Thread Safety:
//
//	Safe to call concurrently (pure function).
//
// Parameters:
//   - uri: The resource URI to parse
//
// Returns:
//   - string: The extracted event ID, or empty string if not found
func extractEventID(uri string) string {
	// Expected format: "bubblyui://events/{id}"
	prefix := "bubblyui://events/"

	if !strings.HasPrefix(uri, prefix) {
		return ""
	}

	// Extract everything after the prefix
	id := strings.TrimPrefix(uri, prefix)
	return id
}
