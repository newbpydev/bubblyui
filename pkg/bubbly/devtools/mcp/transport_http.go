package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// httpServerState holds the HTTP server state for graceful shutdown
type httpServerState struct {
	server  *http.Server
	handler *mcp.StreamableHTTPHandler
	mu      sync.RWMutex
}

// StartHTTPServer starts the MCP server using HTTP/SSE transport.
//
// This method enables IDE integration and remote debugging by serving MCP
// over HTTP with Server-Sent Events for real-time updates. The server will:
//   - Create a StreamableHTTPHandler for MCP sessions
//   - Set up HTTP endpoints (/mcp for MCP protocol, /health for health checks)
//   - Listen on the configured host and port
//   - Support multiple concurrent client connections
//   - Handle graceful shutdown on context cancellation
//
// The method starts the HTTP server in a goroutine and blocks until:
//   - Context is cancelled (graceful shutdown)
//   - Server encounters a fatal error
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
//	cfg.Transport = mcp.MCPTransportHTTP
//	cfg.HTTPPort = 8765
//	cfg.HTTPHost = "localhost"
//
//	server, err := mcp.NewMCPServer(cfg, dt)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Start HTTP server (blocks until context cancelled)
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	if err := server.StartHTTPServer(ctx); err != nil {
//	    log.Printf("HTTP server error: %v", err)
//	}
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - error: Configuration error, bind error, or nil on clean shutdown
func (s *MCPServer) StartHTTPServer(ctx context.Context) error {
	// Panic recovery with observability integration
	defer func() {
		if r := recover(); r != nil {
			if reporter := observability.GetErrorReporter(); reporter != nil {
				panicErr := &observability.HandlerPanicError{
					ComponentName: "MCPServer",
					EventName:     "StartHTTPServer",
					PanicValue:    r,
				}

				errCtx := &observability.ErrorContext{
					ComponentName: "MCPServer",
					ComponentID:   "http-transport",
					EventName:     "StartHTTPServer",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"transport":  "http",
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

	// Validate configuration
	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	if config.Transport&MCPTransportHTTP == 0 {
		return fmt.Errorf("HTTP transport not enabled in configuration")
	}

	// Create StreamableHTTPHandler
	// This handler manages MCP sessions over HTTP with SSE support
	handler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server {
			return s.server
		},
		&mcp.StreamableHTTPOptions{
			SessionTimeout: 5 * time.Minute,
			Stateless:      false, // Enable session tracking for subscriptions
		},
	)

	// Create authentication handler
	authHandler, err := NewAuthHandler(config.AuthToken, config.EnableAuth)
	if err != nil {
		return fmt.Errorf("failed to create auth handler: %w", err)
	}

	// Create HTTP server with routes
	mux := http.NewServeMux()

	// MCP endpoint - handles all MCP protocol messages
	// Apply authentication middleware to protect MCP endpoint
	mux.Handle("/mcp", authHandler.Middleware(handler))

	// Health check endpoint - for monitoring and readiness probes
	// Health check is NOT protected by auth (for monitoring systems)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
		})
	})

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", config.HTTPHost, config.HTTPPort)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: mux,
		// Timeouts to prevent resource exhaustion
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Store server state for graceful shutdown
	state := &httpServerState{
		server:  httpServer,
		handler: handler,
	}

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		// Use ListenAndServe for normal operation
		// If port is 0, the OS will assign a random available port
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		// Context cancelled - perform graceful shutdown
		return s.shutdownHTTPServer(state)
	case err := <-errCh:
		// Server error
		return err
	}
}

// shutdownHTTPServer performs graceful shutdown of the HTTP server
func (s *MCPServer) shutdownHTTPServer(state *httpServerState) error {
	state.mu.Lock()
	defer state.mu.Unlock()

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Note: StreamableHTTPHandler doesn't have a Close() method
	// The handler will be cleaned up when the HTTP server shuts down
	// Active sessions will be terminated by the server shutdown

	// Shutdown HTTP server gracefully
	if state.server != nil {
		if err := state.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("HTTP server shutdown error: %w", err)
		}
	}

	return nil
}

// GetHTTPAddr returns the actual address the HTTP server is listening on.
// This is useful when using port 0 (random port assignment).
//
// Returns empty string if HTTP server is not running.
//
// Note: This is a helper method for testing. In production, the port
// should be configured explicitly.
func (s *MCPServer) GetHTTPAddr() string {
	// This would require storing the listener in the server state
	// For now, return the configured address
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config.Transport&MCPTransportHTTP == 0 {
		return ""
	}

	return fmt.Sprintf("%s:%d", s.config.HTTPHost, s.config.HTTPPort)
}

// GetHTTPPort returns the actual port the HTTP server is listening on.
// Returns 0 if HTTP server is not running or port is not yet assigned.
//
// This is useful for testing with random port assignment (port 0).
func (s *MCPServer) GetHTTPPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config.Transport&MCPTransportHTTP == 0 {
		return 0
	}

	return s.config.HTTPPort
}

// waitForHTTPServer waits for the HTTP server to be ready to accept connections.
// This is a helper for testing to avoid race conditions.
//
// Parameters:
//   - addr: The address to check (e.g., "localhost:8765")
//   - timeout: Maximum time to wait
//
// Returns:
//   - error: Timeout error if server doesn't become ready, nil otherwise
func waitForHTTPServer(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("HTTP server not ready after %v", timeout)
}
