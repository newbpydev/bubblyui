// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewDevToolsIntegration tests the constructor.
func TestNewDevToolsIntegration(t *testing.T) {
	tests := []struct {
		name     string
		profiler *Profiler
		wantNil  bool
	}{
		{
			name:     "with nil profiler creates default",
			profiler: nil,
			wantNil:  false,
		},
		{
			name:     "with valid profiler",
			profiler: New(),
			wantNil:  false,
		},
		{
			name:     "with enabled profiler",
			profiler: New(WithEnabled(true)),
			wantNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dti := NewDevToolsIntegration(tt.profiler)
			if tt.wantNil {
				assert.Nil(t, dti)
			} else {
				assert.NotNil(t, dti)
				assert.NotNil(t, dti.profiler)
				assert.NotNil(t, dti.metricsBuffer)
			}
		})
	}
}

// TestDevToolsIntegration_Enable tests enabling the integration.
func TestDevToolsIntegration_Enable(t *testing.T) {
	dti := NewDevToolsIntegration(nil)
	require.NotNil(t, dti)

	// Initially disabled
	assert.False(t, dti.IsEnabled())

	// Enable
	dti.Enable()
	assert.True(t, dti.IsEnabled())

	// Enable again (idempotent)
	dti.Enable()
	assert.True(t, dti.IsEnabled())
}

// TestDevToolsIntegration_Disable tests disabling the integration.
func TestDevToolsIntegration_Disable(t *testing.T) {
	dti := NewDevToolsIntegration(nil)
	require.NotNil(t, dti)

	// Enable first
	dti.Enable()
	assert.True(t, dti.IsEnabled())

	// Disable
	dti.Disable()
	assert.False(t, dti.IsEnabled())

	// Disable again (idempotent)
	dti.Disable()
	assert.False(t, dti.IsEnabled())
}

// TestDevToolsIntegration_SendMetrics tests sending metrics to devtools.
func TestDevToolsIntegration_SendMetrics(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		setupFunc   func(*DevToolsIntegration)
		wantMetrics bool
	}{
		{
			name:        "disabled does nothing",
			enabled:     false,
			setupFunc:   nil,
			wantMetrics: false,
		},
		{
			name:    "enabled sends metrics",
			enabled: true,
			setupFunc: func(dti *DevToolsIntegration) {
				// Record some metrics
				dti.profiler.collector.GetTimings().Record("test.op", 5*time.Millisecond)
			},
			wantMetrics: true,
		},
		{
			name:        "enabled with no metrics",
			enabled:     true,
			setupFunc:   nil,
			wantMetrics: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prof := New()
			_ = prof.Start()
			dti := NewDevToolsIntegration(prof)

			if tt.enabled {
				dti.Enable()
			}

			if tt.setupFunc != nil {
				tt.setupFunc(dti)
			}

			// SendMetrics should not panic
			dti.SendMetrics()

			// Verify metrics were buffered if enabled
			if tt.wantMetrics {
				assert.Greater(t, dti.GetMetricsCount(), 0)
			}
		})
	}
}

// TestDevToolsIntegration_RegisterPanel tests panel registration.
func TestDevToolsIntegration_RegisterPanel(t *testing.T) {
	tests := []struct {
		name       string
		panelName  string
		wantErr    bool
		wantPanels int
	}{
		{
			name:       "register performance panel",
			panelName:  "Performance",
			wantErr:    false,
			wantPanels: 1,
		},
		{
			name:       "register empty name",
			panelName:  "",
			wantErr:    true,
			wantPanels: 0,
		},
		{
			name:       "register custom panel",
			panelName:  "CustomProfiler",
			wantErr:    false,
			wantPanels: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dti := NewDevToolsIntegration(nil)
			require.NotNil(t, dti)

			err := dti.RegisterPanel(tt.panelName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantPanels, dti.GetPanelCount())
			}
		})
	}
}

// TestDevToolsIntegration_GetMetricsSnapshot tests getting metrics snapshot.
func TestDevToolsIntegration_GetMetricsSnapshot(t *testing.T) {
	prof := New()
	_ = prof.Start()
	dti := NewDevToolsIntegration(prof)
	dti.Enable()

	// Record some metrics
	prof.collector.GetTimings().Record("render.component1", 10*time.Millisecond)
	prof.collector.GetTimings().Record("render.component2", 5*time.Millisecond)
	prof.collector.GetTimings().Record("update.component1", 2*time.Millisecond)

	// Send metrics to buffer
	dti.SendMetrics()

	// Get snapshot
	snapshot := dti.GetMetricsSnapshot()
	assert.NotNil(t, snapshot)
	assert.NotNil(t, snapshot.Timings)
	assert.NotNil(t, snapshot.Timestamp)
}

// TestDevToolsIntegration_ClearMetrics tests clearing metrics buffer.
func TestDevToolsIntegration_ClearMetrics(t *testing.T) {
	prof := New()
	_ = prof.Start()
	dti := NewDevToolsIntegration(prof)
	dti.Enable()

	// Record and send metrics
	prof.collector.GetTimings().Record("test.op", 5*time.Millisecond)
	dti.SendMetrics()
	assert.Greater(t, dti.GetMetricsCount(), 0)

	// Clear metrics
	dti.ClearMetrics()
	assert.Equal(t, 0, dti.GetMetricsCount())
}

// TestDevToolsIntegration_SetUpdateInterval tests setting update interval.
func TestDevToolsIntegration_SetUpdateInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		wantErr  bool
	}{
		{
			name:     "valid interval 100ms",
			interval: 100 * time.Millisecond,
			wantErr:  false,
		},
		{
			name:     "valid interval 1s",
			interval: time.Second,
			wantErr:  false,
		},
		{
			name:     "zero interval uses default",
			interval: 0,
			wantErr:  false,
		},
		{
			name:     "negative interval error",
			interval: -1 * time.Second,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dti := NewDevToolsIntegration(nil)
			require.NotNil(t, dti)

			err := dti.SetUpdateInterval(tt.interval)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.interval > 0 {
					assert.Equal(t, tt.interval, dti.GetUpdateInterval())
				}
			}
		})
	}
}

// TestDevToolsIntegration_ThreadSafety tests concurrent access.
func TestDevToolsIntegration_ThreadSafety(t *testing.T) {
	prof := New()
	_ = prof.Start()
	dti := NewDevToolsIntegration(prof)
	dti.Enable()

	var wg sync.WaitGroup
	const goroutines = 50
	const iterations = 100

	// Concurrent operations
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Mix of operations
				switch j % 5 {
				case 0:
					dti.SendMetrics()
				case 1:
					_ = dti.GetMetricsSnapshot()
				case 2:
					_ = dti.IsEnabled()
				case 3:
					_ = dti.GetMetricsCount()
				case 4:
					_ = dti.GetUpdateInterval()
				}
			}
		}(i)
	}

	wg.Wait()
	// If we get here without deadlock or panic, test passes
}

// TestDevToolsIntegration_PerformanceOverhead tests minimal overhead.
func TestDevToolsIntegration_PerformanceOverhead(t *testing.T) {
	prof := New()
	dti := NewDevToolsIntegration(prof)

	// Measure overhead when disabled
	start := time.Now()
	for i := 0; i < 10000; i++ {
		dti.SendMetrics()
	}
	disabledDuration := time.Since(start)

	// Should be very fast when disabled (< 1ms for 10000 calls)
	assert.Less(t, disabledDuration, 10*time.Millisecond,
		"Disabled overhead too high: %v", disabledDuration)
}

// TestDevToolsIntegration_RealTimeUpdates tests real-time update mechanism.
func TestDevToolsIntegration_RealTimeUpdates(t *testing.T) {
	prof := New()
	_ = prof.Start()
	dti := NewDevToolsIntegration(prof)
	dti.Enable()

	// Set up callback for updates
	updateCount := 0
	var updateMu sync.Mutex
	dti.OnMetricsUpdate(func(snapshot *MetricsSnapshot) {
		updateMu.Lock()
		updateCount++
		updateMu.Unlock()
	})

	// Record metrics and send
	prof.collector.GetTimings().Record("test.op", 5*time.Millisecond)
	dti.SendMetrics()

	// Verify callback was called
	updateMu.Lock()
	count := updateCount
	updateMu.Unlock()
	assert.GreaterOrEqual(t, count, 1)
}

// TestDevToolsIntegration_NoBreakingChanges tests that integration doesn't break profiler.
func TestDevToolsIntegration_NoBreakingChanges(t *testing.T) {
	// Create profiler with integration
	prof := New()
	dti := NewDevToolsIntegration(prof)
	dti.Enable()

	// Start profiler
	err := prof.Start()
	assert.NoError(t, err)

	// Record metrics normally
	prof.collector.GetTimings().Record("test.op", 5*time.Millisecond)

	// Send to devtools
	dti.SendMetrics()

	// Stop profiler
	err = prof.Stop()
	assert.NoError(t, err)

	// Generate report (should still work)
	report := prof.GenerateReport()
	assert.NotNil(t, report)
}

// TestDevToolsIntegration_GetProfiler tests getting the profiler reference.
func TestDevToolsIntegration_GetProfiler(t *testing.T) {
	prof := New()
	dti := NewDevToolsIntegration(prof)

	assert.Equal(t, prof, dti.GetProfiler())
}

// TestDevToolsIntegration_Reset tests resetting the integration.
func TestDevToolsIntegration_Reset(t *testing.T) {
	prof := New()
	_ = prof.Start()
	dti := NewDevToolsIntegration(prof)
	dti.Enable()

	// Record metrics
	prof.collector.GetTimings().Record("test.op", 5*time.Millisecond)
	dti.SendMetrics()
	_ = dti.RegisterPanel("Test")

	// Reset
	dti.Reset()

	// Verify reset state
	assert.Equal(t, 0, dti.GetMetricsCount())
	assert.Equal(t, 0, dti.GetPanelCount())
}

// TestMetricsSnapshot tests the MetricsSnapshot struct.
func TestMetricsSnapshot(t *testing.T) {
	snapshot := &MetricsSnapshot{
		Timings:        make(map[string]*TimingSnapshot),
		Components:     make([]*ComponentMetrics, 0),
		FPS:            60.0,
		DroppedFrames:  0.5,
		MemoryUsage:    1024 * 1024,
		GoroutineCount: 10,
		Timestamp:      time.Now(),
	}

	assert.NotNil(t, snapshot.Timings)
	assert.Equal(t, 60.0, snapshot.FPS)
	assert.Equal(t, 0.5, snapshot.DroppedFrames)
}

// TestTimingSnapshot tests the TimingSnapshot struct.
func TestTimingSnapshot(t *testing.T) {
	snapshot := &TimingSnapshot{
		Name:  "render.component",
		Count: 100,
		Total: 500 * time.Millisecond,
		Min:   1 * time.Millisecond,
		Max:   20 * time.Millisecond,
		Mean:  5 * time.Millisecond,
		P50:   4 * time.Millisecond,
		P95:   15 * time.Millisecond,
		P99:   18 * time.Millisecond,
	}

	assert.Equal(t, "render.component", snapshot.Name)
	assert.Equal(t, int64(100), snapshot.Count)
}

// TestDevToolsIntegration_MultipleCallbacks tests multiple update callbacks.
func TestDevToolsIntegration_MultipleCallbacks(t *testing.T) {
	prof := New()
	_ = prof.Start()
	dti := NewDevToolsIntegration(prof)
	dti.Enable()

	// Register multiple callbacks
	var count1, count2 int
	var mu sync.Mutex

	dti.OnMetricsUpdate(func(snapshot *MetricsSnapshot) {
		mu.Lock()
		count1++
		mu.Unlock()
	})

	dti.OnMetricsUpdate(func(snapshot *MetricsSnapshot) {
		mu.Lock()
		count2++
		mu.Unlock()
	})

	// Send metrics
	prof.collector.GetTimings().Record("test.op", 5*time.Millisecond)
	dti.SendMetrics()

	// Both callbacks should be called
	mu.Lock()
	assert.GreaterOrEqual(t, count1, 1)
	assert.GreaterOrEqual(t, count2, 1)
	mu.Unlock()
}

// TestDevToolsIntegration_PanelExists tests checking if panel exists.
func TestDevToolsIntegration_PanelExists(t *testing.T) {
	dti := NewDevToolsIntegration(nil)

	// Register panel
	err := dti.RegisterPanel("Performance")
	require.NoError(t, err)

	// Check existence
	assert.True(t, dti.PanelExists("Performance"))
	assert.False(t, dti.PanelExists("NonExistent"))
}

// TestDevToolsIntegration_UnregisterPanel tests unregistering a panel.
func TestDevToolsIntegration_UnregisterPanel(t *testing.T) {
	dti := NewDevToolsIntegration(nil)

	// Register panel
	err := dti.RegisterPanel("Performance")
	require.NoError(t, err)
	assert.Equal(t, 1, dti.GetPanelCount())

	// Unregister
	dti.UnregisterPanel("Performance")
	assert.Equal(t, 0, dti.GetPanelCount())
	assert.False(t, dti.PanelExists("Performance"))

	// Unregister non-existent (should not panic)
	dti.UnregisterPanel("NonExistent")
}

// TestDevToolsIntegration_GetPanelNames tests getting panel names.
func TestDevToolsIntegration_GetPanelNames(t *testing.T) {
	dti := NewDevToolsIntegration(nil)

	// Register panels
	_ = dti.RegisterPanel("Performance")
	_ = dti.RegisterPanel("Memory")
	_ = dti.RegisterPanel("CPU")

	names := dti.GetPanelNames()
	assert.Len(t, names, 3)
	assert.Contains(t, names, "Performance")
	assert.Contains(t, names, "Memory")
	assert.Contains(t, names, "CPU")
}

// TestDevToolsIntegration_NilCallback tests nil callback handling.
func TestDevToolsIntegration_NilCallback(t *testing.T) {
	dti := NewDevToolsIntegration(nil)
	dti.Enable()

	// Should not panic with nil callback
	dti.OnMetricsUpdate(nil)
	dti.SendMetrics()
}

// TestDevToolsIntegration_DefaultUpdateInterval tests default interval.
func TestDevToolsIntegration_DefaultUpdateInterval(t *testing.T) {
	dti := NewDevToolsIntegration(nil)

	// Default should be reasonable (e.g., 100ms)
	interval := dti.GetUpdateInterval()
	assert.Greater(t, interval, time.Duration(0))
	assert.LessOrEqual(t, interval, time.Second)
}

// BenchmarkDevToolsIntegration_SendMetrics_Disabled benchmarks disabled path.
func BenchmarkDevToolsIntegration_SendMetrics_Disabled(b *testing.B) {
	dti := NewDevToolsIntegration(nil)
	// Disabled by default

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dti.SendMetrics()
	}
}

// BenchmarkDevToolsIntegration_SendMetrics_Enabled benchmarks enabled path.
func BenchmarkDevToolsIntegration_SendMetrics_Enabled(b *testing.B) {
	prof := New()
	_ = prof.Start()
	dti := NewDevToolsIntegration(prof)
	dti.Enable()

	// Add some metrics
	for i := 0; i < 10; i++ {
		prof.collector.GetTimings().Record("test.op", 5*time.Millisecond)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dti.SendMetrics()
	}
}

// BenchmarkDevToolsIntegration_GetMetricsSnapshot benchmarks snapshot creation.
func BenchmarkDevToolsIntegration_GetMetricsSnapshot(b *testing.B) {
	prof := New()
	_ = prof.Start()
	dti := NewDevToolsIntegration(prof)
	dti.Enable()

	// Add metrics
	for i := 0; i < 100; i++ {
		prof.collector.GetTimings().Record("test.op", 5*time.Millisecond)
	}
	dti.SendMetrics()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dti.GetMetricsSnapshot()
	}
}
