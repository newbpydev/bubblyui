package main

import (
	"fmt"
	"os"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools/mcp"
)

func main() {
	// Print banner to stderr (MCP protocol requires stdout for JSON-RPC only)
	fmt.Fprintln(os.Stderr, "🎯 MCP Server Example 01: Basic Stdio Transport")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "This example demonstrates:")
	fmt.Fprintln(os.Stderr, "  ✅ MCP server with stdio transport")
	fmt.Fprintln(os.Stderr, "  ✅ Composable architecture (UseCounter)")
	fmt.Fprintln(os.Stderr, "  ✅ BubblyUI components (Card, Button, Badge)")
	fmt.Fprintln(os.Stderr, "  ✅ AI-powered debugging via MCP")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Controls:")
	fmt.Fprintln(os.Stderr, "  space: increment counter")
	fmt.Fprintln(os.Stderr, "  r: reset counter")
	fmt.Fprintln(os.Stderr, "  ctrl+c: quit")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "💡 Connect your AI assistant to inspect this app!")
	fmt.Fprintln(os.Stderr, "   See README.md for setup instructions.")
	fmt.Fprintln(os.Stderr, "")

	// LIMITATION: Stdio transport is fundamentally incompatible with BubblyUI components
	// BubblyUI uses Bubbletea which requires stdin/stdout for TUI
	// MCP stdio transport also requires stdin/stdout for JSON-RPC
	// Result: "could not open a new TTY" errors when running as subprocess
	//
	// Use HTTP transport (02-http-server) instead for working MCP integration
	
	// Enable dev tools with MCP server (stdio transport - BROKEN with BubblyUI)
	_, err := mcp.EnableWithMCP(&mcp.MCPConfig{
		Transport:            mcp.MCPTransportStdio,
		MaxClients:           5,
		RateLimit:            60,
		SubscriptionThrottle: 100 * time.Millisecond,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error enabling MCP: %v\n", err)
		os.Exit(1)
	}

	// Verify MCP is enabled (print to stderr for MCP compatibility)
	if devtools.IsEnabled() {
		fmt.Fprintln(os.Stderr, "✅ MCP server enabled on stdio transport")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "⚠️  NOTE: This example demonstrates the stdio transport limitation.")
		fmt.Fprintln(os.Stderr, "    Use HTTP transport (02-http-server) for working MCP integration.")
		fmt.Fprintln(os.Stderr, "")
	}

	// Block forever - MCP stdio server handles communication
	// NOTE: No BubblyUI components can be used with stdio transport
	// due to stdin/stdout conflicts between Bubbletea and MCP protocol
	select {}
}
