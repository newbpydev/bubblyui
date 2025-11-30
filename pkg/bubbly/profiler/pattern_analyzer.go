// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"sync"
	"time"
)

// Pattern defines a performance pattern to detect in component metrics.
//
// Patterns are used to identify common performance anti-patterns such as
// frequent re-renders, slow renders, memory issues, and architectural problems.
// Each pattern has a detection function, severity, and actionable suggestion.
//
// Example:
//
//	pattern := Pattern{
//	    Name:        "slow_render",
//	    Detect:      func(m *ComponentMetrics) bool { return m.AvgRenderTime > 10*time.Millisecond },
//	    Severity:    SeverityHigh,
//	    Description: "Component render is slow",
//	    Suggestion:  "Profile render function, optimize template",
//	}
type Pattern struct {
	// Name is a unique identifier for the pattern
	Name string

	// Detect is the function that determines if the pattern matches
	Detect func(*ComponentMetrics) bool

	// Severity indicates how critical the pattern is
	Severity Severity

	// Description explains what the pattern means
	Description string

	// Suggestion provides actionable advice for fixing the pattern
	Suggestion string
}

// PatternAnalyzer analyzes component metrics to detect performance patterns.
//
// It maintains a list of patterns and applies them to component metrics
// to identify potential performance issues. Custom patterns can be added
// to extend the built-in pattern detection.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	pa := NewPatternAnalyzer()
//	bottlenecks := pa.Analyze(componentMetrics)
//	for _, b := range bottlenecks {
//	    fmt.Printf("Pattern detected: %s - %s\n", b.Location, b.Description)
//	}
type PatternAnalyzer struct {
	// patterns is the list of patterns to check
	patterns []Pattern

	// mu protects concurrent access to analyzer state
	mu sync.RWMutex
}

// NewPatternAnalyzer creates a new PatternAnalyzer with default patterns.
//
// The default patterns include:
//   - frequent_rerender: Component re-renders very frequently (>1000 renders, <1ms avg)
//   - slow_render: Component render is slow (>10ms average)
//   - memory_hog: Component uses excessive memory (>5MB)
//   - render_spike: Component has occasional very slow renders (max >100ms)
//   - inefficient_render: Component renders frequently with moderate time (>500 renders, >5ms avg)
//
// Example:
//
//	pa := NewPatternAnalyzer()
//	bottlenecks := pa.Analyze(metrics)
func NewPatternAnalyzer() *PatternAnalyzer {
	return &PatternAnalyzer{
		patterns: defaultPatterns(),
	}
}

// NewPatternAnalyzerWithPatterns creates a new PatternAnalyzer with custom patterns.
//
// This allows complete control over which patterns are checked.
// Use AddPattern to add patterns to an existing analyzer instead.
//
// Example:
//
//	patterns := []Pattern{
//	    {Name: "custom", Detect: func(m *ComponentMetrics) bool { return m.RenderCount > 100 }, ...},
//	}
//	pa := NewPatternAnalyzerWithPatterns(patterns)
func NewPatternAnalyzerWithPatterns(patterns []Pattern) *PatternAnalyzer {
	if patterns == nil {
		patterns = make([]Pattern, 0)
	}
	return &PatternAnalyzer{
		patterns: patterns,
	}
}

// Analyze checks component metrics against all registered patterns.
//
// Returns a slice of BottleneckInfo for each pattern that matches.
// Returns an empty slice if no patterns match or if metrics is nil.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	pa := NewPatternAnalyzer()
//	metrics := &ComponentMetrics{
//	    ComponentName: "DataTable",
//	    RenderCount:   5000,
//	    AvgRenderTime: 500 * time.Microsecond,
//	}
//	bottlenecks := pa.Analyze(metrics)
//	for _, b := range bottlenecks {
//	    fmt.Printf("Pattern: %s\n", b.Description)
//	}
func (pa *PatternAnalyzer) Analyze(metrics *ComponentMetrics) []*BottleneckInfo {
	if metrics == nil {
		return []*BottleneckInfo{}
	}

	pa.mu.RLock()
	patterns := make([]Pattern, len(pa.patterns))
	copy(patterns, pa.patterns)
	pa.mu.RUnlock()

	bottlenecks := make([]*BottleneckInfo, 0)

	for _, pattern := range patterns {
		if pattern.Detect != nil && pattern.Detect(metrics) {
			bottlenecks = append(bottlenecks, &BottleneckInfo{
				Type:        BottleneckTypePattern,
				Location:    metrics.ComponentName,
				Severity:    pattern.Severity,
				Impact:      calculatePatternImpact(pattern.Severity),
				Description: pattern.Description,
				Suggestion:  pattern.Suggestion,
			})
		}
	}

	return bottlenecks
}

// AnalyzeAll checks multiple component metrics against all registered patterns.
//
// Returns a slice of BottleneckInfo for all patterns that match across all components.
// Returns an empty slice if no patterns match or if metrics is nil/empty.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	pa := NewPatternAnalyzer()
//	allMetrics := []*ComponentMetrics{metrics1, metrics2, metrics3}
//	bottlenecks := pa.AnalyzeAll(allMetrics)
func (pa *PatternAnalyzer) AnalyzeAll(metrics []*ComponentMetrics) []*BottleneckInfo {
	if len(metrics) == 0 {
		return []*BottleneckInfo{}
	}

	bottlenecks := make([]*BottleneckInfo, 0)
	for _, m := range metrics {
		bottlenecks = append(bottlenecks, pa.Analyze(m)...)
	}

	return bottlenecks
}

// AddPattern adds a custom pattern to the analyzer.
//
// Patterns are checked in the order they were added.
// Duplicate pattern names are allowed but not recommended.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	pa.AddPattern(Pattern{
//	    Name:        "custom_pattern",
//	    Detect:      func(m *ComponentMetrics) bool { return m.RenderCount > 10000 },
//	    Severity:    SeverityCritical,
//	    Description: "Component renders excessively",
//	    Suggestion:  "Implement aggressive memoization",
//	})
func (pa *PatternAnalyzer) AddPattern(pattern Pattern) {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	pa.patterns = append(pa.patterns, pattern)
}

// RemovePattern removes a pattern by name.
//
// If multiple patterns have the same name, only the first one is removed.
// Returns true if a pattern was removed, false if no pattern with that name exists.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	removed := pa.RemovePattern("slow_render")
//	if removed {
//	    fmt.Println("Pattern removed")
//	}
func (pa *PatternAnalyzer) RemovePattern(name string) bool {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	for i, p := range pa.patterns {
		if p.Name == name {
			pa.patterns = append(pa.patterns[:i], pa.patterns[i+1:]...)
			return true
		}
	}
	return false
}

// GetPattern returns a pattern by name.
//
// Returns nil if no pattern with that name exists.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	pattern := pa.GetPattern("slow_render")
//	if pattern != nil {
//	    fmt.Printf("Pattern severity: %s\n", pattern.Severity)
//	}
func (pa *PatternAnalyzer) GetPattern(name string) *Pattern {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	for _, p := range pa.patterns {
		if p.Name == name {
			// Return a copy to prevent external modification
			patternCopy := p
			return &patternCopy
		}
	}
	return nil
}

// GetPatterns returns all registered patterns.
//
// Returns a copy of the patterns slice to prevent external modification.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	patterns := pa.GetPatterns()
//	for _, p := range patterns {
//	    fmt.Printf("Pattern: %s - %s\n", p.Name, p.Description)
//	}
func (pa *PatternAnalyzer) GetPatterns() []Pattern {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	result := make([]Pattern, len(pa.patterns))
	copy(result, pa.patterns)
	return result
}

// PatternCount returns the number of registered patterns.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	count := pa.PatternCount()
//	fmt.Printf("Registered patterns: %d\n", count)
func (pa *PatternAnalyzer) PatternCount() int {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	return len(pa.patterns)
}

// Reset removes all patterns and restores default patterns.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	pa.Reset() // Restore default patterns
func (pa *PatternAnalyzer) Reset() {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	pa.patterns = defaultPatterns()
}

// ClearPatterns removes all patterns.
//
// After calling this, Analyze will return empty results until patterns are added.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	pa.ClearPatterns() // Remove all patterns
//	pa.AddPattern(customPattern) // Add only custom patterns
func (pa *PatternAnalyzer) ClearPatterns() {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	pa.patterns = make([]Pattern, 0)
}

// defaultPatterns returns the default set of performance patterns.
func defaultPatterns() []Pattern {
	return []Pattern{
		{
			Name: "frequent_rerender",
			Detect: func(m *ComponentMetrics) bool {
				return m.RenderCount > 1000 && m.AvgRenderTime < time.Millisecond
			},
			Severity:    SeverityMedium,
			Description: "Component re-renders very frequently with minimal work",
			Suggestion:  "Consider memoization or shouldComponentUpdate logic to reduce unnecessary re-renders",
		},
		{
			Name: "slow_render",
			Detect: func(m *ComponentMetrics) bool {
				return m.AvgRenderTime > 10*time.Millisecond
			},
			Severity:    SeverityHigh,
			Description: "Component render is slow",
			Suggestion:  "Profile the render function to identify hot spots. Consider caching expensive computations or optimizing the template",
		},
		{
			Name: "memory_hog",
			Detect: func(m *ComponentMetrics) bool {
				return m.MemoryUsage > 5*1024*1024 // 5MB
			},
			Severity:    SeverityHigh,
			Description: "Component uses excessive memory",
			Suggestion:  "Review memory allocations. Consider object pooling, sync.Pool for temporary objects, or check for memory leaks",
		},
		{
			Name: "render_spike",
			Detect: func(m *ComponentMetrics) bool {
				return m.MaxRenderTime > 100*time.Millisecond
			},
			Severity:    SeverityMedium,
			Description: "Component has occasional very slow renders",
			Suggestion:  "Investigate what causes render spikes. Check for expensive operations that only occur sometimes (e.g., initial data load)",
		},
		{
			Name: "inefficient_render",
			Detect: func(m *ComponentMetrics) bool {
				return m.RenderCount > 500 && m.AvgRenderTime > 5*time.Millisecond
			},
			Severity:    SeverityCritical,
			Description: "Component renders frequently with significant render time",
			Suggestion:  "This component is a major performance bottleneck. Implement memoization and optimize the render function",
		},
	}
}

// calculatePatternImpact converts severity to an impact score (0.0-1.0).
func calculatePatternImpact(severity Severity) float64 {
	switch severity {
	case SeverityCritical:
		return 1.0
	case SeverityHigh:
		return 0.75
	case SeverityMedium:
		return 0.5
	case SeverityLow:
		return 0.25
	default:
		return 0.0
	}
}
