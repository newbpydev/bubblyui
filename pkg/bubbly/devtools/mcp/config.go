package mcp

import (
	"fmt"
	"strings"
	"time"
)

// TransportType defines the transport mechanism for MCP server.
//
// Multiple transports can be enabled simultaneously using bitwise OR.
// For example: MCPTransportStdio | MCPTransportHTTP enables both.
type TransportType int

const (
	// MCPTransportStdio enables stdio transport for local CLI integration
	MCPTransportStdio TransportType = 1 << iota

	// MCPTransportHTTP enables HTTP/SSE transport for IDE integration
	MCPTransportHTTP
)

// String returns the string representation of the transport type.
//
// Returns:
//   - string: Human-readable transport type (e.g., "stdio", "http", "stdio|http")
func (t TransportType) String() string {
	if t == 0 {
		return "none"
	}

	var parts []string
	if t&MCPTransportStdio != 0 {
		parts = append(parts, "stdio")
	}
	if t&MCPTransportHTTP != 0 {
		parts = append(parts, "http")
	}

	return strings.Join(parts, "|")
}

// Config holds configuration options for the MCP server.
//
// Configuration controls transport selection, security settings, performance
// tuning, and write operation permissions. All fields have sensible defaults
// via DefaultMCPConfig().
//
// Thread Safety:
//
//	Config instances are not thread-safe. Create separate instances for
//	concurrent use or protect access with a mutex.
//
// Example:
//
//	// Use defaults (stdio transport, read-only)
//	cfg := mcp.DefaultMCPConfig()
//
//	// Enable HTTP transport with auth
//	cfg := &mcp.Config{
//	    Transport:  mcp.MCPTransportHTTP,
//	    HTTPPort:   8765,
//	    HTTPHost:   "localhost",
//	    EnableAuth: true,
//	    AuthToken:  "secret-token",
//	}
//
//	// Validate before use
//	if err := cfg.Validate(); err != nil {
//	    log.Fatal(err)
//	}
type Config struct {
	// Transport specifies which transport(s) to enable
	// Can be MCPTransportStdio, MCPTransportHTTP, or both (bitwise OR)
	Transport TransportType

	// HTTPPort is the port for HTTP transport (1-65535)
	// Only used when MCPTransportHTTP is enabled
	HTTPPort int

	// HTTPHost is the host to bind HTTP server to
	// Use "localhost" for local-only, "0.0.0.0" for all interfaces
	// Only used when MCPTransportHTTP is enabled
	HTTPHost string

	// WriteEnabled allows state modification tools (set_ref_value, etc.)
	// DANGER: Enable only for testing, not production debugging
	// Default: false (read-only access)
	WriteEnabled bool

	// MaxClients limits concurrent client connections
	// Prevents resource exhaustion from too many clients
	// Default: 5
	MaxClients int

	// SubscriptionThrottle is minimum time between subscription updates
	// Prevents overwhelming clients with high-frequency changes
	// Default: 100ms
	SubscriptionThrottle time.Duration

	// RateLimit is maximum requests per second per client
	// Prevents DoS attacks and resource exhaustion
	// Default: 60 req/sec
	RateLimit int

	// EnableAuth requires bearer token authentication for HTTP transport
	// Recommended for any non-localhost HTTP access
	// Default: false
	EnableAuth bool

	// AuthToken is the bearer token for HTTP authentication
	// Required when EnableAuth is true
	// Should be a strong random string (e.g., UUID)
	AuthToken string

	// SanitizeExports automatically removes PII from exported data
	// Recommended to keep enabled for security
	// Default: true
	SanitizeExports bool
}

// DefaultMCPConfig returns an Config with sensible default values.
//
// The defaults are optimized for local development with stdio transport:
//   - Stdio transport only (no network exposure)
//   - HTTP port 8765 (if HTTP enabled later)
//   - Localhost binding (no remote access)
//   - Read-only access (no state modification)
//   - 5 max concurrent clients
//   - 100ms subscription throttle
//   - 60 requests/second rate limit
//   - No authentication (stdio is local-only)
//   - Sanitization enabled
//
// These defaults can be modified as needed for specific use cases.
//
// Example:
//
//	cfg := mcp.DefaultMCPConfig()
//	cfg.Transport = mcp.MCPTransportHTTP  // Enable HTTP
//	cfg.EnableAuth = true                 // Require auth
//	cfg.AuthToken = "secret-token"        // Set token
//	if err := cfg.Validate(); err != nil {
//	    log.Fatal(err)
//	}
//
// Returns:
//   - *Config: A new config with default values
func DefaultMCPConfig() *Config {
	return &Config{
		Transport:            MCPTransportStdio,
		HTTPPort:             8765,
		HTTPHost:             "localhost",
		WriteEnabled:         false,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            60,
		EnableAuth:           false,
		AuthToken:            "",
		SanitizeExports:      true,
	}
}

// Validate checks that the configuration values are valid.
//
// This method verifies that:
//   - HTTP port is in valid range (1-65535) when HTTP transport enabled
//   - HTTP host is not empty when HTTP transport enabled
//   - Max clients is positive
//   - Rate limit is positive
//   - Subscription throttle is non-negative
//   - Auth token is provided when auth is enabled
//
// Call this after creating or modifying config to ensure validity.
//
// Example:
//
//	cfg := mcp.DefaultMCPConfig()
//	cfg.HTTPPort = 0  // Invalid
//	if err := cfg.Validate(); err != nil {
//	    log.Printf("Invalid config: %v", err)
//	}
//
// Returns:
//   - error: Validation error, or nil if config is valid
func (c *Config) Validate() error {
	// Validate HTTP port if HTTP transport is enabled
	if c.Transport&MCPTransportHTTP != 0 {
		// Port 0 is valid (OS assigns random available port)
		// Ports 1-65535 are valid user-specified ports
		if c.HTTPPort < 0 || c.HTTPPort > 65535 {
			return fmt.Errorf("HTTP port must be between 0 and 65535, got %d", c.HTTPPort)
		}

		// Validate HTTP host
		if c.HTTPHost == "" {
			return fmt.Errorf("HTTP host cannot be empty when HTTP transport is enabled")
		}
	}

	// Validate max clients
	if c.MaxClients <= 0 {
		return fmt.Errorf("max clients must be positive, got %d", c.MaxClients)
	}

	// Validate rate limit
	if c.RateLimit <= 0 {
		return fmt.Errorf("rate limit must be positive, got %d", c.RateLimit)
	}

	// Validate subscription throttle
	if c.SubscriptionThrottle < 0 {
		return fmt.Errorf("subscription throttle must be non-negative, got %v", c.SubscriptionThrottle)
	}

	// Validate auth configuration
	if c.EnableAuth && c.AuthToken == "" {
		return fmt.Errorf("auth token required when auth is enabled")
	}

	return nil
}
