// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFPSCalculator(t *testing.T) {
	t.Run("creates calculator with default window size", func(t *testing.T) {
		fc := NewFPSCalculator()

		assert.NotNil(t, fc)
		assert.Equal(t, DefaultFPSWindowSize, fc.GetWindowSize())
		assert.Empty(t, fc.GetSamples())
	})

	t.Run("creates calculator with custom window size", func(t *testing.T) {
		fc := NewFPSCalculatorWithWindowSize(30)

		assert.NotNil(t, fc)
		assert.Equal(t, 30, fc.GetWindowSize())
	})

	t.Run("uses default for zero window size", func(t *testing.T) {
		fc := NewFPSCalculatorWithWindowSize(0)

		assert.Equal(t, DefaultFPSWindowSize, fc.GetWindowSize())
	})

	t.Run("uses default for negative window size", func(t *testing.T) {
		fc := NewFPSCalculatorWithWindowSize(-10)

		assert.Equal(t, DefaultFPSWindowSize, fc.GetWindowSize())
	})
}

func TestFPSCalculator_AddSample(t *testing.T) {
	tests := []struct {
		name        string
		windowSize  int
		samples     []float64
		wantCount   int
		wantSamples []float64
	}{
		{
			name:        "adds single sample",
			windowSize:  60,
			samples:     []float64{60.0},
			wantCount:   1,
			wantSamples: []float64{60.0},
		},
		{
			name:        "adds multiple samples",
			windowSize:  60,
			samples:     []float64{30.0, 60.0, 90.0},
			wantCount:   3,
			wantSamples: []float64{30.0, 60.0, 90.0},
		},
		{
			name:        "respects window size limit",
			windowSize:  3,
			samples:     []float64{10.0, 20.0, 30.0, 40.0, 50.0},
			wantCount:   3,
			wantSamples: []float64{30.0, 40.0, 50.0}, // Oldest samples removed
		},
		{
			name:        "handles exactly at window size",
			windowSize:  3,
			samples:     []float64{10.0, 20.0, 30.0},
			wantCount:   3,
			wantSamples: []float64{10.0, 20.0, 30.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := NewFPSCalculatorWithWindowSize(tt.windowSize)

			for _, s := range tt.samples {
				fc.AddSample(s)
			}

			assert.Equal(t, tt.wantCount, fc.SampleCount())
			assert.Equal(t, tt.wantSamples, fc.GetSamples())
		})
	}
}

func TestFPSCalculator_GetAverage(t *testing.T) {
	tests := []struct {
		name      string
		samples   []float64
		wantAvg   float64
		tolerance float64
	}{
		{
			name:    "returns zero with no samples",
			samples: nil,
			wantAvg: 0.0,
		},
		{
			name:      "calculates average of single sample",
			samples:   []float64{60.0},
			wantAvg:   60.0,
			tolerance: 0.001,
		},
		{
			name:      "calculates average of multiple samples",
			samples:   []float64{30.0, 60.0, 90.0},
			wantAvg:   60.0, // (30 + 60 + 90) / 3 = 60
			tolerance: 0.001,
		},
		{
			name:      "handles varying FPS values",
			samples:   []float64{55.0, 58.0, 62.0, 59.0, 61.0},
			wantAvg:   59.0, // (55 + 58 + 62 + 59 + 61) / 5 = 59
			tolerance: 0.001,
		},
		{
			name:      "handles very low FPS",
			samples:   []float64{5.0, 10.0, 15.0},
			wantAvg:   10.0,
			tolerance: 0.001,
		},
		{
			name:      "handles very high FPS",
			samples:   []float64{120.0, 144.0, 165.0},
			wantAvg:   143.0, // (120 + 144 + 165) / 3 = 143
			tolerance: 0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := NewFPSCalculator()

			for _, s := range tt.samples {
				fc.AddSample(s)
			}

			avg := fc.GetAverage()

			if len(tt.samples) == 0 {
				assert.Equal(t, 0.0, avg)
			} else {
				assert.InDelta(t, tt.wantAvg, avg, tt.tolerance)
			}
		})
	}
}

func TestFPSCalculator_GetMin(t *testing.T) {
	tests := []struct {
		name    string
		samples []float64
		wantMin float64
	}{
		{
			name:    "returns zero with no samples",
			samples: nil,
			wantMin: 0.0,
		},
		{
			name:    "returns single sample as min",
			samples: []float64{60.0},
			wantMin: 60.0,
		},
		{
			name:    "finds minimum in multiple samples",
			samples: []float64{60.0, 30.0, 90.0, 45.0},
			wantMin: 30.0,
		},
		{
			name:    "handles all same values",
			samples: []float64{60.0, 60.0, 60.0},
			wantMin: 60.0,
		},
		{
			name:    "handles min at start",
			samples: []float64{10.0, 20.0, 30.0},
			wantMin: 10.0,
		},
		{
			name:    "handles min at end",
			samples: []float64{30.0, 20.0, 10.0},
			wantMin: 10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := NewFPSCalculator()

			for _, s := range tt.samples {
				fc.AddSample(s)
			}

			min := fc.GetMin()

			assert.Equal(t, tt.wantMin, min)
		})
	}
}

func TestFPSCalculator_GetMax(t *testing.T) {
	tests := []struct {
		name    string
		samples []float64
		wantMax float64
	}{
		{
			name:    "returns zero with no samples",
			samples: nil,
			wantMax: 0.0,
		},
		{
			name:    "returns single sample as max",
			samples: []float64{60.0},
			wantMax: 60.0,
		},
		{
			name:    "finds maximum in multiple samples",
			samples: []float64{60.0, 30.0, 90.0, 45.0},
			wantMax: 90.0,
		},
		{
			name:    "handles all same values",
			samples: []float64{60.0, 60.0, 60.0},
			wantMax: 60.0,
		},
		{
			name:    "handles max at start",
			samples: []float64{30.0, 20.0, 10.0},
			wantMax: 30.0,
		},
		{
			name:    "handles max at end",
			samples: []float64{10.0, 20.0, 30.0},
			wantMax: 30.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := NewFPSCalculator()

			for _, s := range tt.samples {
				fc.AddSample(s)
			}

			max := fc.GetMax()

			assert.Equal(t, tt.wantMax, max)
		})
	}
}

func TestFPSCalculator_GetMinMax(t *testing.T) {
	t.Run("returns both min and max", func(t *testing.T) {
		fc := NewFPSCalculator()

		fc.AddSample(50.0)
		fc.AddSample(30.0)
		fc.AddSample(90.0)
		fc.AddSample(60.0)

		min, max := fc.GetMinMax()

		assert.Equal(t, 30.0, min)
		assert.Equal(t, 90.0, max)
	})

	t.Run("returns zero for empty calculator", func(t *testing.T) {
		fc := NewFPSCalculator()

		min, max := fc.GetMinMax()

		assert.Equal(t, 0.0, min)
		assert.Equal(t, 0.0, max)
	})
}

func TestFPSCalculator_WindowSizeRespected(t *testing.T) {
	t.Run("old samples removed when window exceeded", func(t *testing.T) {
		fc := NewFPSCalculatorWithWindowSize(3)

		// Add samples that will exceed window
		fc.AddSample(10.0) // Will be removed
		fc.AddSample(20.0) // Will be removed
		fc.AddSample(30.0)
		fc.AddSample(40.0)
		fc.AddSample(50.0)

		// Only last 3 should remain
		assert.Equal(t, 3, fc.SampleCount())
		assert.Equal(t, []float64{30.0, 40.0, 50.0}, fc.GetSamples())

		// Average should be based on remaining samples
		avg := fc.GetAverage()
		assert.InDelta(t, 40.0, avg, 0.001) // (30 + 40 + 50) / 3 = 40
	})

	t.Run("min/max updated after window slides", func(t *testing.T) {
		fc := NewFPSCalculatorWithWindowSize(3)

		// Add samples with min at start
		fc.AddSample(10.0) // Min, will be removed
		fc.AddSample(50.0)
		fc.AddSample(60.0)

		// Verify initial min
		assert.Equal(t, 10.0, fc.GetMin())

		// Add more samples to push out the min
		fc.AddSample(70.0)

		// Min should now be 50.0
		assert.Equal(t, 50.0, fc.GetMin())
	})
}

func TestFPSCalculator_Reset(t *testing.T) {
	t.Run("clears all samples", func(t *testing.T) {
		fc := NewFPSCalculator()

		fc.AddSample(60.0)
		fc.AddSample(60.0)
		fc.AddSample(60.0)

		require.Equal(t, 3, fc.SampleCount())

		fc.Reset()

		assert.Equal(t, 0, fc.SampleCount())
		assert.Equal(t, 0.0, fc.GetAverage())
		assert.Equal(t, 0.0, fc.GetMin())
		assert.Equal(t, 0.0, fc.GetMax())
	})

	t.Run("preserves window size after reset", func(t *testing.T) {
		fc := NewFPSCalculatorWithWindowSize(30)

		fc.AddSample(60.0)
		fc.Reset()

		assert.Equal(t, 30, fc.GetWindowSize())
	})
}

func TestFPSCalculator_SetWindowSize(t *testing.T) {
	t.Run("changes window size", func(t *testing.T) {
		fc := NewFPSCalculator()

		fc.SetWindowSize(30)

		assert.Equal(t, 30, fc.GetWindowSize())
	})

	t.Run("truncates samples when window shrinks", func(t *testing.T) {
		fc := NewFPSCalculatorWithWindowSize(5)

		// Add 5 samples
		for i := 1; i <= 5; i++ {
			fc.AddSample(float64(i * 10))
		}
		require.Equal(t, 5, fc.SampleCount())

		// Shrink window to 3
		fc.SetWindowSize(3)

		// Should keep only last 3 samples
		assert.Equal(t, 3, fc.SampleCount())
		assert.Equal(t, []float64{30.0, 40.0, 50.0}, fc.GetSamples())
	})

	t.Run("uses default for invalid window size", func(t *testing.T) {
		fc := NewFPSCalculator()

		fc.SetWindowSize(0)
		assert.Equal(t, DefaultFPSWindowSize, fc.GetWindowSize())

		fc.SetWindowSize(-5)
		assert.Equal(t, DefaultFPSWindowSize, fc.GetWindowSize())
	})
}

func TestFPSCalculator_GetStandardDeviation(t *testing.T) {
	tests := []struct {
		name       string
		samples    []float64
		wantStdDev float64
		tolerance  float64
	}{
		{
			name:       "returns zero with no samples",
			samples:    nil,
			wantStdDev: 0.0,
		},
		{
			name:       "returns zero with single sample",
			samples:    []float64{60.0},
			wantStdDev: 0.0,
		},
		{
			name:       "calculates std dev for identical samples",
			samples:    []float64{60.0, 60.0, 60.0},
			wantStdDev: 0.0,
		},
		{
			name:       "calculates std dev for varying samples",
			samples:    []float64{10.0, 20.0, 30.0},
			wantStdDev: 8.165, // sqrt(((10-20)^2 + (20-20)^2 + (30-20)^2) / 3) ≈ 8.165
			tolerance:  0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := NewFPSCalculator()

			for _, s := range tt.samples {
				fc.AddSample(s)
			}

			stdDev := fc.GetStandardDeviation()

			if len(tt.samples) <= 1 {
				assert.Equal(t, 0.0, stdDev)
			} else {
				assert.InDelta(t, tt.wantStdDev, stdDev, tt.tolerance)
			}
		})
	}
}

func TestFPSCalculator_GetPercentile(t *testing.T) {
	tests := []struct {
		name       string
		samples    []float64
		percentile float64
		want       float64
	}{
		{
			name:       "returns zero with no samples",
			samples:    nil,
			percentile: 50,
			want:       0.0,
		},
		{
			name:       "returns single sample for any percentile",
			samples:    []float64{60.0},
			percentile: 50,
			want:       60.0,
		},
		{
			name:       "calculates P50 (median)",
			samples:    []float64{10.0, 20.0, 30.0, 40.0, 50.0},
			percentile: 50,
			want:       30.0,
		},
		{
			name:       "calculates P0 (minimum)",
			samples:    []float64{10.0, 20.0, 30.0, 40.0, 50.0},
			percentile: 0,
			want:       10.0,
		},
		{
			name:       "calculates P100 (maximum)",
			samples:    []float64{10.0, 20.0, 30.0, 40.0, 50.0},
			percentile: 100,
			want:       50.0,
		},
		{
			name:       "calculates P95",
			samples:    []float64{10.0, 20.0, 30.0, 40.0, 50.0, 60.0, 70.0, 80.0, 90.0, 100.0},
			percentile: 95,
			want:       100.0, // At index 9 (95% of 10 = 9.5, rounded to 9)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := NewFPSCalculator()

			for _, s := range tt.samples {
				fc.AddSample(s)
			}

			p := fc.GetPercentile(tt.percentile)

			assert.Equal(t, tt.want, p)
		})
	}
}

func TestFPSCalculator_IsStable(t *testing.T) {
	tests := []struct {
		name      string
		samples   []float64
		threshold float64
		want      bool
	}{
		{
			name:      "empty samples are not stable",
			samples:   nil,
			threshold: 5.0,
			want:      false,
		},
		{
			name:      "single sample is stable",
			samples:   []float64{60.0},
			threshold: 5.0,
			want:      true,
		},
		{
			name:      "identical samples are stable",
			samples:   []float64{60.0, 60.0, 60.0},
			threshold: 5.0,
			want:      true,
		},
		{
			name:      "samples within threshold are stable",
			samples:   []float64{58.0, 60.0, 62.0},
			threshold: 5.0,
			want:      true, // std dev ≈ 1.63, within threshold
		},
		{
			name:      "samples outside threshold are unstable",
			samples:   []float64{30.0, 60.0, 90.0},
			threshold: 5.0,
			want:      false, // std dev ≈ 24.5, exceeds threshold
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := NewFPSCalculator()

			for _, s := range tt.samples {
				fc.AddSample(s)
			}

			stable := fc.IsStable(tt.threshold)

			assert.Equal(t, tt.want, stable)
		})
	}
}

func TestFPSCalculator_ThreadSafety(t *testing.T) {
	t.Run("concurrent operations are safe", func(t *testing.T) {
		fc := NewFPSCalculator()
		var wg sync.WaitGroup
		goroutines := 50

		// Concurrent AddSample
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(val float64) {
				defer wg.Done()
				fc.AddSample(val)
			}(float64(i))
		}

		// Concurrent reads
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = fc.GetAverage()
				_ = fc.GetMin()
				_ = fc.GetMax()
				_ = fc.SampleCount()
				_ = fc.GetSamples()
				_ = fc.GetStandardDeviation()
			}()
		}

		wg.Wait()
		// Should complete without race conditions
	})
}

func TestFPSCalculator_AccuracyValidation(t *testing.T) {
	t.Run("maintains accuracy with many samples", func(t *testing.T) {
		fc := NewFPSCalculatorWithWindowSize(1000)

		// Add 1000 samples of exactly 60 FPS
		for i := 0; i < 1000; i++ {
			fc.AddSample(60.0)
		}

		avg := fc.GetAverage()
		assert.InDelta(t, 60.0, avg, 0.0001, "Average should be exactly 60.0")

		min := fc.GetMin()
		assert.Equal(t, 60.0, min, "Min should be exactly 60.0")

		max := fc.GetMax()
		assert.Equal(t, 60.0, max, "Max should be exactly 60.0")
	})

	t.Run("handles floating point precision", func(t *testing.T) {
		fc := NewFPSCalculator()

		// Add samples that could cause floating point issues
		fc.AddSample(0.1)
		fc.AddSample(0.2)
		fc.AddSample(0.3)

		avg := fc.GetAverage()
		// 0.1 + 0.2 + 0.3 = 0.6, / 3 = 0.2
		assert.InDelta(t, 0.2, avg, 0.0001)
	})
}

func TestFPSCalculator_GetSamples(t *testing.T) {
	t.Run("returns copy of samples", func(t *testing.T) {
		fc := NewFPSCalculator()
		fc.AddSample(60.0)
		fc.AddSample(60.0)

		samples := fc.GetSamples()
		originalLen := len(samples)

		// Modify returned slice
		samples = append(samples, 999.0)

		// Original should be unchanged
		assert.Equal(t, originalLen, fc.SampleCount())
	})
}

func TestFPSCalculator_SampleCount(t *testing.T) {
	t.Run("returns correct count", func(t *testing.T) {
		fc := NewFPSCalculator()

		assert.Equal(t, 0, fc.SampleCount())

		fc.AddSample(60.0)
		assert.Equal(t, 1, fc.SampleCount())

		fc.AddSample(60.0)
		fc.AddSample(60.0)
		assert.Equal(t, 3, fc.SampleCount())
	})
}
