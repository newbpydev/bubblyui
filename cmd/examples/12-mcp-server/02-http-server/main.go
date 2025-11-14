package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools/mcp"
)

func main() {
	// Print banner to stderr (keeps stdout clean for potential future stdio use)
	fmt.Fprintln(os.Stderr, "🎯 MCP Server Example 02: HTTP Transport with Auth")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "This example demonstrates:")
	fmt.Fprintln(os.Stderr, "  ✅ MCP server with HTTP/SSE transport")
	fmt.Fprintln(os.Stderr, "  ✅ Authentication with bearer tokens")
	fmt.Fprintln(os.Stderr, "  ✅ Real-time state updates")
	fmt.Fprintln(os.Stderr, "  ✅ Multiple concurrent AI clients")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Controls:")
	fmt.Fprintln(os.Stderr, "  ctrl+n: new todo")
	fmt.Fprintln(os.Stderr, "  space: toggle completion")
	fmt.Fprintln(os.Stderr, "  ctrl+d: delete todo")
	fmt.Fprintln(os.Stderr, "  ↑/↓: navigate todos")
	fmt.Fprintln(os.Stderr, "  ctrl+c: quit")
	fmt.Fprintln(os.Stderr, "")

	// Enable dev tools with MCP server (HTTP transport)
	dt, err := mcp.EnableWithMCP(&mcp.MCPConfig{
		Transport:            mcp.MCPTransportHTTP,
		HTTPPort:             8765,
		HTTPHost:             "localhost",
		EnableAuth:           true,
		AuthToken:            "demo-token-12345",
		MaxClients:           5,
		RateLimit:            60,
		SubscriptionThrottle: 100 * time.Millisecond,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error enabling MCP: %v\n", err)
		os.Exit(1)
	}

	// Verify MCP is enabled (print to stderr)
	if devtools.IsEnabled() {
		fmt.Fprintln(os.Stderr, "✅ MCP server enabled on http://localhost:8765")
		fmt.Fprintln(os.Stderr, "🔐 Auth token: demo-token-12345")
		fmt.Fprintln(os.Stderr, "💡 Connect your AI assistant to inspect this app!")
		fmt.Fprintln(os.Stderr, "   See README.md for setup instructions.")
		fmt.Fprintln(os.Stderr, "")
	}

	// Get the MCP server and start it
	mcpServer := dt.GetMCPServer().(*mcp.MCPServer)
	
	// Start HTTP server in goroutine (blocks until shutdown)
	ctx := context.Background()
	go func() {
		if err := mcpServer.StartHTTPServer(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		}
	}()
	
	fmt.Fprintln(os.Stderr, "🔌 MCP server started on http://localhost:8765")
	fmt.Fprintln(os.Stderr, "   Ready for AI assistant connections!")
	
	// Create app component with state for MCP to expose
	app, err := CreateApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}

	// Initialize component state (MCP can now inspect this)
	app.Init()

	// Block forever - MCP server runs in background with component state
	select {}
}
