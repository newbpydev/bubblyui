// Package mcp provides Model Context Protocol (MCP) server integration for BubblyUI DevTools.
//
// The MCP server enables remote state inspection, component manipulation, and
// real-time debugging through the Model Context Protocol. It supports rate limiting,
// authentication, batching, and subscription-based updates.
//
// This package is an alias for github.com/newbpydev/bubblyui/pkg/bubbly/devtools/mcp,
// providing a cleaner import path for users.
//
// # Features
//
//   - Remote state inspection and manipulation
//   - Component search and filtering
//   - Event filtering and export
//   - Performance monitoring
//   - Rate limiting and authentication
//   - Real-time subscription updates
//   - HTTP and stdio transports
//
// # Example
//
//	import "github.com/newbpydev/bubblyui/devtools/mcp"
//
//	func main() {
//	    config := mcp.DefaultMCPConfig()
//	    dt, err := mcp.EnableWithMCP(config)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer dt.Shutdown()
//	}
package mcp

import (
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools/mcp"
)

// =============================================================================
// Initialization
// =============================================================================

// EnableWithMCP enables devtools with MCP server integration.
func EnableWithMCP(config *Config) (*devtools.DevTools, error) {
	return mcp.EnableWithMCP(config)
}

// =============================================================================
// Configuration
// =============================================================================

// Config holds MCP server configuration.
type Config = mcp.Config

// DefaultMCPConfig returns the default MCP configuration.
var DefaultMCPConfig = mcp.DefaultMCPConfig

// TransportType specifies the transport protocol.
type TransportType = mcp.TransportType

// =============================================================================
// Server
// =============================================================================

// Server is the MCP server instance.
type Server = mcp.Server

// NewMCPServer creates a new MCP server.
func NewMCPServer(config *Config, dt *devtools.DevTools) (*Server, error) {
	return mcp.NewMCPServer(config, dt)
}

// =============================================================================
// Authentication
// =============================================================================

// AuthHandler handles authentication for MCP requests.
type AuthHandler = mcp.AuthHandler

// NewAuthHandler creates a new authentication handler.
var NewAuthHandler = mcp.NewAuthHandler

// =============================================================================
// Rate Limiting
// =============================================================================

// RateLimiter limits request rates per client.
type RateLimiter = mcp.RateLimiter

// NewRateLimiter creates a new rate limiter.
var NewRateLimiter = mcp.NewRateLimiter

// Throttler throttles update notifications.
type Throttler = mcp.Throttler

// NewThrottler creates a new throttler.
func NewThrottler(minInterval time.Duration) (*Throttler, error) {
	return mcp.NewThrottler(minInterval)
}

// =============================================================================
// Update Batching
// =============================================================================

// UpdateBatcher batches update notifications.
type UpdateBatcher = mcp.UpdateBatcher

// NewUpdateBatcher creates a new update batcher.
func NewUpdateBatcher(flushInterval time.Duration, maxBatchSize int) (*UpdateBatcher, error) {
	return mcp.NewUpdateBatcher(flushInterval, maxBatchSize)
}

// UpdateNotification represents an update notification.
type UpdateNotification = mcp.UpdateNotification

// FlushHandler handles batched update flushes.
type FlushHandler = mcp.FlushHandler

// NotificationSender sends notifications to subscribers.
type NotificationSender = mcp.NotificationSender

// NewNotificationSender creates a new notification sender.
var NewNotificationSender = mcp.NewNotificationSender

// =============================================================================
// Subscriptions
// =============================================================================

// SubscriptionManager manages client subscriptions.
type SubscriptionManager = mcp.SubscriptionManager

// NewSubscriptionManager creates a new subscription manager.
var NewSubscriptionManager = mcp.NewSubscriptionManager

// Subscription represents a client subscription.
type Subscription = mcp.Subscription

// =============================================================================
// State Change Detection
// =============================================================================

// StateChangeDetector detects state changes for notifications.
type StateChangeDetector = mcp.StateChangeDetector

// NewStateChangeDetector creates a new state change detector.
var NewStateChangeDetector = mcp.NewStateChangeDetector

// =============================================================================
// Validation
// =============================================================================

// ValidateResourceURI validates a resource URI.
var ValidateResourceURI = mcp.ValidateResourceURI

// ValidateToolParams validates tool parameters.
var ValidateToolParams = mcp.ValidateToolParams

// SanitizeInput sanitizes user input.
var SanitizeInput = mcp.SanitizeInput

// =============================================================================
// Resources
// =============================================================================

// ComponentsResource provides component inspection resources.
type ComponentsResource = mcp.ComponentsResource

// StateResource provides state inspection resources.
type StateResource = mcp.StateResource

// EventsResource provides event inspection resources.
type EventsResource = mcp.EventsResource

// PerformanceResource provides performance monitoring resources.
type PerformanceResource = mcp.PerformanceResource

// =============================================================================
// Tool Parameters
// =============================================================================

// SearchComponentsParams configures component search.
type SearchComponentsParams = mcp.SearchComponentsParams

// SearchComponentsResult contains search results.
type SearchComponentsResult = mcp.SearchComponentsResult

// ComponentMatch represents a matched component.
type ComponentMatch = mcp.ComponentMatch

// FilterEventsParams configures event filtering.
type FilterEventsParams = mcp.FilterEventsParams

// FilterEventsResult contains filter results.
type FilterEventsResult = mcp.FilterEventsResult

// ExportParams configures data export.
type ExportParams = mcp.ExportParams

// ExportResult contains export results.
type ExportResult = mcp.ExportResult

// ClearStateHistoryParams configures state history clearing.
type ClearStateHistoryParams = mcp.ClearStateHistoryParams

// ClearEventLogParams configures event log clearing.
type ClearEventLogParams = mcp.ClearEventLogParams

// ClearResult contains clear operation results.
type ClearResult = mcp.ClearResult

// SetRefValueParams configures ref value setting.
type SetRefValueParams = mcp.SetRefValueParams

// SetRefResult contains set ref results.
type SetRefResult = mcp.SetRefResult

// =============================================================================
// Data Types
// =============================================================================

// RefInfo contains information about a ref.
type RefInfo = mcp.RefInfo

// ComputedInfo contains information about a computed value.
type ComputedInfo = mcp.ComputedInfo

// PerformanceSummary contains performance summary data.
type PerformanceSummary = mcp.PerformanceSummary
