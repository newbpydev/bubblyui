package mcp

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// StartStdioServer starts the MCP server using stdio transport.
//
// This method enables local CLI integration by communicating over stdin/stdout
// using newline-delimited JSON-RPC messages. The server will:
//   - Create a stdio transport (uses os.Stdin/os.Stdout)
//   - Connect to the MCP SDK server
//   - Complete the initialization handshake with the client
//   - Negotiate protocol version (2025-06-18)
//   - Declare server capabilities (resources, tools, subscriptions)
//   - Block until the client disconnects or context is canceled
//
// The method blocks until one of the following occurs:
//   - Client disconnects gracefully
//   - Context is canceled
//   - Transport error occurs
//
// Thread Safety:
//
//	Safe to call concurrently (uses internal mutex for state access).
//
// Error Handling:
//
//	All errors are wrapped with context. Panics are recovered and reported
//	to the observability system. MCP failures never crash the host application.
//
// Example:
//
//	dt := devtools.Enable()
//	cfg := mcp.DefaultMCPConfig()
//	server, err := mcp.NewMCPServer(cfg, dt)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Start stdio server (blocks until client disconnects)
//	ctx := context.Background()
//	if err := server.StartStdioServer(ctx); err != nil {
//	    log.Printf("Stdio server error: %v", err)
//	}
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - error: Connection error, session error, or nil on clean shutdown
func (s *Server) StartStdioServer(ctx context.Context) error {
	// Panic recovery with observability integration
	defer func() {
		if r := recover(); r != nil {
			if reporter := observability.GetErrorReporter(); reporter != nil {
				panicErr := &observability.HandlerPanicError{
					ComponentName: "Server",
					EventName:     "StartStdioServer",
					PanicValue:    r,
				}

				errCtx := &observability.ErrorContext{
					ComponentName: "Server",
					ComponentID:   "stdio-transport",
					EventName:     "StartStdioServer",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"transport":  "stdio",
						"error_type": "panic",
					},
					Extra: map[string]interface{}{
						"panic_value": r,
					},
				}

				reporter.ReportPanic(panicErr, errCtx)
			}
		}
	}()

	// Create stdio transport
	// The SDK's StdioTransport uses os.Stdin and os.Stdout automatically
	transport := &mcp.StdioTransport{}

	// Connect server to client via stdio transport
	// This establishes the JSON-RPC connection and performs the handshake
	session, err := s.server.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect stdio transport: %w", err)
	}

	// Wait for session to complete
	// This blocks until:
	// - Client disconnects
	// - Context is canceled
	// - Transport error occurs
	err = session.Wait()
	if err != nil {
		return fmt.Errorf("stdio session ended with error: %w", err)
	}

	// Clean shutdown
	return nil
}
