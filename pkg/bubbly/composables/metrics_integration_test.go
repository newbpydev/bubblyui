package composables

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUseState_MetricsIntegration tests that UseState records metrics when enabled
func TestUseState_MetricsIntegration(t *testing.T) {
	// Create custom metrics
	reg := prometheus.NewRegistry()
	metrics := monitoring.NewPrometheusMetrics(reg)
	
	// Enable metrics
	monitoring.SetGlobalMetrics(metrics)
	defer monitoring.SetGlobalMetrics(&monitoring.NoOpMetrics{})
	
	// Create context
	ctx := createTestContext()
	
	// Use composable
	state := UseState(ctx, 42)
	_ = state
	
	// Gather metrics
	families, err := reg.Gather()
	require.NoError(t, err)
	
	// Find composable_creations metric
	var found bool
	for _, family := range families {
		if family.GetName() == "bubblyui_composable_creations_total" {
			for _, metric := range family.GetMetric() {
				for _, label := range metric.GetLabel() {
					if label.GetName() == "name" && label.GetValue() == "UseState" {
						found = true
						assert.Equal(t, float64(1), metric.GetCounter().GetValue(),
							"Should record 1 UseState creation")
					}
				}
			}
		}
	}
	
	assert.True(t, found, "Should record UseState creation metric")
}

// TestUseForm_MetricsIntegration tests that UseForm records metrics when enabled
func TestUseForm_MetricsIntegration(t *testing.T) {
	type TestForm struct {
		Name  string
		Email string
	}
	
	// Create custom metrics
	reg := prometheus.NewRegistry()
	metrics := monitoring.NewPrometheusMetrics(reg)
	
	// Enable metrics
	monitoring.SetGlobalMetrics(metrics)
	defer monitoring.SetGlobalMetrics(&monitoring.NoOpMetrics{})
	
	// Create context
	ctx := createTestContext()
	
	// Use composable
	form := UseForm(ctx, TestForm{}, func(f TestForm) map[string]string {
		return make(map[string]string)
	})
	_ = form
	
	// Gather metrics
	families, err := reg.Gather()
	require.NoError(t, err)
	
	// Find composable_creations metric
	var found bool
	for _, family := range families {
		if family.GetName() == "bubblyui_composable_creations_total" {
			for _, metric := range family.GetMetric() {
				for _, label := range metric.GetLabel() {
					if label.GetName() == "name" && label.GetValue() == "UseForm" {
						found = true
						assert.Equal(t, float64(1), metric.GetCounter().GetValue(),
							"Should record 1 UseForm creation")
					}
				}
			}
		}
	}
	
	assert.True(t, found, "Should record UseForm creation metric")
}

// TestUseAsync_MetricsIntegration tests that UseAsync records metrics when enabled
func TestUseAsync_MetricsIntegration(t *testing.T) {
	// Create custom metrics
	reg := prometheus.NewRegistry()
	metrics := monitoring.NewPrometheusMetrics(reg)
	
	// Enable metrics
	monitoring.SetGlobalMetrics(metrics)
	defer monitoring.SetGlobalMetrics(&monitoring.NoOpMetrics{})
	
	// Create context
	ctx := createTestContext()
	
	// Use composable
	async := UseAsync(ctx, func() (*string, error) {
		result := "test"
		return &result, nil
	})
	_ = async
	
	// Gather metrics
	families, err := reg.Gather()
	require.NoError(t, err)
	
	// Find composable_creations metric
	var found bool
	for _, family := range families {
		if family.GetName() == "bubblyui_composable_creations_total" {
			for _, metric := range family.GetMetric() {
				for _, label := range metric.GetLabel() {
					if label.GetName() == "name" && label.GetValue() == "UseAsync" {
						found = true
						assert.Equal(t, float64(1), metric.GetCounter().GetValue(),
							"Should record 1 UseAsync creation")
					}
				}
			}
		}
	}
	
	assert.True(t, found, "Should record UseAsync creation metric")
}

// TestMetricsIntegration_ZeroOverheadWhenDisabled verifies no overhead when monitoring disabled
func TestMetricsIntegration_ZeroOverheadWhenDisabled(t *testing.T) {
	// Ensure monitoring is disabled (default NoOp)
	monitoring.SetGlobalMetrics(&monitoring.NoOpMetrics{})
	
	ctx := createTestContext()
	
	// Should not panic or have issues
	assert.NotPanics(t, func() {
		state := UseState(ctx, 42)
		_ = state
		
		form := UseForm(ctx, struct{ Name string }{}, func(f struct{ Name string }) map[string]string {
			return nil
		})
		_ = form
		
		async := UseAsync(ctx, func() (*int, error) {
			val := 123
			return &val, nil
		})
		_ = async
	}, "Composables should work normally with NoOp metrics")
}

// TestMetricsIntegration_MultipleCreations tests multiple composable creations
func TestMetricsIntegration_MultipleCreations(t *testing.T) {
	// Create custom metrics
	reg := prometheus.NewRegistry()
	metrics := monitoring.NewPrometheusMetrics(reg)
	
	// Enable metrics
	monitoring.SetGlobalMetrics(metrics)
	defer monitoring.SetGlobalMetrics(&monitoring.NoOpMetrics{})
	
	ctx := createTestContext()
	
	// Create multiple composables
	_ = UseState(ctx, 1)
	_ = UseState(ctx, 2)
	_ = UseState(ctx, 3)
	
	_ = UseForm(ctx, struct{ Name string }{}, func(f struct{ Name string }) map[string]string {
		return nil
	})
	
	// Gather metrics
	families, err := reg.Gather()
	require.NoError(t, err)
	
	// Find composable_creations metric
	for _, family := range families {
		if family.GetName() == "bubblyui_composable_creations_total" {
			for _, metric := range family.GetMetric() {
				for _, label := range metric.GetLabel() {
					if label.GetName() == "name" {
						if label.GetValue() == "UseState" {
							assert.Equal(t, float64(3), metric.GetCounter().GetValue(),
								"Should record 3 UseState creations")
						}
						if label.GetValue() == "UseForm" {
							assert.Equal(t, float64(1), metric.GetCounter().GetValue(),
								"Should record 1 UseForm creation")
						}
					}
				}
			}
		}
	}
}
