package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

// TestStartHTTPServer_Success tests successful HTTP server startup
func TestStartHTTPServer_Success(t *testing.T) {
	// Create devtools and MCP server
	dt := devtools.Enable()

	cfg := DefaultMCPConfig()
	cfg.Transport = MCPTransportHTTP
	cfg.HTTPPort = 0 // Use random available port
	cfg.HTTPHost = "localhost"

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Start HTTP server in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StartHTTPServer(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context to stop server
	cancel()

	// Wait for server to stop
	select {
	case err := <-errCh:
		// Server should stop gracefully (context canceled)
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Server did not stop within timeout")
	}
}

// TestStartHTTPServer_HealthCheck tests the health check endpoint
func TestStartHTTPServer_HealthCheck(t *testing.T) {
	dt := devtools.Enable()

	cfg := DefaultMCPConfig()
	cfg.Transport = MCPTransportHTTP
	cfg.HTTPPort = 0 // Random port
	cfg.HTTPHost = "localhost"

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = server.StartHTTPServer(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	// Test health check (will fail until we implement port discovery)
	// This test documents the expected behavior
	t.Skip("Skipping until port discovery is implemented")
}

// TestStartHTTPServer_InvalidConfig tests error handling for invalid config
func TestStartHTTPServer_InvalidConfig(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func(*MCPConfig)
		wantErr     bool
		errContains string
	}{
		{
			name: "HTTP transport not enabled",
			setupConfig: func(cfg *MCPConfig) {
				cfg.Transport = MCPTransportStdio // Wrong transport
			},
			wantErr:     true,
			errContains: "HTTP transport not enabled",
		},
		{
			name: "Invalid port",
			setupConfig: func(cfg *MCPConfig) {
				cfg.HTTPPort = 0 // Will be assigned by OS
			},
			wantErr: false, // Port 0 is valid (random port)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := devtools.Enable()

			cfg := DefaultMCPConfig()
			cfg.Transport = MCPTransportHTTP
			cfg.HTTPHost = "localhost"
			tt.setupConfig(cfg)

			server, err := NewMCPServer(cfg, dt)
			if tt.name == "HTTP transport not enabled" {
				// Server creation succeeds, but StartHTTPServer should fail
				require.NoError(t, err)

				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel()

				err = server.StartHTTPServer(ctx)
				if tt.wantErr {
					assert.Error(t, err)
					if tt.errContains != "" {
						assert.Contains(t, err.Error(), tt.errContains)
					}
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

// TestStartHTTPServer_GracefulShutdown tests graceful shutdown
func TestStartHTTPServer_GracefulShutdown(t *testing.T) {
	dt := devtools.Enable()

	cfg := DefaultMCPConfig()
	cfg.Transport = MCPTransportHTTP
	cfg.HTTPPort = 0
	cfg.HTTPHost = "localhost"

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StartHTTPServer(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Server should stop gracefully
	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Server did not stop gracefully")
	}
}

// TestStartHTTPServer_ConcurrentAccess tests thread safety
func TestStartHTTPServer_ConcurrentAccess(t *testing.T) {
	dt := devtools.Enable()

	cfg := DefaultMCPConfig()
	cfg.Transport = MCPTransportHTTP
	cfg.HTTPPort = 0
	cfg.HTTPHost = "localhost"

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server
	go func() {
		_ = server.StartHTTPServer(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	// Concurrent access to server methods
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = server.GetConfig()
			_ = server.GetDevTools()
			_ = server.GetStore()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	cancel()
}

// TestStartHTTPServer_MultipleClients tests multiple client connections
func TestStartHTTPServer_MultipleClients(t *testing.T) {
	t.Skip("Requires MCP client implementation - will be tested in integration tests")
}

// TestStartHTTPServer_SessionTimeout tests session timeout handling
func TestStartHTTPServer_SessionTimeout(t *testing.T) {
	t.Skip("Requires MCP client implementation - will be tested in integration tests")
}

// TestHTTPTransport_HealthEndpoint tests the /health endpoint specifically
func TestHTTPTransport_HealthEndpoint(t *testing.T) {
	dt := devtools.Enable()

	cfg := DefaultMCPConfig()
	cfg.Transport = MCPTransportHTTP
	cfg.HTTPPort = 18765 // Fixed port for testing
	cfg.HTTPHost = "localhost"

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server
	go func() {
		_ = server.StartHTTPServer(ctx)
	}()

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	// Test health endpoint
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", cfg.HTTPPort))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var health map[string]string
	err = json.Unmarshal(body, &health)
	require.NoError(t, err)

	assert.Equal(t, "healthy", health["status"])
}

// TestStartHTTPServer_PortInUse tests handling of port already in use
func TestStartHTTPServer_PortInUse(t *testing.T) {
	dt := devtools.Enable()

	cfg := DefaultMCPConfig()
	cfg.Transport = MCPTransportHTTP
	cfg.HTTPPort = 18766 // Fixed port
	cfg.HTTPHost = "localhost"

	// Start first server
	server1, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()

	errCh1 := make(chan error, 1)
	go func() {
		errCh1 <- server1.StartHTTPServer(ctx1)
	}()

	time.Sleep(100 * time.Millisecond)

	// Try to start second server on same port
	server2, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel2()

	err = server2.StartHTTPServer(ctx2)
	// Should fail because port is in use
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "address already in use")

	// Cleanup first server
	cancel1()
	<-errCh1
}

// TestStartHTTPServer_NilContext tests error handling for nil context
func TestStartHTTPServer_NilContext(t *testing.T) {
	dt := devtools.Enable()

	cfg := DefaultMCPConfig()
	cfg.Transport = MCPTransportHTTP
	cfg.HTTPPort = 0
	cfg.HTTPHost = "localhost"

	_, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// This should panic or return error
	// Go's http.Server.Shutdown requires non-nil context
	// We'll document expected behavior
	t.Skip("Nil context handling - implementation decision needed")
}

// TestGetHTTPAddr tests GetHTTPAddr method
func TestGetHTTPAddr(t *testing.T) {
	dt := devtools.Enable()

	tests := []struct {
		name      string
		transport MCPTransportType
		port      int
		host      string
		wantAddr  string
	}{
		{
			name:      "HTTP transport enabled",
			transport: MCPTransportHTTP,
			port:      8765,
			host:      "localhost",
			wantAddr:  "localhost:8765",
		},
		{
			name:      "Stdio transport only",
			transport: MCPTransportStdio,
			port:      8765,
			host:      "localhost",
			wantAddr:  "",
		},
		{
			name:      "Both transports",
			transport: MCPTransportStdio | MCPTransportHTTP,
			port:      9000,
			host:      "0.0.0.0",
			wantAddr:  "0.0.0.0:9000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultMCPConfig()
			cfg.Transport = tt.transport
			cfg.HTTPPort = tt.port
			cfg.HTTPHost = tt.host

			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)

			addr := server.GetHTTPAddr()
			assert.Equal(t, tt.wantAddr, addr)
		})
	}
}

// TestGetHTTPPort tests GetHTTPPort method
func TestGetHTTPPort(t *testing.T) {
	dt := devtools.Enable()

	tests := []struct {
		name      string
		transport MCPTransportType
		port      int
		wantPort  int
	}{
		{
			name:      "HTTP transport with port",
			transport: MCPTransportHTTP,
			port:      8765,
			wantPort:  8765,
		},
		{
			name:      "Stdio transport only",
			transport: MCPTransportStdio,
			port:      8765,
			wantPort:  0,
		},
		{
			name:      "Random port (0)",
			transport: MCPTransportHTTP,
			port:      0,
			wantPort:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultMCPConfig()
			cfg.Transport = tt.transport
			cfg.HTTPPort = tt.port
			cfg.HTTPHost = "localhost"

			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)

			port := server.GetHTTPPort()
			assert.Equal(t, tt.wantPort, port)
		})
	}
}

// TestWaitForHTTPServer tests the waitForHTTPServer helper
func TestWaitForHTTPServer(t *testing.T) {
	dt := devtools.Enable()

	cfg := DefaultMCPConfig()
	cfg.Transport = MCPTransportHTTP
	cfg.HTTPPort = 18767 // Fixed port for testing
	cfg.HTTPHost = "localhost"

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = server.StartHTTPServer(ctx)
	}()

	// Wait for server to be ready
	addr := fmt.Sprintf("%s:%d", cfg.HTTPHost, cfg.HTTPPort)
	err = waitForHTTPServer(addr, 2*time.Second)
	assert.NoError(t, err, "Server should become ready")

	// Test timeout case
	err = waitForHTTPServer("localhost:99999", 100*time.Millisecond)
	assert.Error(t, err, "Should timeout for non-existent server")
	assert.Contains(t, err.Error(), "not ready")
}
