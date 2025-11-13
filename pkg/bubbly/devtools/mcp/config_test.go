package mcp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultMCPConfig verifies that default configuration has sensible values
func TestDefaultMCPConfig(t *testing.T) {
	cfg := DefaultMCPConfig()

	require.NotNil(t, cfg, "DefaultMCPConfig should not return nil")

	// Verify sensible defaults
	assert.Equal(t, MCPTransportStdio, cfg.Transport, "Default transport should be stdio")
	assert.Equal(t, 8765, cfg.HTTPPort, "Default HTTP port should be 8765")
	assert.Equal(t, "localhost", cfg.HTTPHost, "Default HTTP host should be localhost")
	assert.False(t, cfg.WriteEnabled, "Write operations should be disabled by default")
	assert.Equal(t, 5, cfg.MaxClients, "Default max clients should be 5")
	assert.Equal(t, 100*time.Millisecond, cfg.SubscriptionThrottle, "Default throttle should be 100ms")
	assert.Equal(t, 60, cfg.RateLimit, "Default rate limit should be 60 req/sec")
	assert.False(t, cfg.EnableAuth, "Auth should be disabled by default")
	assert.Equal(t, "", cfg.AuthToken, "Auth token should be empty by default")
	assert.True(t, cfg.SanitizeExports, "Sanitization should be enabled by default")
}

// TestMCPConfig_Validate tests configuration validation
func TestMCPConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    *MCPConfig
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid default config",
			config:    DefaultMCPConfig(),
			wantError: false,
		},
		{
			name: "valid stdio transport",
			config: &MCPConfig{
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
			},
			wantError: false,
		},
		{
			name: "valid HTTP transport",
			config: &MCPConfig{
				Transport:            MCPTransportHTTP,
				HTTPPort:             9000,
				HTTPHost:             "0.0.0.0",
				WriteEnabled:         false,
				MaxClients:           10,
				SubscriptionThrottle: 200 * time.Millisecond,
				RateLimit:            120,
				EnableAuth:           true,
				AuthToken:            "secret-token",
				SanitizeExports:      true,
			},
			wantError: false,
		},
		{
			name: "valid both transports",
			config: &MCPConfig{
				Transport:            MCPTransportStdio | MCPTransportHTTP,
				HTTPPort:             8765,
				HTTPHost:             "localhost",
				WriteEnabled:         false,
				MaxClients:           5,
				SubscriptionThrottle: 100 * time.Millisecond,
				RateLimit:            60,
				EnableAuth:           false,
				AuthToken:            "",
				SanitizeExports:      true,
			},
			wantError: false,
		},
		{
			name: "valid HTTP port - zero (random port)",
			config: &MCPConfig{
				Transport:            MCPTransportHTTP,
				HTTPPort:             0, // Port 0 is valid (OS assigns random port)
				HTTPHost:             "localhost",
				WriteEnabled:         false,
				MaxClients:           5,
				SubscriptionThrottle: 100 * time.Millisecond,
				RateLimit:            60,
				EnableAuth:           false,
				AuthToken:            "",
				SanitizeExports:      true,
			},
			wantError: false,
		},
		{
			name: "invalid HTTP port - negative",
			config: &MCPConfig{
				Transport:            MCPTransportHTTP,
				HTTPPort:             -1,
				HTTPHost:             "localhost",
				WriteEnabled:         false,
				MaxClients:           5,
				SubscriptionThrottle: 100 * time.Millisecond,
				RateLimit:            60,
				EnableAuth:           false,
				AuthToken:            "",
				SanitizeExports:      true,
			},
			wantError: true,
			errorMsg:  "HTTP port must be between 0 and 65535",
		},
		{
			name: "invalid HTTP port - too high",
			config: &MCPConfig{
				Transport:            MCPTransportHTTP,
				HTTPPort:             70000,
				HTTPHost:             "localhost",
				WriteEnabled:         false,
				MaxClients:           5,
				SubscriptionThrottle: 100 * time.Millisecond,
				RateLimit:            60,
				EnableAuth:           false,
				AuthToken:            "",
				SanitizeExports:      true,
			},
			wantError: true,
			errorMsg:  "HTTP port must be between 0 and 65535",
		},
		{
			name: "invalid max clients - zero",
			config: &MCPConfig{
				Transport:            MCPTransportStdio,
				HTTPPort:             8765,
				HTTPHost:             "localhost",
				WriteEnabled:         false,
				MaxClients:           0,
				SubscriptionThrottle: 100 * time.Millisecond,
				RateLimit:            60,
				EnableAuth:           false,
				AuthToken:            "",
				SanitizeExports:      true,
			},
			wantError: true,
			errorMsg:  "max clients must be positive",
		},
		{
			name: "invalid max clients - negative",
			config: &MCPConfig{
				Transport:            MCPTransportStdio,
				HTTPPort:             8765,
				HTTPHost:             "localhost",
				WriteEnabled:         false,
				MaxClients:           -1,
				SubscriptionThrottle: 100 * time.Millisecond,
				RateLimit:            60,
				EnableAuth:           false,
				AuthToken:            "",
				SanitizeExports:      true,
			},
			wantError: true,
			errorMsg:  "max clients must be positive",
		},
		{
			name: "invalid rate limit - zero",
			config: &MCPConfig{
				Transport:            MCPTransportStdio,
				HTTPPort:             8765,
				HTTPHost:             "localhost",
				WriteEnabled:         false,
				MaxClients:           5,
				SubscriptionThrottle: 100 * time.Millisecond,
				RateLimit:            0,
				EnableAuth:           false,
				AuthToken:            "",
				SanitizeExports:      true,
			},
			wantError: true,
			errorMsg:  "rate limit must be positive",
		},
		{
			name: "invalid subscription throttle - negative",
			config: &MCPConfig{
				Transport:            MCPTransportStdio,
				HTTPPort:             8765,
				HTTPHost:             "localhost",
				WriteEnabled:         false,
				MaxClients:           5,
				SubscriptionThrottle: -1 * time.Millisecond,
				RateLimit:            60,
				EnableAuth:           false,
				AuthToken:            "",
				SanitizeExports:      true,
			},
			wantError: true,
			errorMsg:  "subscription throttle must be non-negative",
		},
		{
			name: "HTTP transport without auth token when auth enabled",
			config: &MCPConfig{
				Transport:            MCPTransportHTTP,
				HTTPPort:             8765,
				HTTPHost:             "localhost",
				WriteEnabled:         false,
				MaxClients:           5,
				SubscriptionThrottle: 100 * time.Millisecond,
				RateLimit:            60,
				EnableAuth:           true,
				AuthToken:            "",
				SanitizeExports:      true,
			},
			wantError: true,
			errorMsg:  "auth token required when auth is enabled",
		},
		{
			name: "empty HTTP host",
			config: &MCPConfig{
				Transport:            MCPTransportHTTP,
				HTTPPort:             8765,
				HTTPHost:             "",
				WriteEnabled:         false,
				MaxClients:           5,
				SubscriptionThrottle: 100 * time.Millisecond,
				RateLimit:            60,
				EnableAuth:           false,
				AuthToken:            "",
				SanitizeExports:      true,
			},
			wantError: true,
			errorMsg:  "HTTP host cannot be empty when HTTP transport is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantError {
				require.Error(t, err, "Expected validation error")
				assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
			} else {
				require.NoError(t, err, "Expected no validation error")
			}
		})
	}
}

// TestMCPTransportType_String tests transport type string representation
func TestMCPTransportType_String(t *testing.T) {
	tests := []struct {
		name      string
		transport MCPTransportType
		want      string
	}{
		{
			name:      "stdio transport",
			transport: MCPTransportStdio,
			want:      "stdio",
		},
		{
			name:      "HTTP transport",
			transport: MCPTransportHTTP,
			want:      "http",
		},
		{
			name:      "both transports",
			transport: MCPTransportStdio | MCPTransportHTTP,
			want:      "stdio|http",
		},
		{
			name:      "no transport",
			transport: 0,
			want:      "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.transport.String()
			assert.Equal(t, tt.want, got, "String representation should match")
		})
	}
}
