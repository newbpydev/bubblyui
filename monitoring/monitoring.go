// Package monitoring provides pluggable metrics collection for BubblyUI composables.
//
// The monitoring system is entirely optional and has zero overhead when disabled.
// By default, a NoOp implementation is used which performs no operations.
//
// This package is an alias for github.com/newbpydev/bubblyui/pkg/bubbly/monitoring,
// providing a cleaner import path for users.
//
// # Features
//
//   - Composable creation count and duration tracking
//   - Provide/Inject tree depth monitoring
//   - Memory allocation tracking per composable
//   - Cache hit/miss rates for performance optimization
//   - Prometheus metrics integration
//   - pprof profiling endpoints
//
// # Example
//
//	import "github.com/newbpydev/bubblyui/monitoring"
//
//	func main() {
//	    // Enable Prometheus metrics
//	    metrics := monitoring.NewPrometheusMetrics(prometheus.DefaultRegisterer)
//	    monitoring.SetGlobalMetrics(metrics)
//
//	    // Enable pprof profiling on port 6060
//	    monitoring.EnableProfiling(":6060")
//	    defer monitoring.StopProfiling()
//	}
//
// # Zero Overhead
//
// When monitoring is disabled (default), there is zero overhead:
//   - No allocations
//   - No mutex contention
//   - No function calls (inlined NoOp methods)
//   - No performance impact
package monitoring

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// =============================================================================
// Global Metrics
// =============================================================================

// ComposableMetrics defines the interface for composable metrics collection.
type ComposableMetrics = monitoring.ComposableMetrics

// GetGlobalMetrics returns the current global metrics implementation.
var GetGlobalMetrics = monitoring.GetGlobalMetrics

// SetGlobalMetrics sets the global metrics implementation.
var SetGlobalMetrics = monitoring.SetGlobalMetrics

// NoOpMetrics is a no-op implementation with zero overhead.
type NoOpMetrics = monitoring.NoOpMetrics

// =============================================================================
// Prometheus Integration
// =============================================================================

// PrometheusMetrics implements ComposableMetrics using Prometheus.
type PrometheusMetrics = monitoring.PrometheusMetrics

// NewPrometheusMetrics creates a new Prometheus metrics implementation.
func NewPrometheusMetrics(reg prometheus.Registerer) *PrometheusMetrics {
	return monitoring.NewPrometheusMetrics(reg)
}

// =============================================================================
// Profiling
// =============================================================================

// ProfileComposables runs composable profiling for the specified duration.
func ProfileComposables(duration time.Duration) *ComposableProfile {
	return monitoring.ProfileComposables(duration)
}

// ComposableProfile contains profiling results for composables.
type ComposableProfile = monitoring.ComposableProfile

// CallStats contains statistics about composable calls.
type CallStats = monitoring.CallStats

// =============================================================================
// pprof Profiling Endpoints
// =============================================================================

// EnableProfiling starts a pprof HTTP server on the specified address.
// Returns an error if profiling is already enabled or the server fails to start.
var EnableProfiling = monitoring.EnableProfiling

// StopProfiling stops the pprof HTTP server if running.
var StopProfiling = monitoring.StopProfiling

// IsProfilingEnabled returns whether pprof profiling is currently enabled.
var IsProfilingEnabled = monitoring.IsProfilingEnabled

// GetProfilingAddress returns the address of the pprof server if enabled.
var GetProfilingAddress = monitoring.GetProfilingAddress
