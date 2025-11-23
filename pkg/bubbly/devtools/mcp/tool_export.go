package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// ExportParams defines the parameters for the export_session tool.
//
// This structure is used by AI agents to specify export options when
// calling the export_session tool via MCP.
//
// Example:
//
//	{
//	  "format": "json",
//	  "compress": true,
//	  "sanitize": true,
//	  "include": ["components", "state", "events"],
//	  "destination": "/tmp/debug-session.json.gz"
//	}
type ExportParams struct {
	// Format specifies the export format: "json", "yaml", or "msgpack"
	Format string `json:"format"`

	// Compress enables gzip compression of the export file
	Compress bool `json:"compress"`

	// Sanitize enables redaction of sensitive data
	Sanitize bool `json:"sanitize"`

	// Include specifies which data sections to include
	// Valid values: "components", "state", "events", "performance"
	Include []string `json:"include"`

	// Destination is the file path where the export will be saved
	// Use "stdout" to write to standard output (not recommended for large exports)
	Destination string `json:"destination"`
}

// ExportResult contains the result of an export operation.
//
// This structure is returned to AI agents after a successful export.
//
// Example:
//
//	{
//	  "path": "/tmp/debug-session.json.gz",
//	  "size": 245678,
//	  "format": "json",
//	  "compressed": true,
//	  "timestamp": "2025-01-13T14:30:22Z"
//	}
type ExportResult struct {
	// Path is the absolute path to the exported file
	Path string `json:"path"`

	// Size is the file size in bytes
	Size int64 `json:"size"`

	// Format is the export format used
	Format string `json:"format"`

	// Compressed indicates if the file is gzip compressed
	Compressed bool `json:"compressed"`

	// Timestamp is when the export was created
	Timestamp time.Time `json:"timestamp"`
}

// RegisterExportTool registers the export_session tool with the MCP server.
//
// This tool allows AI agents to export debug data with compression and
// sanitization. It supports multiple formats (JSON, YAML, MessagePack) and
// selective inclusion of data sections.
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
//	err := server.RegisterExportTool()
//	if err != nil {
//	    log.Fatalf("Failed to register export tool: %v", err)
//	}
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
func (s *Server) RegisterExportTool() error {
	// Panic recovery with observability integration
	defer func() {
		if r := recover(); r != nil {
			if reporter := observability.GetErrorReporter(); reporter != nil {
				panicErr := &observability.HandlerPanicError{
					ComponentName: "Server",
					EventName:     "RegisterExportTool",
					PanicValue:    r,
				}

				ctx := &observability.ErrorContext{
					ComponentName: "Server",
					EventName:     "RegisterExportTool",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"tool": "export_session",
					},
				}

				reporter.ReportPanic(panicErr, ctx)
			}
		}
	}()

	// Define tool metadata
	tool := &mcp.Tool{
		Name:        "export_session",
		Description: "Export debug session data with compression and sanitization. Supports JSON, YAML, and MessagePack formats.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"format": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"json", "yaml", "msgpack"},
					"description": "Export format (json, yaml, or msgpack)",
					"default":     "json",
				},
				"compress": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable gzip compression",
					"default":     false,
				},
				"sanitize": map[string]interface{}{
					"type":        "boolean",
					"description": "Redact sensitive data (passwords, tokens, etc.)",
					"default":     false,
				},
				"include": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string", "enum": []string{"components", "state", "events", "performance"}},
					"description": "Data sections to include",
					"default":     []string{"components", "state", "events", "performance"},
				},
				"destination": map[string]interface{}{
					"type":        "string",
					"description": "File path for the export (use 'stdout' for standard output)",
				},
			},
			"required": []string{"destination"},
		},
	}

	// Register tool handler (AddTool doesn't return error)
	s.server.AddTool(tool, s.handleExportTool)

	return nil
}

// handleExportTool handles the export_session tool execution.
//
// This is the internal handler called by the MCP SDK when an AI agent
// invokes the export_session tool.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses DevTools' thread-safe export methods.
func (s *Server) handleExportTool(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Panic recovery with observability integration
	defer func() {
		if r := recover(); r != nil {
			if reporter := observability.GetErrorReporter(); reporter != nil {
				panicErr := &observability.HandlerPanicError{
					ComponentName: "Server",
					EventName:     "handleExportTool",
					PanicValue:    r,
				}

				errorCtx := &observability.ErrorContext{
					ComponentName: "Server",
					EventName:     "handleExportTool",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"tool": "export_session",
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
	params, err := parseExportParams(argsMap)
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

	// Validate parameters
	if err := validateExportParams(params); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Validation failed: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Build export options
	opts := devtools.ExportOptions{
		IncludeComponents:  contains(params.Include, "components"),
		IncludeState:       contains(params.Include, "state"),
		IncludeEvents:      contains(params.Include, "events"),
		IncludePerformance: contains(params.Include, "performance"),
		Compress:           params.Compress,
		Sanitize:           params.Sanitize || s.config.SanitizeExports,
	}

	// Determine export path
	exportPath := params.Destination
	if exportPath == "stdout" {
		// For stdout, use a temporary file and read it back
		tmpFile, err := os.CreateTemp("", "bubblyui-export-*.json")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to create temp file: %v", err),
					},
				},
				IsError: true,
			}, nil
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()
		exportPath = tmpFile.Name()
	}

	// Ensure export directory exists
	exportDir := filepath.Dir(exportPath)
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to create export directory: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Perform export using the specified format
	var exportErr error
	if params.Format == "json" {
		// Use standard Export method for JSON
		exportErr = s.devtools.Export(exportPath, opts)
	} else {
		// Use ExportFormat for YAML and MessagePack
		exportErr = s.devtools.ExportFormat(exportPath, params.Format, opts)
	}

	if exportErr != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Export failed: %v", exportErr),
				},
			},
			IsError: true,
		}, nil
	}

	// Get file info
	fileInfo, err := os.Stat(exportPath)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to stat export file: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Build result
	result := ExportResult{
		Path:       exportPath,
		Size:       fileInfo.Size(),
		Format:     params.Format,
		Compressed: params.Compress,
		Timestamp:  time.Now(),
	}

	// If stdout was requested, read file and return content
	if params.Destination == "stdout" {
		content, err := os.ReadFile(exportPath)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to read export file: %v", err),
					},
				},
				IsError: true,
			}, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: string(content),
				},
			},
			IsError: false,
		}, nil
	}

	// Return result as text
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf(
					"Export completed successfully:\n"+
						"  Path: %s\n"+
						"  Size: %d bytes\n"+
						"  Format: %s\n"+
						"  Compressed: %v\n"+
						"  Timestamp: %s",
					result.Path,
					result.Size,
					result.Format,
					result.Compressed,
					result.Timestamp.Format(time.RFC3339),
				),
			},
		},
		IsError: false,
	}, nil
}

// parseExportParams parses the raw parameters into ExportParams struct.
func parseExportParams(args map[string]interface{}) (*ExportParams, error) {
	params := &ExportParams{
		Format:   "json", // Default
		Compress: false,
		Sanitize: false,
	}

	// Only set default include if not explicitly provided
	if _, includeProvided := args["include"]; !includeProvided {
		params.Include = []string{"components", "state", "events", "performance"} // Default: all
	}

	// Parse format
	if format, ok := args["format"].(string); ok {
		params.Format = format
	}

	// Parse compress
	if compress, ok := args["compress"].(bool); ok {
		params.Compress = compress
	}

	// Parse sanitize
	if sanitize, ok := args["sanitize"].(bool); ok {
		params.Sanitize = sanitize
	}

	// Parse include array
	if includeRaw, ok := args["include"]; ok {
		if includeSlice, ok := includeRaw.([]interface{}); ok {
			params.Include = make([]string, 0, len(includeSlice))
			for _, item := range includeSlice {
				if str, ok := item.(string); ok {
					params.Include = append(params.Include, str)
				}
			}
		}
	}

	// Parse destination (required)
	if dest, ok := args["destination"].(string); ok {
		params.Destination = dest
	} else {
		return nil, fmt.Errorf("destination is required")
	}

	return params, nil
}

// validateExportParams validates the export parameters.
func validateExportParams(params *ExportParams) error {
	// Validate format
	validFormats := map[string]bool{
		"json":    true,
		"yaml":    true,
		"msgpack": true,
	}
	if !validFormats[params.Format] {
		return fmt.Errorf("invalid format: %s (must be json, yaml, or msgpack)", params.Format)
	}

	// Validate destination
	if params.Destination == "" {
		return fmt.Errorf("destination cannot be empty")
	}

	// Validate include array
	if len(params.Include) == 0 {
		return fmt.Errorf("include array cannot be empty (must specify at least one section)")
	}

	validSections := map[string]bool{
		"components":  true,
		"state":       true,
		"events":      true,
		"performance": true,
	}

	for _, section := range params.Include {
		if !validSections[section] {
			return fmt.Errorf("invalid section: %s (must be components, state, events, or performance)", section)
		}
	}

	return nil
}

// contains checks if a string slice contains a specific string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
