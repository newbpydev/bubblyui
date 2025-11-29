// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"fmt"
	"runtime"
	"sync"
)

// LeakDetector detects memory leaks from runtime memory snapshots.
//
// It analyzes memory snapshots to identify heap growth, goroutine leaks,
// and heap object accumulation. The detector uses configurable thresholds
// to distinguish between normal memory fluctuations and actual leaks.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	ld := NewLeakDetector()
//
//	// Collect snapshots over time
//	snapshots := []*runtime.MemStats{...}
//
//	// Detect leaks
//	leaks := ld.DetectLeaks(snapshots)
//	for _, leak := range leaks {
//	    fmt.Printf("LEAK: %s - %s\n", leak.Type, leak.Description)
//	}
//
//	// Check for goroutine leaks
//	before := runtime.NumGoroutine()
//	// ... run workload ...
//	after := runtime.NumGoroutine()
//	if leak := ld.DetectGoroutineLeaks(before, after); leak != nil {
//	    fmt.Printf("GOROUTINE LEAK: %s\n", leak.Description)
//	}
type LeakDetector struct {
	// thresholds defines the leak detection thresholds
	thresholds *LeakThresholds

	// mu protects concurrent access to detector state
	mu sync.RWMutex
}

// LeakThresholds defines configurable thresholds for leak detection.
//
// Thresholds help filter out normal memory fluctuations from actual leaks.
// Values below these thresholds are not reported as leaks.
type LeakThresholds struct {
	// HeapGrowthBytes is the minimum heap growth to report as a leak
	HeapGrowthBytes int64

	// GoroutineGrowth is the minimum goroutine count increase to report as a leak
	GoroutineGrowth int

	// HeapObjectGrowth is the minimum heap object count increase to report as a leak
	HeapObjectGrowth int64

	// GCPauseThreshold is the minimum GC pause duration to flag (reserved for future use)
	GCPauseThreshold int64

	// SeverityHighBytes is the threshold for high severity (bytes)
	SeverityHighBytes int64

	// SeverityCriticalBytes is the threshold for critical severity (bytes)
	SeverityCriticalBytes int64

	// SeverityHighGoroutines is the threshold for high severity (goroutine count)
	SeverityHighGoroutines int

	// SeverityCriticalGoroutines is the threshold for critical severity (goroutine count)
	SeverityCriticalGoroutines int
}

// LeakInfo describes a detected memory leak.
//
// It contains information about the type of leak, severity,
// and a human-readable description with suggestions.
type LeakInfo struct {
	// Type categorizes the leak (e.g., "heap_growth", "goroutine_leak", "heap_object_growth")
	Type string

	// BytesLeaked is the amount of memory leaked in bytes (for heap leaks)
	BytesLeaked int64

	// Count is the number of leaked items (e.g., goroutines, objects)
	Count int

	// Description provides a human-readable explanation of the leak
	Description string

	// Severity indicates how critical the leak is
	Severity Severity
}

// DefaultLeakThresholds returns sensible default thresholds for leak detection.
//
// Default values:
//   - HeapGrowthBytes: 1MB (1,048,576 bytes)
//   - GoroutineGrowth: 10 goroutines
//   - HeapObjectGrowth: 10,000 objects
//   - SeverityHighBytes: 10MB
//   - SeverityCriticalBytes: 100MB
//   - SeverityHighGoroutines: 50
//   - SeverityCriticalGoroutines: 500
func DefaultLeakThresholds() *LeakThresholds {
	return &LeakThresholds{
		HeapGrowthBytes:            1 * 1024 * 1024,   // 1MB
		GoroutineGrowth:            10,                // 10 goroutines
		HeapObjectGrowth:           10000,             // 10,000 objects
		GCPauseThreshold:           0,                 // Reserved
		SeverityHighBytes:          10 * 1024 * 1024,  // 10MB
		SeverityCriticalBytes:      100 * 1024 * 1024, // 100MB
		SeverityHighGoroutines:     50,
		SeverityCriticalGoroutines: 500,
	}
}

// NewLeakDetector creates a new leak detector with default thresholds.
//
// Example:
//
//	ld := NewLeakDetector()
//	leaks := ld.DetectLeaks(snapshots)
func NewLeakDetector() *LeakDetector {
	return &LeakDetector{
		thresholds: DefaultLeakThresholds(),
	}
}

// NewLeakDetectorWithThresholds creates a new leak detector with custom thresholds.
//
// Example:
//
//	thresholds := &LeakThresholds{
//	    HeapGrowthBytes: 5 * 1024 * 1024, // 5MB
//	    GoroutineGrowth: 20,
//	}
//	ld := NewLeakDetectorWithThresholds(thresholds)
func NewLeakDetectorWithThresholds(thresholds *LeakThresholds) *LeakDetector {
	return &LeakDetector{
		thresholds: thresholds,
	}
}

// DetectLeaks analyzes memory snapshots to detect potential memory leaks.
//
// It compares the first and last snapshots to identify:
//   - Heap growth exceeding the threshold
//   - Heap object count growth exceeding the threshold
//
// Returns an empty slice if no leaks are detected or if there are
// insufficient snapshots (less than 2).
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	mp := NewMemoryProfiler()
//	mp.TakeSnapshot()
//	// ... run workload ...
//	mp.TakeSnapshot()
//
//	ld := NewLeakDetector()
//	leaks := ld.DetectLeaks(mp.GetSnapshots())
//	for _, leak := range leaks {
//	    fmt.Printf("LEAK: %s (severity: %s)\n", leak.Description, leak.Severity)
//	}
func (ld *LeakDetector) DetectLeaks(snapshots []*runtime.MemStats) []*LeakInfo {
	ld.mu.RLock()
	thresholds := ld.thresholds
	ld.mu.RUnlock()

	if len(snapshots) < 2 {
		return []*LeakInfo{}
	}

	first := snapshots[0]
	last := snapshots[len(snapshots)-1]

	leaks := make([]*LeakInfo, 0)

	// Check for heap growth
	heapGrowth := int64(last.HeapAlloc) - int64(first.HeapAlloc)
	if heapGrowth > thresholds.HeapGrowthBytes {
		leaks = append(leaks, &LeakInfo{
			Type:        "heap_growth",
			BytesLeaked: heapGrowth,
			Count:       0,
			Description: fmt.Sprintf("Heap grew by %s (%d bytes) from %s to %s",
				formatBytes(heapGrowth),
				heapGrowth,
				formatBytes(int64(first.HeapAlloc)),
				formatBytes(int64(last.HeapAlloc))),
			Severity: ld.calculateSeverity(heapGrowth),
		})
	}

	// Check for heap object growth
	objectGrowth := int64(last.HeapObjects) - int64(first.HeapObjects)
	if objectGrowth > thresholds.HeapObjectGrowth {
		leaks = append(leaks, &LeakInfo{
			Type:        "heap_object_growth",
			BytesLeaked: 0,
			Count:       int(objectGrowth),
			Description: fmt.Sprintf("Heap objects grew by %d (from %d to %d)",
				objectGrowth, first.HeapObjects, last.HeapObjects),
			Severity: ld.calculateObjectSeverity(objectGrowth),
		})
	}

	return leaks
}

// DetectGoroutineLeaks checks for goroutine leaks by comparing counts.
//
// Returns nil if the goroutine growth is below the threshold.
// Returns a LeakInfo if a potential goroutine leak is detected.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	before := runtime.NumGoroutine()
//	// ... run workload that should clean up goroutines ...
//	after := runtime.NumGoroutine()
//
//	ld := NewLeakDetector()
//	if leak := ld.DetectGoroutineLeaks(before, after); leak != nil {
//	    fmt.Printf("WARNING: %s\n", leak.Description)
//	}
func (ld *LeakDetector) DetectGoroutineLeaks(before, after int) *LeakInfo {
	ld.mu.RLock()
	thresholds := ld.thresholds
	ld.mu.RUnlock()

	growth := after - before
	if growth <= thresholds.GoroutineGrowth {
		return nil
	}

	return &LeakInfo{
		Type:        "goroutine_leak",
		BytesLeaked: 0,
		Count:       growth,
		Description: fmt.Sprintf("%d goroutines leaked (from %d to %d)",
			growth, before, after),
		Severity: ld.calculateGoroutineSeverity(growth),
	}
}

// calculateSeverity determines the severity of a heap leak based on bytes leaked.
func (ld *LeakDetector) calculateSeverity(bytesLeaked int64) Severity {
	ld.mu.RLock()
	thresholds := ld.thresholds
	ld.mu.RUnlock()

	if bytesLeaked >= thresholds.SeverityCriticalBytes {
		return SeverityCritical
	}
	if bytesLeaked >= thresholds.SeverityHighBytes {
		return SeverityHigh
	}
	// Medium threshold is half of high
	if bytesLeaked >= thresholds.SeverityHighBytes/2 {
		return SeverityMedium
	}
	return SeverityLow
}

// calculateObjectSeverity determines the severity based on object count growth.
func (ld *LeakDetector) calculateObjectSeverity(objectGrowth int64) Severity {
	// Object severity thresholds (relative to heap object threshold)
	ld.mu.RLock()
	thresholds := ld.thresholds
	ld.mu.RUnlock()

	criticalThreshold := thresholds.HeapObjectGrowth * 100 // 100x threshold
	highThreshold := thresholds.HeapObjectGrowth * 10      // 10x threshold
	mediumThreshold := thresholds.HeapObjectGrowth * 2     // 2x threshold

	if objectGrowth >= criticalThreshold {
		return SeverityCritical
	}
	if objectGrowth >= highThreshold {
		return SeverityHigh
	}
	if objectGrowth >= mediumThreshold {
		return SeverityMedium
	}
	return SeverityLow
}

// calculateGoroutineSeverity determines the severity of a goroutine leak.
func (ld *LeakDetector) calculateGoroutineSeverity(growth int) Severity {
	ld.mu.RLock()
	thresholds := ld.thresholds
	ld.mu.RUnlock()

	if growth >= thresholds.SeverityCriticalGoroutines {
		return SeverityCritical
	}
	if growth >= thresholds.SeverityHighGoroutines {
		return SeverityHigh
	}
	// Medium threshold is half of high
	if growth >= thresholds.SeverityHighGoroutines/2 {
		return SeverityMedium
	}
	return SeverityLow
}

// GetThresholds returns a copy of the current leak detection thresholds.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (ld *LeakDetector) GetThresholds() *LeakThresholds {
	ld.mu.RLock()
	defer ld.mu.RUnlock()

	// Return a copy to prevent external modification
	return &LeakThresholds{
		HeapGrowthBytes:            ld.thresholds.HeapGrowthBytes,
		GoroutineGrowth:            ld.thresholds.GoroutineGrowth,
		HeapObjectGrowth:           ld.thresholds.HeapObjectGrowth,
		GCPauseThreshold:           ld.thresholds.GCPauseThreshold,
		SeverityHighBytes:          ld.thresholds.SeverityHighBytes,
		SeverityCriticalBytes:      ld.thresholds.SeverityCriticalBytes,
		SeverityHighGoroutines:     ld.thresholds.SeverityHighGoroutines,
		SeverityCriticalGoroutines: ld.thresholds.SeverityCriticalGoroutines,
	}
}

// SetThresholds updates the leak detection thresholds.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	ld := NewLeakDetector()
//	ld.SetThresholds(&LeakThresholds{
//	    HeapGrowthBytes: 5 * 1024 * 1024, // 5MB
//	    GoroutineGrowth: 20,
//	})
func (ld *LeakDetector) SetThresholds(thresholds *LeakThresholds) {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	ld.thresholds = thresholds
}

// Reset restores the leak detector to default thresholds.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (ld *LeakDetector) Reset() {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	ld.thresholds = DefaultLeakThresholds()
}

// formatBytes formats a byte count as a human-readable string.
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	absBytes := bytes
	if absBytes < 0 {
		absBytes = -absBytes
	}

	switch {
	case absBytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case absBytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case absBytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
