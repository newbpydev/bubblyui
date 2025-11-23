package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
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
	dt, err := mcp.EnableWithMCP(&mcp.Config{
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

	// Get the MCP server and register resources/tools
	mcpServer := dt.GetMCPServer().(*mcp.Server)

	// Register MCP resources (what AI can inspect)
	if err := mcpServer.RegisterComponentsResource(); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering components resource: %v\n", err)
		os.Exit(1)
	}
	if err := mcpServer.RegisterComponentResource(); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering component resource: %v\n", err)
		os.Exit(1)
	}
	if err := mcpServer.RegisterStateResource(); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering state resource: %v\n", err)
		os.Exit(1)
	}
	if err := mcpServer.RegisterEventsResource(); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering events resource: %v\n", err)
		os.Exit(1)
	}
	if err := mcpServer.RegisterPerformanceResource(); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering performance resource: %v\n", err)
		os.Exit(1)
	}

	// Register MCP tools (what AI can do)
	if err := mcpServer.RegisterSearchComponentsTool(); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering search tool: %v\n", err)
		os.Exit(1)
	}
	if err := mcpServer.RegisterFilterEventsTool(); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering filter tool: %v\n", err)
		os.Exit(1)
	}
	if err := mcpServer.RegisterExportTool(); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering export tool: %v\n", err)
		os.Exit(1)
	}
	if err := mcpServer.RegisterClearStateHistoryTool(); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering clear state tool: %v\n", err)
		os.Exit(1)
	}
	if err := mcpServer.RegisterClearEventLogTool(); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering clear events tool: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "📋 Registered 5 resources and 5 tools for MCP")

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

	// Run Bubbletea TUI - MCP HTTP server runs concurrently in goroutine
	// The TUI renders in terminal while MCP serves HTTP requests on port 8765
	// User can interact with TUI (keyboard) while AI inspects state via MCP
	p := tea.NewProgram(
		bubbly.Wrap(app),
		tea.WithAltScreen(), // Full screen TUI
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
