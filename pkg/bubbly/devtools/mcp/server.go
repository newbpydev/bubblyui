package mcp

import (
	"fmt"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

// MCPServer is the main MCP server instance that exposes BubblyUI devtools
// data and capabilities to AI agents via the Model Context Protocol.
//
// The server provides:
//   - Resources: Read-only access to component tree, state, events, performance
//   - Tools: Actions like export, search, clear history, state modification
//   - Subscriptions: Real-time updates on component/state/event changes
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Lifecycle:
//
//  1. NewMCPServer() - Creates and initializes the server
//  2. StartStdioServer() or StartHTTPServer() - Starts transport (Task 1.2/1.3)
//  3. ... AI agents connect and interact ...
//  4. Shutdown() - Graceful cleanup (Task 7.1)
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
//	// Start transport (Task 1.2)
//	if err := server.StartStdioServer(ctx); err != nil {
//	    log.Fatal(err)
//	}
type MCPServer struct {
	// server is the MCP SDK server instance
	// Handles JSON-RPC protocol, resource/tool registration, subscriptions
	server *mcp.Server

	// config holds MCP server configuration
	// Immutable after creation
	config *MCPConfig

	// devtools is a reference to the DevTools instance
	// Used to access collected debug data
	devtools *devtools.DevTools

	// store is a reference to the DevToolsStore
	// Provides direct access to component/state/event data
	store *devtools.DevToolsStore

	// mu protects concurrent access to server fields
	// Currently only used for getters, but prepared for future state
	mu sync.RWMutex
}

// NewMCPServer creates and initializes a new MCP server.
//
// This function:
//   - Validates the configuration
//   - Validates the devtools instance
//   - Creates the MCP SDK server with proper implementation details
//   - Stores references to devtools and store for resource handlers
//
// The server is created but not started. Call StartStdioServer() or
// StartHTTPServer() to begin accepting connections (Task 1.2/1.3).
//
// Thread Safety:
//
//	Safe to call concurrently (creates new instance each time).
//
// Example:
//
//	dt := devtools.Enable()
//	cfg := mcp.DefaultMCPConfig()
//
//	server, err := mcp.NewMCPServer(cfg, dt)
//	if err != nil {
//	    log.Fatalf("Failed to create MCP server: %v", err)
//	}
//
// Parameters:
//   - config: MCP server configuration (transport, security, performance)
//   - dt: DevTools instance to expose via MCP
//
// Returns:
//   - *MCPServer: Initialized server ready to start transport
//   - error: Validation error, or nil on success
func NewMCPServer(config *MCPConfig, dt *devtools.DevTools) (*MCPServer, error) {
	// Validate inputs
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if dt == nil {
		return nil, fmt.Errorf("devtools cannot be nil")
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Get DevToolsStore from DevTools
	// This is the data source for all MCP resources
	store := dt.GetStore()
	if store == nil {
		return nil, fmt.Errorf("devtools store is nil")
	}

	// Create MCP SDK server with implementation details
	// This declares the server's identity and version
	impl := &mcp.Implementation{
		Name:    "bubblyui-devtools",
		Version: "1.0.0",
	}

	// Create server with options
	// Resources, tools, and subscriptions will be registered in future tasks
	opts := &mcp.ServerOptions{
		// Capabilities will be declared as features are added
		// Task 2.x: Resources
		// Task 3.x: Tools
		// Task 4.x: Subscriptions
	}

	mcpServer := mcp.NewServer(impl, opts)

	// Create and return MCPServer wrapper
	return &MCPServer{
		server:   mcpServer,
		config:   config,
		devtools: dt,
		store:    store,
	}, nil
}

// GetConfig returns the server's configuration.
//
// The returned config is the same instance passed to NewMCPServer().
// Do not modify the returned config - it's shared.
//
// Thread Safety:
//
//	Safe to call concurrently.
//
// Example:
//
//	cfg := server.GetConfig()
//	fmt.Printf("Transport: %s\n", cfg.Transport)
//
// Returns:
//   - *MCPConfig: The server's configuration
func (s *MCPServer) GetConfig() *MCPConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// GetDevTools returns the DevTools instance.
//
// This provides access to the full DevTools API for advanced use cases.
//
// Thread Safety:
//
//	Safe to call concurrently.
//
// Example:
//
//	dt := server.GetDevTools()
//	if dt.IsEnabled() {
//	    fmt.Println("DevTools active")
//	}
//
// Returns:
//   - *devtools.DevTools: The DevTools instance
func (s *MCPServer) GetDevTools() *devtools.DevTools {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.devtools
}

// GetStore returns the DevToolsStore instance.
//
// This provides direct access to collected debug data for resource handlers.
// Used internally by resource/tool implementations (Task 2.x, 3.x).
//
// Thread Safety:
//
//	Safe to call concurrently.
//
// Example:
//
//	store := server.GetStore()
//	components := store.GetAllComponents()
//	fmt.Printf("Tracking %d components\n", len(components))
//
// Returns:
//   - *devtools.DevToolsStore: The DevToolsStore instance
func (s *MCPServer) GetStore() *devtools.DevToolsStore {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.store
}
