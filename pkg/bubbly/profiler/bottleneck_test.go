// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBottleneckDetector(t *testing.T) {
	t.Run("creates new bottleneck detector with default thresholds", func(t *testing.T) {
		bd := NewBottleneckDetector()

		assert.NotNil(t, bd)
		assert.NotNil(t, bd.thresholds)
		assert.NotNil(t, bd.violations)
		assert.NotNil(t, bd.config)
	})

	t.Run("creates bottleneck detector with custom thresholds", func(t *testing.T) {
		config := &BottleneckThresholds{
			DefaultOperationThreshold: 5 * time.Millisecond,
			RenderThreshold:           10 * time.Millisecond,
			UpdateThreshold:           3 * time.Millisecond,
		}
		bd := NewBottleneckDetectorWithThresholds(config)

		assert.NotNil(t, bd)
		assert.Equal(t, 5*time.Millisecond, bd.config.DefaultOperationThreshold)
		assert.Equal(t, 10*time.Millisecond, bd.config.RenderThreshold)
	})
}

func TestDefaultBottleneckThresholds(t *testing.T) {
	t.Run("returns sensible defaults", func(t *testing.T) {
		thresholds := DefaultBottleneckThresholds()

		assert.NotNil(t, thresholds)
		assert.Greater(t, thresholds.DefaultOperationThreshold, time.Duration(0))
		assert.Greater(t, thresholds.RenderThreshold, time.Duration(0))
		assert.Greater(t, thresholds.UpdateThreshold, time.Duration(0))
		assert.Greater(t, thresholds.EventThreshold, time.Duration(0))
		assert.Greater(t, thresholds.FrequentRenderThreshold, int64(0))
		assert.Greater(t, thresholds.MemoryThreshold, uint64(0))
	})
}

func TestBottleneckDetector_Check(t *testing.T) {
	tests := []struct {
		name           string
		operation      string
		duration       time.Duration
		threshold      time.Duration
		wantBottleneck bool
		wantSeverity   Severity
	}{
		{
			name:           "no bottleneck when duration below threshold",
			operation:      "render",
			duration:       5 * time.Millisecond,
			threshold:      16 * time.Millisecond,
			wantBottleneck: false,
		},
		{
			name:           "no bottleneck when duration equals threshold",
			operation:      "render",
			duration:       16 * time.Millisecond,
			threshold:      16 * time.Millisecond,
			wantBottleneck: false,
		},
		{
			name:           "bottleneck when duration exceeds threshold - low severity",
			operation:      "render",
			duration:       20 * time.Millisecond,
			threshold:      16 * time.Millisecond,
			wantBottleneck: true,
			wantSeverity:   SeverityLow,
		},
		{
			name:           "bottleneck with medium severity (2x threshold)",
			operation:      "render",
			duration:       35 * time.Millisecond,
			threshold:      16 * time.Millisecond,
			wantBottleneck: true,
			wantSeverity:   SeverityMedium,
		},
		{
			name:           "bottleneck with high severity (3x threshold)",
			operation:      "render",
			duration:       50 * time.Millisecond,
			threshold:      16 * time.Millisecond,
			wantBottleneck: true,
			wantSeverity:   SeverityHigh,
		},
		{
			name:           "bottleneck with critical severity (5x threshold)",
			operation:      "render",
			duration:       100 * time.Millisecond,
			threshold:      16 * time.Millisecond,
			wantBottleneck: true,
			wantSeverity:   SeverityCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bd := NewBottleneckDetector()
			bd.SetThreshold(tt.operation, tt.threshold)

			bottleneck := bd.Check(tt.operation, tt.duration)

			if tt.wantBottleneck {
				require.NotNil(t, bottleneck)
				assert.Equal(t, tt.wantSeverity, bottleneck.Severity)
				assert.Equal(t, tt.operation, bottleneck.Location)
				assert.Equal(t, BottleneckTypeSlow, bottleneck.Type)
				assert.NotEmpty(t, bottleneck.Description)
				assert.NotEmpty(t, bottleneck.Suggestion)
			} else {
				assert.Nil(t, bottleneck)
			}
		})
	}
}

func TestBottleneckDetector_ViolationTracking(t *testing.T) {
	t.Run("tracks violations count", func(t *testing.T) {
		bd := NewBottleneckDetector()
		bd.SetThreshold("render", 10*time.Millisecond)

		// First violation
		bd.Check("render", 20*time.Millisecond)
		assert.Equal(t, 1, bd.GetViolations("render"))

		// Second violation
		bd.Check("render", 30*time.Millisecond)
		assert.Equal(t, 2, bd.GetViolations("render"))

		// No violation (below threshold)
		bd.Check("render", 5*time.Millisecond)
		assert.Equal(t, 2, bd.GetViolations("render"))
	})

	t.Run("tracks violations per operation", func(t *testing.T) {
		bd := NewBottleneckDetector()
		bd.SetThreshold("render", 10*time.Millisecond)
		bd.SetThreshold("update", 5*time.Millisecond)

		bd.Check("render", 20*time.Millisecond)
		bd.Check("render", 30*time.Millisecond)
		bd.Check("update", 10*time.Millisecond)

		assert.Equal(t, 2, bd.GetViolations("render"))
		assert.Equal(t, 1, bd.GetViolations("update"))
		assert.Equal(t, 0, bd.GetViolations("nonexistent"))
	})
}

func TestBottleneckDetector_ImpactCalculation(t *testing.T) {
	tests := []struct {
		name       string
		duration   time.Duration
		threshold  time.Duration
		wantImpact float64
	}{
		{
			name:       "impact just above threshold",
			duration:   20 * time.Millisecond,
			threshold:  16 * time.Millisecond,
			wantImpact: 0.125, // 1.25 / 10 = 0.125
		},
		{
			name:       "impact at 2x threshold",
			duration:   32 * time.Millisecond,
			threshold:  16 * time.Millisecond,
			wantImpact: 0.2, // 2.0 / 10 = 0.2
		},
		{
			name:       "impact at 5x threshold",
			duration:   80 * time.Millisecond,
			threshold:  16 * time.Millisecond,
			wantImpact: 0.5, // 5.0 / 10 = 0.5
		},
		{
			name:       "impact capped at 1.0 for extreme cases",
			duration:   200 * time.Millisecond,
			threshold:  16 * time.Millisecond,
			wantImpact: 1.0, // 12.5 / 10 = 1.25, capped at 1.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bd := NewBottleneckDetector()
			bd.SetThreshold("render", tt.threshold)

			bottleneck := bd.Check("render", tt.duration)

			require.NotNil(t, bottleneck)
			assert.InDelta(t, tt.wantImpact, bottleneck.Impact, 0.01)
		})
	}
}

func TestBottleneckDetector_SuggestionGeneration(t *testing.T) {
	t.Run("generates suggestion for slow render", func(t *testing.T) {
		bd := NewBottleneckDetector()
		bd.SetThreshold("render", 10*time.Millisecond)

		bottleneck := bd.Check("render", 50*time.Millisecond)

		require.NotNil(t, bottleneck)
		assert.NotEmpty(t, bottleneck.Suggestion)
		assert.Contains(t, bottleneck.Suggestion, "render")
	})

	t.Run("generates suggestion for slow update", func(t *testing.T) {
		bd := NewBottleneckDetector()
		bd.SetThreshold("update", 5*time.Millisecond)

		bottleneck := bd.Check("update", 25*time.Millisecond)

		require.NotNil(t, bottleneck)
		assert.NotEmpty(t, bottleneck.Suggestion)
	})
}

func TestBottleneckDetector_Detect(t *testing.T) {
	t.Run("detects bottlenecks from component metrics", func(t *testing.T) {
		bd := NewBottleneckDetector()

		metrics := &PerformanceMetrics{
			Components: []*ComponentMetrics{
				{
					ComponentID:     "comp-1",
					ComponentName:   "SlowComponent",
					RenderCount:     100,
					TotalRenderTime: 5 * time.Second,
					AvgRenderTime:   50 * time.Millisecond, // Very slow
					MaxRenderTime:   100 * time.Millisecond,
				},
				{
					ComponentID:     "comp-2",
					ComponentName:   "FastComponent",
					RenderCount:     100,
					TotalRenderTime: 500 * time.Millisecond,
					AvgRenderTime:   5 * time.Millisecond, // Fast
					MaxRenderTime:   10 * time.Millisecond,
				},
			},
		}

		bottlenecks := bd.Detect(metrics)

		// Should detect slow component
		assert.GreaterOrEqual(t, len(bottlenecks), 1)

		// Find the slow component bottleneck
		var slowBottleneck *BottleneckInfo
		for _, b := range bottlenecks {
			if b.Location == "SlowComponent" {
				slowBottleneck = b
				break
			}
		}
		require.NotNil(t, slowBottleneck)
		assert.Equal(t, BottleneckTypeSlow, slowBottleneck.Type)
	})

	t.Run("detects frequent render bottleneck", func(t *testing.T) {
		bd := NewBottleneckDetector()

		metrics := &PerformanceMetrics{
			Components: []*ComponentMetrics{
				{
					ComponentID:     "comp-1",
					ComponentName:   "FrequentComponent",
					RenderCount:     10000, // Very frequent
					TotalRenderTime: 1 * time.Second,
					AvgRenderTime:   100 * time.Microsecond,
					MaxRenderTime:   1 * time.Millisecond,
				},
			},
		}

		bottlenecks := bd.Detect(metrics)

		// Should detect frequent render pattern
		var frequentBottleneck *BottleneckInfo
		for _, b := range bottlenecks {
			if b.Type == BottleneckTypeFrequent {
				frequentBottleneck = b
				break
			}
		}
		require.NotNil(t, frequentBottleneck)
		assert.Contains(t, frequentBottleneck.Suggestion, "memoization")
	})

	t.Run("handles nil metrics gracefully", func(t *testing.T) {
		bd := NewBottleneckDetector()

		bottlenecks := bd.Detect(nil)

		assert.Empty(t, bottlenecks)
	})

	t.Run("handles empty metrics gracefully", func(t *testing.T) {
		bd := NewBottleneckDetector()

		metrics := &PerformanceMetrics{
			Components: []*ComponentMetrics{},
		}

		bottlenecks := bd.Detect(metrics)

		assert.Empty(t, bottlenecks)
	})

	t.Run("detects multiple bottlenecks", func(t *testing.T) {
		bd := NewBottleneckDetector()

		metrics := &PerformanceMetrics{
			Components: []*ComponentMetrics{
				{
					ComponentID:   "comp-1",
					ComponentName: "SlowComponent",
					RenderCount:   100,
					AvgRenderTime: 50 * time.Millisecond,
					MaxRenderTime: 100 * time.Millisecond,
				},
				{
					ComponentID:   "comp-2",
					ComponentName: "FrequentComponent",
					RenderCount:   10000,
					AvgRenderTime: 100 * time.Microsecond,
					MaxRenderTime: 1 * time.Millisecond,
				},
			},
		}

		bottlenecks := bd.Detect(metrics)

		// Should detect both slow and frequent bottlenecks
		assert.GreaterOrEqual(t, len(bottlenecks), 2)
	})
}

func TestBottleneckDetector_ThresholdManagement(t *testing.T) {
	t.Run("SetThreshold sets operation threshold", func(t *testing.T) {
		bd := NewBottleneckDetector()

		bd.SetThreshold("render", 20*time.Millisecond)

		assert.Equal(t, 20*time.Millisecond, bd.GetThreshold("render"))
	})

	t.Run("GetThreshold returns default for unknown operation", func(t *testing.T) {
		bd := NewBottleneckDetector()

		threshold := bd.GetThreshold("unknown")

		assert.Equal(t, bd.config.DefaultOperationThreshold, threshold)
	})

	t.Run("GetThreshold returns operation-specific threshold", func(t *testing.T) {
		bd := NewBottleneckDetector()
		bd.SetThreshold("custom", 100*time.Millisecond)

		threshold := bd.GetThreshold("custom")

		assert.Equal(t, 100*time.Millisecond, threshold)
	})
}

func TestBottleneckDetector_Reset(t *testing.T) {
	t.Run("resets violations and custom thresholds", func(t *testing.T) {
		bd := NewBottleneckDetector()
		bd.SetThreshold("render", 5*time.Millisecond)
		bd.Check("render", 20*time.Millisecond)
		bd.Check("render", 30*time.Millisecond)

		assert.Equal(t, 2, bd.GetViolations("render"))

		bd.Reset()

		assert.Equal(t, 0, bd.GetViolations("render"))
		// Custom thresholds should be cleared
		assert.Equal(t, bd.config.DefaultOperationThreshold, bd.GetThreshold("render"))
	})
}

func TestBottleneckDetector_ThreadSafety(t *testing.T) {
	t.Run("concurrent operations are safe", func(t *testing.T) {
		bd := NewBottleneckDetector()
		bd.SetThreshold("render", 10*time.Millisecond)
		var wg sync.WaitGroup
		goroutines := 50

		// Concurrent Check operations
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = bd.Check("render", 20*time.Millisecond)
			}()
		}

		// Concurrent SetThreshold operations
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				bd.SetThreshold("render", time.Duration(i+1)*time.Millisecond)
			}(i)
		}

		// Concurrent GetViolations operations
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = bd.GetViolations("render")
			}()
		}

		wg.Wait()
		// Should complete without race conditions
	})

	t.Run("concurrent Detect operations are safe", func(t *testing.T) {
		bd := NewBottleneckDetector()
		var wg sync.WaitGroup
		goroutines := 50

		metrics := &PerformanceMetrics{
			Components: []*ComponentMetrics{
				{
					ComponentID:   "comp-1",
					ComponentName: "TestComponent",
					RenderCount:   100,
					AvgRenderTime: 50 * time.Millisecond,
				},
			},
		}

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = bd.Detect(metrics)
			}()
		}

		wg.Wait()
		// Should complete without race conditions
	})
}

func TestBottleneckDetector_BottleneckInfoFields(t *testing.T) {
	t.Run("bottleneck info has all required fields", func(t *testing.T) {
		bd := NewBottleneckDetector()
		bd.SetThreshold("render", 10*time.Millisecond)

		bottleneck := bd.Check("render", 100*time.Millisecond)

		require.NotNil(t, bottleneck)
		assert.NotEqual(t, BottleneckType(""), bottleneck.Type)
		assert.NotEmpty(t, bottleneck.Location)
		assert.NotEqual(t, Severity(""), bottleneck.Severity)
		assert.Greater(t, bottleneck.Impact, float64(0))
		assert.NotEmpty(t, bottleneck.Description)
		assert.NotEmpty(t, bottleneck.Suggestion)
	})
}

func TestBottleneckDetector_GetAllViolations(t *testing.T) {
	t.Run("returns all violations", func(t *testing.T) {
		bd := NewBottleneckDetector()
		bd.SetThreshold("render", 10*time.Millisecond)
		bd.SetThreshold("update", 5*time.Millisecond)

		bd.Check("render", 20*time.Millisecond)
		bd.Check("render", 30*time.Millisecond)
		bd.Check("update", 10*time.Millisecond)

		violations := bd.GetAllViolations()

		assert.Equal(t, 2, violations["render"])
		assert.Equal(t, 1, violations["update"])
	})
}

func TestBottleneckDetector_GetConfig(t *testing.T) {
	t.Run("returns current config", func(t *testing.T) {
		config := &BottleneckThresholds{
			DefaultOperationThreshold: 20 * time.Millisecond,
			RenderThreshold:           30 * time.Millisecond,
		}
		bd := NewBottleneckDetectorWithThresholds(config)

		got := bd.GetConfig()

		assert.Equal(t, 20*time.Millisecond, got.DefaultOperationThreshold)
		assert.Equal(t, 30*time.Millisecond, got.RenderThreshold)
	})
}

func TestBottleneckDetector_DetectMemoryBottleneck(t *testing.T) {
	t.Run("detects memory bottleneck", func(t *testing.T) {
		bd := NewBottleneckDetector()

		metrics := &PerformanceMetrics{
			Components: []*ComponentMetrics{
				{
					ComponentID:   "comp-1",
					ComponentName: "MemoryHeavyComponent",
					RenderCount:   10,
					AvgRenderTime: 1 * time.Millisecond,
					MemoryUsage:   50 * 1024 * 1024, // 50MB - exceeds 10MB threshold
				},
			},
		}

		bottlenecks := bd.Detect(metrics)

		// Should detect memory bottleneck
		var memoryBottleneck *BottleneckInfo
		for _, b := range bottlenecks {
			if b.Type == BottleneckTypeMemory {
				memoryBottleneck = b
				break
			}
		}
		require.NotNil(t, memoryBottleneck)
		assert.Contains(t, memoryBottleneck.Suggestion, "memory")
	})
}

func TestBottleneckDetector_SuggestionTypes(t *testing.T) {
	tests := []struct {
		name           string
		operation      string
		bottleneckType BottleneckType
		wantContains   string
	}{
		{
			name:           "render suggestion",
			operation:      "render",
			bottleneckType: BottleneckTypeSlow,
			wantContains:   "render",
		},
		{
			name:           "update suggestion",
			operation:      "update",
			bottleneckType: BottleneckTypeSlow,
			wantContains:   "update",
		},
		{
			name:           "generic operation suggestion",
			operation:      "custom_operation",
			bottleneckType: BottleneckTypeSlow,
			wantContains:   "custom_operation",
		},
		{
			name:           "frequent suggestion",
			operation:      "render",
			bottleneckType: BottleneckTypeFrequent,
			wantContains:   "memoization",
		},
		{
			name:           "memory suggestion",
			operation:      "memory",
			bottleneckType: BottleneckTypeMemory,
			wantContains:   "memory",
		},
		{
			name:           "pattern suggestion",
			operation:      "pattern",
			bottleneckType: BottleneckTypePattern,
			wantContains:   "architecture",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := generateSuggestion(tt.operation, tt.bottleneckType)
			assert.Contains(t, suggestion, tt.wantContains)
		})
	}
}

func TestBottleneckDetector_UnknownBottleneckType(t *testing.T) {
	t.Run("handles unknown bottleneck type", func(t *testing.T) {
		suggestion := generateSuggestion("unknown", BottleneckType("unknown"))
		assert.NotEmpty(t, suggestion)
		assert.Contains(t, suggestion, "optimization")
	})
}
