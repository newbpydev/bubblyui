// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"sync"
	"time"
)

// RenderProfiler tracks render performance and FPS for BubblyUI applications.
//
// It records frame timing information, calculates frames per second (FPS),
// and detects dropped frames that exceed the target frame time.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	rp := NewRenderProfiler()
//
//	// In your render loop
//	start := time.Now()
//	// ... render frame ...
//	rp.RecordFrame(time.Since(start))
//
//	// Get performance metrics
//	fmt.Printf("FPS: %.1f\n", rp.GetFPS())
//	fmt.Printf("Dropped frames: %.1f%%\n", rp.GetDroppedFramePercent())
type RenderProfiler struct {
	// frames stores recorded frame information
	frames []FrameInfo

	// lastFrame is the timestamp of the last recorded frame
	lastFrame time.Time

	// fpsSamples stores recent FPS calculations for averaging
	fpsSamples []float64

	// config holds render profiler configuration
	config *RenderConfig

	// mu protects concurrent access to profiler state
	mu sync.RWMutex
}

// RenderConfig defines configuration for the render profiler.
type RenderConfig struct {
	// TargetFPS is the target frames per second (default: 60)
	TargetFPS int

	// MaxFrames is the maximum number of frames to store (default: 1000)
	MaxFrames int

	// MaxFPSSamples is the maximum number of FPS samples to keep (default: 60)
	MaxFPSSamples int

	// DroppedFrameThreshold is the duration above which a frame is considered dropped
	// Default: ~16.67ms for 60fps
	DroppedFrameThreshold time.Duration
}

// FrameInfo contains information about a single rendered frame.
type FrameInfo struct {
	// Timestamp is when the frame was recorded
	Timestamp time.Time

	// Duration is how long the frame took to render
	Duration time.Duration

	// Dropped indicates if this frame exceeded the target frame time
	Dropped bool
}

// DefaultRenderConfig returns sensible default configuration for render profiling.
//
// Default values:
//   - TargetFPS: 60
//   - MaxFrames: 1000
//   - MaxFPSSamples: 60
//   - DroppedFrameThreshold: ~16.67ms (1000ms / 60fps)
func DefaultRenderConfig() *RenderConfig {
	return &RenderConfig{
		TargetFPS:             60,
		MaxFrames:             1000,
		MaxFPSSamples:         60,
		DroppedFrameThreshold: time.Second / 60, // ~16.67ms for 60fps
	}
}

// NewRenderProfiler creates a new render profiler with default configuration.
//
// Example:
//
//	rp := NewRenderProfiler()
//	rp.RecordFrame(frameTime)
//	fmt.Printf("FPS: %.1f\n", rp.GetFPS())
func NewRenderProfiler() *RenderProfiler {
	return NewRenderProfilerWithConfig(DefaultRenderConfig())
}

// NewRenderProfilerWithConfig creates a new render profiler with custom configuration.
//
// Example:
//
//	config := &RenderConfig{
//	    TargetFPS: 30,
//	    MaxFrames: 500,
//	    DroppedFrameThreshold: 33 * time.Millisecond,
//	}
//	rp := NewRenderProfilerWithConfig(config)
func NewRenderProfilerWithConfig(config *RenderConfig) *RenderProfiler {
	return &RenderProfiler{
		frames:     make([]FrameInfo, 0, config.MaxFrames),
		fpsSamples: make([]float64, 0, config.MaxFPSSamples),
		config:     config,
	}
}

// RecordFrame records a frame with the given render duration.
//
// It automatically:
//   - Detects if the frame was dropped (exceeded threshold)
//   - Calculates instantaneous FPS from frame intervals
//   - Maintains frame and FPS sample limits
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	start := time.Now()
//	// ... render frame ...
//	rp.RecordFrame(time.Since(start))
func (rp *RenderProfiler) RecordFrame(duration time.Duration) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	now := time.Now()

	frame := FrameInfo{
		Timestamp: now,
		Duration:  duration,
		Dropped:   duration > rp.config.DroppedFrameThreshold,
	}

	rp.frames = append(rp.frames, frame)

	// Limit stored frames
	if len(rp.frames) > rp.config.MaxFrames {
		rp.frames = rp.frames[1:]
	}

	// Calculate FPS from frame intervals
	if !rp.lastFrame.IsZero() {
		frameDelta := now.Sub(rp.lastFrame)
		if frameDelta > 0 {
			fps := 1.0 / frameDelta.Seconds()
			rp.fpsSamples = append(rp.fpsSamples, fps)

			// Limit FPS samples
			if len(rp.fpsSamples) > rp.config.MaxFPSSamples {
				rp.fpsSamples = rp.fpsSamples[1:]
			}
		}
	}

	rp.lastFrame = now
}

// GetFPS returns the average frames per second based on recent samples.
//
// Returns 0 if no FPS samples have been recorded yet.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	fps := rp.GetFPS()
//	fmt.Printf("Current FPS: %.1f\n", fps)
func (rp *RenderProfiler) GetFPS() float64 {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	if len(rp.fpsSamples) == 0 {
		return 0
	}

	sum := 0.0
	for _, fps := range rp.fpsSamples {
		sum += fps
	}

	return sum / float64(len(rp.fpsSamples))
}

// GetDroppedFramePercent returns the percentage of frames that were dropped.
//
// A frame is considered "dropped" if its render duration exceeded the
// configured DroppedFrameThreshold (default: ~16.67ms for 60fps).
//
// Returns 0 if no frames have been recorded.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	percent := rp.GetDroppedFramePercent()
//	if percent > 5 {
//	    fmt.Printf("WARNING: %.1f%% frames dropped\n", percent)
//	}
func (rp *RenderProfiler) GetDroppedFramePercent() float64 {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	if len(rp.frames) == 0 {
		return 0
	}

	dropped := 0
	for _, frame := range rp.frames {
		if frame.Dropped {
			dropped++
		}
	}

	return float64(dropped) / float64(len(rp.frames)) * 100
}

// GetFrames returns a copy of all recorded frames.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (rp *RenderProfiler) GetFrames() []FrameInfo {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	// Return a copy to prevent external modification
	frames := make([]FrameInfo, len(rp.frames))
	copy(frames, rp.frames)
	return frames
}

// GetFrameCount returns the number of recorded frames.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (rp *RenderProfiler) GetFrameCount() int {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	return len(rp.frames)
}

// GetAverageFrameDuration returns the average frame render duration.
//
// Returns 0 if no frames have been recorded.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (rp *RenderProfiler) GetAverageFrameDuration() time.Duration {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	if len(rp.frames) == 0 {
		return 0
	}

	var total time.Duration
	for _, frame := range rp.frames {
		total += frame.Duration
	}

	return total / time.Duration(len(rp.frames))
}

// GetMinMaxFrameDuration returns the minimum and maximum frame durations.
//
// Returns (0, 0) if no frames have been recorded.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (rp *RenderProfiler) GetMinMaxFrameDuration() (min, max time.Duration) {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	if len(rp.frames) == 0 {
		return 0, 0
	}

	min = rp.frames[0].Duration
	max = rp.frames[0].Duration

	for _, frame := range rp.frames[1:] {
		if frame.Duration < min {
			min = frame.Duration
		}
		if frame.Duration > max {
			max = frame.Duration
		}
	}

	return min, max
}

// Reset clears all recorded frames and FPS samples.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (rp *RenderProfiler) Reset() {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	rp.frames = make([]FrameInfo, 0, rp.config.MaxFrames)
	rp.fpsSamples = make([]float64, 0, rp.config.MaxFPSSamples)
	rp.lastFrame = time.Time{}
}

// GetConfig returns a copy of the current configuration.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (rp *RenderProfiler) GetConfig() *RenderConfig {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	return &RenderConfig{
		TargetFPS:             rp.config.TargetFPS,
		MaxFrames:             rp.config.MaxFrames,
		MaxFPSSamples:         rp.config.MaxFPSSamples,
		DroppedFrameThreshold: rp.config.DroppedFrameThreshold,
	}
}

// SetConfig updates the render profiler configuration.
//
// Note: This does not resize existing data. Use Reset() after SetConfig()
// if you want to clear existing data.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (rp *RenderProfiler) SetConfig(config *RenderConfig) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	rp.config = config
}
