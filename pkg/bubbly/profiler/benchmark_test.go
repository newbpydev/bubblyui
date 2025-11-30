package profiler

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewBenchmarkProfiler tests the constructor
func TestNewBenchmarkProfiler(t *testing.T) {
	tests := []struct {
		name     string
		b        *testing.B
		wantNil  bool
		wantName string
	}{
		{
			name:     "with valid testing.B",
			b:        &testing.B{},
			wantNil:  false,
			wantName: "",
		},
		{
			name:     "with nil testing.B",
			b:        nil,
			wantNil:  false,
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := NewBenchmarkProfiler(tt.b)

			if tt.wantNil {
				assert.Nil(t, bp)
			} else {
				assert.NotNil(t, bp)
				assert.NotNil(t, bp.metrics)
				assert.NotNil(t, bp.measurements)
				assert.Equal(t, 0, len(bp.measurements))
			}
		})
	}
}

// TestBenchmarkProfiler_Measure tests the Measure method
func TestBenchmarkProfiler_Measure(t *testing.T) {
	tests := []struct {
		name          string
		fn            func()
		expectedCount int
		minDuration   time.Duration
		maxDuration   time.Duration
	}{
		{
			name: "fast function",
			fn: func() {
				_ = 1 + 1
			},
			expectedCount: 1,
			minDuration:   0,
			maxDuration:   1 * time.Millisecond,
		},
		{
			name: "slow function",
			fn: func() {
				time.Sleep(10 * time.Millisecond)
			},
			expectedCount: 1,
			minDuration:   10 * time.Millisecond,
			maxDuration:   50 * time.Millisecond,
		},
		{
			name:          "nil function",
			fn:            nil,
			expectedCount: 0,
			minDuration:   0,
			maxDuration:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := NewBenchmarkProfiler(nil)

			bp.Measure(tt.fn)

			assert.Equal(t, tt.expectedCount, bp.MeasureCount())

			if tt.expectedCount > 0 {
				measurements := bp.GetMeasurements()
				assert.Len(t, measurements, tt.expectedCount)

				if tt.minDuration > 0 {
					assert.GreaterOrEqual(t, measurements[0], tt.minDuration)
				}
				if tt.maxDuration > 0 {
					assert.LessOrEqual(t, measurements[0], tt.maxDuration)
				}
			}
		})
	}
}

// TestBenchmarkProfiler_MeasureMultiple tests multiple measurements
func TestBenchmarkProfiler_MeasureMultiple(t *testing.T) {
	bp := NewBenchmarkProfiler(nil)

	// Record multiple measurements
	for i := 0; i < 100; i++ {
		bp.Measure(func() {
			time.Sleep(1 * time.Millisecond)
		})
	}

	assert.Equal(t, 100, bp.MeasureCount())

	stats := bp.GetStats()
	require.NotNil(t, stats)
	assert.Equal(t, 100, stats.Iterations)
	assert.GreaterOrEqual(t, stats.Mean, 1*time.Millisecond)
}

// TestBenchmarkProfiler_StartMeasurement tests the StartMeasurement method
func TestBenchmarkProfiler_StartMeasurement(t *testing.T) {
	bp := NewBenchmarkProfiler(nil)

	stop := bp.StartMeasurement()
	time.Sleep(5 * time.Millisecond)
	stop()

	assert.Equal(t, 1, bp.MeasureCount())

	measurements := bp.GetMeasurements()
	assert.Len(t, measurements, 1)
	assert.GreaterOrEqual(t, measurements[0], 5*time.Millisecond)
}

// TestBenchmarkProfiler_GetStats tests statistics calculation
func TestBenchmarkProfiler_GetStats(t *testing.T) {
	tests := []struct {
		name         string
		measurements []time.Duration
		wantNil      bool
		checkStats   func(t *testing.T, stats *BenchmarkStats)
	}{
		{
			name:         "no measurements",
			measurements: nil,
			wantNil:      true,
		},
		{
			name:         "single measurement",
			measurements: []time.Duration{10 * time.Millisecond},
			wantNil:      false,
			checkStats: func(t *testing.T, stats *BenchmarkStats) {
				assert.Equal(t, 1, stats.Iterations)
				assert.Equal(t, 10*time.Millisecond, stats.Mean)
				assert.Equal(t, 10*time.Millisecond, stats.Min)
				assert.Equal(t, 10*time.Millisecond, stats.Max)
			},
		},
		{
			name: "multiple measurements",
			measurements: []time.Duration{
				1 * time.Millisecond,
				2 * time.Millisecond,
				3 * time.Millisecond,
				4 * time.Millisecond,
				5 * time.Millisecond,
			},
			wantNil: false,
			checkStats: func(t *testing.T, stats *BenchmarkStats) {
				assert.Equal(t, 5, stats.Iterations)
				assert.Equal(t, 3*time.Millisecond, stats.Mean)
				assert.Equal(t, 1*time.Millisecond, stats.Min)
				assert.Equal(t, 5*time.Millisecond, stats.Max)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := NewBenchmarkProfiler(nil)

			// Add measurements directly
			if tt.measurements != nil {
				bp.mu.Lock()
				bp.measurements = tt.measurements
				bp.mu.Unlock()
			}

			stats := bp.GetStats()

			if tt.wantNil {
				assert.Nil(t, stats)
			} else {
				require.NotNil(t, stats)
				if tt.checkStats != nil {
					tt.checkStats(t, stats)
				}
			}
		})
	}
}

// TestBenchmarkProfiler_GetMeasurements tests measurement retrieval
func TestBenchmarkProfiler_GetMeasurements(t *testing.T) {
	bp := NewBenchmarkProfiler(nil)

	// Empty measurements
	measurements := bp.GetMeasurements()
	assert.Empty(t, measurements)

	// Add some measurements
	bp.Measure(func() { time.Sleep(1 * time.Millisecond) })
	bp.Measure(func() { time.Sleep(2 * time.Millisecond) })

	measurements = bp.GetMeasurements()
	assert.Len(t, measurements, 2)

	// Verify it's a copy (modifying returned slice doesn't affect original)
	measurements[0] = 0
	originalMeasurements := bp.GetMeasurements()
	assert.NotEqual(t, time.Duration(0), originalMeasurements[0])
}

// TestBenchmarkProfiler_Reset tests the Reset method
func TestBenchmarkProfiler_Reset(t *testing.T) {
	bp := NewBenchmarkProfiler(nil)

	// Add measurements
	bp.Measure(func() { time.Sleep(1 * time.Millisecond) })
	bp.Measure(func() { time.Sleep(1 * time.Millisecond) })
	bp.SetAllocStats(1024, 10)

	assert.Equal(t, 2, bp.MeasureCount())

	// Reset
	bp.Reset()

	assert.Equal(t, 0, bp.MeasureCount())
	stats := bp.GetStats()
	assert.Nil(t, stats)
}

// TestBenchmarkProfiler_Baseline tests baseline management
func TestBenchmarkProfiler_Baseline(t *testing.T) {
	bp := NewBenchmarkProfiler(nil)

	// Initially no baseline
	assert.Nil(t, bp.GetBaseline())

	// Set baseline
	baseline := &Baseline{
		Name:    "test",
		NsPerOp: 1000,
	}
	bp.SetBaseline(baseline)

	// Get baseline
	got := bp.GetBaseline()
	assert.Equal(t, baseline, got)
}

// TestBenchmarkProfiler_NewBaseline tests creating baseline from stats
func TestBenchmarkProfiler_NewBaseline(t *testing.T) {
	tests := []struct {
		name         string
		measurements []time.Duration
		wantNil      bool
	}{
		{
			name:         "no measurements",
			measurements: nil,
			wantNil:      true,
		},
		{
			name:         "with measurements",
			measurements: []time.Duration{1 * time.Millisecond, 2 * time.Millisecond},
			wantNil:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := NewBenchmarkProfiler(nil)

			if tt.measurements != nil {
				bp.mu.Lock()
				bp.measurements = tt.measurements
				bp.mu.Unlock()
			}

			baseline := bp.NewBaseline("TestBenchmark")

			if tt.wantNil {
				assert.Nil(t, baseline)
			} else {
				require.NotNil(t, baseline)
				assert.Equal(t, "TestBenchmark", baseline.Name)
				assert.Equal(t, runtime.Version(), baseline.GoVersion)
				assert.Equal(t, runtime.GOOS, baseline.GOOS)
				assert.Equal(t, runtime.GOARCH, baseline.GOARCH)
				assert.NotZero(t, baseline.Timestamp)
			}
		})
	}
}

// TestBaseline_SaveAndLoad tests saving and loading baselines
func TestBaseline_SaveAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "baseline.json")

	// Create baseline
	original := &Baseline{
		Name:        "TestBenchmark",
		NsPerOp:     1500,
		AllocBytes:  256,
		AllocsPerOp: 5,
		Iterations:  1000,
		Timestamp:   time.Now().Truncate(time.Second),
		GoVersion:   runtime.Version(),
		GOOS:        runtime.GOOS,
		GOARCH:      runtime.GOARCH,
		Metadata: map[string]string{
			"key": "value",
		},
	}

	// Save
	err := original.SaveToFile(filename)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filename)
	require.NoError(t, err)

	// Load
	loaded, err := LoadBaseline(filename)
	require.NoError(t, err)
	require.NotNil(t, loaded)

	// Compare
	assert.Equal(t, original.Name, loaded.Name)
	assert.Equal(t, original.NsPerOp, loaded.NsPerOp)
	assert.Equal(t, original.AllocBytes, loaded.AllocBytes)
	assert.Equal(t, original.AllocsPerOp, loaded.AllocsPerOp)
	assert.Equal(t, original.Iterations, loaded.Iterations)
	assert.Equal(t, original.GoVersion, loaded.GoVersion)
	assert.Equal(t, original.GOOS, loaded.GOOS)
	assert.Equal(t, original.GOARCH, loaded.GOARCH)
	assert.Equal(t, original.Metadata, loaded.Metadata)
}

// TestLoadBaseline_Errors tests error cases for LoadBaseline
func TestLoadBaseline_Errors(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		setup    func(t *testing.T, filename string)
		wantErr  bool
	}{
		{
			name:     "file not found",
			filename: "/nonexistent/path/baseline.json",
			wantErr:  true,
		},
		{
			name:     "invalid JSON",
			filename: "invalid.json",
			setup: func(t *testing.T, filename string) {
				err := os.WriteFile(filename, []byte("not valid json"), 0644)
				require.NoError(t, err)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filename := tt.filename
			if tt.setup != nil {
				filename = filepath.Join(tmpDir, tt.filename)
				tt.setup(t, filename)
			}

			_, err := LoadBaseline(filename)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBaseline_SaveToFile_Errors tests error cases for SaveToFile
func TestBaseline_SaveToFile_Errors(t *testing.T) {
	baseline := &Baseline{
		Name:    "test",
		NsPerOp: 1000,
	}

	// Test invalid path
	err := baseline.SaveToFile("/nonexistent/directory/baseline.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write baseline file")
}

// TestBenchmarkProfiler_SaveBaseline tests SaveBaseline method
func TestBenchmarkProfiler_SaveBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "baseline.json")

	bp := NewBenchmarkProfiler(nil)

	// No measurements - should error
	err := bp.SaveBaseline(filename)
	assert.Error(t, err)
	assert.Equal(t, ErrNoMeasurements, err)

	// Add measurements
	bp.mu.Lock()
	bp.measurements = []time.Duration{1 * time.Millisecond, 2 * time.Millisecond}
	bp.mu.Unlock()

	// Should succeed
	err = bp.SaveBaseline(filename)
	assert.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filename)
	assert.NoError(t, err)
}

// TestBenchmarkProfiler_HasRegression tests regression detection
func TestBenchmarkProfiler_HasRegression(t *testing.T) {
	tests := []struct {
		name         string
		measurements []time.Duration
		baseline     *Baseline
		threshold    float64
		want         bool
	}{
		{
			name:         "nil baseline",
			measurements: []time.Duration{1 * time.Millisecond},
			baseline:     nil,
			threshold:    0.10,
			want:         false,
		},
		{
			name:         "no measurements",
			measurements: nil,
			baseline:     &Baseline{NsPerOp: 1000000},
			threshold:    0.10,
			want:         false,
		},
		{
			name:         "no regression - same performance",
			measurements: []time.Duration{1 * time.Millisecond},
			baseline:     &Baseline{NsPerOp: 1000000}, // 1ms
			threshold:    0.10,
			want:         false,
		},
		{
			name:         "no regression - within threshold",
			measurements: []time.Duration{1050 * time.Microsecond}, // 5% slower
			baseline:     &Baseline{NsPerOp: 1000000},              // 1ms
			threshold:    0.10,                                     // 10% allowed
			want:         false,
		},
		{
			name:         "regression detected - exceeds threshold",
			measurements: []time.Duration{1200 * time.Microsecond}, // 20% slower
			baseline:     &Baseline{NsPerOp: 1000000},              // 1ms
			threshold:    0.10,                                     // 10% allowed
			want:         true,
		},
		{
			name:         "improvement - negative regression",
			measurements: []time.Duration{800 * time.Microsecond}, // 20% faster
			baseline:     &Baseline{NsPerOp: 1000000},             // 1ms
			threshold:    0.10,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := NewBenchmarkProfiler(nil)

			if tt.measurements != nil {
				bp.mu.Lock()
				bp.measurements = tt.measurements
				bp.mu.Unlock()
			}

			got := bp.HasRegression(tt.baseline, tt.threshold)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestBenchmarkProfiler_AssertNoRegression tests AssertNoRegression method
func TestBenchmarkProfiler_AssertNoRegression(t *testing.T) {
	tests := []struct {
		name         string
		measurements []time.Duration
		baseline     *Baseline
		threshold    float64
		wantErr      error
	}{
		{
			name:         "nil baseline",
			measurements: []time.Duration{1 * time.Millisecond},
			baseline:     nil,
			threshold:    0.10,
			wantErr:      ErrNilBaseline,
		},
		{
			name:         "invalid threshold - negative",
			measurements: []time.Duration{1 * time.Millisecond},
			baseline:     &Baseline{NsPerOp: 1000000},
			threshold:    -0.10,
			wantErr:      ErrInvalidThreshold,
		},
		{
			name:         "invalid threshold - too high",
			measurements: []time.Duration{1 * time.Millisecond},
			baseline:     &Baseline{NsPerOp: 1000000},
			threshold:    1.5,
			wantErr:      ErrInvalidThreshold,
		},
		{
			name:         "no measurements",
			measurements: nil,
			baseline:     &Baseline{NsPerOp: 1000000},
			threshold:    0.10,
			wantErr:      ErrNoMeasurements,
		},
		{
			name:         "no regression",
			measurements: []time.Duration{1 * time.Millisecond},
			baseline:     &Baseline{NsPerOp: 1000000},
			threshold:    0.10,
			wantErr:      nil,
		},
		{
			name:         "regression detected",
			measurements: []time.Duration{1500 * time.Microsecond}, // 50% slower
			baseline:     &Baseline{NsPerOp: 1000000},
			threshold:    0.10,
			wantErr:      ErrRegressionDetected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := NewBenchmarkProfiler(nil)

			if tt.measurements != nil {
				bp.mu.Lock()
				bp.measurements = tt.measurements
				bp.mu.Unlock()
			}

			err := bp.AssertNoRegression(tt.baseline, tt.threshold)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBenchmarkProfiler_GetRegressionInfo tests GetRegressionInfo method
func TestBenchmarkProfiler_GetRegressionInfo(t *testing.T) {
	tests := []struct {
		name         string
		measurements []time.Duration
		allocBytes   int64
		allocsPerOp  int64
		baseline     *Baseline
		wantNil      bool
		checkInfo    func(t *testing.T, info *RegressionInfo)
	}{
		{
			name:         "no measurements",
			measurements: nil,
			baseline:     &Baseline{NsPerOp: 1000000},
			wantNil:      true,
		},
		{
			name:         "nil baseline",
			measurements: []time.Duration{1 * time.Millisecond},
			baseline:     nil,
			wantNil:      false,
			checkInfo: func(t *testing.T, info *RegressionInfo) {
				assert.False(t, info.HasRegression)
				assert.Contains(t, info.Details, "no baseline")
			},
		},
		{
			name:         "time regression",
			measurements: []time.Duration{2 * time.Millisecond}, // 100% slower
			baseline:     &Baseline{NsPerOp: 1000000},           // 1ms
			wantNil:      false,
			checkInfo: func(t *testing.T, info *RegressionInfo) {
				assert.True(t, info.HasRegression)
				assert.InDelta(t, 1.0, info.TimeRegression, 0.01) // 100% regression
			},
		},
		{
			name:         "memory regression",
			measurements: []time.Duration{1 * time.Millisecond},
			allocBytes:   2048,
			baseline:     &Baseline{NsPerOp: 1000000, AllocBytes: 1024},
			wantNil:      false,
			checkInfo: func(t *testing.T, info *RegressionInfo) {
				assert.True(t, info.HasRegression)
				assert.InDelta(t, 1.0, info.MemoryRegression, 0.01) // 100% regression
			},
		},
		{
			name:         "alloc regression",
			measurements: []time.Duration{1 * time.Millisecond},
			allocsPerOp:  20,
			baseline:     &Baseline{NsPerOp: 1000000, AllocsPerOp: 10},
			wantNil:      false,
			checkInfo: func(t *testing.T, info *RegressionInfo) {
				assert.True(t, info.HasRegression)
				assert.InDelta(t, 1.0, info.AllocRegression, 0.01) // 100% regression
			},
		},
		{
			name:         "improvement - no regression",
			measurements: []time.Duration{500 * time.Microsecond}, // 50% faster
			baseline:     &Baseline{NsPerOp: 1000000},
			wantNil:      false,
			checkInfo: func(t *testing.T, info *RegressionInfo) {
				assert.False(t, info.HasRegression)
				assert.InDelta(t, -0.5, info.TimeRegression, 0.01) // 50% improvement
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := NewBenchmarkProfiler(nil)

			if tt.measurements != nil {
				bp.mu.Lock()
				bp.measurements = tt.measurements
				bp.allocBytes = tt.allocBytes
				bp.allocsPerOp = tt.allocsPerOp
				bp.mu.Unlock()
			}

			info := bp.GetRegressionInfo(tt.baseline)

			if tt.wantNil {
				assert.Nil(t, info)
			} else {
				require.NotNil(t, info)
				if tt.checkInfo != nil {
					tt.checkInfo(t, info)
				}
			}
		})
	}
}

// TestBenchmarkProfiler_ReportMetrics tests ReportMetrics method
func TestBenchmarkProfiler_ReportMetrics(t *testing.T) {
	// With nil testing.B - should not panic
	bp := NewBenchmarkProfiler(nil)
	bp.mu.Lock()
	bp.measurements = []time.Duration{1 * time.Millisecond, 2 * time.Millisecond}
	bp.mu.Unlock()

	// Should not panic
	assert.NotPanics(t, func() {
		bp.ReportMetrics()
	})

	// With no measurements - should not panic
	bp2 := NewBenchmarkProfiler(nil)
	assert.NotPanics(t, func() {
		bp2.ReportMetrics()
	})
}

// TestBenchmarkProfiler_String tests String method
func TestBenchmarkProfiler_String(t *testing.T) {
	tests := []struct {
		name         string
		measurements []time.Duration
		wantContains string
	}{
		{
			name:         "no measurements",
			measurements: nil,
			wantContains: "no measurements",
		},
		{
			name:         "with measurements",
			measurements: []time.Duration{1 * time.Millisecond, 2 * time.Millisecond},
			wantContains: "iterations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := NewBenchmarkProfiler(nil)

			if tt.measurements != nil {
				bp.mu.Lock()
				bp.measurements = tt.measurements
				bp.mu.Unlock()
			}

			got := bp.String()
			assert.Contains(t, got, tt.wantContains)
		})
	}
}

// TestBenchmarkProfiler_SetAllocStats tests SetAllocStats method
func TestBenchmarkProfiler_SetAllocStats(t *testing.T) {
	bp := NewBenchmarkProfiler(nil)

	bp.SetAllocStats(1024, 10)

	bp.mu.RLock()
	assert.Equal(t, int64(1024), bp.allocBytes)
	assert.Equal(t, int64(10), bp.allocsPerOp)
	bp.mu.RUnlock()
}

// TestBenchmarkProfiler_GetSetName tests GetName and SetName methods
func TestBenchmarkProfiler_GetSetName(t *testing.T) {
	bp := NewBenchmarkProfiler(nil)

	// Initially empty
	assert.Equal(t, "", bp.GetName())

	// Set name
	bp.SetName("TestBenchmark")
	assert.Equal(t, "TestBenchmark", bp.GetName())
}

// TestBaseline_Clone tests Clone method
func TestBaseline_Clone(t *testing.T) {
	tests := []struct {
		name     string
		baseline *Baseline
		wantNil  bool
	}{
		{
			name:     "nil baseline",
			baseline: nil,
			wantNil:  true,
		},
		{
			name: "with metadata",
			baseline: &Baseline{
				Name:        "test",
				NsPerOp:     1000,
				AllocBytes:  256,
				AllocsPerOp: 5,
				Metadata:    map[string]string{"key": "value"},
			},
			wantNil: false,
		},
		{
			name: "without metadata",
			baseline: &Baseline{
				Name:    "test",
				NsPerOp: 1000,
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clone := tt.baseline.Clone()

			if tt.wantNil {
				assert.Nil(t, clone)
			} else {
				require.NotNil(t, clone)
				assert.Equal(t, tt.baseline.Name, clone.Name)
				assert.Equal(t, tt.baseline.NsPerOp, clone.NsPerOp)

				// Verify it's a deep copy
				if tt.baseline.Metadata != nil {
					assert.Equal(t, tt.baseline.Metadata, clone.Metadata)
					// Modify clone metadata
					clone.Metadata["new"] = "value"
					assert.NotEqual(t, tt.baseline.Metadata, clone.Metadata)
				}
			}
		})
	}
}

// TestBaseline_String tests String method
func TestBaseline_String(t *testing.T) {
	tests := []struct {
		name         string
		baseline     *Baseline
		wantContains string
	}{
		{
			name:         "nil baseline",
			baseline:     nil,
			wantContains: "nil",
		},
		{
			name: "valid baseline",
			baseline: &Baseline{
				Name:      "test",
				NsPerOp:   1000,
				GoVersion: "go1.22",
				GOOS:      "linux",
				GOARCH:    "amd64",
			},
			wantContains: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.baseline.String()
			assert.Contains(t, got, tt.wantContains)
		})
	}
}

// TestRegressionInfo_String tests String method
func TestRegressionInfo_String(t *testing.T) {
	tests := []struct {
		name         string
		info         *RegressionInfo
		wantContains string
	}{
		{
			name:         "nil info",
			info:         nil,
			wantContains: "nil",
		},
		{
			name: "no regression",
			info: &RegressionInfo{
				HasRegression: false,
			},
			wantContains: "no regression",
		},
		{
			name: "with regression",
			info: &RegressionInfo{
				HasRegression:  true,
				TimeRegression: 0.5,
			},
			wantContains: "time=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.info.String()
			assert.Contains(t, got, tt.wantContains)
		})
	}
}

// TestBenchmarkStats_String tests String method
func TestBenchmarkStats_String(t *testing.T) {
	tests := []struct {
		name         string
		stats        *BenchmarkStats
		wantContains string
	}{
		{
			name:         "nil stats",
			stats:        nil,
			wantContains: "nil",
		},
		{
			name: "valid stats",
			stats: &BenchmarkStats{
				Name:       "test",
				Iterations: 100,
				NsPerOp:    1000,
			},
			wantContains: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stats.String()
			assert.Contains(t, got, tt.wantContains)
		})
	}
}

// TestBenchmarkProfiler_ConcurrentAccess tests thread safety
func TestBenchmarkProfiler_ConcurrentAccess(t *testing.T) {
	bp := NewBenchmarkProfiler(nil)

	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrent Measure calls
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bp.Measure(func() {
				time.Sleep(1 * time.Microsecond)
			})
		}()
	}

	// Concurrent GetStats calls
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = bp.GetStats()
		}()
	}

	// Concurrent GetMeasurements calls
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = bp.GetMeasurements()
		}()
	}

	wg.Wait()

	// Should have recorded all measurements
	assert.Equal(t, numGoroutines, bp.MeasureCount())
}

// TestBenchmarkProfiler_PercentileCalculation tests percentile accuracy
func TestBenchmarkProfiler_PercentileCalculation(t *testing.T) {
	bp := NewBenchmarkProfiler(nil)

	// Add 100 measurements from 1ms to 100ms
	measurements := make([]time.Duration, 100)
	for i := 0; i < 100; i++ {
		measurements[i] = time.Duration(i+1) * time.Millisecond
	}

	bp.mu.Lock()
	bp.measurements = measurements
	bp.mu.Unlock()

	stats := bp.GetStats()
	require.NotNil(t, stats)

	// P50 should be around 50ms
	assert.InDelta(t, 50*time.Millisecond, stats.P50, float64(5*time.Millisecond))

	// P95 should be around 95ms
	assert.InDelta(t, 95*time.Millisecond, stats.P95, float64(5*time.Millisecond))

	// P99 should be around 99ms
	assert.InDelta(t, 99*time.Millisecond, stats.P99, float64(5*time.Millisecond))
}

// TestBenchmarkProfiler_IntegrationWorkflow tests full workflow
func TestBenchmarkProfiler_IntegrationWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	baselineFile := filepath.Join(tmpDir, "baseline.json")

	// Step 1: Create baseline
	bp1 := NewBenchmarkProfiler(nil)
	bp1.SetName("BenchmarkIntegration")

	for i := 0; i < 100; i++ {
		bp1.Measure(func() {
			time.Sleep(1 * time.Millisecond)
		})
	}

	err := bp1.SaveBaseline(baselineFile)
	require.NoError(t, err)

	// Step 2: Load baseline
	baseline, err := LoadBaseline(baselineFile)
	require.NoError(t, err)
	require.NotNil(t, baseline)

	// Step 3: Run new benchmark with similar performance
	bp2 := NewBenchmarkProfiler(nil)
	bp2.SetName("BenchmarkIntegration")

	for i := 0; i < 100; i++ {
		bp2.Measure(func() {
			time.Sleep(1 * time.Millisecond)
		})
	}

	// Step 4: Check for regression (should pass with 20% threshold)
	err = bp2.AssertNoRegression(baseline, 0.20)
	assert.NoError(t, err)

	// Step 5: Get regression info
	info := bp2.GetRegressionInfo(baseline)
	require.NotNil(t, info)
	t.Logf("Regression info: %s", info.String())
}

// BenchmarkBenchmarkProfiler_Measure benchmarks the Measure method
func BenchmarkBenchmarkProfiler_Measure(b *testing.B) {
	bp := NewBenchmarkProfiler(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bp.Measure(func() {
			// Empty function to measure overhead
		})
	}
}

// BenchmarkBenchmarkProfiler_GetStats benchmarks the GetStats method
func BenchmarkBenchmarkProfiler_GetStats(b *testing.B) {
	bp := NewBenchmarkProfiler(b)

	// Pre-populate with measurements
	for i := 0; i < 1000; i++ {
		bp.Measure(func() {})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bp.GetStats()
	}
}
