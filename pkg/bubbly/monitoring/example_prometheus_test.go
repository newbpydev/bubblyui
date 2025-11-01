package monitoring_test

import (
	"fmt"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
	"github.com/prometheus/client_golang/prometheus"
)

// ExampleNewPrometheusMetrics demonstrates creating Prometheus metrics with a custom registry.
func ExampleNewPrometheusMetrics() {
	// Create custom registry to avoid conflicts
	reg := prometheus.NewRegistry()
	
	// Create Prometheus metrics using custom registry
	metrics := monitoring.NewPrometheusMetrics(reg)
	
	// Set as global metrics
	monitoring.SetGlobalMetrics(metrics)
	
	// In a real app, expose metrics endpoint:
	// http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	// http.ListenAndServe(":2112", nil)
	
	fmt.Println("Prometheus metrics initialized")
	// Output: Prometheus metrics initialized
}

// ExampleNewPrometheusMetrics_customRegistry demonstrates using a custom registry.
func ExampleNewPrometheusMetrics_customRegistry() {
	// Create a custom registry for isolated metrics
	reg := prometheus.NewRegistry()
	
	// Create Prometheus metrics with custom registry
	metrics := monitoring.NewPrometheusMetrics(reg)
	
	// Set as global metrics
	monitoring.SetGlobalMetrics(metrics)
	
	// Use the registry with your metrics
	_ = metrics // Metrics ready to use
	
	fmt.Println("Custom Prometheus registry initialized")
	// Output: Custom Prometheus registry initialized
}

// Example_prometheusMetricsRecordComposableCreation demonstrates recording composable creations.
func Example_prometheusMetricsRecordComposableCreation() {
	reg := prometheus.NewRegistry()
	metrics := monitoring.NewPrometheusMetrics(reg)

	// Record composable creations
	metrics.RecordComposableCreation("UseState", 100*time.Nanosecond)
	metrics.RecordComposableCreation("UseForm", 250*time.Nanosecond)
	metrics.RecordComposableCreation("UseState", 150*time.Nanosecond)

	// Metrics are now available at /metrics endpoint
	// Example output in Prometheus format:
	// bubblyui_composable_creations_total{name="UseState"} 2
	// bubblyui_composable_creations_total{name="UseForm"} 1

	fmt.Println("Recorded composable creations")
	// Output: Recorded composable creations
}

// Example_prometheusMetricsRecordCacheMetrics demonstrates tracking cache performance.
func Example_prometheusMetricsRecordCacheMetrics() {
	reg := prometheus.NewRegistry()
	metrics := monitoring.NewPrometheusMetrics(reg)

	// Simulate cache hits and misses
	metrics.RecordCacheHit("reflection")
	metrics.RecordCacheHit("reflection")
	metrics.RecordCacheHit("reflection")
	metrics.RecordCacheMiss("reflection")

	metrics.RecordCacheHit("timer")
	metrics.RecordCacheMiss("timer")
	metrics.RecordCacheMiss("timer")

	// Calculate hit rates in Prometheus queries:
	// rate(bubblyui_cache_hits_total[5m]) / (rate(bubblyui_cache_hits_total[5m]) + rate(bubblyui_cache_misses_total[5m]))

	fmt.Println("Recorded cache metrics")
	// Output: Recorded cache metrics
}

// Example_prometheusMetricsRecordProvideInjectDepth demonstrates tracking component tree depth.
func Example_prometheusMetricsRecordProvideInjectDepth() {
	reg := prometheus.NewRegistry()
	metrics := monitoring.NewPrometheusMetrics(reg)

	// Record various tree depths
	metrics.RecordProvideInjectDepth(1) // Shallow nesting
	metrics.RecordProvideInjectDepth(3)
	metrics.RecordProvideInjectDepth(5)
	metrics.RecordProvideInjectDepth(12) // Deep nesting - may need refactoring

	// Use Prometheus histogram_quantile to analyze:
	// histogram_quantile(0.95, rate(bubblyui_provide_inject_depth_bucket[5m]))
	// This shows 95th percentile tree depth

	fmt.Println("Recorded tree depth observations")
	// Output: Recorded tree depth observations
}

// Example_prometheusMetricsRecordAllocationBytes demonstrates tracking memory allocations.
func Example_prometheusMetricsRecordAllocationBytes() {
	reg := prometheus.NewRegistry()
	metrics := monitoring.NewPrometheusMetrics(reg)

	// Record memory allocations by composable
	metrics.RecordAllocationBytes("UseState", 128)
	metrics.RecordAllocationBytes("UseForm", 2048)
	metrics.RecordAllocationBytes("UseAsync", 512)
	metrics.RecordAllocationBytes("UseForm", 1024)

	// Analyze allocation patterns in Prometheus:
	// histogram_quantile(0.99, sum(rate(bubblyui_allocation_bytes_bucket[5m])) by (composable, le))

	fmt.Println("Recorded allocation metrics")
	// Output: Recorded allocation metrics
}

// Example_prometheusMetricsComplete demonstrates a complete setup with metrics endpoint.
func Example_prometheusMetricsComplete() {
	// Create custom registry
	reg := prometheus.NewRegistry()
	
	// Create Prometheus metrics
	metrics := monitoring.NewPrometheusMetrics(reg)
	
	// Set as global metrics so composables automatically record
	monitoring.SetGlobalMetrics(metrics)
	
	// Simulate some composable usage
	metrics.RecordComposableCreation("UseState", 100*time.Nanosecond)
	metrics.RecordComposableCreation("UseForm", 250*time.Nanosecond)
	metrics.RecordProvideInjectDepth(3)
	metrics.RecordCacheHit("reflection")
	metrics.RecordAllocationBytes("UseState", 128)
	
	// In a real application, expose metrics endpoint:
	// http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	// http.ListenAndServe(":2112", nil)
	//
	// Then configure Prometheus to scrape:
	// scrape_configs:
	//   - job_name: 'bubblyui-app'
	//     static_configs:
	//       - targets: ['localhost:2112']
	
	fmt.Println("Complete Prometheus setup initialized")
	// Output: Complete Prometheus setup initialized
}
