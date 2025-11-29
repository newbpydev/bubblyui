// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"math"
	"sort"
	"sync"
)

// DefaultFPSWindowSize is the default number of FPS samples to keep for averaging.
const DefaultFPSWindowSize = 60

// FPSCalculator calculates frames per second metrics from FPS samples.
//
// It maintains a sliding window of FPS samples and provides various
// statistical calculations including average, min, max, standard deviation,
// and percentiles.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	fc := NewFPSCalculator()
//
//	// Add FPS samples from your render loop
//	fc.AddSample(60.0)
//	fc.AddSample(58.5)
//	fc.AddSample(61.2)
//
//	// Get statistics
//	fmt.Printf("Average FPS: %.1f\n", fc.GetAverage())
//	fmt.Printf("Min FPS: %.1f\n", fc.GetMin())
//	fmt.Printf("Max FPS: %.1f\n", fc.GetMax())
type FPSCalculator struct {
	// samples stores the FPS values
	samples []float64

	// windowSize is the maximum number of samples to keep
	windowSize int

	// mu protects concurrent access to calculator state
	mu sync.RWMutex
}

// NewFPSCalculator creates a new FPS calculator with default window size (60 samples).
//
// Example:
//
//	fc := NewFPSCalculator()
//	fc.AddSample(60.0)
//	fmt.Printf("Average: %.1f\n", fc.GetAverage())
func NewFPSCalculator() *FPSCalculator {
	return NewFPSCalculatorWithWindowSize(DefaultFPSWindowSize)
}

// NewFPSCalculatorWithWindowSize creates a new FPS calculator with a custom window size.
//
// If windowSize is <= 0, DefaultFPSWindowSize is used.
//
// Example:
//
//	// Keep last 30 samples for a smoother average
//	fc := NewFPSCalculatorWithWindowSize(30)
func NewFPSCalculatorWithWindowSize(windowSize int) *FPSCalculator {
	if windowSize <= 0 {
		windowSize = DefaultFPSWindowSize
	}

	return &FPSCalculator{
		samples:    make([]float64, 0, windowSize),
		windowSize: windowSize,
	}
}

// AddSample adds an FPS sample to the calculator.
//
// If the number of samples exceeds the window size, the oldest sample is removed.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	fc.AddSample(60.0)
//	fc.AddSample(58.5)
func (fc *FPSCalculator) AddSample(fps float64) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.samples = append(fc.samples, fps)

	// Remove oldest sample if we exceed window size
	if len(fc.samples) > fc.windowSize {
		fc.samples = fc.samples[1:]
	}
}

// GetAverage returns the average FPS from all samples in the window.
//
// Returns 0 if no samples have been added.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	avg := fc.GetAverage()
//	fmt.Printf("Average FPS: %.1f\n", avg)
func (fc *FPSCalculator) GetAverage() float64 {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	if len(fc.samples) == 0 {
		return 0
	}

	sum := 0.0
	for _, fps := range fc.samples {
		sum += fps
	}

	return sum / float64(len(fc.samples))
}

// GetMin returns the minimum FPS from all samples in the window.
//
// Returns 0 if no samples have been added.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	min := fc.GetMin()
//	fmt.Printf("Min FPS: %.1f\n", min)
func (fc *FPSCalculator) GetMin() float64 {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	if len(fc.samples) == 0 {
		return 0
	}

	min := fc.samples[0]
	for _, fps := range fc.samples[1:] {
		if fps < min {
			min = fps
		}
	}

	return min
}

// GetMax returns the maximum FPS from all samples in the window.
//
// Returns 0 if no samples have been added.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	max := fc.GetMax()
//	fmt.Printf("Max FPS: %.1f\n", max)
func (fc *FPSCalculator) GetMax() float64 {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	if len(fc.samples) == 0 {
		return 0
	}

	max := fc.samples[0]
	for _, fps := range fc.samples[1:] {
		if fps > max {
			max = fps
		}
	}

	return max
}

// GetMinMax returns both the minimum and maximum FPS from all samples.
//
// Returns (0, 0) if no samples have been added.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	min, max := fc.GetMinMax()
//	fmt.Printf("FPS range: %.1f - %.1f\n", min, max)
func (fc *FPSCalculator) GetMinMax() (min, max float64) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	if len(fc.samples) == 0 {
		return 0, 0
	}

	min = fc.samples[0]
	max = fc.samples[0]

	for _, fps := range fc.samples[1:] {
		if fps < min {
			min = fps
		}
		if fps > max {
			max = fps
		}
	}

	return min, max
}

// GetStandardDeviation returns the population standard deviation of FPS samples.
//
// Returns 0 if there are fewer than 2 samples.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	stdDev := fc.GetStandardDeviation()
//	if stdDev > 10 {
//	    fmt.Println("Warning: FPS is unstable")
//	}
func (fc *FPSCalculator) GetStandardDeviation() float64 {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	if len(fc.samples) <= 1 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, fps := range fc.samples {
		sum += fps
	}
	mean := sum / float64(len(fc.samples))

	// Calculate variance
	variance := 0.0
	for _, fps := range fc.samples {
		diff := fps - mean
		variance += diff * diff
	}
	variance /= float64(len(fc.samples))

	return math.Sqrt(variance)
}

// GetPercentile returns the FPS value at the given percentile (0-100).
//
// Uses the nearest-rank method for percentile calculation.
// Returns 0 if no samples have been added.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	p50 := fc.GetPercentile(50) // Median
//	p95 := fc.GetPercentile(95) // 95th percentile
func (fc *FPSCalculator) GetPercentile(percentile float64) float64 {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	if len(fc.samples) == 0 {
		return 0
	}

	if len(fc.samples) == 1 {
		return fc.samples[0]
	}

	// Create sorted copy
	sorted := make([]float64, len(fc.samples))
	copy(sorted, fc.samples)
	sort.Float64s(sorted)

	// Clamp percentile to valid range
	if percentile < 0 {
		percentile = 0
	}
	if percentile > 100 {
		percentile = 100
	}

	// Calculate index using nearest-rank method
	index := int(math.Ceil(percentile/100*float64(len(sorted)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// IsStable returns true if the FPS is stable (standard deviation below threshold).
//
// Returns false if there are no samples.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	if fc.IsStable(5.0) {
//	    fmt.Println("FPS is stable")
//	}
func (fc *FPSCalculator) IsStable(threshold float64) bool {
	fc.mu.RLock()
	n := len(fc.samples)
	fc.mu.RUnlock()

	if n == 0 {
		return false
	}

	if n == 1 {
		return true
	}

	return fc.GetStandardDeviation() <= threshold
}

// SampleCount returns the number of samples currently stored.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (fc *FPSCalculator) SampleCount() int {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	return len(fc.samples)
}

// GetSamples returns a copy of all current samples.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (fc *FPSCalculator) GetSamples() []float64 {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	// Return a copy to prevent external modification
	samples := make([]float64, len(fc.samples))
	copy(samples, fc.samples)
	return samples
}

// GetWindowSize returns the current window size.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (fc *FPSCalculator) GetWindowSize() int {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	return fc.windowSize
}

// SetWindowSize changes the window size.
//
// If the new window size is smaller than the current number of samples,
// the oldest samples are removed to fit within the new window.
//
// If windowSize is <= 0, DefaultFPSWindowSize is used.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	fc.SetWindowSize(30) // Keep only last 30 samples
func (fc *FPSCalculator) SetWindowSize(windowSize int) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if windowSize <= 0 {
		windowSize = DefaultFPSWindowSize
	}

	fc.windowSize = windowSize

	// Truncate samples if needed
	if len(fc.samples) > windowSize {
		fc.samples = fc.samples[len(fc.samples)-windowSize:]
	}
}

// Reset clears all samples while preserving the window size.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	fc.Reset() // Clear all samples
func (fc *FPSCalculator) Reset() {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.samples = make([]float64, 0, fc.windowSize)
}
