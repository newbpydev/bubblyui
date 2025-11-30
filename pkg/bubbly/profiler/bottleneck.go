// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"fmt"
	"time"
)

// BottleneckThresholds defines configurable thresholds for bottleneck detection.
//
// These thresholds determine when an operation is considered a bottleneck.
// Operations exceeding their threshold are flagged and tracked.
type BottleneckThresholds struct {
	// DefaultOperationThreshold is the default threshold for operations
	// without a specific threshold configured.
	DefaultOperationThreshold time.Duration

	// RenderThreshold is the threshold for render operations.
	// Renders exceeding this duration are flagged as slow.
	RenderThreshold time.Duration

	// UpdateThreshold is the threshold for update operations.
	UpdateThreshold time.Duration

	// EventThreshold is the threshold for event handling operations.
	EventThreshold time.Duration

	// FrequentRenderThreshold is the count threshold for detecting
	// components that render too frequently.
	FrequentRenderThreshold int64

	// MemoryThreshold is the threshold for memory usage (bytes).
	// Components exceeding this are flagged for memory issues.
	MemoryThreshold uint64
}

// PerformanceMetrics aggregates metrics for bottleneck detection.
//
// This struct is passed to Detect() to analyze performance data
// and identify bottlenecks across the application.
type PerformanceMetrics struct {
	// Components contains per-component performance data
	Components []*ComponentMetrics

	// Timings contains operation timing data (reserved for future use)
	Timings map[string]*TimingStats

	// MemoryUsage is current memory usage in bytes
	MemoryUsage uint64

	// GoroutineCount is current goroutine count
	GoroutineCount int
}

// DefaultBottleneckThresholds returns sensible default thresholds for bottleneck detection.
//
// Default values:
//   - DefaultOperationThreshold: 16ms (60 FPS frame budget)
//   - RenderThreshold: 16ms
//   - UpdateThreshold: 5ms
//   - EventThreshold: 10ms
//   - FrequentRenderThreshold: 1000 renders
//   - MemoryThreshold: 10MB
func DefaultBottleneckThresholds() *BottleneckThresholds {
	return &BottleneckThresholds{
		DefaultOperationThreshold: 16 * time.Millisecond, // 60 FPS frame budget
		RenderThreshold:           16 * time.Millisecond,
		UpdateThreshold:           5 * time.Millisecond,
		EventThreshold:            10 * time.Millisecond,
		FrequentRenderThreshold:   1000,
		MemoryThreshold:           10 * 1024 * 1024, // 10MB
	}
}

// NewBottleneckDetector creates a new bottleneck detector with default thresholds.
//
// Example:
//
//	bd := NewBottleneckDetector()
//	bd.SetThreshold("render", 16*time.Millisecond)
//	if bottleneck := bd.Check("render", duration); bottleneck != nil {
//	    fmt.Printf("Bottleneck: %s\n", bottleneck.Description)
//	}
func NewBottleneckDetector() *BottleneckDetector {
	return &BottleneckDetector{
		thresholds: make(map[string]time.Duration),
		violations: make(map[string]int),
		config:     DefaultBottleneckThresholds(),
	}
}

// NewBottleneckDetectorWithThresholds creates a new bottleneck detector with custom thresholds.
//
// Example:
//
//	config := &BottleneckThresholds{
//	    DefaultOperationThreshold: 10 * time.Millisecond,
//	    RenderThreshold:           20 * time.Millisecond,
//	}
//	bd := NewBottleneckDetectorWithThresholds(config)
func NewBottleneckDetectorWithThresholds(config *BottleneckThresholds) *BottleneckDetector {
	return &BottleneckDetector{
		thresholds: make(map[string]time.Duration),
		violations: make(map[string]int),
		config:     config,
	}
}

// Check checks if an operation duration exceeds its threshold.
//
// Returns nil if the duration is at or below the threshold.
// Returns a BottleneckInfo if the duration exceeds the threshold.
//
// The severity is calculated based on how much the duration exceeds the threshold:
//   - < 2x threshold: Low
//   - 2-3x threshold: Medium
//   - 3-5x threshold: High
//   - > 5x threshold: Critical
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	bd := NewBottleneckDetector()
//	bd.SetThreshold("render", 16*time.Millisecond)
//
//	start := time.Now()
//	// ... render operation ...
//	duration := time.Since(start)
//
//	if bottleneck := bd.Check("render", duration); bottleneck != nil {
//	    log.Printf("Slow render: %s", bottleneck.Description)
//	}
func (bd *BottleneckDetector) Check(operation string, duration time.Duration) *BottleneckInfo {
	bd.mu.RLock()
	threshold := bd.getThresholdLocked(operation)
	bd.mu.RUnlock()

	if duration <= threshold {
		return nil
	}

	// Track violation
	bd.mu.Lock()
	bd.violations[operation]++
	bd.mu.Unlock()

	// Calculate severity and impact
	ratio := float64(duration) / float64(threshold)
	severity := calculateSeverityFromRatio(ratio)
	impact := calculateImpact(ratio)

	return &BottleneckInfo{
		Type:        BottleneckTypeSlow,
		Location:    operation,
		Severity:    severity,
		Impact:      impact,
		Description: fmt.Sprintf("%s took %v (threshold: %v, %.1fx slower)", operation, duration, threshold, ratio),
		Suggestion:  generateSuggestion(operation, BottleneckTypeSlow),
	}
}

// Detect analyzes performance metrics to detect bottlenecks.
//
// It examines component metrics to identify:
//   - Slow components (average render time exceeds threshold)
//   - Frequently rendering components (render count exceeds threshold)
//   - Memory-heavy components (memory usage exceeds threshold)
//
// Returns an empty slice if no bottlenecks are detected or if metrics is nil.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	bd := NewBottleneckDetector()
//	metrics := &PerformanceMetrics{
//	    Components: componentTracker.GetAllMetrics(),
//	}
//	bottlenecks := bd.Detect(metrics)
//	for _, b := range bottlenecks {
//	    fmt.Printf("Bottleneck: %s - %s\n", b.Location, b.Description)
//	}
func (bd *BottleneckDetector) Detect(metrics *PerformanceMetrics) []*BottleneckInfo {
	if metrics == nil {
		return []*BottleneckInfo{}
	}

	bd.mu.RLock()
	config := bd.config
	bd.mu.RUnlock()

	bottlenecks := make([]*BottleneckInfo, 0)

	for _, comp := range metrics.Components {
		// Check for slow renders
		if comp.AvgRenderTime > config.RenderThreshold {
			ratio := float64(comp.AvgRenderTime) / float64(config.RenderThreshold)
			bottlenecks = append(bottlenecks, &BottleneckInfo{
				Type:        BottleneckTypeSlow,
				Location:    comp.ComponentName,
				Severity:    calculateSeverityFromRatio(ratio),
				Impact:      calculateImpact(ratio),
				Description: fmt.Sprintf("Component %s has slow average render time: %v (threshold: %v)", comp.ComponentName, comp.AvgRenderTime, config.RenderThreshold),
				Suggestion:  generateSuggestion("render", BottleneckTypeSlow),
			})
		}

		// Check for frequent renders
		if comp.RenderCount > config.FrequentRenderThreshold {
			ratio := float64(comp.RenderCount) / float64(config.FrequentRenderThreshold)
			bottlenecks = append(bottlenecks, &BottleneckInfo{
				Type:        BottleneckTypeFrequent,
				Location:    comp.ComponentName,
				Severity:    calculateSeverityFromRatio(ratio),
				Impact:      calculateImpact(ratio),
				Description: fmt.Sprintf("Component %s renders too frequently: %d renders (threshold: %d)", comp.ComponentName, comp.RenderCount, config.FrequentRenderThreshold),
				Suggestion:  generateSuggestion("render", BottleneckTypeFrequent),
			})
		}

		// Check for memory issues
		if comp.MemoryUsage > config.MemoryThreshold {
			ratio := float64(comp.MemoryUsage) / float64(config.MemoryThreshold)
			bottlenecks = append(bottlenecks, &BottleneckInfo{
				Type:        BottleneckTypeMemory,
				Location:    comp.ComponentName,
				Severity:    calculateSeverityFromRatio(ratio),
				Impact:      calculateImpact(ratio),
				Description: fmt.Sprintf("Component %s uses excessive memory: %d bytes (threshold: %d)", comp.ComponentName, comp.MemoryUsage, config.MemoryThreshold),
				Suggestion:  generateSuggestion("memory", BottleneckTypeMemory),
			})
		}
	}

	return bottlenecks
}

// SetThreshold sets a custom threshold for a specific operation.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	bd.SetThreshold("render", 20*time.Millisecond)
//	bd.SetThreshold("update", 5*time.Millisecond)
func (bd *BottleneckDetector) SetThreshold(operation string, threshold time.Duration) {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	bd.thresholds[operation] = threshold
}

// GetThreshold returns the threshold for an operation.
//
// If no custom threshold is set for the operation, returns the default threshold.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	threshold := bd.GetThreshold("render")
//	fmt.Printf("Render threshold: %v\n", threshold)
func (bd *BottleneckDetector) GetThreshold(operation string) time.Duration {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	return bd.getThresholdLocked(operation)
}

// getThresholdLocked returns the threshold for an operation.
// Caller must hold at least a read lock.
func (bd *BottleneckDetector) getThresholdLocked(operation string) time.Duration {
	if threshold, ok := bd.thresholds[operation]; ok {
		return threshold
	}
	return bd.config.DefaultOperationThreshold
}

// GetViolations returns the number of threshold violations for an operation.
//
// Returns 0 if the operation has no recorded violations.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	count := bd.GetViolations("render")
//	fmt.Printf("Render violations: %d\n", count)
func (bd *BottleneckDetector) GetViolations(operation string) int {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	return bd.violations[operation]
}

// GetAllViolations returns all recorded violations.
//
// Returns a copy of the violations map to prevent external modification.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	violations := bd.GetAllViolations()
//	for op, count := range violations {
//	    fmt.Printf("%s: %d violations\n", op, count)
//	}
func (bd *BottleneckDetector) GetAllViolations() map[string]int {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	result := make(map[string]int, len(bd.violations))
	for k, v := range bd.violations {
		result[k] = v
	}
	return result
}

// GetConfig returns a copy of the current configuration.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (bd *BottleneckDetector) GetConfig() *BottleneckThresholds {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	return &BottleneckThresholds{
		DefaultOperationThreshold: bd.config.DefaultOperationThreshold,
		RenderThreshold:           bd.config.RenderThreshold,
		UpdateThreshold:           bd.config.UpdateThreshold,
		EventThreshold:            bd.config.EventThreshold,
		FrequentRenderThreshold:   bd.config.FrequentRenderThreshold,
		MemoryThreshold:           bd.config.MemoryThreshold,
	}
}

// Reset clears all violations and custom thresholds.
//
// The configuration is preserved; only runtime state is cleared.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	bd.Reset() // Clear all tracking data
func (bd *BottleneckDetector) Reset() {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	bd.thresholds = make(map[string]time.Duration)
	bd.violations = make(map[string]int)
}

// calculateSeverityFromRatio determines severity based on the ratio of actual to threshold.
func calculateSeverityFromRatio(ratio float64) Severity {
	switch {
	case ratio > 5.0:
		return SeverityCritical
	case ratio > 3.0:
		return SeverityHigh
	case ratio > 2.0:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

// calculateImpact normalizes the ratio to a 0.0-1.0 impact score.
func calculateImpact(ratio float64) float64 {
	impact := ratio / 10.0
	if impact > 1.0 {
		return 1.0
	}
	return impact
}

// generateSuggestion generates an actionable suggestion based on the bottleneck type.
func generateSuggestion(operation string, bottleneckType BottleneckType) string {
	switch bottleneckType {
	case BottleneckTypeSlow:
		switch operation {
		case "render":
			return "Profile the render function to identify hot spots. Consider caching expensive computations or optimizing the template."
		case "update":
			return "Review the update logic for unnecessary computations. Consider batching state updates."
		default:
			return fmt.Sprintf("Profile the %s operation to identify performance bottlenecks. Consider caching or lazy evaluation.", operation)
		}
	case BottleneckTypeFrequent:
		return "Consider implementing memoization to reduce unnecessary re-renders. Use shouldComponentUpdate logic or batch multiple updates together."
	case BottleneckTypeMemory:
		return "Review memory allocations in this component. Consider object pooling, sync.Pool for temporary objects, or check for memory leaks."
	case BottleneckTypePattern:
		return "Review the component architecture. Consider splitting into smaller components or optimizing state management."
	default:
		return "Review the operation for potential optimizations."
	}
}
