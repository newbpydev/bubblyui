package mcp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

// TestEnableWithMCP tests that EnableWithMCP creates DevTools and starts MCP server.
func TestEnableWithMCP(t *testing.T) {
	// Create config
	cfg := DefaultMCPConfig()
	cfg.Transport = MCPTransportStdio

	// Enable with MCP
	dt, err := EnableWithMCP(cfg)

	// Verify no error
	require.NoError(t, err, "EnableWithMCP should not return error")
	require.NotNil(t, dt, "EnableWithMCP should return DevTools instance")

	// Verify DevTools is enabled
	assert.True(t, dt.IsEnabled(), "DevTools should be enabled")

	// Verify MCP is enabled
	assert.True(t, dt.MCPEnabled(), "MCP should be enabled")

	// Verify MCP server exists
	mcpServer := dt.GetMCPServer()
	assert.NotNil(t, mcpServer, "MCP server should exist")

	// Type assert to verify it's the right type
	server, ok := mcpServer.(*Server)
	assert.True(t, ok, "MCP server should be *Server type")
	assert.NotNil(t, server, "Type-asserted server should not be nil")

	// Verify MCP server has access to store
	store := server.GetStore()
	assert.NotNil(t, store, "MCP server should have access to store")

	// Cleanup
	devtools.Disable()
}

// TestEnableWithMCP_NilConfig tests that nil config returns error.
func TestEnableWithMCP_NilConfig(t *testing.T) {
	// Enable with nil config
	dt, err := EnableWithMCP(nil)

	// Verify error
	assert.Error(t, err, "EnableWithMCP should return error for nil config")
	assert.Nil(t, dt, "EnableWithMCP should return nil for nil config")
}

// TestEnableWithMCP_InvalidConfig tests that invalid config returns error.
func TestEnableWithMCP_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "No transport",
			config: &Config{
				Transport: 0, // No transport
			},
		},
		{
			name: "Invalid HTTP port",
			config: &Config{
				Transport: MCPTransportHTTP,
				HTTPPort:  0, // Invalid port
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Enable with invalid config
			dt, err := EnableWithMCP(tt.config)

			// Verify error
			assert.Error(t, err, "EnableWithMCP should return error for invalid config")
			assert.Nil(t, dt, "EnableWithMCP should return nil for invalid config")
		})
	}
}

// TestMCPServer_AccessesDevToolsStore tests that MCP server can access Store.
func TestMCPServer_AccessesDevToolsStore(t *testing.T) {
	// Enable with MCP
	cfg := DefaultMCPConfig()
	dt, err := EnableWithMCP(cfg)
	require.NoError(t, err)
	defer devtools.Disable()

	// Get MCP server
	mcpServer := dt.GetMCPServer()
	require.NotNil(t, mcpServer, "MCP server should exist")

	// Type assert
	server, ok := mcpServer.(*Server)
	require.True(t, ok, "Should be *Server")

	// Get store from DevTools
	dtStore := dt.GetStore()
	require.NotNil(t, dtStore, "DevTools store should exist")

	// Get store from MCP server
	mcpStore := server.GetStore()
	require.NotNil(t, mcpStore, "MCP server store should exist")

	// Verify they are the same instance
	assert.Equal(t, dtStore, mcpStore, "MCP server should use same store as DevTools")
}

// TestMCPShutdownOnDisable tests that MCP server reference is cleared when DevTools is disabled.
func TestMCPShutdownOnDisable(t *testing.T) {
	// Enable with MCP
	cfg := DefaultMCPConfig()
	dt, err := EnableWithMCP(cfg)
	require.NoError(t, err)

	// Verify MCP is enabled
	assert.True(t, dt.MCPEnabled(), "MCP should be enabled")

	// Disable DevTools
	devtools.Disable()

	// Verify DevTools is disabled
	assert.False(t, dt.IsEnabled(), "DevTools should be disabled")

	// Verify MCP reference is cleared
	assert.False(t, dt.MCPEnabled(), "MCP should be disabled after Disable()")
	assert.Nil(t, dt.GetMCPServer(), "MCP server reference should be nil")
}

// TestEnableWithMCP_Idempotent tests that EnableWithMCP is idempotent.
func TestEnableWithMCP_Idempotent(t *testing.T) {
	// Enable first time
	cfg := DefaultMCPConfig()
	dt1, err := EnableWithMCP(cfg)
	require.NoError(t, err)
	require.NotNil(t, dt1)
	defer devtools.Disable()

	// Enable second time
	dt2, err := EnableWithMCP(cfg)
	require.NoError(t, err)
	require.NotNil(t, dt2)

	// Verify same instance
	assert.Equal(t, dt1, dt2, "EnableWithMCP should return same instance on subsequent calls")
}

// TestEnableWithMCP_NoConflictWithExistingDevtools tests that MCP doesn't conflict.
func TestEnableWithMCP_NoConflictWithExistingDevtools(t *testing.T) {
	// Enable regular DevTools first
	dt1 := devtools.Enable()
	require.NotNil(t, dt1)
	require.True(t, dt1.IsEnabled())
	require.False(t, dt1.MCPEnabled(), "MCP should not be enabled initially")

	// Now enable MCP
	cfg := DefaultMCPConfig()
	dt2, err := EnableWithMCP(cfg)
	require.NoError(t, err)
	require.NotNil(t, dt2)
	defer devtools.Disable()

	// Verify same instance
	assert.Equal(t, dt1, dt2, "Should return same DevTools instance")

	// Verify MCP is now enabled
	assert.True(t, dt2.MCPEnabled(), "MCP should be enabled after EnableWithMCP")

	// Verify DevTools still works
	assert.True(t, dt2.IsEnabled(), "DevTools should still be enabled")
	assert.NotNil(t, dt2.GetStore(), "DevTools store should still work")
}

// TestEnableWithMCP_ThreadSafe tests concurrent access to MCP server integration.
func TestEnableWithMCP_ThreadSafe(t *testing.T) {
	// Enable with MCP
	cfg := DefaultMCPConfig()
	dt, err := EnableWithMCP(cfg)
	require.NoError(t, err)
	defer devtools.Disable()

	// Concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			// Access MCP server methods concurrently
			_ = dt.MCPEnabled()
			_ = dt.GetMCPServer()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for goroutines")
		}
	}
}
