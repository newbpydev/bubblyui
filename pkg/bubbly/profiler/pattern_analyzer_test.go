package profiler

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPatternAnalyzer(t *testing.T) {
	tests := []struct {
		name         string
		wantPatterns int
		wantNotEmpty bool
	}{
		{
			name:         "creates analyzer with default patterns",
			wantPatterns: 5, // 5 default patterns
			wantNotEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa := NewPatternAnalyzer()
			require.NotNil(t, pa)
			assert.Equal(t, tt.wantPatterns, pa.PatternCount())
			if tt.wantNotEmpty {
				assert.NotEmpty(t, pa.GetPatterns())
			}
		})
	}
}

func TestNewPatternAnalyzerWithPatterns(t *testing.T) {
	tests := []struct {
		name         string
		patterns     []Pattern
		wantPatterns int
	}{
		{
			name:         "creates analyzer with nil patterns",
			patterns:     nil,
			wantPatterns: 0,
		},
		{
			name:         "creates analyzer with empty patterns",
			patterns:     []Pattern{},
			wantPatterns: 0,
		},
		{
			name: "creates analyzer with custom patterns",
			patterns: []Pattern{
				{Name: "custom1", Severity: SeverityLow},
				{Name: "custom2", Severity: SeverityHigh},
			},
			wantPatterns: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa := NewPatternAnalyzerWithPatterns(tt.patterns)
			require.NotNil(t, pa)
			assert.Equal(t, tt.wantPatterns, pa.PatternCount())
		})
	}
}

func TestPatternAnalyzer_Analyze(t *testing.T) {
	tests := []struct {
		name            string
		metrics         *ComponentMetrics
		wantBottlenecks int
		wantPatterns    []string
	}{
		{
			name:            "returns empty for nil metrics",
			metrics:         nil,
			wantBottlenecks: 0,
		},
		{
			name: "returns empty for healthy component",
			metrics: &ComponentMetrics{
				ComponentName: "HealthyComponent",
				RenderCount:   100,
				AvgRenderTime: 1 * time.Millisecond,
				MaxRenderTime: 5 * time.Millisecond,
				MemoryUsage:   1024,
			},
			wantBottlenecks: 0,
		},
		{
			name: "detects frequent_rerender pattern",
			metrics: &ComponentMetrics{
				ComponentName: "FrequentComponent",
				RenderCount:   5000,
				AvgRenderTime: 100 * time.Microsecond, // < 1ms
				MaxRenderTime: 500 * time.Microsecond,
				MemoryUsage:   1024,
			},
			wantBottlenecks: 1,
			wantPatterns:    []string{"frequent_rerender"},
		},
		{
			name: "detects slow_render pattern",
			metrics: &ComponentMetrics{
				ComponentName: "SlowComponent",
				RenderCount:   50,
				AvgRenderTime: 15 * time.Millisecond, // > 10ms
				MaxRenderTime: 20 * time.Millisecond,
				MemoryUsage:   1024,
			},
			wantBottlenecks: 1,
			wantPatterns:    []string{"slow_render"},
		},
		{
			name: "detects memory_hog pattern",
			metrics: &ComponentMetrics{
				ComponentName: "MemoryHog",
				RenderCount:   50,
				AvgRenderTime: 1 * time.Millisecond,
				MaxRenderTime: 5 * time.Millisecond,
				MemoryUsage:   10 * 1024 * 1024, // 10MB > 5MB threshold
			},
			wantBottlenecks: 1,
			wantPatterns:    []string{"memory_hog"},
		},
		{
			name: "detects render_spike pattern",
			metrics: &ComponentMetrics{
				ComponentName: "SpikyComponent",
				RenderCount:   50,
				AvgRenderTime: 1 * time.Millisecond,
				MaxRenderTime: 150 * time.Millisecond, // > 100ms
				MemoryUsage:   1024,
			},
			wantBottlenecks: 1,
			wantPatterns:    []string{"render_spike"},
		},
		{
			name: "detects inefficient_render pattern",
			metrics: &ComponentMetrics{
				ComponentName: "InefficientComponent",
				RenderCount:   1000,                 // > 500
				AvgRenderTime: 8 * time.Millisecond, // > 5ms
				MaxRenderTime: 20 * time.Millisecond,
				MemoryUsage:   1024,
			},
			wantBottlenecks: 1,
			wantPatterns:    []string{"inefficient_render"},
		},
		{
			name: "detects multiple patterns",
			metrics: &ComponentMetrics{
				ComponentName: "ProblematicComponent",
				RenderCount:   2000,                   // > 500 for inefficient_render
				AvgRenderTime: 15 * time.Millisecond,  // > 10ms for slow_render, > 5ms for inefficient
				MaxRenderTime: 200 * time.Millisecond, // > 100ms for render_spike
				MemoryUsage:   10 * 1024 * 1024,       // > 5MB for memory_hog
			},
			wantBottlenecks: 4, // slow_render, memory_hog, render_spike, inefficient_render
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa := NewPatternAnalyzer()
			bottlenecks := pa.Analyze(tt.metrics)

			assert.Len(t, bottlenecks, tt.wantBottlenecks)

			// Verify all bottlenecks have correct type
			for _, b := range bottlenecks {
				assert.Equal(t, BottleneckTypePattern, b.Type)
				if tt.metrics != nil {
					assert.Equal(t, tt.metrics.ComponentName, b.Location)
				}
				assert.NotEmpty(t, b.Description)
				assert.NotEmpty(t, b.Suggestion)
				assert.True(t, b.Impact >= 0 && b.Impact <= 1.0)
			}
		})
	}
}

func TestPatternAnalyzer_AnalyzeAll(t *testing.T) {
	tests := []struct {
		name            string
		metrics         []*ComponentMetrics
		wantBottlenecks int
	}{
		{
			name:            "returns empty for nil metrics",
			metrics:         nil,
			wantBottlenecks: 0,
		},
		{
			name:            "returns empty for empty metrics",
			metrics:         []*ComponentMetrics{},
			wantBottlenecks: 0,
		},
		{
			name: "analyzes multiple components",
			metrics: []*ComponentMetrics{
				{
					ComponentName: "SlowComponent",
					RenderCount:   50,
					AvgRenderTime: 15 * time.Millisecond,
				},
				{
					ComponentName: "HealthyComponent",
					RenderCount:   100,
					AvgRenderTime: 1 * time.Millisecond,
				},
				{
					ComponentName: "MemoryHog",
					MemoryUsage:   10 * 1024 * 1024,
				},
			},
			wantBottlenecks: 2, // slow_render + memory_hog
		},
		{
			name: "handles nil component in slice",
			metrics: []*ComponentMetrics{
				nil,
				{
					ComponentName: "SlowComponent",
					RenderCount:   50,
					AvgRenderTime: 15 * time.Millisecond,
				},
			},
			wantBottlenecks: 1, // Only slow_render from non-nil component
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa := NewPatternAnalyzer()
			bottlenecks := pa.AnalyzeAll(tt.metrics)

			assert.Len(t, bottlenecks, tt.wantBottlenecks)
		})
	}
}

func TestPatternAnalyzer_AddPattern(t *testing.T) {
	tests := []struct {
		name           string
		pattern        Pattern
		wantCount      int
		wantDetectable bool
	}{
		{
			name: "adds custom pattern",
			pattern: Pattern{
				Name: "custom_pattern",
				Detect: func(m *ComponentMetrics) bool {
					return m.RenderCount > 10000
				},
				Severity:    SeverityCritical,
				Description: "Custom pattern detected",
				Suggestion:  "Custom suggestion",
			},
			wantCount:      6, // 5 default + 1 custom
			wantDetectable: true,
		},
		{
			name: "adds pattern with nil detect function",
			pattern: Pattern{
				Name:        "nil_detect",
				Detect:      nil,
				Severity:    SeverityLow,
				Description: "Pattern with nil detect",
			},
			wantCount:      6,
			wantDetectable: false, // Won't detect anything
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa := NewPatternAnalyzer()
			initialCount := pa.PatternCount()

			pa.AddPattern(tt.pattern)

			assert.Equal(t, tt.wantCount, pa.PatternCount())
			assert.Equal(t, initialCount+1, pa.PatternCount())

			// Verify pattern was added
			pattern := pa.GetPattern(tt.pattern.Name)
			require.NotNil(t, pattern)
			assert.Equal(t, tt.pattern.Name, pattern.Name)
			assert.Equal(t, tt.pattern.Severity, pattern.Severity)

			// Test detection if applicable
			if tt.wantDetectable {
				metrics := &ComponentMetrics{
					ComponentName: "TestComponent",
					RenderCount:   20000, // Triggers custom pattern
				}
				bottlenecks := pa.Analyze(metrics)
				found := false
				for _, b := range bottlenecks {
					if b.Description == tt.pattern.Description {
						found = true
						break
					}
				}
				assert.True(t, found, "Custom pattern should be detected")
			}
		})
	}
}

func TestPatternAnalyzer_RemovePattern(t *testing.T) {
	tests := []struct {
		name        string
		patternName string
		wantRemoved bool
		wantCount   int
	}{
		{
			name:        "removes existing pattern",
			patternName: "slow_render",
			wantRemoved: true,
			wantCount:   4, // 5 default - 1
		},
		{
			name:        "returns false for non-existent pattern",
			patternName: "non_existent",
			wantRemoved: false,
			wantCount:   5, // unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa := NewPatternAnalyzer()

			removed := pa.RemovePattern(tt.patternName)

			assert.Equal(t, tt.wantRemoved, removed)
			assert.Equal(t, tt.wantCount, pa.PatternCount())

			// Verify pattern is gone
			if tt.wantRemoved {
				pattern := pa.GetPattern(tt.patternName)
				assert.Nil(t, pattern)
			}
		})
	}
}

func TestPatternAnalyzer_GetPattern(t *testing.T) {
	tests := []struct {
		name        string
		patternName string
		wantNil     bool
		wantName    string
	}{
		{
			name:        "returns existing pattern",
			patternName: "slow_render",
			wantNil:     false,
			wantName:    "slow_render",
		},
		{
			name:        "returns nil for non-existent pattern",
			patternName: "non_existent",
			wantNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa := NewPatternAnalyzer()

			pattern := pa.GetPattern(tt.patternName)

			if tt.wantNil {
				assert.Nil(t, pattern)
			} else {
				require.NotNil(t, pattern)
				assert.Equal(t, tt.wantName, pattern.Name)
			}
		})
	}
}

func TestPatternAnalyzer_GetPatterns(t *testing.T) {
	pa := NewPatternAnalyzer()

	patterns := pa.GetPatterns()

	assert.Len(t, patterns, 5)

	// Verify it's a copy by modifying the returned slice
	originalLen := pa.PatternCount()
	_ = append(patterns, Pattern{Name: "extra"})
	assert.Equal(t, originalLen, pa.PatternCount())
}

func TestPatternAnalyzer_PatternCount(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*PatternAnalyzer)
		wantCount int
	}{
		{
			name:      "returns count for default patterns",
			setup:     func(pa *PatternAnalyzer) {},
			wantCount: 5,
		},
		{
			name: "returns count after adding pattern",
			setup: func(pa *PatternAnalyzer) {
				pa.AddPattern(Pattern{Name: "custom"})
			},
			wantCount: 6,
		},
		{
			name: "returns count after removing pattern",
			setup: func(pa *PatternAnalyzer) {
				pa.RemovePattern("slow_render")
			},
			wantCount: 4,
		},
		{
			name: "returns zero after clearing",
			setup: func(pa *PatternAnalyzer) {
				pa.ClearPatterns()
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa := NewPatternAnalyzer()
			tt.setup(pa)

			assert.Equal(t, tt.wantCount, pa.PatternCount())
		})
	}
}

func TestPatternAnalyzer_Reset(t *testing.T) {
	pa := NewPatternAnalyzer()

	// Modify patterns
	pa.AddPattern(Pattern{Name: "custom"})
	pa.RemovePattern("slow_render")
	assert.Equal(t, 5, pa.PatternCount()) // 5 - 1 + 1 = 5

	// Reset
	pa.Reset()

	// Verify default patterns restored
	assert.Equal(t, 5, pa.PatternCount())
	assert.NotNil(t, pa.GetPattern("slow_render"))
	assert.Nil(t, pa.GetPattern("custom"))
}

func TestPatternAnalyzer_ClearPatterns(t *testing.T) {
	pa := NewPatternAnalyzer()
	assert.Equal(t, 5, pa.PatternCount())

	pa.ClearPatterns()

	assert.Equal(t, 0, pa.PatternCount())
	assert.Empty(t, pa.GetPatterns())

	// Analyze should return empty
	metrics := &ComponentMetrics{
		ComponentName: "SlowComponent",
		AvgRenderTime: 100 * time.Millisecond,
	}
	bottlenecks := pa.Analyze(metrics)
	assert.Empty(t, bottlenecks)
}

func TestPatternAnalyzer_SeverityAssignment(t *testing.T) {
	tests := []struct {
		name         string
		metrics      *ComponentMetrics
		wantSeverity Severity
	}{
		{
			name: "frequent_rerender has medium severity",
			metrics: &ComponentMetrics{
				ComponentName: "Test",
				RenderCount:   5000,
				AvgRenderTime: 100 * time.Microsecond,
			},
			wantSeverity: SeverityMedium,
		},
		{
			name: "slow_render has high severity",
			metrics: &ComponentMetrics{
				ComponentName: "Test",
				AvgRenderTime: 15 * time.Millisecond,
			},
			wantSeverity: SeverityHigh,
		},
		{
			name: "memory_hog has high severity",
			metrics: &ComponentMetrics{
				ComponentName: "Test",
				MemoryUsage:   10 * 1024 * 1024,
			},
			wantSeverity: SeverityHigh,
		},
		{
			name: "render_spike has medium severity",
			metrics: &ComponentMetrics{
				ComponentName: "Test",
				MaxRenderTime: 150 * time.Millisecond,
			},
			wantSeverity: SeverityMedium,
		},
		{
			name: "inefficient_render has critical severity",
			metrics: &ComponentMetrics{
				ComponentName: "Test",
				RenderCount:   1000,
				AvgRenderTime: 8 * time.Millisecond,
			},
			wantSeverity: SeverityCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa := NewPatternAnalyzer()
			bottlenecks := pa.Analyze(tt.metrics)

			require.NotEmpty(t, bottlenecks)
			assert.Equal(t, tt.wantSeverity, bottlenecks[0].Severity)
		})
	}
}

func TestPatternAnalyzer_ImpactCalculation(t *testing.T) {
	tests := []struct {
		name       string
		severity   Severity
		wantImpact float64
	}{
		{name: "critical impact", severity: SeverityCritical, wantImpact: 1.0},
		{name: "high impact", severity: SeverityHigh, wantImpact: 0.75},
		{name: "medium impact", severity: SeverityMedium, wantImpact: 0.5},
		{name: "low impact", severity: SeverityLow, wantImpact: 0.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impact := calculatePatternImpact(tt.severity)
			assert.Equal(t, tt.wantImpact, impact)
		})
	}
}

func TestPatternAnalyzer_ThreadSafety(t *testing.T) {
	pa := NewPatternAnalyzer()
	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrent reads and writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(4)

		// Concurrent Analyze
		go func() {
			defer wg.Done()
			metrics := &ComponentMetrics{
				ComponentName: "TestComponent",
				RenderCount:   5000,
				AvgRenderTime: 100 * time.Microsecond,
			}
			_ = pa.Analyze(metrics)
		}()

		// Concurrent GetPatterns
		go func() {
			defer wg.Done()
			_ = pa.GetPatterns()
		}()

		// Concurrent PatternCount
		go func() {
			defer wg.Done()
			_ = pa.PatternCount()
		}()

		// Concurrent GetPattern
		go func() {
			defer wg.Done()
			_ = pa.GetPattern("slow_render")
		}()
	}

	wg.Wait()
	// If we get here without deadlock or panic, thread safety is working
}

func TestPatternAnalyzer_DefaultPatterns(t *testing.T) {
	patterns := defaultPatterns()

	assert.Len(t, patterns, 5)

	expectedNames := []string{
		"frequent_rerender",
		"slow_render",
		"memory_hog",
		"render_spike",
		"inefficient_render",
	}

	for _, name := range expectedNames {
		found := false
		for _, p := range patterns {
			if p.Name == name {
				found = true
				assert.NotNil(t, p.Detect, "Pattern %s should have Detect function", name)
				assert.NotEmpty(t, p.Description, "Pattern %s should have Description", name)
				assert.NotEmpty(t, p.Suggestion, "Pattern %s should have Suggestion", name)
				break
			}
		}
		assert.True(t, found, "Pattern %s should exist", name)
	}
}

func TestPatternAnalyzer_CustomPatternIntegration(t *testing.T) {
	// Create analyzer with only custom patterns
	customPatterns := []Pattern{
		{
			Name: "high_render_count",
			Detect: func(m *ComponentMetrics) bool {
				return m.RenderCount > 100
			},
			Severity:    SeverityLow,
			Description: "Component has high render count",
			Suggestion:  "Consider reducing renders",
		},
		{
			Name: "very_high_render_count",
			Detect: func(m *ComponentMetrics) bool {
				return m.RenderCount > 1000
			},
			Severity:    SeverityHigh,
			Description: "Component has very high render count",
			Suggestion:  "Implement memoization",
		},
	}

	pa := NewPatternAnalyzerWithPatterns(customPatterns)

	// Test with metrics that trigger both patterns
	metrics := &ComponentMetrics{
		ComponentName: "TestComponent",
		RenderCount:   5000,
	}

	bottlenecks := pa.Analyze(metrics)

	assert.Len(t, bottlenecks, 2)

	// Verify both patterns detected
	severities := make(map[Severity]bool)
	for _, b := range bottlenecks {
		severities[b.Severity] = true
	}
	assert.True(t, severities[SeverityLow])
	assert.True(t, severities[SeverityHigh])
}

func TestPatternAnalyzer_EdgeCases(t *testing.T) {
	t.Run("zero values in metrics", func(t *testing.T) {
		pa := NewPatternAnalyzer()
		metrics := &ComponentMetrics{
			ComponentName: "ZeroComponent",
			// All other fields are zero
		}
		bottlenecks := pa.Analyze(metrics)
		assert.Empty(t, bottlenecks)
	})

	t.Run("boundary values", func(t *testing.T) {
		pa := NewPatternAnalyzer()

		// Exactly at threshold (should not trigger)
		// Note: frequent_rerender requires RenderCount > 1000 AND AvgRenderTime < 1ms
		// inefficient_render requires RenderCount > 500 AND AvgRenderTime > 5ms
		// We set AvgRenderTime to exactly 5ms to avoid triggering inefficient_render
		metrics := &ComponentMetrics{
			ComponentName: "BoundaryComponent",
			RenderCount:   1000,                   // Exactly at frequent_rerender threshold (needs > 1000)
			AvgRenderTime: 5 * time.Millisecond,   // Between thresholds (not < 1ms, not > 10ms, not > 5ms for inefficient)
			MaxRenderTime: 100 * time.Millisecond, // Exactly at render_spike threshold (needs > 100ms)
			MemoryUsage:   5 * 1024 * 1024,        // Exactly at memory_hog threshold (needs > 5MB)
		}
		bottlenecks := pa.Analyze(metrics)
		// At boundary, patterns should NOT trigger (> not >=)
		assert.Empty(t, bottlenecks)
	})

	t.Run("just above threshold", func(t *testing.T) {
		pa := NewPatternAnalyzer()

		// Just above threshold (should trigger)
		metrics := &ComponentMetrics{
			ComponentName: "AboveThresholdComponent",
			AvgRenderTime: 11 * time.Millisecond, // Just above 10ms
		}
		bottlenecks := pa.Analyze(metrics)
		assert.Len(t, bottlenecks, 1)
		assert.Equal(t, "slow_render", pa.GetPatterns()[1].Name) // Verify pattern name
	})
}

func TestCalculatePatternImpact_UnknownSeverity(t *testing.T) {
	// Test with an unknown severity value
	impact := calculatePatternImpact(Severity("unknown"))
	assert.Equal(t, 0.0, impact)
}
