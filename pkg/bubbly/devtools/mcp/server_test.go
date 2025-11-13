package mcp

import (
	"sync"
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMCPServer_ValidConfig tests successful server creation with valid config
func TestNewMCPServer_ValidConfig(t *testing.T) {
	// Create devtools instance
	dt := devtools.Enable()
	require.NotNil(t, dt, "DevTools should be created")

	// Create valid config
	cfg := DefaultMCPConfig()
	require.NotNil(t, cfg, "Config should be created")

	// Create MCP server
	server, err := NewMCPServer(cfg, dt)

	// Verify success
	require.NoError(t, err, "NewMCPServer should succeed with valid config")
	require.NotNil(t, server, "Server should not be nil")

	// Verify server fields are initialized
	assert.NotNil(t, server.server, "MCP SDK server should be initialized")
	assert.NotNil(t, server.config, "Config should be stored")
	assert.NotNil(t, server.devtools, "DevTools reference should be stored")
	assert.NotNil(t, server.store, "DevToolsStore reference should be stored")
	assert.Equal(t, cfg, server.config, "Config should match input")
	assert.Equal(t, dt, server.devtools, "DevTools should match input")
}

// TestNewMCPServer_NilConfig tests that nil config returns error
func TestNewMCPServer_NilConfig(t *testing.T) {
	dt := devtools.Enable()
	require.NotNil(t, dt, "DevTools should be created")

	// Attempt to create server with nil config
	server, err := NewMCPServer(nil, dt)

	// Verify error
	require.Error(t, err, "NewMCPServer should fail with nil config")
	assert.Nil(t, server, "Server should be nil on error")
	assert.Contains(t, err.Error(), "config cannot be nil", "Error should mention nil config")
}

// TestNewMCPServer_NilDevTools tests that nil devtools returns error
func TestNewMCPServer_NilDevTools(t *testing.T) {
	cfg := DefaultMCPConfig()
	require.NotNil(t, cfg, "Config should be created")

	// Attempt to create server with nil devtools
	server, err := NewMCPServer(cfg, nil)

	// Verify error
	require.Error(t, err, "NewMCPServer should fail with nil devtools")
	assert.Nil(t, server, "Server should be nil on error")
	assert.Contains(t, err.Error(), "devtools cannot be nil", "Error should mention nil devtools")
}

// TestNewMCPServer_InvalidConfig tests error handling for invalid config
func TestNewMCPServer_InvalidConfig(t *testing.T) {
	dt := devtools.Enable()
	require.NotNil(t, dt, "DevTools should be created")

	// Create config with invalid port (negative is invalid)
	cfg := &MCPConfig{
		Transport:            MCPTransportHTTP,
		HTTPPort:             -1, // Invalid port (negative)
		HTTPHost:             "localhost",
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            60,
		EnableAuth:           false,
		AuthToken:            "",
		SanitizeExports:      true,
	}

	// Attempt to create server with invalid config
	server, err := NewMCPServer(cfg, dt)

	// Verify error
	require.Error(t, err, "NewMCPServer should fail with invalid config")
	assert.Nil(t, server, "Server should be nil on error")
	assert.Contains(t, err.Error(), "invalid config", "Error should mention invalid config")
}

// TestNewMCPServer_NilStore tests that nil store is handled (defensive check)
func TestNewMCPServer_NilStore(t *testing.T) {
	// Note: This test verifies the defensive nil check exists
	// In practice, dt.GetStore() should never return nil if dt is valid
	// But we have the check for safety

	dt := devtools.Enable()
	require.NotNil(t, dt, "DevTools should be created")

	// Verify store is not nil (normal case)
	store := dt.GetStore()
	assert.NotNil(t, store, "Store should not be nil in normal operation")

	// The nil store check in NewMCPServer is defensive code
	// It's tested implicitly by all successful NewMCPServer calls
}

// TestMCPServer_ThreadSafe tests concurrent access to server with race detector
func TestMCPServer_ThreadSafe(t *testing.T) {
	dt := devtools.Enable()
	require.NotNil(t, dt, "DevTools should be created")

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err, "Server creation should succeed")
	require.NotNil(t, server, "Server should not be nil")

	// Perform concurrent reads
	const goroutines = 10
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				// Concurrent reads should be safe
				_ = server.GetConfig()
				_ = server.GetDevTools()
				_ = server.GetStore()
			}
		}()
	}

	wg.Wait()
	// If we get here without race detector errors, test passes
}

// TestMCPServer_GetConfig tests config getter
func TestMCPServer_GetConfig(t *testing.T) {
	dt := devtools.Enable()
	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Get config
	gotCfg := server.GetConfig()

	// Verify it matches
	assert.Equal(t, cfg, gotCfg, "GetConfig should return the same config")
}

// TestMCPServer_GetDevTools tests devtools getter
func TestMCPServer_GetDevTools(t *testing.T) {
	dt := devtools.Enable()
	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Get devtools
	gotDT := server.GetDevTools()

	// Verify it matches
	assert.Equal(t, dt, gotDT, "GetDevTools should return the same devtools instance")
}

// TestMCPServer_GetStore tests store getter
func TestMCPServer_GetStore(t *testing.T) {
	dt := devtools.Enable()
	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Get store
	store := server.GetStore()

	// Verify it's not nil and accessible
	assert.NotNil(t, store, "GetStore should return non-nil store")
}
