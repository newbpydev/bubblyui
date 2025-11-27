package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// recoverToolRegistration provides panic recovery with observability integration for tool registration.
// This helper function is used by all tool registration methods to avoid code duplication.
func recoverToolRegistration(eventName, toolName string) {
	if r := recover(); r != nil {
		if reporter := observability.GetErrorReporter(); reporter != nil {
			panicErr := &observability.HandlerPanicError{
				ComponentName: "Server",
				EventName:     eventName,
				PanicValue:    r,
			}

			ctx := &observability.ErrorContext{
				ComponentName: "Server",
				EventName:     eventName,
				Timestamp:     time.Now(),
				StackTrace:    debug.Stack(),
				Tags: map[string]string{
					"tool": toolName,
				},
			}

			reporter.ReportPanic(panicErr, ctx)
		}
	}
}

// ClearStateHistoryParams defines the parameters for the clear_state_history tool.
//
// This structure is used by AI agents to specify options when clearing state history.
//
// Example:
//
//	{
//	  "confirm": true
//	}
type ClearStateHistoryParams struct {
	// Confirm is required to be true to prevent accidental deletion
	Confirm bool `json:"confirm"`
}

// ClearEventLogParams defines the parameters for the clear_event_log tool.
//
// This structure is used by AI agents to specify options when clearing the event log.
//
// Example:
//
//	{
//	  "confirm": true
//	}
type ClearEventLogParams struct {
	// Confirm is required to be true to prevent accidental deletion
	Confirm bool `json:"confirm"`
}

// ClearResult contains the result of a clear operation.
//
// This structure is returned to AI agents after a successful clear operation.
//
// Example:
//
//	{
//	  "cleared": 150,
//	  "timestamp": "2025-01-13T14:30:22Z"
//	}
type ClearResult struct {
	// Cleared is the number of items that were cleared
	Cleared int `json:"cleared"`

	// Timestamp is when the clear operation completed
	Timestamp time.Time `json:"timestamp"`
}

// RegisterClearStateHistoryTool registers the clear_state_history tool with the MCP server.
//
// This tool allows AI agents to clear the state change history. It requires explicit
// confirmation to prevent accidental data loss.
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
//	err := server.RegisterClearStateHistoryTool()
//	if err != nil {
//	    log.Fatalf("Failed to register clear state history tool: %v", err)
//	}
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
func (s *Server) RegisterClearStateHistoryTool() error {
	// Panic recovery with observability integration
	defer recoverToolRegistration("RegisterClearStateHistoryTool", "clear_state_history")

	// Define tool metadata
	tool := &mcp.Tool{
		Name:        "clear_state_history",
		Description: "Clear all state change history. Requires confirmation to prevent accidental data loss.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"confirm": map[string]interface{}{
					"type":        "boolean",
					"description": "Must be true to confirm the destructive operation",
				},
			},
			"required": []string{"confirm"},
		},
	}

	// Register tool handler (AddTool doesn't return error)
	s.server.AddTool(tool, s.handleClearStateHistoryTool)

	return nil
}

// RegisterClearEventLogTool registers the clear_event_log tool with the MCP server.
//
// This tool allows AI agents to clear the event log. It requires explicit
// confirmation to prevent accidental data loss.
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
//	err := server.RegisterClearEventLogTool()
//	if err != nil {
//	    log.Fatalf("Failed to register clear event log tool: %v", err)
//	}
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
func (s *Server) RegisterClearEventLogTool() error {
	// Panic recovery with observability integration
	defer recoverToolRegistration("RegisterClearEventLogTool", "clear_event_log")

	// Define tool metadata
	tool := &mcp.Tool{
		Name:        "clear_event_log",
		Description: "Clear all events from the event log. Requires confirmation to prevent accidental data loss.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"confirm": map[string]interface{}{
					"type":        "boolean",
					"description": "Must be true to confirm the destructive operation",
				},
			},
			"required": []string{"confirm"},
		},
	}

	// Register tool handler (AddTool doesn't return error)
	s.server.AddTool(tool, s.handleClearEventLogTool)

	return nil
}

// handleClearStateHistoryTool handles the clear_state_history tool execution.
//
// This is the internal handler called by the MCP SDK when an AI agent
// invokes the clear_state_history tool.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses Store's thread-safe Clear method.
func (s *Server) handleClearStateHistoryTool(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Panic recovery with observability integration
	defer func() {
		if r := recover(); r != nil {
			if reporter := observability.GetErrorReporter(); reporter != nil {
				panicErr := &observability.HandlerPanicError{
					ComponentName: "Server",
					EventName:     "handleClearStateHistoryTool",
					PanicValue:    r,
				}

				errorCtx := &observability.ErrorContext{
					ComponentName: "Server",
					EventName:     "handleClearStateHistoryTool",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"tool": "clear_state_history",
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
	params, err := parseClearStateHistoryParams(argsMap)
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

	// Validate confirmation
	if !params.Confirm {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Confirmation required: set 'confirm' to true to clear state history",
				},
			},
			IsError: true,
		}, nil
	}

	// Get count before clearing
	stateHistory := s.store.GetStateHistory()
	countBefore := len(stateHistory.GetAll())

	// Clear state history (atomic operation)
	stateHistory.Clear()

	// Build result
	result := ClearResult{
		Cleared:   countBefore,
		Timestamp: time.Now(),
	}

	// Return result as text
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf(
					"State history cleared successfully:\n"+
						"  Items cleared: %d\n"+
						"  Timestamp: %s",
					result.Cleared,
					result.Timestamp.Format(time.RFC3339),
				),
			},
		},
		IsError: false,
	}, nil
}

// handleClearEventLogTool handles the clear_event_log tool execution.
//
// This is the internal handler called by the MCP SDK when an AI agent
// invokes the clear_event_log tool.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses Store's thread-safe Clear method.
func (s *Server) handleClearEventLogTool(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Panic recovery with observability integration
	defer func() {
		if r := recover(); r != nil {
			if reporter := observability.GetErrorReporter(); reporter != nil {
				panicErr := &observability.HandlerPanicError{
					ComponentName: "Server",
					EventName:     "handleClearEventLogTool",
					PanicValue:    r,
				}

				errorCtx := &observability.ErrorContext{
					ComponentName: "Server",
					EventName:     "handleClearEventLogTool",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"tool": "clear_event_log",
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
	params, err := parseClearEventLogParams(argsMap)
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

	// Validate confirmation
	if !params.Confirm {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Confirmation required: set 'confirm' to true to clear event log",
				},
			},
			IsError: true,
		}, nil
	}

	// Get count before clearing
	eventLog := s.store.GetEventLog()
	countBefore := eventLog.Len()

	// Clear event log (atomic operation)
	eventLog.Clear()

	// Build result
	result := ClearResult{
		Cleared:   countBefore,
		Timestamp: time.Now(),
	}

	// Return result as text
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf(
					"Event log cleared successfully:\n"+
						"  Items cleared: %d\n"+
						"  Timestamp: %s",
					result.Cleared,
					result.Timestamp.Format(time.RFC3339),
				),
			},
		},
		IsError: false,
	}, nil
}

// parseClearStateHistoryParams parses the raw parameters into ClearStateHistoryParams struct.
func parseClearStateHistoryParams(args map[string]interface{}) (*ClearStateHistoryParams, error) {
	params := &ClearStateHistoryParams{
		Confirm: false, // Default to false for safety
	}

	// Parse confirm
	if confirm, ok := args["confirm"].(bool); ok {
		params.Confirm = confirm
	} else {
		return nil, fmt.Errorf("confirm parameter is required and must be a boolean")
	}

	return params, nil
}

// parseClearEventLogParams parses the raw parameters into ClearEventLogParams struct.
func parseClearEventLogParams(args map[string]interface{}) (*ClearEventLogParams, error) {
	params := &ClearEventLogParams{
		Confirm: false, // Default to false for safety
	}

	// Parse confirm
	if confirm, ok := args["confirm"].(bool); ok {
		params.Confirm = confirm
	} else {
		return nil, fmt.Errorf("confirm parameter is required and must be a boolean")
	}

	return params, nil
}
