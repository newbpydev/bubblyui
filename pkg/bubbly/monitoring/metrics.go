// Package monitoring provides pluggable metrics collection for BubblyUI composables.
//
// The monitoring system is entirely optional and has zero overhead when disabled.
// By default, a NoOp implementation is used which performs no operations and makes no allocations.
//
// To enable monitoring, create a metrics implementation (e.g., PrometheusMetrics) and set it globally:
//
//	metrics := NewPrometheusMetrics(prometheus.DefaultRegisterer)
//	monitoring.SetGlobalMetrics(metrics)
//
// Once enabled, composables will automatically record metrics about their usage:
//   - Composable creation count and duration
//   - Provide/Inject tree depth for detecting deep nesting
//   - Memory allocation bytes per composable
//   - Cache hit/miss rates for performance optimization
//
// The metrics interface is designed to be lightweight and non-intrusive. All metric
// recording happens asynchronously and doesn't block composable execution.
//
// Example usage:
//
//	// In your main.go
//	func main() {
//	    // Enable Prometheus metrics
//	    metrics := monitoring.NewPrometheusMetrics(prometheus.DefaultRegisterer)
//	    monitoring.SetGlobalMetrics(metrics)
//
//	    // Composables automatically record metrics
//	    // ... your application code ...
//	}
//
// Zero Overhead:
//
// When monitoring is disabled (default), there is zero overhead:
//   - No allocations
//   - No mutex contention
//   - No function calls (inlined NoOp methods)
//   - No performance impact
//
// Thread Safety:
//
// All operations are thread-safe. Multiple goroutines can safely call SetGlobalMetrics
// and GetGlobalMetrics concurrently. Implementations should also be thread-safe.
package monitoring

import (
	"sync"
	"time"
)

// ComposableMetrics defines the interface for collecting metrics from composables.
//
// Implementations of this interface can export metrics to various backends:
//   - Prometheus (recommended for production)
//   - StatsD
//   - CloudWatch
//   - Datadog
//   - Custom backends
//
// All methods should be thread-safe and non-blocking. Implementations should
// handle errors internally rather than returning them, as metric recording
// should never fail the composable operation.
//
// Example implementation:
//
//	type MyMetrics struct {
//	    counter *prometheus.CounterVec
//	}
//
//	func (m *MyMetrics) RecordComposableCreation(name string, duration time.Duration) {
//	    m.counter.WithLabelValues(name).Inc()
//	}
type ComposableMetrics interface {
	// RecordComposableCreation records when a composable is created.
	//
	// Parameters:
	//   - name: The composable name (e.g., "UseState", "UseAsync", "UseForm")
	//   - duration: How long the composable took to initialize
	//
	// This metric helps track:
	//   - Which composables are used most frequently
	//   - Performance trends over time
	//   - Initialization overhead
	RecordComposableCreation(name string, duration time.Duration)

	// RecordProvideInjectDepth records the depth of the provide/inject tree.
	//
	// Parameters:
	//   - depth: The current tree depth (0 = root, 1 = first level child, etc.)
	//
	// This metric helps:
	//   - Detect deeply nested component trees (potential performance issue)
	//   - Track component hierarchy complexity
	//   - Identify areas for refactoring
	//
	// A depth > 10 typically indicates overly complex nesting that should be simplified.
	RecordProvideInjectDepth(depth int)

	// RecordAllocationBytes records memory allocation for a composable.
	//
	// Parameters:
	//   - composable: The composable name
	//   - bytes: Number of bytes allocated
	//
	// This metric helps:
	//   - Track memory usage patterns
	//   - Identify memory-heavy composables
	//   - Detect memory leaks or excessive allocation
	RecordAllocationBytes(composable string, bytes int64)

	// RecordCacheHit records a cache hit.
	//
	// Parameters:
	//   - cache: The cache name (e.g., "reflection", "timer")
	//
	// This metric helps:
	//   - Monitor cache effectiveness
	//   - Calculate hit rates
	//   - Optimize cache strategies
	RecordCacheHit(cache string)

	// RecordCacheMiss records a cache miss.
	//
	// Parameters:
	//   - cache: The cache name (e.g., "reflection", "timer")
	//
	// This metric helps:
	//   - Monitor cache effectiveness
	//   - Calculate miss rates
	//   - Identify opportunities for cache warming
	RecordCacheMiss(cache string)
}

// NoOpMetrics is a zero-overhead implementation that does nothing.
//
// This is the default implementation when monitoring is not enabled.
// All methods are no-ops and will be inlined by the compiler, resulting
// in zero runtime overhead.
//
// NoOpMetrics is safe for concurrent use and makes no allocations.
//
// Example:
//
//	// Default behavior - no metrics collected
//	metrics := &NoOpMetrics{}
//	metrics.RecordComposableCreation("UseState", 100*time.Nanosecond) // Does nothing
type NoOpMetrics struct{}

// RecordComposableCreation does nothing (no-op).
func (n *NoOpMetrics) RecordComposableCreation(name string, duration time.Duration) {
	// No-op: Intentionally empty for zero overhead
}

// RecordProvideInjectDepth does nothing (no-op).
func (n *NoOpMetrics) RecordProvideInjectDepth(depth int) {
	// No-op: Intentionally empty for zero overhead
}

// RecordAllocationBytes does nothing (no-op).
func (n *NoOpMetrics) RecordAllocationBytes(composable string, bytes int64) {
	// No-op: Intentionally empty for zero overhead
}

// RecordCacheHit does nothing (no-op).
func (n *NoOpMetrics) RecordCacheHit(cache string) {
	// No-op: Intentionally empty for zero overhead
}

// RecordCacheMiss does nothing (no-op).
func (n *NoOpMetrics) RecordCacheMiss(cache string) {
	// No-op: Intentionally empty for zero overhead
}

// globalMetrics holds the current metrics implementation.
// Defaults to NoOpMetrics for zero overhead when monitoring is disabled.
var globalMetrics ComposableMetrics = &NoOpMetrics{}

// globalMetricsMu protects access to globalMetrics for thread safety.
var globalMetricsMu sync.RWMutex

// SetGlobalMetrics sets the global metrics implementation.
//
// This should be called once at application startup to enable monitoring.
// Setting to nil will reset to NoOpMetrics for safety.
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	func main() {
//	    // Enable Prometheus metrics
//	    metrics := monitoring.NewPrometheusMetrics(prometheus.DefaultRegisterer)
//	    monitoring.SetGlobalMetrics(metrics)
//
//	    // ... rest of application ...
//	}
//
// To disable metrics:
//
//	monitoring.SetGlobalMetrics(nil) // Resets to NoOp
func SetGlobalMetrics(m ComposableMetrics) {
	globalMetricsMu.Lock()
	defer globalMetricsMu.Unlock()

	if m == nil {
		// Safety: never allow nil metrics to prevent panics
		globalMetrics = &NoOpMetrics{}
		return
	}

	globalMetrics = m
}

// GetGlobalMetrics returns the current global metrics implementation.
//
// This function is called by composables to record metrics. It never returns nil.
// If monitoring is disabled, returns NoOpMetrics which has zero overhead.
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
//
// Example (internal use by composables):
//
//	func UseState[T any](ctx *Context, initial T) UseStateReturn[T] {
//	    start := time.Now()
//	    defer func() {
//	        if metrics := monitoring.GetGlobalMetrics(); metrics != nil {
//	            metrics.RecordComposableCreation("UseState", time.Since(start))
//	        }
//	    }()
//	    // ... composable implementation ...
//	}
//
// Returns:
//   - ComposableMetrics: The current metrics implementation (never nil)
func GetGlobalMetrics() ComposableMetrics {
	globalMetricsMu.RLock()
	defer globalMetricsMu.RUnlock()

	return globalMetrics
}
