package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// SearchComponentsParams defines the parameters for the search_components tool.
//
// This structure is used by AI agents to specify search criteria when
// searching for components in the component tree.
//
// Example:
//
//	{
//	  "query": "counter",
//	  "fields": ["name", "type"],
//	  "max_results": 10
//	}
type SearchComponentsParams struct {
	// Query is the search term to match against component fields
	Query string `json:"query"`

	// Fields specifies which fields to search in: "name", "type", "id"
	// If empty, searches all fields
	Fields []string `json:"fields"`

	// MaxResults limits the number of results returned (default: 50)
	MaxResults int `json:"max_results"`
}

// SearchComponentsResult contains the result of a component search operation.
//
// This structure is returned to AI agents after a successful search.
//
// Example:
//
//	{
//	  "matches": [
//	    {
//	      "id": "comp-1",
//	      "name": "Counter",
//	      "type": "Counter",
//	      "match_score": 1.0,
//	      "matched_field": "name"
//	    }
//	  ],
//	  "total_matches": 1,
//	  "query": "counter",
//	  "timestamp": "2025-01-13T14:30:22Z"
//	}
type SearchComponentsResult struct {
	// Matches contains the components that matched the search
	Matches []ComponentMatch `json:"matches"`

	// TotalMatches is the total number of matches found
	TotalMatches int `json:"total_matches"`

	// Query is the search query that was used
	Query string `json:"query"`

	// Timestamp is when the search was performed
	Timestamp time.Time `json:"timestamp"`
}

// ComponentMatch represents a single component that matched the search.
type ComponentMatch struct {
	// ID is the component's unique identifier
	ID string `json:"id"`

	// Name is the component's name
	Name string `json:"name"`

	// Type is the component's type
	Type string `json:"type"`

	// Status is the component's lifecycle status
	Status string `json:"status"`

	// MatchScore is a relevance score (0.0 to 1.0)
	MatchScore float64 `json:"match_score"`

	// MatchedField is which field matched the query
	MatchedField string `json:"matched_field"`
}

// FilterEventsParams defines the parameters for the filter_events tool.
//
// This structure is used by AI agents to specify filtering criteria when
// querying the event log.
//
// Example:
//
//	{
//	  "event_names": ["click", "submit"],
//	  "source_ids": ["comp-1"],
//	  "start_time": "2025-01-13T14:00:00Z",
//	  "end_time": "2025-01-13T15:00:00Z",
//	  "limit": 100
//	}
type FilterEventsParams struct {
	// EventNames filters by event name (empty = all events)
	EventNames []string `json:"event_names"`

	// SourceIDs filters by source component ID (empty = all sources)
	SourceIDs []string `json:"source_ids"`

	// StartTime filters events after this time (optional)
	StartTime *time.Time `json:"start_time"`

	// EndTime filters events before this time (optional)
	EndTime *time.Time `json:"end_time"`

	// Limit limits the number of results returned (default: 100)
	Limit int `json:"limit"`
}

// FilterEventsResult contains the result of an event filtering operation.
//
// This structure is returned to AI agents after a successful filter.
//
// Example:
//
//	{
//	  "events": [
//	    {
//	      "id": "event-1",
//	      "name": "click",
//	      "source_id": "comp-1",
//	      "timestamp": "2025-01-13T14:30:22Z"
//	    }
//	  ],
//	  "total_matches": 1,
//	  "filtered_from": 150,
//	  "timestamp": "2025-01-13T14:30:22Z"
//	}
type FilterEventsResult struct {
	// Events contains the filtered events
	Events []devtools.EventRecord `json:"events"`

	// TotalMatches is the number of events that matched the filter
	TotalMatches int `json:"total_matches"`

	// FilteredFrom is the total number of events before filtering
	FilteredFrom int `json:"filtered_from"`

	// Timestamp is when the filter was performed
	Timestamp time.Time `json:"timestamp"`
}

// RegisterSearchComponentsTool registers the search_components tool with the MCP server.
//
// This tool allows AI agents to search for components by name, type, or ID with
// fuzzy matching support. Results are ranked by relevance.
//
// The tool is registered with JSON Schema validation for parameters.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses MCP SDK's thread-safe registration.
//
// Example:
//
//	server, _ := NewMCPServer(cfg, dt)
//	err := server.RegisterSearchComponentsTool()
//	if err != nil {
//	    log.Fatalf("Failed to register search components tool: %v", err)
//	}
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
func (s *MCPServer) RegisterSearchComponentsTool() error {
	// Panic recovery with observability integration
	defer func() {
		if r := recover(); r != nil {
			if reporter := observability.GetErrorReporter(); reporter != nil {
				panicErr := &observability.HandlerPanicError{
					ComponentName: "MCPServer",
					EventName:     "RegisterSearchComponentsTool",
					PanicValue:    r,
				}

				ctx := &observability.ErrorContext{
					ComponentName: "MCPServer",
					EventName:     "RegisterSearchComponentsTool",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"tool": "search_components",
					},
				}

				reporter.ReportPanic(panicErr, ctx)
			}
		}
	}()

	// Define tool metadata
	tool := &mcp.Tool{
		Name:        "search_components",
		Description: "Search for components by name, type, or ID with fuzzy matching. Returns ranked results.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search term to match against component fields",
				},
				"fields": map[string]interface{}{
					"type":        "array",
					"description": "Fields to search in: 'name', 'type', 'id'. Empty = all fields",
					"items": map[string]interface{}{
						"type": "string",
						"enum": []string{"name", "type", "id"},
					},
				},
				"max_results": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of results to return (default: 50)",
					"minimum":     1,
					"maximum":     1000,
				},
			},
			"required": []string{"query"},
		},
	}

	// Register tool handler (AddTool doesn't return error)
	s.server.AddTool(tool, s.handleSearchComponentsTool)

	return nil
}

// RegisterFilterEventsTool registers the filter_events tool with the MCP server.
//
// This tool allows AI agents to filter events by name, source, and time range.
// Results can be limited to prevent overwhelming responses.
//
// The tool is registered with JSON Schema validation for parameters.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses MCP SDK's thread-safe registration.
//
// Example:
//
//	server, _ := NewMCPServer(cfg, dt)
//	err := server.RegisterFilterEventsTool()
//	if err != nil {
//	    log.Fatalf("Failed to register filter events tool: %v", err)
//	}
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
func (s *MCPServer) RegisterFilterEventsTool() error {
	// Panic recovery with observability integration
	defer func() {
		if r := recover(); r != nil {
			if reporter := observability.GetErrorReporter(); reporter != nil {
				panicErr := &observability.HandlerPanicError{
					ComponentName: "MCPServer",
					EventName:     "RegisterFilterEventsTool",
					PanicValue:    r,
				}

				ctx := &observability.ErrorContext{
					ComponentName: "MCPServer",
					EventName:     "RegisterFilterEventsTool",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"tool": "filter_events",
					},
				}

				reporter.ReportPanic(panicErr, ctx)
			}
		}
	}()

	// Define tool metadata
	tool := &mcp.Tool{
		Name:        "filter_events",
		Description: "Filter events by name, source component, and time range. Returns matching events.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"event_names": map[string]interface{}{
					"type":        "array",
					"description": "Filter by event names (empty = all events)",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"source_ids": map[string]interface{}{
					"type":        "array",
					"description": "Filter by source component IDs (empty = all sources)",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"start_time": map[string]interface{}{
					"type":        "string",
					"description": "Filter events after this time (ISO 8601 format)",
					"format":      "date-time",
				},
				"end_time": map[string]interface{}{
					"type":        "string",
					"description": "Filter events before this time (ISO 8601 format)",
					"format":      "date-time",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of results to return (default: 100)",
					"minimum":     1,
					"maximum":     10000,
				},
			},
		},
	}

	// Register tool handler (AddTool doesn't return error)
	s.server.AddTool(tool, s.handleFilterEventsTool)

	return nil
}

// handleSearchComponentsTool handles the search_components tool execution.
//
// This is the internal handler called by the MCP SDK when an AI agent
// invokes the search_components tool.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses DevToolsStore's thread-safe methods.
func (s *MCPServer) handleSearchComponentsTool(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Panic recovery with observability integration
	defer func() {
		if r := recover(); r != nil {
			if reporter := observability.GetErrorReporter(); reporter != nil {
				panicErr := &observability.HandlerPanicError{
					ComponentName: "MCPServer",
					EventName:     "handleSearchComponentsTool",
					PanicValue:    r,
				}

				errorCtx := &observability.ErrorContext{
					ComponentName: "MCPServer",
					EventName:     "handleSearchComponentsTool",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"tool": "search_components",
					},
				}

				reporter.ReportPanic(panicErr, errorCtx)
			}
		}
	}()

	// Unmarshal JSON parameters
	var argsMap map[string]interface{}
	if err := json.Unmarshal(request.Params.Arguments, &argsMap); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to parse parameters: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Parse parameters
	params, err := parseSearchComponentsParams(argsMap)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Invalid parameters: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Get all components from store
	allComponents := s.store.GetAllComponents()

	// Perform search with fuzzy matching
	matches := searchComponents(allComponents, params)

	// Build result
	result := SearchComponentsResult{
		Matches:      matches,
		TotalMatches: len(matches),
		Query:        params.Query,
		Timestamp:    time.Now(),
	}

	// Return result as text
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: formatSearchComponentsResult(result),
			},
		},
		IsError: false,
	}, nil
}

// handleFilterEventsTool handles the filter_events tool execution.
//
// This is the internal handler called by the MCP SDK when an AI agent
// invokes the filter_events tool.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses DevToolsStore's thread-safe methods.
func (s *MCPServer) handleFilterEventsTool(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Panic recovery with observability integration
	defer func() {
		if r := recover(); r != nil {
			if reporter := observability.GetErrorReporter(); reporter != nil {
				panicErr := &observability.HandlerPanicError{
					ComponentName: "MCPServer",
					EventName:     "handleFilterEventsTool",
					PanicValue:    r,
				}

				errorCtx := &observability.ErrorContext{
					ComponentName: "MCPServer",
					EventName:     "handleFilterEventsTool",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"tool": "filter_events",
					},
				}

				reporter.ReportPanic(panicErr, errorCtx)
			}
		}
	}()

	// Unmarshal JSON parameters
	var argsMap map[string]interface{}
	if err := json.Unmarshal(request.Params.Arguments, &argsMap); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to parse parameters: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Parse parameters
	params, err := parseFilterEventsParams(argsMap)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Invalid parameters: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Get all events from store
	eventLog := s.store.GetEventLog()
	allEvents := eventLog.GetRecent(eventLog.Len())

	// Apply filters
	filteredEvents := filterEvents(allEvents, params)

	// Build result
	result := FilterEventsResult{
		Events:       filteredEvents,
		TotalMatches: len(filteredEvents),
		FilteredFrom: len(allEvents),
		Timestamp:    time.Now(),
	}

	// Return result as text
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: formatFilterEventsResult(result),
			},
		},
		IsError: false,
	}, nil
}

// parseSearchComponentsParams parses the raw parameters into SearchComponentsParams struct.
func parseSearchComponentsParams(args map[string]interface{}) (*SearchComponentsParams, error) {
	params := &SearchComponentsParams{
		Fields:     []string{}, // Default: search all fields
		MaxResults: 50,         // Default: 50 results
	}

	// Parse query (required)
	if query, ok := args["query"].(string); ok {
		params.Query = query
	} else {
		return nil, fmt.Errorf("query parameter is required and must be a string")
	}

	// Parse fields (optional)
	if fields, ok := args["fields"].([]interface{}); ok {
		for _, field := range fields {
			if fieldStr, ok := field.(string); ok {
				params.Fields = append(params.Fields, fieldStr)
			}
		}
	}

	// Parse max_results (optional)
	if maxResults, ok := args["max_results"].(float64); ok {
		params.MaxResults = int(maxResults)
	}

	// Validate max_results
	if params.MaxResults < 1 || params.MaxResults > 1000 {
		return nil, fmt.Errorf("max_results must be between 1 and 1000")
	}

	return params, nil
}

// parseFilterEventsParams parses the raw parameters into FilterEventsParams struct.
func parseFilterEventsParams(args map[string]interface{}) (*FilterEventsParams, error) {
	params := &FilterEventsParams{
		EventNames: []string{},
		SourceIDs:  []string{},
		Limit:      100, // Default: 100 results
	}

	// Parse event_names (optional)
	if eventNames, ok := args["event_names"].([]interface{}); ok {
		for _, name := range eventNames {
			if nameStr, ok := name.(string); ok {
				params.EventNames = append(params.EventNames, nameStr)
			}
		}
	}

	// Parse source_ids (optional)
	if sourceIDs, ok := args["source_ids"].([]interface{}); ok {
		for _, id := range sourceIDs {
			if idStr, ok := id.(string); ok {
				params.SourceIDs = append(params.SourceIDs, idStr)
			}
		}
	}

	// Parse start_time (optional)
	if startTimeStr, ok := args["start_time"].(string); ok {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start_time format: %w", err)
		}
		params.StartTime = &startTime
	}

	// Parse end_time (optional)
	if endTimeStr, ok := args["end_time"].(string); ok {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time format: %w", err)
		}
		params.EndTime = &endTime
	}

	// Parse limit (optional)
	if limit, ok := args["limit"].(float64); ok {
		params.Limit = int(limit)
	}

	// Validate limit
	if params.Limit < 1 || params.Limit > 10000 {
		return nil, fmt.Errorf("limit must be between 1 and 10000")
	}

	return params, nil
}

// checkFieldMatch checks if a field matches the query and updates best score.
func checkFieldMatch(fieldValue, query, fieldName string, currentScore float64, currentField string) (float64, string) {
	if strings.Contains(strings.ToLower(fieldValue), query) {
		score := calculateMatchScore(fieldValue, query)
		if score > currentScore {
			return score, fieldName
		}
	}
	return currentScore, currentField
}

// sortMatchesByScore sorts matches in descending order by score.
func sortMatchesByScore(matches []ComponentMatch) {
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].MatchScore > matches[i].MatchScore {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
}

// searchComponents performs fuzzy search on components.
func searchComponents(components []*devtools.ComponentSnapshot, params *SearchComponentsParams) []ComponentMatch {
	matches := make([]ComponentMatch, 0)
	query := strings.ToLower(params.Query)

	searchAll := len(params.Fields) == 0
	searchName := searchAll || contains(params.Fields, "name")
	searchType := searchAll || contains(params.Fields, "type")
	searchID := searchAll || contains(params.Fields, "id")

	for _, comp := range components {
		var matchScore float64
		var matchedField string

		if searchName {
			matchScore, matchedField = checkFieldMatch(comp.Name, query, "name", matchScore, matchedField)
		}
		if searchType {
			matchScore, matchedField = checkFieldMatch(comp.Type, query, "type", matchScore, matchedField)
		}
		if searchID {
			matchScore, matchedField = checkFieldMatch(comp.ID, query, "id", matchScore, matchedField)
		}

		if matchScore > 0 {
			matches = append(matches, ComponentMatch{
				ID:           comp.ID,
				Name:         comp.Name,
				Type:         comp.Type,
				Status:       comp.Status,
				MatchScore:   matchScore,
				MatchedField: matchedField,
			})
		}
	}

	sortMatchesByScore(matches)

	if len(matches) > params.MaxResults {
		matches = matches[:params.MaxResults]
	}

	return matches
}

// filterEvents filters events based on the provided criteria.
func filterEvents(events []devtools.EventRecord, params *FilterEventsParams) []devtools.EventRecord {
	filtered := make([]devtools.EventRecord, 0)

	for _, event := range events {
		// Filter by event name
		if len(params.EventNames) > 0 && !contains(params.EventNames, event.Name) {
			continue
		}

		// Filter by source ID
		if len(params.SourceIDs) > 0 && !contains(params.SourceIDs, event.SourceID) {
			continue
		}

		// Filter by start time
		if params.StartTime != nil && event.Timestamp.Before(*params.StartTime) {
			continue
		}

		// Filter by end time
		if params.EndTime != nil && event.Timestamp.After(*params.EndTime) {
			continue
		}

		// Event passed all filters
		filtered = append(filtered, event)

		// Check limit
		if len(filtered) >= params.Limit {
			break
		}
	}

	return filtered
}

// calculateMatchScore calculates a relevance score for a match.
// Returns a value between 0.0 and 1.0, where 1.0 is an exact match.
func calculateMatchScore(text, query string) float64 {
	textLower := strings.ToLower(text)
	queryLower := strings.ToLower(query)

	// Exact match
	if textLower == queryLower {
		return 1.0
	}

	// Starts with query
	if strings.HasPrefix(textLower, queryLower) {
		return 0.9
	}

	// Contains query
	if strings.Contains(textLower, queryLower) {
		// Score based on position and length
		index := strings.Index(textLower, queryLower)
		positionScore := 1.0 - (float64(index) / float64(len(textLower)))
		lengthScore := float64(len(queryLower)) / float64(len(textLower))
		return 0.5 + (positionScore * 0.2) + (lengthScore * 0.2)
	}

	return 0.0
}

// formatSearchComponentsResult formats the search result as human-readable text.
func formatSearchComponentsResult(result SearchComponentsResult) string {
	if result.TotalMatches == 0 {
		return fmt.Sprintf("No components found matching '%s'", result.Query)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d component(s) matching '%s':\n\n", result.TotalMatches, result.Query))

	for i, match := range result.Matches {
		sb.WriteString(fmt.Sprintf("%d. %s (ID: %s)\n", i+1, match.Name, match.ID))
		sb.WriteString(fmt.Sprintf("   Type: %s\n", match.Type))
		sb.WriteString(fmt.Sprintf("   Status: %s\n", match.Status))
		sb.WriteString(fmt.Sprintf("   Match: %.2f (field: %s)\n", match.MatchScore, match.MatchedField))
		if i < len(result.Matches)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// formatFilterEventsResult formats the filter result as human-readable text.
func formatFilterEventsResult(result FilterEventsResult) string {
	if result.TotalMatches == 0 {
		return fmt.Sprintf("No events found matching the filter criteria (filtered from %d total events)", result.FilteredFrom)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d event(s) matching filter (from %d total):\n\n", result.TotalMatches, result.FilteredFrom))

	for i, event := range result.Events {
		sb.WriteString(fmt.Sprintf("%d. %s (ID: %s)\n", i+1, event.Name, event.ID))
		sb.WriteString(fmt.Sprintf("   Source: %s\n", event.SourceID))
		if event.TargetID != "" {
			sb.WriteString(fmt.Sprintf("   Target: %s\n", event.TargetID))
		}
		sb.WriteString(fmt.Sprintf("   Time: %s\n", event.Timestamp.Format(time.RFC3339)))
		if event.Duration > 0 {
			sb.WriteString(fmt.Sprintf("   Duration: %v\n", event.Duration))
		}
		if i < len(result.Events)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
