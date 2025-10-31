package monitoring

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusMetrics implements ComposableMetrics using Prometheus for metric collection.
//
// This implementation exposes metrics in the Prometheus format, allowing them to be
// scraped by a Prometheus server and visualized in dashboards like Grafana.
//
// All metrics are prefixed with "bubblyui_" to avoid naming conflicts.
//
// Metrics exposed:
//   - bubblyui_composable_creations_total: Counter of composable creations by name
//   - bubblyui_provide_inject_depth: Histogram of provide/inject tree depth
//   - bubblyui_allocation_bytes: Histogram of memory allocations by composable
//   - bubblyui_cache_hits_total: Counter of cache hits by cache name
//   - bubblyui_cache_misses_total: Counter of cache misses by cache name
//
// Thread-safe: All Prometheus collectors are thread-safe by design.
//
// Example:
//
//	func main() {
//	    // Create Prometheus metrics
//	    metrics := monitoring.NewPrometheusMetrics(prometheus.DefaultRegisterer)
//	    monitoring.SetGlobalMetrics(metrics)
//
//	    // Expose metrics endpoint
//	    http.Handle("/metrics", promhttp.Handler())
//	    http.ListenAndServe(":2112", nil)
//	}
type PrometheusMetrics struct {
	composableCreations *prometheus.CounterVec
	provideInjectDepth  prometheus.Histogram
	allocationBytes     *prometheus.HistogramVec
	cacheHits           *prometheus.CounterVec
	cacheMisses         *prometheus.CounterVec
	registry            prometheus.Registerer
}

// NewPrometheusMetrics creates a new Prometheus metrics collector and registers all metrics.
//
// The provided Registerer is used to register all metrics. You can use:
//   - prometheus.DefaultRegisterer for the global default registry
//   - prometheus.NewRegistry() for a custom isolated registry
//
// All metrics are registered immediately. If any metric fails to register (e.g., duplicate),
// this function will panic. This is intentional for fail-fast behavior at startup.
//
// Parameters:
//   - reg: The Prometheus Registerer to use for metric registration
//
// Returns:
//   - *PrometheusMetrics: A new Prometheus metrics collector
//
// Example:
//
//	// Use default registry
//	metrics := monitoring.NewPrometheusMetrics(prometheus.DefaultRegisterer)
//
//	// Use custom registry
//	reg := prometheus.NewRegistry()
//	metrics := monitoring.NewPrometheusMetrics(reg)
//	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
func NewPrometheusMetrics(reg prometheus.Registerer) *PrometheusMetrics {
	// Create composable creations counter
	// Labels: name (composable name like "UseState", "UseForm", etc.)
	composableCreations := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bubblyui_composable_creations_total",
			Help: "Total number of composable creations, partitioned by composable name.",
		},
		[]string{"name"},
	)

	// Create provide/inject depth histogram
	// Buckets: 0, 1, 2, 3, 5, 7, 10, 15, 20 (reasonable tree depths)
	provideInjectDepth := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "bubblyui_provide_inject_depth",
			Help:    "Histogram of provide/inject tree depth, indicating component nesting levels.",
			Buckets: []float64{0, 1, 2, 3, 5, 7, 10, 15, 20},
		},
	)

	// Create allocation bytes histogram
	// Labels: composable (composable name)
	// Buckets: 64B, 128B, 256B, 512B, 1KB, 2KB, 4KB, 8KB (typical allocation sizes)
	allocationBytes := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bubblyui_allocation_bytes",
			Help:    "Histogram of memory allocation sizes in bytes, partitioned by composable.",
			Buckets: []float64{64, 128, 256, 512, 1024, 2048, 4096, 8192},
		},
		[]string{"composable"},
	)

	// Create cache hits counter
	// Labels: cache (cache name like "reflection", "timer", etc.)
	cacheHits := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bubblyui_cache_hits_total",
			Help: "Total number of cache hits, partitioned by cache name.",
		},
		[]string{"cache"},
	)

	// Create cache misses counter
	// Labels: cache (cache name like "reflection", "timer", etc.)
	cacheMisses := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bubblyui_cache_misses_total",
			Help: "Total number of cache misses, partitioned by cache name.",
		},
		[]string{"cache"},
	)

	// Register all metrics (will panic on duplicate registration - fail fast)
	reg.MustRegister(composableCreations)
	reg.MustRegister(provideInjectDepth)
	reg.MustRegister(allocationBytes)
	reg.MustRegister(cacheHits)
	reg.MustRegister(cacheMisses)

	return &PrometheusMetrics{
		composableCreations: composableCreations,
		provideInjectDepth:  provideInjectDepth,
		allocationBytes:     allocationBytes,
		cacheHits:           cacheHits,
		cacheMisses:         cacheMisses,
		registry:            reg,
	}
}

// RecordComposableCreation records when a composable is created.
//
// Increments the bubblyui_composable_creations_total counter for the given composable name.
// The duration parameter is currently not used but available for future enhancements
// (e.g., recording creation time histograms).
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - name: The composable name (e.g., "UseState", "UseForm", "UseAsync")
//   - duration: How long the composable took to initialize (informational, not currently recorded)
//
// Example:
//
//	metrics.RecordComposableCreation("UseState", 150*time.Nanosecond)
func (pm *PrometheusMetrics) RecordComposableCreation(name string, duration time.Duration) {
	pm.composableCreations.WithLabelValues(name).Inc()
	// Note: duration is available for future enhancements (e.g., creation time histogram)
	// For now, we only count creations
}

// RecordProvideInjectDepth records the depth of the provide/inject tree.
//
// Adds an observation to the bubblyui_provide_inject_depth histogram.
// Tree depth indicates component nesting levels - high values (>10) may indicate
// overly complex component hierarchies that should be refactored.
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - depth: The current tree depth (0 = root, 1 = first level child, etc.)
//
// Example:
//
//	metrics.RecordProvideInjectDepth(5) // 5 levels deep
func (pm *PrometheusMetrics) RecordProvideInjectDepth(depth int) {
	pm.provideInjectDepth.Observe(float64(depth))
}

// RecordAllocationBytes records memory allocation for a composable.
//
// Adds an observation to the bubblyui_allocation_bytes histogram for the given composable.
// Helps track memory usage patterns and identify memory-heavy composables.
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - composable: The composable name (e.g., "UseForm", "UseState")
//   - bytes: Number of bytes allocated
//
// Example:
//
//	metrics.RecordAllocationBytes("UseForm", 2048) // 2KB allocated
func (pm *PrometheusMetrics) RecordAllocationBytes(composable string, bytes int64) {
	pm.allocationBytes.WithLabelValues(composable).Observe(float64(bytes))
}

// RecordCacheHit records a cache hit.
//
// Increments the bubblyui_cache_hits_total counter for the given cache.
// Used to monitor cache effectiveness (compare hits vs misses).
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - cache: The cache name (e.g., "reflection", "timer")
//
// Example:
//
//	metrics.RecordCacheHit("reflection") // Cache hit for reflection cache
func (pm *PrometheusMetrics) RecordCacheHit(cache string) {
	pm.cacheHits.WithLabelValues(cache).Inc()
}

// RecordCacheMiss records a cache miss.
//
// Increments the bubblyui_cache_misses_total counter for the given cache.
// Used to monitor cache effectiveness (compare hits vs misses).
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - cache: The cache name (e.g., "reflection", "timer")
//
// Example:
//
//	metrics.RecordCacheMiss("timer") // Cache miss for timer cache
func (pm *PrometheusMetrics) RecordCacheMiss(cache string) {
	pm.cacheMisses.WithLabelValues(cache).Inc()
}
