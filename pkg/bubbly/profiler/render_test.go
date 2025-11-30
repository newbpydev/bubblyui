// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRenderProfiler(t *testing.T) {
	t.Run("creates new render profiler with defaults", func(t *testing.T) {
		rp := NewRenderProfiler()

		assert.NotNil(t, rp)
		assert.NotNil(t, rp.frames)
		assert.Empty(t, rp.frames)
		assert.NotNil(t, rp.fpsSamples)
		assert.Empty(t, rp.fpsSamples)
		assert.True(t, rp.lastFrame.IsZero())
	})

	t.Run("creates render profiler with custom config", func(t *testing.T) {
		config := &RenderConfig{
			TargetFPS:             30,
			MaxFrames:             100,
			MaxFPSSamples:         30,
			DroppedFrameThreshold: 33 * time.Millisecond, // 30fps threshold
		}
		rp := NewRenderProfilerWithConfig(config)

		assert.NotNil(t, rp)
		assert.Equal(t, 30, rp.config.TargetFPS)
		assert.Equal(t, 100, rp.config.MaxFrames)
	})
}

func TestDefaultRenderConfig(t *testing.T) {
	t.Run("returns sensible defaults", func(t *testing.T) {
		config := DefaultRenderConfig()

		assert.NotNil(t, config)
		assert.Equal(t, 60, config.TargetFPS)
		assert.Greater(t, config.MaxFrames, 0)
		assert.Greater(t, config.MaxFPSSamples, 0)
		// Default threshold for 60fps is ~16.67ms
		assert.True(t, config.DroppedFrameThreshold > 16*time.Millisecond)
		assert.True(t, config.DroppedFrameThreshold < 20*time.Millisecond)
	})
}

func TestRenderProfiler_RecordFrame(t *testing.T) {
	tests := []struct {
		name           string
		durations      []time.Duration
		wantFrameCount int
		wantDropped    int
	}{
		{
			name:           "records single fast frame",
			durations:      []time.Duration{5 * time.Millisecond},
			wantFrameCount: 1,
			wantDropped:    0,
		},
		{
			name:           "records multiple fast frames",
			durations:      []time.Duration{5 * time.Millisecond, 8 * time.Millisecond, 10 * time.Millisecond},
			wantFrameCount: 3,
			wantDropped:    0,
		},
		{
			name:           "detects dropped frame",
			durations:      []time.Duration{20 * time.Millisecond}, // > 16.67ms
			wantFrameCount: 1,
			wantDropped:    1,
		},
		{
			name: "mixed fast and dropped frames",
			durations: []time.Duration{
				5 * time.Millisecond,  // fast
				25 * time.Millisecond, // dropped
				10 * time.Millisecond, // fast
				30 * time.Millisecond, // dropped
			},
			wantFrameCount: 4,
			wantDropped:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := NewRenderProfiler()

			for _, d := range tt.durations {
				rp.RecordFrame(d)
			}

			frames := rp.GetFrames()
			assert.Len(t, frames, tt.wantFrameCount)

			droppedCount := 0
			for _, f := range frames {
				if f.Dropped {
					droppedCount++
				}
			}
			assert.Equal(t, tt.wantDropped, droppedCount)
		})
	}
}

func TestRenderProfiler_RecordFrame_FPSCalculation(t *testing.T) {
	t.Run("calculates FPS from frame intervals", func(t *testing.T) {
		rp := NewRenderProfiler()

		// Record first frame
		rp.RecordFrame(10 * time.Millisecond)

		// Wait a bit and record second frame
		time.Sleep(20 * time.Millisecond)
		rp.RecordFrame(10 * time.Millisecond)

		// Should have at least one FPS sample
		assert.GreaterOrEqual(t, len(rp.fpsSamples), 1)
	})

	t.Run("limits FPS samples to max", func(t *testing.T) {
		config := &RenderConfig{
			TargetFPS:             60,
			MaxFrames:             1000,
			MaxFPSSamples:         5, // Very small limit
			DroppedFrameThreshold: 17 * time.Millisecond,
		}
		rp := NewRenderProfilerWithConfig(config)

		// Record many frames
		for i := 0; i < 20; i++ {
			rp.RecordFrame(10 * time.Millisecond)
			time.Sleep(1 * time.Millisecond)
		}

		// Should be limited to MaxFPSSamples
		assert.LessOrEqual(t, len(rp.fpsSamples), config.MaxFPSSamples)
	})
}

func TestRenderProfiler_GetFPS(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*RenderProfiler)
		wantFPS   float64
		wantZero  bool
		tolerance float64
	}{
		{
			name:     "returns zero with no samples",
			setup:    func(rp *RenderProfiler) {},
			wantZero: true,
		},
		{
			name: "returns zero with single frame",
			setup: func(rp *RenderProfiler) {
				rp.RecordFrame(10 * time.Millisecond)
			},
			wantZero: true,
		},
		{
			name: "calculates average FPS",
			setup: func(rp *RenderProfiler) {
				// Manually set FPS samples for predictable test
				rp.fpsSamples = []float64{60.0, 60.0, 60.0}
			},
			wantFPS:   60.0,
			tolerance: 0.01,
		},
		{
			name: "handles varying FPS",
			setup: func(rp *RenderProfiler) {
				rp.fpsSamples = []float64{30.0, 60.0, 90.0}
			},
			wantFPS:   60.0, // Average
			tolerance: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := NewRenderProfiler()
			tt.setup(rp)

			fps := rp.GetFPS()

			if tt.wantZero {
				assert.Equal(t, 0.0, fps)
			} else {
				assert.InDelta(t, tt.wantFPS, fps, tt.tolerance)
			}
		})
	}
}

func TestRenderProfiler_GetDroppedFramePercent(t *testing.T) {
	tests := []struct {
		name        string
		frames      []FrameInfo
		wantPercent float64
		tolerance   float64
	}{
		{
			name:        "returns zero with no frames",
			frames:      nil,
			wantPercent: 0.0,
		},
		{
			name: "returns zero with no dropped frames",
			frames: []FrameInfo{
				{Dropped: false},
				{Dropped: false},
				{Dropped: false},
			},
			wantPercent: 0.0,
		},
		{
			name: "calculates 100% dropped",
			frames: []FrameInfo{
				{Dropped: true},
				{Dropped: true},
			},
			wantPercent: 100.0,
		},
		{
			name: "calculates 50% dropped",
			frames: []FrameInfo{
				{Dropped: true},
				{Dropped: false},
				{Dropped: true},
				{Dropped: false},
			},
			wantPercent: 50.0,
		},
		{
			name: "calculates 25% dropped",
			frames: []FrameInfo{
				{Dropped: true},
				{Dropped: false},
				{Dropped: false},
				{Dropped: false},
			},
			wantPercent: 25.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := NewRenderProfiler()
			rp.frames = tt.frames

			percent := rp.GetDroppedFramePercent()

			assert.InDelta(t, tt.wantPercent, percent, 0.01)
		})
	}
}

func TestRenderProfiler_GetFrames(t *testing.T) {
	t.Run("returns copy of frames", func(t *testing.T) {
		rp := NewRenderProfiler()
		rp.RecordFrame(10 * time.Millisecond)
		rp.RecordFrame(15 * time.Millisecond)

		frames := rp.GetFrames()
		assert.Len(t, frames, 2)

		// Verify it's a copy
		originalLen := len(rp.GetFrames())
		_ = append(frames, FrameInfo{})
		assert.Len(t, rp.GetFrames(), originalLen)
	})
}

func TestRenderProfiler_GetFrameCount(t *testing.T) {
	t.Run("returns correct count", func(t *testing.T) {
		rp := NewRenderProfiler()

		assert.Equal(t, 0, rp.GetFrameCount())

		rp.RecordFrame(10 * time.Millisecond)
		assert.Equal(t, 1, rp.GetFrameCount())

		rp.RecordFrame(10 * time.Millisecond)
		rp.RecordFrame(10 * time.Millisecond)
		assert.Equal(t, 3, rp.GetFrameCount())
	})
}

func TestRenderProfiler_GetAverageFrameDuration(t *testing.T) {
	tests := []struct {
		name         string
		durations    []time.Duration
		wantDuration time.Duration
		wantZero     bool
	}{
		{
			name:     "returns zero with no frames",
			wantZero: true,
		},
		{
			name:         "calculates average duration",
			durations:    []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond},
			wantDuration: 20 * time.Millisecond,
		},
		{
			name:         "single frame returns its duration",
			durations:    []time.Duration{15 * time.Millisecond},
			wantDuration: 15 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := NewRenderProfiler()

			for _, d := range tt.durations {
				rp.RecordFrame(d)
			}

			avg := rp.GetAverageFrameDuration()

			if tt.wantZero {
				assert.Equal(t, time.Duration(0), avg)
			} else {
				assert.Equal(t, tt.wantDuration, avg)
			}
		})
	}
}

func TestRenderProfiler_GetMinMaxFrameDuration(t *testing.T) {
	t.Run("returns zero with no frames", func(t *testing.T) {
		rp := NewRenderProfiler()

		min, max := rp.GetMinMaxFrameDuration()

		assert.Equal(t, time.Duration(0), min)
		assert.Equal(t, time.Duration(0), max)
	})

	t.Run("returns correct min and max", func(t *testing.T) {
		rp := NewRenderProfiler()

		rp.RecordFrame(20 * time.Millisecond)
		rp.RecordFrame(5 * time.Millisecond)
		rp.RecordFrame(30 * time.Millisecond)
		rp.RecordFrame(10 * time.Millisecond)

		min, max := rp.GetMinMaxFrameDuration()

		assert.Equal(t, 5*time.Millisecond, min)
		assert.Equal(t, 30*time.Millisecond, max)
	})
}

func TestRenderProfiler_Reset(t *testing.T) {
	t.Run("clears all data", func(t *testing.T) {
		rp := NewRenderProfiler()

		// Add some data
		rp.RecordFrame(10 * time.Millisecond)
		time.Sleep(10 * time.Millisecond)
		rp.RecordFrame(10 * time.Millisecond)

		assert.Greater(t, rp.GetFrameCount(), 0)

		rp.Reset()

		assert.Equal(t, 0, rp.GetFrameCount())
		assert.Equal(t, 0.0, rp.GetFPS())
		assert.True(t, rp.lastFrame.IsZero())
	})
}

func TestRenderProfiler_MaxFramesLimit(t *testing.T) {
	t.Run("limits stored frames to max", func(t *testing.T) {
		config := &RenderConfig{
			TargetFPS:             60,
			MaxFrames:             5, // Very small limit
			MaxFPSSamples:         60,
			DroppedFrameThreshold: 17 * time.Millisecond,
		}
		rp := NewRenderProfilerWithConfig(config)

		// Record more frames than max
		for i := 0; i < 20; i++ {
			rp.RecordFrame(10 * time.Millisecond)
		}

		// Should be limited to MaxFrames
		assert.LessOrEqual(t, rp.GetFrameCount(), config.MaxFrames)
	})
}

func TestRenderProfiler_ThreadSafety(t *testing.T) {
	t.Run("concurrent operations are safe", func(t *testing.T) {
		rp := NewRenderProfiler()
		var wg sync.WaitGroup
		goroutines := 50

		// Concurrent RecordFrame
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				rp.RecordFrame(10 * time.Millisecond)
			}()
		}

		// Concurrent reads
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = rp.GetFPS()
				_ = rp.GetDroppedFramePercent()
				_ = rp.GetFrames()
				_ = rp.GetFrameCount()
				_ = rp.GetAverageFrameDuration()
			}()
		}

		wg.Wait()
		// Should complete without race conditions
	})
}

func TestRenderProfiler_GetConfig(t *testing.T) {
	t.Run("returns current config", func(t *testing.T) {
		config := &RenderConfig{
			TargetFPS:             30,
			MaxFrames:             100,
			MaxFPSSamples:         30,
			DroppedFrameThreshold: 33 * time.Millisecond,
		}
		rp := NewRenderProfilerWithConfig(config)

		got := rp.GetConfig()

		assert.Equal(t, 30, got.TargetFPS)
		assert.Equal(t, 100, got.MaxFrames)
	})
}

func TestRenderProfiler_SetConfig(t *testing.T) {
	t.Run("updates config", func(t *testing.T) {
		rp := NewRenderProfiler()

		newConfig := &RenderConfig{
			TargetFPS:             30,
			MaxFrames:             50,
			MaxFPSSamples:         30,
			DroppedFrameThreshold: 33 * time.Millisecond,
		}
		rp.SetConfig(newConfig)

		got := rp.GetConfig()
		assert.Equal(t, 30, got.TargetFPS)
	})
}

func TestRenderProfiler_IsDroppedFrame(t *testing.T) {
	tests := []struct {
		name      string
		duration  time.Duration
		threshold time.Duration
		want      bool
	}{
		{
			name:      "fast frame not dropped",
			duration:  10 * time.Millisecond,
			threshold: 17 * time.Millisecond,
			want:      false,
		},
		{
			name:      "slow frame is dropped",
			duration:  20 * time.Millisecond,
			threshold: 17 * time.Millisecond,
			want:      true,
		},
		{
			name:      "exactly at threshold not dropped",
			duration:  17 * time.Millisecond,
			threshold: 17 * time.Millisecond,
			want:      false,
		},
		{
			name:      "just over threshold is dropped",
			duration:  18 * time.Millisecond,
			threshold: 17 * time.Millisecond,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &RenderConfig{
				TargetFPS:             60,
				MaxFrames:             1000,
				MaxFPSSamples:         60,
				DroppedFrameThreshold: tt.threshold,
			}
			rp := NewRenderProfilerWithConfig(config)

			rp.RecordFrame(tt.duration)

			frames := rp.GetFrames()
			require.Len(t, frames, 1)
			assert.Equal(t, tt.want, frames[0].Dropped)
		})
	}
}

func TestRenderProfiler_PerformanceAcceptable(t *testing.T) {
	t.Run("RecordFrame has minimal overhead", func(t *testing.T) {
		rp := NewRenderProfiler()

		// Measure overhead of RecordFrame
		iterations := 10000
		start := time.Now()

		for i := 0; i < iterations; i++ {
			rp.RecordFrame(10 * time.Millisecond)
		}

		elapsed := time.Since(start)
		avgOverhead := elapsed / time.Duration(iterations)

		// RecordFrame should take less than 1ms per call
		assert.Less(t, avgOverhead, 1*time.Millisecond,
			"RecordFrame overhead too high: %v per call", avgOverhead)
	})

	t.Run("GetFPS has minimal overhead", func(t *testing.T) {
		rp := NewRenderProfiler()
		// Pre-populate with samples
		rp.fpsSamples = make([]float64, 60)
		for i := range rp.fpsSamples {
			rp.fpsSamples[i] = 60.0
		}

		iterations := 10000
		start := time.Now()

		for i := 0; i < iterations; i++ {
			_ = rp.GetFPS()
		}

		elapsed := time.Since(start)
		avgOverhead := elapsed / time.Duration(iterations)

		// GetFPS should take less than 100µs per call
		assert.Less(t, avgOverhead, 100*time.Microsecond,
			"GetFPS overhead too high: %v per call", avgOverhead)
	})
}

func TestFrameInfo_Fields(t *testing.T) {
	t.Run("has all required fields", func(t *testing.T) {
		now := time.Now()
		frame := FrameInfo{
			Timestamp: now,
			Duration:  16 * time.Millisecond,
			Dropped:   false,
		}

		assert.Equal(t, now, frame.Timestamp)
		assert.Equal(t, 16*time.Millisecond, frame.Duration)
		assert.False(t, frame.Dropped)
	})
}
