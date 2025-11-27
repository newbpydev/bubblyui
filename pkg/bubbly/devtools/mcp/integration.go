package mcp

import (
	"fmt"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

// EnableWithMCP creates and enables the dev tools singleton with MCP server integration.
//
// This function combines devtools.Enable() with MCP server initialization, providing
// a convenient way to start dev tools with AI-assisted debugging capabilities enabled.
//
// The MCP server exposes devtools data via the Model Context Protocol, allowing AI agents
// to inspect components, state, events, and performance metrics in real-time.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	// Enable with stdio transport (default)
//	cfg := mcp.DefaultMCPConfig()
//	dt, err := mcp.EnableWithMCP(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Enable with HTTP transport
//	cfg := &mcp.Config{
//	    Transport:  mcp.MCPTransportHTTP,
//	    HTTPPort:   8765,
//	    EnableAuth: true,
//	    AuthToken:  "secret-token",
//	}
//	dt, err := mcp.EnableWithMCP(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - config: MCP server configuration
//
// Returns:
//   - *devtools.DevTools: The singleton dev tools instance with MCP enabled
//   - error: Configuration validation error, or MCP server creation error
func EnableWithMCP(config *Config) (*devtools.DevTools, error) {
	// Validate config
	if config == nil {
		return nil, fmt.Errorf("mcp config cannot be nil")
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid mcp config: %w", err)
	}

	// Enable dev tools (or get existing instance)
	dt := devtools.Enable()

	// Check if MCP is already enabled
	if dt.GetMCPServer() != nil {
		// MCP already enabled, return existing instance
		return dt, nil
	}

	// Create MCP server
	mcpServer, err := NewMCPServer(config, dt)
	if err != nil {
		return nil, fmt.Errorf("failed to create mcp server: %w", err)
	}

	// Store MCP server in DevTools
	dt.SetMCPServer(mcpServer)

	return dt, nil
}
