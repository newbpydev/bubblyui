package monitoring

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPrometheusMetrics_ImplementsInterface tests that PrometheusMetrics implements ComposableMetrics
func TestPrometheusMetrics_ImplementsInterface(t *testing.T) {
	var _ ComposableMetrics = (*PrometheusMetrics)(nil)
}

// TestNewPrometheusMetrics tests creating new Prometheus metrics
func TestNewPrometheusMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()

	metrics := NewPrometheusMetrics(reg)

	require.NotNil(t, metrics, "NewPrometheusMetrics should return non-nil")
	require.NotNil(t, metrics.registry, "registry should be set")
}

// TestPrometheusMetrics_MetricsRegistered tests that all metrics are registered
func TestPrometheusMetrics_MetricsRegistered(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewPrometheusMetrics(reg)

	// Record at least one value for each metric so they show up in Gather()
	// (Vec metrics don't appear until they have at least one label combination)
	metrics.RecordComposableCreation("UseState", 100*time.Nanosecond)
	metrics.RecordProvideInjectDepth(5)
	metrics.RecordAllocationBytes("UseForm", 128)
	metrics.RecordCacheHit("reflection")
	metrics.RecordCacheMiss("timer")

	// Gather metrics to verify registration
	families, err := reg.Gather()
	require.NoError(t, err, "Should gather metrics without error")

	// Verify expected metrics are registered
	expectedMetrics := []string{
		"bubblyui_composable_creations_total",
		"bubblyui_provide_inject_depth",
		"bubblyui_allocation_bytes",
		"bubblyui_cache_hits_total",
		"bubblyui_cache_misses_total",
	}

	metricNames := make([]string, len(families))
	for i, family := range families {
		metricNames[i] = family.GetName()
	}

	for _, expected := range expectedMetrics {
		assert.Contains(t, metricNames, expected, "Should have registered metric: %s", expected)
	}
}

// TestPrometheusMetrics_RecordComposableCreation tests recording composable creation
func TestPrometheusMetrics_RecordComposableCreation(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewPrometheusMetrics(reg)

	// Record some creations
	metrics.RecordComposableCreation("UseState", 100*time.Nanosecond)
	metrics.RecordComposableCreation("UseState", 150*time.Nanosecond)
	metrics.RecordComposableCreation("UseForm", 200*time.Nanosecond)

	// Gather and verify
	families, err := reg.Gather()
	require.NoError(t, err)

	// Find composable_creations metric
	var creationsMetric *dto.MetricFamily
	for _, family := range families {
		if family.GetName() == "bubblyui_composable_creations_total" {
			creationsMetric = family
			break
		}
	}

	require.NotNil(t, creationsMetric, "Should find composable_creations metric")

	// Verify UseState count (2 creations)
	var useStateValue float64
	var useFormValue float64

	for _, metric := range creationsMetric.GetMetric() {
		for _, label := range metric.GetLabel() {
			if label.GetName() == "name" && label.GetValue() == "UseState" {
				useStateValue = metric.GetCounter().GetValue()
			}
			if label.GetName() == "name" && label.GetValue() == "UseForm" {
				useFormValue = metric.GetCounter().GetValue()
			}
		}
	}

	assert.Equal(t, float64(2), useStateValue, "UseState should have 2 creations")
	assert.Equal(t, float64(1), useFormValue, "UseForm should have 1 creation")
}

// TestPrometheusMetrics_RecordProvideInjectDepth tests recording tree depth
func TestPrometheusMetrics_RecordProvideInjectDepth(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewPrometheusMetrics(reg)

	// Record various depths
	metrics.RecordProvideInjectDepth(1)
	metrics.RecordProvideInjectDepth(3)
	metrics.RecordProvideInjectDepth(5)
	metrics.RecordProvideInjectDepth(2)

	// Gather and verify histogram exists
	families, err := reg.Gather()
	require.NoError(t, err)

	var depthMetric *dto.MetricFamily
	for _, family := range families {
		if family.GetName() == "bubblyui_provide_inject_depth" {
			depthMetric = family
			break
		}
	}

	require.NotNil(t, depthMetric, "Should find provide_inject_depth metric")
	require.Len(t, depthMetric.GetMetric(), 1, "Should have one histogram")

	histogram := depthMetric.GetMetric()[0].GetHistogram()
	assert.Equal(t, uint64(4), histogram.GetSampleCount(), "Should have 4 observations")
}

// TestPrometheusMetrics_RecordAllocationBytes tests recording allocation bytes
func TestPrometheusMetrics_RecordAllocationBytes(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewPrometheusMetrics(reg)

	// Record allocations
	metrics.RecordAllocationBytes("UseState", 128)
	metrics.RecordAllocationBytes("UseForm", 512)
	metrics.RecordAllocationBytes("UseForm", 1024)

	// Gather and verify
	families, err := reg.Gather()
	require.NoError(t, err)

	var allocMetric *dto.MetricFamily
	for _, family := range families {
		if family.GetName() == "bubblyui_allocation_bytes" {
			allocMetric = family
			break
		}
	}

	require.NotNil(t, allocMetric, "Should find allocation_bytes metric")

	// Verify observations were recorded
	var useStateCount, useFormCount uint64
	for _, metric := range allocMetric.GetMetric() {
		for _, label := range metric.GetLabel() {
			if label.GetName() == "composable" && label.GetValue() == "UseState" {
				useStateCount = metric.GetHistogram().GetSampleCount()
			}
			if label.GetName() == "composable" && label.GetValue() == "UseForm" {
				useFormCount = metric.GetHistogram().GetSampleCount()
			}
		}
	}

	assert.Equal(t, uint64(1), useStateCount, "UseState should have 1 observation")
	assert.Equal(t, uint64(2), useFormCount, "UseForm should have 2 observations")
}

// TestPrometheusMetrics_RecordCacheHit tests recording cache hits
func TestPrometheusMetrics_RecordCacheHit(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewPrometheusMetrics(reg)

	// Record cache hits
	metrics.RecordCacheHit("reflection")
	metrics.RecordCacheHit("reflection")
	metrics.RecordCacheHit("timer")

	// Gather and verify
	families, err := reg.Gather()
	require.NoError(t, err)

	var hitsMetric *dto.MetricFamily
	for _, family := range families {
		if family.GetName() == "bubblyui_cache_hits_total" {
			hitsMetric = family
			break
		}
	}

	require.NotNil(t, hitsMetric, "Should find cache_hits metric")

	var reflectionHits, timerHits float64
	for _, metric := range hitsMetric.GetMetric() {
		for _, label := range metric.GetLabel() {
			if label.GetName() == "cache" && label.GetValue() == "reflection" {
				reflectionHits = metric.GetCounter().GetValue()
			}
			if label.GetName() == "cache" && label.GetValue() == "timer" {
				timerHits = metric.GetCounter().GetValue()
			}
		}
	}

	assert.Equal(t, float64(2), reflectionHits, "reflection should have 2 hits")
	assert.Equal(t, float64(1), timerHits, "timer should have 1 hit")
}

// TestPrometheusMetrics_RecordCacheMiss tests recording cache misses
func TestPrometheusMetrics_RecordCacheMiss(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewPrometheusMetrics(reg)

	// Record cache misses
	metrics.RecordCacheMiss("reflection")
	metrics.RecordCacheMiss("timer")
	metrics.RecordCacheMiss("timer")
	metrics.RecordCacheMiss("timer")

	// Gather and verify
	families, err := reg.Gather()
	require.NoError(t, err)

	var missesMetric *dto.MetricFamily
	for _, family := range families {
		if family.GetName() == "bubblyui_cache_misses_total" {
			missesMetric = family
			break
		}
	}

	require.NotNil(t, missesMetric, "Should find cache_misses metric")

	var reflectionMisses, timerMisses float64
	for _, metric := range missesMetric.GetMetric() {
		for _, label := range metric.GetLabel() {
			if label.GetName() == "cache" && label.GetValue() == "reflection" {
				reflectionMisses = metric.GetCounter().GetValue()
			}
			if label.GetName() == "cache" && label.GetValue() == "timer" {
				timerMisses = metric.GetCounter().GetValue()
			}
		}
	}

	assert.Equal(t, float64(1), reflectionMisses, "reflection should have 1 miss")
	assert.Equal(t, float64(3), timerMisses, "timer should have 3 misses")
}

// TestPrometheusMetrics_DefaultRegistry tests using default registry
func TestPrometheusMetrics_DefaultRegistry(t *testing.T) {
	// Create with default registry
	metrics := NewPrometheusMetrics(prometheus.DefaultRegisterer)

	require.NotNil(t, metrics, "Should create with default registry")

	// Should be able to record metrics
	assert.NotPanics(t, func() {
		metrics.RecordComposableCreation("UseState", 100*time.Nanosecond)
		metrics.RecordCacheHit("test")
	}, "Should not panic with default registry")
}

// TestPrometheusMetrics_MetricNaming tests metric naming conventions
func TestPrometheusMetrics_MetricNaming(t *testing.T) {
	reg := prometheus.NewRegistry()
	_ = NewPrometheusMetrics(reg)

	families, err := reg.Gather()
	require.NoError(t, err)

	for _, family := range families {
		name := family.GetName()

		// All metrics should start with bubblyui_
		assert.True(t, strings.HasPrefix(name, "bubblyui_"),
			"Metric %s should have bubblyui_ prefix", name)

		// Counter metrics should end with _total
		if family.GetType() == dto.MetricType_COUNTER {
			assert.True(t, strings.HasSuffix(name, "_total"),
				"Counter metric %s should end with _total", name)
		}

		// Should have help text
		assert.NotEmpty(t, family.GetHelp(), "Metric %s should have help text", name)
	}
}

// TestPrometheusMetrics_HistogramBuckets tests histogram bucket configuration
func TestPrometheusMetrics_HistogramBuckets(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewPrometheusMetrics(reg)

	// Record observations across different ranges
	metrics.RecordProvideInjectDepth(1)
	metrics.RecordProvideInjectDepth(5)
	metrics.RecordProvideInjectDepth(10)
	metrics.RecordProvideInjectDepth(15)

	families, err := reg.Gather()
	require.NoError(t, err)

	var depthMetric *dto.MetricFamily
	for _, family := range families {
		if family.GetName() == "bubblyui_provide_inject_depth" {
			depthMetric = family
			break
		}
	}

	require.NotNil(t, depthMetric)
	histogram := depthMetric.GetMetric()[0].GetHistogram()

	// Should have buckets
	assert.NotEmpty(t, histogram.GetBucket(), "Histogram should have buckets")

	// Verify we have reasonable bucket boundaries
	bucketBounds := make([]float64, len(histogram.GetBucket()))
	for i, bucket := range histogram.GetBucket() {
		bucketBounds[i] = bucket.GetUpperBound()
	}

	// Should have some buckets that make sense for tree depth (0-20 range)
	assert.Contains(t, bucketBounds, float64(5), "Should have bucket for depth 5")
	assert.Contains(t, bucketBounds, float64(10), "Should have bucket for depth 10")
}
