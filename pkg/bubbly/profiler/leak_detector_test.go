// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLeakDetector(t *testing.T) {
	t.Run("creates new leak detector with default thresholds", func(t *testing.T) {
		ld := NewLeakDetector()

		assert.NotNil(t, ld)
		assert.NotNil(t, ld.thresholds)
	})

	t.Run("creates leak detector with custom thresholds", func(t *testing.T) {
		thresholds := &LeakThresholds{
			HeapGrowthBytes:       1024 * 1024, // 1MB
			GoroutineGrowth:       5,
			HeapObjectGrowth:      100,
			GCPauseThreshold:      0,
			SeverityHighBytes:     10 * 1024 * 1024,
			SeverityCriticalBytes: 100 * 1024 * 1024,
		}
		ld := NewLeakDetectorWithThresholds(thresholds)

		assert.NotNil(t, ld)
		assert.Equal(t, int64(1024*1024), ld.thresholds.HeapGrowthBytes)
		assert.Equal(t, 5, ld.thresholds.GoroutineGrowth)
	})
}

func TestDefaultLeakThresholds(t *testing.T) {
	t.Run("returns sensible defaults", func(t *testing.T) {
		thresholds := DefaultLeakThresholds()

		assert.NotNil(t, thresholds)
		assert.Greater(t, thresholds.HeapGrowthBytes, int64(0))
		assert.Greater(t, thresholds.GoroutineGrowth, 0)
		assert.Greater(t, thresholds.HeapObjectGrowth, int64(0))
		assert.Greater(t, thresholds.SeverityHighBytes, int64(0))
		assert.Greater(t, thresholds.SeverityCriticalBytes, thresholds.SeverityHighBytes)
	})
}

func TestLeakDetector_DetectLeaks(t *testing.T) {
	tests := []struct {
		name           string
		setupSnapshots func() []*runtime.MemStats
		wantLeakCount  int
		wantLeakTypes  []string
	}{
		{
			name: "no leaks with stable memory",
			setupSnapshots: func() []*runtime.MemStats {
				return []*runtime.MemStats{
					{HeapAlloc: 1000, HeapObjects: 10},
					{HeapAlloc: 1000, HeapObjects: 10},
				}
			},
			wantLeakCount: 0,
			wantLeakTypes: nil,
		},
		{
			name: "detects heap growth leak",
			setupSnapshots: func() []*runtime.MemStats {
				return []*runtime.MemStats{
					{HeapAlloc: 1000, HeapObjects: 10},
					{HeapAlloc: 10 * 1024 * 1024, HeapObjects: 10}, // 10MB growth
				}
			},
			wantLeakCount: 1,
			wantLeakTypes: []string{"heap_growth"},
		},
		{
			name: "detects heap object growth leak",
			setupSnapshots: func() []*runtime.MemStats {
				return []*runtime.MemStats{
					{HeapAlloc: 1000, HeapObjects: 10},
					{HeapAlloc: 1000, HeapObjects: 100000}, // 99990 object growth
				}
			},
			wantLeakCount: 1,
			wantLeakTypes: []string{"heap_object_growth"},
		},
		{
			name: "detects multiple leak types",
			setupSnapshots: func() []*runtime.MemStats {
				return []*runtime.MemStats{
					{HeapAlloc: 1000, HeapObjects: 10},
					{HeapAlloc: 50 * 1024 * 1024, HeapObjects: 100000}, // Both grow
				}
			},
			wantLeakCount: 2,
			wantLeakTypes: []string{"heap_growth", "heap_object_growth"},
		},
		{
			name: "nil snapshots returns empty",
			setupSnapshots: func() []*runtime.MemStats {
				return nil
			},
			wantLeakCount: 0,
			wantLeakTypes: nil,
		},
		{
			name: "single snapshot returns empty",
			setupSnapshots: func() []*runtime.MemStats {
				return []*runtime.MemStats{
					{HeapAlloc: 1000, HeapObjects: 10},
				}
			},
			wantLeakCount: 0,
			wantLeakTypes: nil,
		},
		{
			name: "empty snapshots returns empty",
			setupSnapshots: func() []*runtime.MemStats {
				return []*runtime.MemStats{}
			},
			wantLeakCount: 0,
			wantLeakTypes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ld := NewLeakDetector()
			snapshots := tt.setupSnapshots()

			leaks := ld.DetectLeaks(snapshots)

			assert.Len(t, leaks, tt.wantLeakCount)

			if tt.wantLeakTypes != nil {
				for i, wantType := range tt.wantLeakTypes {
					if i < len(leaks) {
						assert.Equal(t, wantType, leaks[i].Type)
					}
				}
			}
		})
	}
}

func TestLeakDetector_DetectGoroutineLeaks(t *testing.T) {
	tests := []struct {
		name         string
		before       int
		after        int
		wantLeak     bool
		wantCount    int
		wantSeverity Severity
	}{
		{
			name:      "no leak with same count",
			before:    10,
			after:     10,
			wantLeak:  false,
			wantCount: 0,
		},
		{
			name:      "no leak with small growth",
			before:    10,
			after:     12,
			wantLeak:  false,
			wantCount: 0,
		},
		{
			name:         "detects goroutine leak low severity",
			before:       10,
			after:        25,
			wantLeak:     true,
			wantCount:    15,
			wantSeverity: SeverityLow,
		},
		{
			name:         "detects goroutine leak medium severity",
			before:       10,
			after:        45,
			wantLeak:     true,
			wantCount:    35,
			wantSeverity: SeverityMedium,
		},
		{
			name:         "detects large goroutine leak",
			before:       10,
			after:        110,
			wantLeak:     true,
			wantCount:    100,
			wantSeverity: SeverityHigh,
		},
		{
			name:         "detects critical goroutine leak",
			before:       10,
			after:        1010,
			wantLeak:     true,
			wantCount:    1000,
			wantSeverity: SeverityCritical,
		},
		{
			name:      "no leak with decrease",
			before:    20,
			after:     10,
			wantLeak:  false,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ld := NewLeakDetector()

			leak := ld.DetectGoroutineLeaks(tt.before, tt.after)

			if tt.wantLeak {
				require.NotNil(t, leak)
				assert.Equal(t, "goroutine_leak", leak.Type)
				assert.Equal(t, tt.wantCount, leak.Count)
				assert.Equal(t, tt.wantSeverity, leak.Severity)
				assert.NotEmpty(t, leak.Description)
			} else {
				assert.Nil(t, leak)
			}
		})
	}
}

func TestLeakDetector_SeverityCalculation(t *testing.T) {
	tests := []struct {
		name         string
		bytesLeaked  int64
		wantSeverity Severity
	}{
		{
			name:         "low severity for small leak",
			bytesLeaked:  1 * 1024 * 1024, // 1MB
			wantSeverity: SeverityLow,
		},
		{
			name:         "medium severity for moderate leak",
			bytesLeaked:  5 * 1024 * 1024, // 5MB
			wantSeverity: SeverityMedium,
		},
		{
			name:         "high severity for large leak",
			bytesLeaked:  50 * 1024 * 1024, // 50MB
			wantSeverity: SeverityHigh,
		},
		{
			name:         "critical severity for huge leak",
			bytesLeaked:  500 * 1024 * 1024, // 500MB
			wantSeverity: SeverityCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ld := NewLeakDetector()

			severity := ld.calculateSeverity(tt.bytesLeaked)

			assert.Equal(t, tt.wantSeverity, severity)
		})
	}
}

func TestLeakDetector_FalsePositiveFiltering(t *testing.T) {
	t.Run("filters out small fluctuations", func(t *testing.T) {
		ld := NewLeakDetector()

		// Small memory fluctuation (below threshold)
		snapshots := []*runtime.MemStats{
			{HeapAlloc: 1000000, HeapObjects: 100},
			{HeapAlloc: 1001000, HeapObjects: 101}, // 1KB growth, 1 object
		}

		leaks := ld.DetectLeaks(snapshots)

		// Should not report as leak (below threshold)
		assert.Empty(t, leaks)
	})

	t.Run("filters out temporary spikes with multiple snapshots", func(t *testing.T) {
		ld := NewLeakDetector()

		// Memory spike then return to normal
		snapshots := []*runtime.MemStats{
			{HeapAlloc: 1000000, HeapObjects: 100},
			{HeapAlloc: 50000000, HeapObjects: 5000}, // Spike
			{HeapAlloc: 1000000, HeapObjects: 100},   // Back to normal
		}

		leaks := ld.DetectLeaks(snapshots)

		// Should not report as leak (first and last are similar)
		assert.Empty(t, leaks)
	})
}

func TestLeakDetector_ThresholdConfiguration(t *testing.T) {
	t.Run("respects custom heap growth threshold", func(t *testing.T) {
		thresholds := &LeakThresholds{
			HeapGrowthBytes:       100, // Very low threshold
			GoroutineGrowth:       10,
			HeapObjectGrowth:      1000,
			SeverityHighBytes:     1000,
			SeverityCriticalBytes: 10000,
		}
		ld := NewLeakDetectorWithThresholds(thresholds)

		snapshots := []*runtime.MemStats{
			{HeapAlloc: 1000, HeapObjects: 10},
			{HeapAlloc: 1200, HeapObjects: 10}, // 200 bytes growth
		}

		leaks := ld.DetectLeaks(snapshots)

		// Should detect with low threshold
		assert.Len(t, leaks, 1)
		assert.Equal(t, "heap_growth", leaks[0].Type)
	})

	t.Run("respects custom goroutine threshold", func(t *testing.T) {
		thresholds := &LeakThresholds{
			HeapGrowthBytes:       1024 * 1024,
			GoroutineGrowth:       2, // Very low threshold
			HeapObjectGrowth:      1000,
			SeverityHighBytes:     10 * 1024 * 1024,
			SeverityCriticalBytes: 100 * 1024 * 1024,
		}
		ld := NewLeakDetectorWithThresholds(thresholds)

		leak := ld.DetectGoroutineLeaks(10, 13) // 3 goroutine growth

		require.NotNil(t, leak)
		assert.Equal(t, 3, leak.Count)
	})
}

func TestLeakDetector_LeakInfoFields(t *testing.T) {
	t.Run("heap leak info has all required fields", func(t *testing.T) {
		ld := NewLeakDetector()

		snapshots := []*runtime.MemStats{
			{HeapAlloc: 1000, HeapObjects: 10},
			{HeapAlloc: 100 * 1024 * 1024, HeapObjects: 10}, // 100MB growth
		}

		leaks := ld.DetectLeaks(snapshots)

		require.Len(t, leaks, 1)
		leak := leaks[0]

		assert.Equal(t, "heap_growth", leak.Type)
		assert.Greater(t, leak.BytesLeaked, int64(0))
		assert.NotEmpty(t, leak.Description)
		assert.NotEqual(t, Severity(""), leak.Severity)
	})

	t.Run("goroutine leak info has all required fields", func(t *testing.T) {
		ld := NewLeakDetector()

		leak := ld.DetectGoroutineLeaks(10, 100)

		require.NotNil(t, leak)
		assert.Equal(t, "goroutine_leak", leak.Type)
		assert.Equal(t, 90, leak.Count)
		assert.NotEmpty(t, leak.Description)
		assert.NotEqual(t, Severity(""), leak.Severity)
	})
}

func TestLeakDetector_ThreadSafety(t *testing.T) {
	t.Run("concurrent operations are safe", func(t *testing.T) {
		ld := NewLeakDetector()
		var wg sync.WaitGroup
		goroutines := 50

		snapshots := []*runtime.MemStats{
			{HeapAlloc: 1000, HeapObjects: 10},
			{HeapAlloc: 100 * 1024 * 1024, HeapObjects: 1000},
		}

		// Concurrent DetectLeaks
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = ld.DetectLeaks(snapshots)
			}()
		}

		// Concurrent DetectGoroutineLeaks
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = ld.DetectGoroutineLeaks(10, 100)
			}()
		}

		wg.Wait()
		// Should complete without race conditions
	})
}

func TestLeakDetector_GetThresholds(t *testing.T) {
	t.Run("returns current thresholds", func(t *testing.T) {
		thresholds := &LeakThresholds{
			HeapGrowthBytes:       5 * 1024 * 1024,
			GoroutineGrowth:       20,
			HeapObjectGrowth:      5000,
			SeverityHighBytes:     50 * 1024 * 1024,
			SeverityCriticalBytes: 500 * 1024 * 1024,
		}
		ld := NewLeakDetectorWithThresholds(thresholds)

		got := ld.GetThresholds()

		assert.Equal(t, thresholds.HeapGrowthBytes, got.HeapGrowthBytes)
		assert.Equal(t, thresholds.GoroutineGrowth, got.GoroutineGrowth)
	})
}

func TestLeakDetector_SetThresholds(t *testing.T) {
	t.Run("updates thresholds", func(t *testing.T) {
		ld := NewLeakDetector()
		originalThreshold := ld.GetThresholds().HeapGrowthBytes

		newThresholds := &LeakThresholds{
			HeapGrowthBytes:       originalThreshold * 2,
			GoroutineGrowth:       50,
			HeapObjectGrowth:      10000,
			SeverityHighBytes:     100 * 1024 * 1024,
			SeverityCriticalBytes: 1024 * 1024 * 1024,
		}
		ld.SetThresholds(newThresholds)

		got := ld.GetThresholds()
		assert.Equal(t, newThresholds.HeapGrowthBytes, got.HeapGrowthBytes)
		assert.Equal(t, 50, got.GoroutineGrowth)
	})
}

func TestLeakDetector_Reset(t *testing.T) {
	t.Run("resets to default thresholds", func(t *testing.T) {
		customThresholds := &LeakThresholds{
			HeapGrowthBytes:       1,
			GoroutineGrowth:       1,
			HeapObjectGrowth:      1,
			SeverityHighBytes:     1,
			SeverityCriticalBytes: 1,
		}
		ld := NewLeakDetectorWithThresholds(customThresholds)

		ld.Reset()

		defaults := DefaultLeakThresholds()
		got := ld.GetThresholds()
		assert.Equal(t, defaults.HeapGrowthBytes, got.HeapGrowthBytes)
		assert.Equal(t, defaults.GoroutineGrowth, got.GoroutineGrowth)
	})
}
