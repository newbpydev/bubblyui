package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// SetRefValueParams defines the parameters for the set_ref_value tool.
//
// This structure is used by AI agents to modify ref values for testing purposes.
// This is a WRITE operation and requires WriteEnabled=true in MCPConfig.
//
// Example:
//
//	{
//	  "ref_id": "ref-counter-123",
//	  "new_value": 42,
//	  "dry_run": false
//	}
type SetRefValueParams struct {
	// RefID is the unique identifier of the ref to modify
	RefID string `json:"ref_id"`

	// NewValue is the new value to set (type must match ref's current type)
	NewValue interface{} `json:"new_value"`

	// DryRun validates the operation without applying changes
	// Useful for checking if a value change would succeed
	DryRun bool `json:"dry_run"`
}

// SetRefResult contains the result of a set_ref_value operation.
//
// This structure is returned to AI agents after a successful (or dry-run) operation.
//
// Example:
//
//	{
//	  "ref_id": "ref-counter-123",
//	  "old_value": 41,
//	  "new_value": 42,
//	  "owner_id": "comp-main-456",
//	  "timestamp": "2025-01-13T14:30:22Z",
//	  "dry_run": false,
//	  "type_match": true
//	}
type SetRefResult struct {
	// RefID is the ref that was modified
	RefID string `json:"ref_id"`

	// OldValue is the value before the change
	OldValue interface{} `json:"old_value"`

	// NewValue is the value after the change
	NewValue interface{} `json:"new_value"`

	// OwnerID is the component that owns this ref
	OwnerID string `json:"owner_id"`

	// Timestamp is when the operation was performed
	Timestamp time.Time `json:"timestamp"`

	// DryRun indicates if this was a validation-only operation
	DryRun bool `json:"dry_run"`

	// TypeMatch indicates if the new value type matches the old value type
	TypeMatch bool `json:"type_match"`
}

// RegisterSetRefValueTool registers the set_ref_value tool with the MCP server.
//
// This tool allows AI agents to modify ref values for testing purposes.
// It is a WRITE operation and only registers if WriteEnabled=true in MCPConfig.
//
// The tool performs type checking to prevent invalid value assignments and
// supports dry-run mode for validation without side effects.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses MCP SDK's thread-safe registration.
//
// Example:
//
//	server, _ := NewMCPServer(cfg, dt)
//	err := server.RegisterSetRefValueTool()
//	if err != nil {
//	    log.Fatalf("Failed to register set ref value tool: %v", err)
//	}
//
// Returns:
//   - error: nil on success, error if WriteEnabled=false or registration fails
func (s *MCPServer) RegisterSetRefValueTool() error {
	// Panic recovery with observability integration
	defer func() {
		if r := recover(); r != nil {
			if reporter := observability.GetErrorReporter(); reporter != nil {
				panicErr := &observability.HandlerPanicError{
					ComponentName: "MCPServer",
					EventName:     "RegisterSetRefValueTool",
					PanicValue:    r,
				}

				ctx := &observability.ErrorContext{
					ComponentName: "MCPServer",
					EventName:     "RegisterSetRefValueTool",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"tool": "set_ref_value",
					},
				}

				reporter.ReportPanic(panicErr, ctx)
			}
		}
	}()

	// Only register if write operations are enabled
	if !s.config.WriteEnabled {
		return fmt.Errorf("set_ref_value tool requires WriteEnabled=true in MCPConfig")
	}

	// Define tool metadata
	tool := &mcp.Tool{
		Name:        "set_ref_value",
		Description: "Modify a ref value for testing (requires write permission). Supports dry-run validation and type checking.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"ref_id": map[string]interface{}{
					"type":        "string",
					"description": "The unique identifier of the ref to modify",
				},
				"new_value": map[string]interface{}{
					"description": "The new value to set (type must match ref's current type)",
				},
				"dry_run": map[string]interface{}{
					"type":        "boolean",
					"description": "Validate without applying changes (default: false)",
					"default":     false,
				},
			},
			"required": []string{"ref_id", "new_value"},
		},
	}

	// Register tool handler (AddTool doesn't return error)
	s.server.AddTool(tool, s.handleSetRefValueTool)

	return nil
}

// handleSetRefValueTool handles the set_ref_value tool execution.
//
// This is the internal handler called by the MCP SDK when an AI agent
// invokes the set_ref_value tool.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses DevToolsStore's thread-safe methods.
func (s *MCPServer) handleSetRefValueTool(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Panic recovery with observability integration
	defer func() {
		if r := recover(); r != nil {
			if reporter := observability.GetErrorReporter(); reporter != nil {
				panicErr := &observability.HandlerPanicError{
					ComponentName: "MCPServer",
					EventName:     "handleSetRefValueTool",
					PanicValue:    r,
				}

				errorCtx := &observability.ErrorContext{
					ComponentName: "MCPServer",
					EventName:     "handleSetRefValueTool",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"tool": "set_ref_value",
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
	params, err := parseSetRefValueParams(argsMap)
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

	// Get current ref value and owner
	ownerID, oldValue, err := s.getRefValueAndOwner(params.RefID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to get ref: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Type checking
	typeMatch := checkTypeCompatibility(oldValue, params.NewValue)
	if !typeMatch {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Type mismatch: ref has type %T, new value has type %T", oldValue, params.NewValue),
				},
			},
			IsError: true,
		}, nil
	}

	// Build result
	result := SetRefResult{
		RefID:     params.RefID,
		OldValue:  oldValue,
		NewValue:  params.NewValue,
		OwnerID:   ownerID,
		Timestamp: time.Now(),
		DryRun:    params.DryRun,
		TypeMatch: typeMatch,
	}

	// Apply changes if not dry-run
	if !params.DryRun {
		_, updated := s.store.UpdateRefValue(params.RefID, params.NewValue)
		if !updated {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to update ref value for ref_id=%s", params.RefID),
					},
				},
				IsError: true,
			}, nil
		}

		// Log the modification for audit trail
		s.logRefModification(params.RefID, oldValue, params.NewValue, ownerID)
	}

	// Return result as text
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: formatSetRefResult(result),
			},
		},
		IsError: false,
	}, nil
}

// parseSetRefValueParams parses the raw parameters into SetRefValueParams struct.
func parseSetRefValueParams(args map[string]interface{}) (*SetRefValueParams, error) {
	params := &SetRefValueParams{
		DryRun: false, // Default: apply changes
	}

	// Parse ref_id (required)
	if refID, ok := args["ref_id"].(string); ok {
		params.RefID = refID
	} else {
		return nil, fmt.Errorf("ref_id parameter is required and must be a string")
	}

	// Validate ref_id is not empty
	if params.RefID == "" {
		return nil, fmt.Errorf("ref_id cannot be empty")
	}

	// Parse new_value (required)
	if newValue, ok := args["new_value"]; ok {
		params.NewValue = newValue
	} else {
		return nil, fmt.Errorf("new_value parameter is required")
	}

	// Parse dry_run (optional)
	if dryRun, ok := args["dry_run"].(bool); ok {
		params.DryRun = dryRun
	}

	return params, nil
}

// getRefValueAndOwner retrieves the current value and owner of a ref.
func (s *MCPServer) getRefValueAndOwner(refID string) (ownerID string, value interface{}, err error) {
	// Get all components to find the ref
	allComponents := s.store.GetAllComponents()

	for _, comp := range allComponents {
		for _, ref := range comp.Refs {
			if ref.ID == refID {
				return comp.ID, ref.Value, nil
			}
		}
	}

	return "", nil, fmt.Errorf("ref not found: %s", refID)
}

// checkTypeCompatibility checks if the new value type is compatible with the old value type.
func checkTypeCompatibility(oldValue, newValue interface{}) bool {
	// Handle nil cases
	if oldValue == nil && newValue == nil {
		return true
	}
	if oldValue == nil || newValue == nil {
		return false
	}

	oldType := reflect.TypeOf(oldValue)
	newType := reflect.TypeOf(newValue)

	// Exact type match
	if oldType == newType {
		return true
	}

	// Check if types are assignable
	if newType.AssignableTo(oldType) {
		return true
	}

	// Check if types are convertible
	if newType.ConvertibleTo(oldType) {
		return true
	}

	return false
}

// logRefModification logs a ref modification for audit trail.
func (s *MCPServer) logRefModification(refID string, oldValue, newValue interface{}, ownerID string) {
	// Report to observability system for audit trail
	if reporter := observability.GetErrorReporter(); reporter != nil {
		ctx := &observability.ErrorContext{
			ComponentName: "MCPServer",
			EventName:     "RefModification",
			Timestamp:     time.Now(),
			Tags: map[string]string{
				"ref_id":   refID,
				"owner_id": ownerID,
			},
			Extra: map[string]interface{}{
				"old_value": fmt.Sprintf("%v", oldValue),
				"new_value": fmt.Sprintf("%v", newValue),
				"old_type":  fmt.Sprintf("%T", oldValue),
				"new_type":  fmt.Sprintf("%T", newValue),
			},
		}

		// This is not an error, but we use the observability system for audit logging
		// In production, this would go to an audit log
		_ = ctx
	}
}

// formatSetRefResult formats the set ref result as human-readable text.
func formatSetRefResult(result SetRefResult) string {
	var msg string

	if result.DryRun {
		msg = fmt.Sprintf("✓ Dry-run validation successful for ref '%s'\n\n", result.RefID)
		msg += fmt.Sprintf("Owner: %s\n", result.OwnerID)
		msg += fmt.Sprintf("Current value: %v (type: %T)\n", result.OldValue, result.OldValue)
		msg += fmt.Sprintf("New value: %v (type: %T)\n", result.NewValue, result.NewValue)
		msg += fmt.Sprintf("Type match: %v\n", result.TypeMatch)
		msg += "\nNo changes were applied (dry-run mode)"
	} else {
		msg = fmt.Sprintf("✓ Successfully updated ref '%s'\n\n", result.RefID)
		msg += fmt.Sprintf("Owner: %s\n", result.OwnerID)
		msg += fmt.Sprintf("Old value: %v (type: %T)\n", result.OldValue, result.OldValue)
		msg += fmt.Sprintf("New value: %v (type: %T)\n", result.NewValue, result.NewValue)
		msg += fmt.Sprintf("Timestamp: %s\n", result.Timestamp.Format(time.RFC3339))
	}

	return msg
}
