// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewThresholdMonitor(t *testing.T) {
	t.Run("creates new threshold monitor with default config", func(t *testing.T) {
		tm := NewThresholdMonitor()

		assert.NotNil(t, tm)
		assert.NotNil(t, tm.thresholds)
		assert.NotNil(t, tm.violations)
		assert.NotNil(t, tm.alerts)
		assert.NotNil(t, tm.config)
	})

	t.Run("creates threshold monitor with custom config", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 5 * time.Millisecond,
			AlertCooldown:    1 * time.Second,
			MaxAlerts:        50,
			EnableAlerts:     true,
		}
		tm := NewThresholdMonitorWithConfig(config)

		assert.NotNil(t, tm)
		assert.Equal(t, 5*time.Millisecond, tm.config.DefaultThreshold)
		assert.Equal(t, 1*time.Second, tm.config.AlertCooldown)
		assert.Equal(t, 50, tm.config.MaxAlerts)
		assert.True(t, tm.config.EnableAlerts)
	})

	t.Run("handles nil config gracefully", func(t *testing.T) {
		tm := NewThresholdMonitorWithConfig(nil)

		assert.NotNil(t, tm)
		assert.NotNil(t, tm.config)
	})
}

func TestDefaultThresholdConfig(t *testing.T) {
	t.Run("returns sensible defaults", func(t *testing.T) {
		config := DefaultThresholdConfig()

		assert.NotNil(t, config)
		assert.Greater(t, config.DefaultThreshold, time.Duration(0))
		assert.Greater(t, config.AlertCooldown, time.Duration(0))
		assert.Greater(t, config.MaxAlerts, 0)
	})
}

func TestThresholdMonitor_Check(t *testing.T) {
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
			tm := NewThresholdMonitor()
			tm.SetThreshold(tt.operation, tt.threshold)

			bottleneck := tm.Check(tt.operation, tt.duration)

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

func TestThresholdMonitor_ViolationTracking(t *testing.T) {
	t.Run("tracks violations count", func(t *testing.T) {
		tm := NewThresholdMonitor()
		tm.SetThreshold("render", 10*time.Millisecond)

		// First violation
		tm.Check("render", 20*time.Millisecond)
		assert.Equal(t, 1, tm.GetViolations("render"))

		// Second violation
		tm.Check("render", 30*time.Millisecond)
		assert.Equal(t, 2, tm.GetViolations("render"))

		// No violation (below threshold)
		tm.Check("render", 5*time.Millisecond)
		assert.Equal(t, 2, tm.GetViolations("render"))
	})

	t.Run("tracks violations per operation", func(t *testing.T) {
		tm := NewThresholdMonitor()
		tm.SetThreshold("render", 10*time.Millisecond)
		tm.SetThreshold("update", 5*time.Millisecond)

		tm.Check("render", 20*time.Millisecond)
		tm.Check("render", 30*time.Millisecond)
		tm.Check("update", 10*time.Millisecond)

		assert.Equal(t, 2, tm.GetViolations("render"))
		assert.Equal(t, 1, tm.GetViolations("update"))
		assert.Equal(t, 0, tm.GetViolations("nonexistent"))
	})
}

func TestThresholdMonitor_ThresholdManagement(t *testing.T) {
	t.Run("SetThreshold sets operation threshold", func(t *testing.T) {
		tm := NewThresholdMonitor()

		tm.SetThreshold("render", 20*time.Millisecond)

		assert.Equal(t, 20*time.Millisecond, tm.GetThreshold("render"))
	})

	t.Run("GetThreshold returns default for unknown operation", func(t *testing.T) {
		tm := NewThresholdMonitor()

		threshold := tm.GetThreshold("unknown")

		assert.Equal(t, tm.config.DefaultThreshold, threshold)
	})

	t.Run("GetThreshold returns operation-specific threshold", func(t *testing.T) {
		tm := NewThresholdMonitor()
		tm.SetThreshold("custom", 100*time.Millisecond)

		threshold := tm.GetThreshold("custom")

		assert.Equal(t, 100*time.Millisecond, threshold)
	})
}

func TestThresholdMonitor_MultipleOperations(t *testing.T) {
	t.Run("handles multiple operations independently", func(t *testing.T) {
		tm := NewThresholdMonitor()
		tm.SetThreshold("render", 10*time.Millisecond)
		tm.SetThreshold("update", 5*time.Millisecond)
		tm.SetThreshold("event", 2*time.Millisecond)

		// Check render - should pass
		result1 := tm.Check("render", 5*time.Millisecond)
		assert.Nil(t, result1)

		// Check update - should fail
		result2 := tm.Check("update", 10*time.Millisecond)
		assert.NotNil(t, result2)

		// Check event - should fail
		result3 := tm.Check("event", 5*time.Millisecond)
		assert.NotNil(t, result3)

		// Verify violations are tracked independently
		assert.Equal(t, 0, tm.GetViolations("render"))
		assert.Equal(t, 1, tm.GetViolations("update"))
		assert.Equal(t, 1, tm.GetViolations("event"))
	})
}

func TestThresholdMonitor_AlertGeneration(t *testing.T) {
	t.Run("generates alerts when enabled", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 10 * time.Millisecond,
			AlertCooldown:    0, // No cooldown for testing
			MaxAlerts:        100,
			EnableAlerts:     true,
		}
		tm := NewThresholdMonitorWithConfig(config)
		tm.SetThreshold("render", 10*time.Millisecond)

		// Trigger violation
		tm.Check("render", 50*time.Millisecond)

		alerts := tm.GetAlerts()
		require.Len(t, alerts, 1)
		assert.Equal(t, "render", alerts[0].Operation)
		assert.Equal(t, 50*time.Millisecond, alerts[0].Duration)
		assert.Equal(t, 10*time.Millisecond, alerts[0].Threshold)
		assert.NotEmpty(t, alerts[0].Description)
	})

	t.Run("does not generate alerts when disabled", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 10 * time.Millisecond,
			EnableAlerts:     false,
		}
		tm := NewThresholdMonitorWithConfig(config)
		tm.SetThreshold("render", 10*time.Millisecond)

		// Trigger violation
		tm.Check("render", 50*time.Millisecond)

		alerts := tm.GetAlerts()
		assert.Empty(t, alerts)
	})
}

func TestThresholdMonitor_AlertHandler(t *testing.T) {
	t.Run("calls handler on alert", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 10 * time.Millisecond,
			AlertCooldown:    0,
			MaxAlerts:        100,
			EnableAlerts:     true,
		}
		tm := NewThresholdMonitorWithConfig(config)
		tm.SetThreshold("render", 10*time.Millisecond)

		var receivedAlert *Alert
		tm.SetAlertHandler(func(alert *Alert) {
			receivedAlert = alert
		})

		// Trigger violation
		tm.Check("render", 50*time.Millisecond)

		require.NotNil(t, receivedAlert)
		assert.Equal(t, "render", receivedAlert.Operation)
		assert.Equal(t, 50*time.Millisecond, receivedAlert.Duration)
	})

	t.Run("handler not called when alerts disabled", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 10 * time.Millisecond,
			EnableAlerts:     false,
		}
		tm := NewThresholdMonitorWithConfig(config)
		tm.SetThreshold("render", 10*time.Millisecond)

		handlerCalled := false
		tm.SetAlertHandler(func(alert *Alert) {
			handlerCalled = true
		})

		// Trigger violation
		tm.Check("render", 50*time.Millisecond)

		assert.False(t, handlerCalled)
	})
}

func TestThresholdMonitor_AlertCooldown(t *testing.T) {
	t.Run("respects cooldown between alerts for same operation", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 10 * time.Millisecond,
			AlertCooldown:    100 * time.Millisecond,
			MaxAlerts:        100,
			EnableAlerts:     true,
		}
		tm := NewThresholdMonitorWithConfig(config)
		tm.SetThreshold("render", 10*time.Millisecond)

		// First violation - should generate alert
		tm.Check("render", 50*time.Millisecond)

		// Second violation immediately - should NOT generate alert (cooldown)
		tm.Check("render", 60*time.Millisecond)

		alerts := tm.GetAlerts()
		assert.Len(t, alerts, 1) // Only one alert due to cooldown
	})

	t.Run("allows alerts after cooldown expires", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 10 * time.Millisecond,
			AlertCooldown:    10 * time.Millisecond, // Short cooldown for testing
			MaxAlerts:        100,
			EnableAlerts:     true,
		}
		tm := NewThresholdMonitorWithConfig(config)
		tm.SetThreshold("render", 10*time.Millisecond)

		// First violation
		tm.Check("render", 50*time.Millisecond)

		// Wait for cooldown
		time.Sleep(15 * time.Millisecond)

		// Second violation after cooldown
		tm.Check("render", 60*time.Millisecond)

		alerts := tm.GetAlerts()
		assert.Len(t, alerts, 2) // Both alerts should be generated
	})

	t.Run("cooldown is per-operation", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 10 * time.Millisecond,
			AlertCooldown:    100 * time.Millisecond,
			MaxAlerts:        100,
			EnableAlerts:     true,
		}
		tm := NewThresholdMonitorWithConfig(config)
		tm.SetThreshold("render", 10*time.Millisecond)
		tm.SetThreshold("update", 10*time.Millisecond)

		// Violation on render
		tm.Check("render", 50*time.Millisecond)

		// Violation on update - should generate alert (different operation)
		tm.Check("update", 50*time.Millisecond)

		alerts := tm.GetAlerts()
		assert.Len(t, alerts, 2) // Both alerts should be generated
	})
}

func TestThresholdMonitor_MaxAlerts(t *testing.T) {
	t.Run("limits alert history size", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 10 * time.Millisecond,
			AlertCooldown:    0, // No cooldown
			MaxAlerts:        5,
			EnableAlerts:     true,
		}
		tm := NewThresholdMonitorWithConfig(config)
		tm.SetThreshold("render", 10*time.Millisecond)

		// Generate 10 alerts
		for i := 0; i < 10; i++ {
			tm.Check("render", 50*time.Millisecond)
		}

		alerts := tm.GetAlerts()
		assert.Len(t, alerts, 5) // Limited to MaxAlerts
	})

	t.Run("keeps most recent alerts", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 10 * time.Millisecond,
			AlertCooldown:    0,
			MaxAlerts:        3,
			EnableAlerts:     true,
		}
		tm := NewThresholdMonitorWithConfig(config)

		// Generate alerts with different operations to distinguish them
		tm.SetThreshold("op1", 10*time.Millisecond)
		tm.SetThreshold("op2", 10*time.Millisecond)
		tm.SetThreshold("op3", 10*time.Millisecond)
		tm.SetThreshold("op4", 10*time.Millisecond)
		tm.SetThreshold("op5", 10*time.Millisecond)

		tm.Check("op1", 50*time.Millisecond)
		tm.Check("op2", 50*time.Millisecond)
		tm.Check("op3", 50*time.Millisecond)
		tm.Check("op4", 50*time.Millisecond)
		tm.Check("op5", 50*time.Millisecond)

		alerts := tm.GetAlerts()
		require.Len(t, alerts, 3)

		// Should have the most recent alerts (op3, op4, op5)
		operations := make([]string, len(alerts))
		for i, a := range alerts {
			operations[i] = a.Operation
		}
		assert.Contains(t, operations, "op3")
		assert.Contains(t, operations, "op4")
		assert.Contains(t, operations, "op5")
	})
}

func TestThresholdMonitor_ClearAlerts(t *testing.T) {
	t.Run("clears alert history", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 10 * time.Millisecond,
			AlertCooldown:    0,
			MaxAlerts:        100,
			EnableAlerts:     true,
		}
		tm := NewThresholdMonitorWithConfig(config)
		tm.SetThreshold("render", 10*time.Millisecond)

		// Generate alerts
		tm.Check("render", 50*time.Millisecond)
		tm.Check("render", 60*time.Millisecond)

		assert.Len(t, tm.GetAlerts(), 2)

		tm.ClearAlerts()

		assert.Empty(t, tm.GetAlerts())
	})

	t.Run("does not clear violations", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 10 * time.Millisecond,
			AlertCooldown:    0,
			MaxAlerts:        100,
			EnableAlerts:     true,
		}
		tm := NewThresholdMonitorWithConfig(config)
		tm.SetThreshold("render", 10*time.Millisecond)

		// Generate violations
		tm.Check("render", 50*time.Millisecond)
		tm.Check("render", 60*time.Millisecond)

		tm.ClearAlerts()

		// Violations should still be tracked
		assert.Equal(t, 2, tm.GetViolations("render"))
	})
}

func TestThresholdMonitor_Reset(t *testing.T) {
	t.Run("resets all state", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 10 * time.Millisecond,
			AlertCooldown:    0,
			MaxAlerts:        100,
			EnableAlerts:     true,
		}
		tm := NewThresholdMonitorWithConfig(config)
		tm.SetThreshold("render", 5*time.Millisecond)
		tm.Check("render", 20*time.Millisecond)
		tm.Check("render", 30*time.Millisecond)

		assert.Equal(t, 2, tm.GetViolations("render"))
		assert.Len(t, tm.GetAlerts(), 2)

		tm.Reset()

		assert.Equal(t, 0, tm.GetViolations("render"))
		assert.Empty(t, tm.GetAlerts())
		// Custom thresholds should be cleared
		assert.Equal(t, tm.config.DefaultThreshold, tm.GetThreshold("render"))
	})
}

func TestThresholdMonitor_GetAllViolations(t *testing.T) {
	t.Run("returns all violations", func(t *testing.T) {
		tm := NewThresholdMonitor()
		tm.SetThreshold("render", 10*time.Millisecond)
		tm.SetThreshold("update", 5*time.Millisecond)

		tm.Check("render", 20*time.Millisecond)
		tm.Check("render", 30*time.Millisecond)
		tm.Check("update", 10*time.Millisecond)

		violations := tm.GetAllViolations()

		assert.Equal(t, 2, violations["render"])
		assert.Equal(t, 1, violations["update"])
	})

	t.Run("returns copy of violations map", func(t *testing.T) {
		tm := NewThresholdMonitor()
		tm.SetThreshold("render", 10*time.Millisecond)
		tm.Check("render", 20*time.Millisecond)

		violations := tm.GetAllViolations()
		violations["render"] = 999 // Modify the copy

		// Original should be unchanged
		assert.Equal(t, 1, tm.GetViolations("render"))
	})
}

func TestThresholdMonitor_ThreadSafety(t *testing.T) {
	t.Run("concurrent operations are safe", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 10 * time.Millisecond,
			AlertCooldown:    0,
			MaxAlerts:        1000,
			EnableAlerts:     true,
		}
		tm := NewThresholdMonitorWithConfig(config)
		tm.SetThreshold("render", 10*time.Millisecond)
		var wg sync.WaitGroup
		goroutines := 50

		// Concurrent Check operations
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = tm.Check("render", 20*time.Millisecond)
			}()
		}

		// Concurrent SetThreshold operations
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				tm.SetThreshold("render", time.Duration(i+1)*time.Millisecond)
			}(i)
		}

		// Concurrent GetViolations operations
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = tm.GetViolations("render")
			}()
		}

		// Concurrent GetAlerts operations
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = tm.GetAlerts()
			}()
		}

		wg.Wait()
		// Should complete without race conditions
	})

	t.Run("concurrent alert handler calls are safe", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 10 * time.Millisecond,
			AlertCooldown:    0,
			MaxAlerts:        1000,
			EnableAlerts:     true,
		}
		tm := NewThresholdMonitorWithConfig(config)
		tm.SetThreshold("render", 10*time.Millisecond)

		var alertCount int64
		tm.SetAlertHandler(func(alert *Alert) {
			atomic.AddInt64(&alertCount, 1)
		})

		var wg sync.WaitGroup
		goroutines := 50

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				tm.Check("render", 20*time.Millisecond)
			}()
		}

		wg.Wait()

		// All alerts should have been handled
		assert.Equal(t, int64(goroutines), atomic.LoadInt64(&alertCount))
	})
}

func TestThresholdMonitor_SeverityCalculation(t *testing.T) {
	tests := []struct {
		name         string
		duration     time.Duration
		threshold    time.Duration
		wantSeverity Severity
	}{
		{
			name:         "low severity (< 2x)",
			duration:     18 * time.Millisecond,
			threshold:    16 * time.Millisecond,
			wantSeverity: SeverityLow,
		},
		{
			name:         "medium severity (2-3x)",
			duration:     40 * time.Millisecond,
			threshold:    16 * time.Millisecond,
			wantSeverity: SeverityMedium,
		},
		{
			name:         "high severity (3-5x)",
			duration:     64 * time.Millisecond,
			threshold:    16 * time.Millisecond,
			wantSeverity: SeverityHigh,
		},
		{
			name:         "critical severity (> 5x)",
			duration:     100 * time.Millisecond,
			threshold:    16 * time.Millisecond,
			wantSeverity: SeverityCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewThresholdMonitor()
			tm.SetThreshold("render", tt.threshold)

			bottleneck := tm.Check("render", tt.duration)

			require.NotNil(t, bottleneck)
			assert.Equal(t, tt.wantSeverity, bottleneck.Severity)
		})
	}
}

func TestThresholdMonitor_AlertTimestamp(t *testing.T) {
	t.Run("alert has valid timestamp", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 10 * time.Millisecond,
			AlertCooldown:    0,
			MaxAlerts:        100,
			EnableAlerts:     true,
		}
		tm := NewThresholdMonitorWithConfig(config)
		tm.SetThreshold("render", 10*time.Millisecond)

		before := time.Now()
		tm.Check("render", 50*time.Millisecond)
		after := time.Now()

		alerts := tm.GetAlerts()
		require.Len(t, alerts, 1)

		// Timestamp should be between before and after
		assert.True(t, alerts[0].Timestamp.After(before) || alerts[0].Timestamp.Equal(before))
		assert.True(t, alerts[0].Timestamp.Before(after) || alerts[0].Timestamp.Equal(after))
	})
}

func TestThresholdMonitor_AlertSeverity(t *testing.T) {
	t.Run("alert severity matches bottleneck severity", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 10 * time.Millisecond,
			AlertCooldown:    0,
			MaxAlerts:        100,
			EnableAlerts:     true,
		}
		tm := NewThresholdMonitorWithConfig(config)
		tm.SetThreshold("render", 10*time.Millisecond)

		// Trigger critical severity (> 5x threshold)
		bottleneck := tm.Check("render", 100*time.Millisecond)

		alerts := tm.GetAlerts()
		require.Len(t, alerts, 1)
		require.NotNil(t, bottleneck)

		assert.Equal(t, bottleneck.Severity, alerts[0].Severity)
		assert.Equal(t, SeverityCritical, alerts[0].Severity)
	})
}

func TestThresholdMonitor_GetConfig(t *testing.T) {
	t.Run("returns current config", func(t *testing.T) {
		config := &ThresholdConfig{
			DefaultThreshold: 20 * time.Millisecond,
			AlertCooldown:    5 * time.Second,
			MaxAlerts:        50,
			EnableAlerts:     true,
		}
		tm := NewThresholdMonitorWithConfig(config)

		got := tm.GetConfig()

		assert.Equal(t, 20*time.Millisecond, got.DefaultThreshold)
		assert.Equal(t, 5*time.Second, got.AlertCooldown)
		assert.Equal(t, 50, got.MaxAlerts)
		assert.True(t, got.EnableAlerts)
	})

	t.Run("returns copy of config", func(t *testing.T) {
		tm := NewThresholdMonitor()

		got := tm.GetConfig()
		got.DefaultThreshold = 999 * time.Second // Modify the copy

		// Original should be unchanged
		assert.NotEqual(t, 999*time.Second, tm.config.DefaultThreshold)
	})
}

func TestThresholdMonitor_ImpactCalculation(t *testing.T) {
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
			name:       "impact capped at 1.0",
			duration:   200 * time.Millisecond,
			threshold:  16 * time.Millisecond,
			wantImpact: 1.0, // Capped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewThresholdMonitor()
			tm.SetThreshold("render", tt.threshold)

			bottleneck := tm.Check("render", tt.duration)

			require.NotNil(t, bottleneck)
			assert.InDelta(t, tt.wantImpact, bottleneck.Impact, 0.01)
		})
	}
}
