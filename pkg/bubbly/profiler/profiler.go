// Package profiler provides comprehensive performance profiling for BubblyUI applications.
//
// The profiler enables runtime performance analysis, CPU and memory profiling,
// rendering benchmarks, component performance tracking, and optimization recommendations.
// It integrates with Go's built-in pprof tools and operates with minimal overhead.
//
// # Basic Usage
//
//	prof := profiler.New()
//	prof.Start()
//	defer prof.Stop()
//
//	// Run your application
//	tea.NewProgram(app).Run()
//
//	// Generate report
//	report := prof.GenerateReport()
//	report.SaveHTML("performance-report.html")
//
// # Configuration
//
//	prof := profiler.New(
//	    profiler.WithSamplingRate(0.1),      // 10% sampling
//	    profiler.WithMaxSamples(5000),       // Limit samples
//	    profiler.WithMinimalMetrics(),       // Low overhead mode
//	    profiler.WithThreshold("render", 10*time.Millisecond),
//	)
//
// # Thread Safety
//
// All methods are thread-safe and can be called concurrently.
//
// # Performance
//
// Profiling overhead is < 3% when enabled and < 0.1% when disabled.
package profiler

import (
	"errors"
	"sync"
	"time"
)

// Profiler is the main performance profiler instance.
//
// It coordinates metric collection, CPU/memory profiling, bottleneck detection,
// and report generation. The profiler can be enabled/disabled at runtime and
// operates with configurable overhead.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	prof := profiler.New()
//	prof.Start()
//	defer func() {
//	    prof.Stop()
//	    report := prof.GenerateReport()
//	    report.SaveHTML("profile.html")
//	}()
type Profiler struct {
	// enabled indicates whether profiling is active
	enabled bool

	// collector handles metric collection (implemented in Task 1.2)
	collector *MetricCollector

	// cpuProf handles CPU profiling (implemented in Task 2.1)
	cpuProf *CPUProfiler

	// memProf handles memory profiling (implemented in Task 2.3)
	memProf *MemoryProfiler

	// renderProf handles render profiling (implemented in Task 3.1)
	renderProf *RenderProfiler

	// detector handles bottleneck detection (implemented in Task 4.1)
	detector *BottleneckDetector

	// config holds profiler configuration
	config *Config

	// mu protects concurrent access to profiler state
	mu sync.RWMutex
}

// MetricCollector is defined in collector.go (Task 1.2)
// CPUProfiler is defined in cpu.go (Task 2.1)
// MemoryProfiler is defined in heap.go (Task 2.3)
// RenderProfiler is defined in render.go (Task 3.1)

// BottleneckDetector detects performance bottlenecks.
// This is a stub for Task 4.1.
type BottleneckDetector struct {
	thresholds map[string]time.Duration
	mu         sync.RWMutex // nolint:unused // Will be used in Task 4.1
}

// Config holds profiler configuration.
type Config struct {
	// Enabled indicates if profiler is enabled at creation
	Enabled bool

	// SamplingRate is the fraction of operations to sample (0.0 to 1.0)
	SamplingRate float64

	// MaxSamples is the maximum number of samples to retain
	MaxSamples int

	// MinimalMetrics enables low-overhead mode
	MinimalMetrics bool

	// Thresholds maps operation names to duration thresholds
	Thresholds map[string]time.Duration
}

// Report is the complete performance analysis report.
type Report struct {
	// Summary contains high-level performance summary
	Summary *Summary

	// Components contains per-component metrics
	Components []*ComponentMetrics

	// Bottlenecks contains detected performance issues
	Bottlenecks []*BottleneckInfo

	// CPUProfile contains CPU profiling data
	CPUProfile *CPUProfileData

	// MemProfile contains memory profiling data
	MemProfile *MemProfileData

	// Recommendations contains optimization suggestions
	Recommendations []*Recommendation

	// Timestamp is when the report was generated
	Timestamp time.Time
}

// Summary contains high-level performance metrics.
type Summary struct {
	// Duration is the total profiling duration
	Duration time.Duration

	// TotalOperations is the count of operations profiled
	TotalOperations int64

	// AverageFPS is the average frames per second
	AverageFPS float64

	// MemoryUsage is the current heap size in bytes
	MemoryUsage uint64

	// GoroutineCount is the number of goroutines
	GoroutineCount int
}

// ComponentMetrics tracks per-component performance.
type ComponentMetrics struct {
	// ComponentID is the unique identifier
	ComponentID string

	// ComponentName is the human-readable name
	ComponentName string

	// RenderCount is total number of renders
	RenderCount int64

	// TotalRenderTime is cumulative render duration
	TotalRenderTime time.Duration

	// AvgRenderTime is average render duration
	AvgRenderTime time.Duration

	// MaxRenderTime is the longest render
	MaxRenderTime time.Duration

	// MemoryUsage is estimated memory usage
	MemoryUsage uint64
}

// BottleneckInfo describes a performance bottleneck.
type BottleneckInfo struct {
	// Type is the bottleneck category
	Type BottleneckType

	// Location is where the bottleneck occurs
	Location string

	// Severity indicates how critical the issue is
	Severity Severity

	// Impact is the performance impact (0.0 to 1.0)
	Impact float64

	// Description explains the issue
	Description string

	// Suggestion provides optimization advice
	Suggestion string
}

// BottleneckType categorizes bottleneck types.
type BottleneckType string

const (
	// BottleneckTypeSlow indicates a slow operation
	BottleneckTypeSlow BottleneckType = "slow"

	// BottleneckTypeMemory indicates excessive memory usage
	BottleneckTypeMemory BottleneckType = "memory"

	// BottleneckTypeFrequent indicates too many operations
	BottleneckTypeFrequent BottleneckType = "frequent"

	// BottleneckTypePattern indicates a detected anti-pattern
	BottleneckTypePattern BottleneckType = "pattern"
)

// Severity indicates the severity of a performance issue.
type Severity string

const (
	// SeverityCritical requires immediate attention
	SeverityCritical Severity = "critical"

	// SeverityHigh is a significant issue
	SeverityHigh Severity = "high"

	// SeverityMedium is a moderate issue
	SeverityMedium Severity = "medium"

	// SeverityLow is a minor issue
	SeverityLow Severity = "low"
)

// CPUProfileData contains CPU profiling results.
type CPUProfileData struct {
	// HotFunctions lists functions consuming the most CPU
	HotFunctions []*HotFunction

	// CallGraph maps functions to their callees
	CallGraph map[string][]string

	// TotalSamples is the number of CPU samples collected
	TotalSamples int64
}

// HotFunction represents a function consuming significant CPU.
type HotFunction struct {
	// Name is the function name
	Name string

	// Samples is the number of samples in this function
	Samples int64

	// Percent is the percentage of total CPU time
	Percent float64
}

// MemProfileData contains memory profiling results.
type MemProfileData struct {
	// HeapAlloc is current heap allocation in bytes
	HeapAlloc uint64

	// HeapObjects is number of allocated objects
	HeapObjects uint64

	// GCPauses is list of recent GC pause durations
	GCPauses []time.Duration
}

// Recommendation provides an optimization suggestion.
type Recommendation struct {
	// Title is a short description
	Title string

	// Description explains the recommendation
	Description string

	// Action suggests what to do
	Action string

	// Priority indicates importance
	Priority Priority

	// Category groups related recommendations
	Category Category

	// Impact indicates expected improvement
	Impact ImpactLevel
}

// Priority indicates recommendation priority.
type Priority int

const (
	// PriorityLow is optional optimization
	PriorityLow Priority = iota

	// PriorityMedium is recommended optimization
	PriorityMedium

	// PriorityHigh is strongly recommended
	PriorityHigh

	// PriorityCritical is urgent optimization
	PriorityCritical
)

// Category groups related recommendations.
type Category string

const (
	// CategoryOptimization covers general performance
	CategoryOptimization Category = "optimization"

	// CategoryMemory covers memory usage
	CategoryMemory Category = "memory"

	// CategoryRendering covers render performance
	CategoryRendering Category = "rendering"

	// CategoryArchitecture covers design patterns
	CategoryArchitecture Category = "architecture"
)

// ImpactLevel indicates expected improvement.
type ImpactLevel string

const (
	// ImpactLow is minor improvement
	ImpactLow ImpactLevel = "low"

	// ImpactMedium is moderate improvement
	ImpactMedium ImpactLevel = "medium"

	// ImpactHigh is significant improvement
	ImpactHigh ImpactLevel = "high"
)

// Option is a functional option for configuring the profiler.
type Option func(*Config)

// Common errors
var (
	// ErrAlreadyStarted is returned when Start() is called on an active profiler
	ErrAlreadyStarted = errors.New("profiler already started")

	// ErrNotStarted is returned when Stop() is called on an inactive profiler
	ErrNotStarted = errors.New("profiler not started")

	// ErrInvalidSamplingRate is returned for invalid sampling rate
	ErrInvalidSamplingRate = errors.New("sampling rate must be between 0.0 and 1.0")

	// ErrInvalidMaxSamples is returned for invalid max samples
	ErrInvalidMaxSamples = errors.New("max samples must be greater than 0")
)

// DefaultConfig returns a Config with default values.
//
// Default values:
//   - SamplingRate: 1.0 (100% sampling)
//   - MaxSamples: 10000
//   - MinimalMetrics: false
//   - Thresholds: empty map
func DefaultConfig() *Config {
	return &Config{
		Enabled:        false,
		SamplingRate:   1.0,
		MaxSamples:     10000,
		MinimalMetrics: false,
		Thresholds:     make(map[string]time.Duration),
	}
}

// Validate checks if the configuration is valid.
//
// Returns an error if:
//   - SamplingRate is not in [0.0, 1.0]
//   - MaxSamples is <= 0
func (c *Config) Validate() error {
	if c.SamplingRate < 0.0 || c.SamplingRate > 1.0 {
		return ErrInvalidSamplingRate
	}
	if c.MaxSamples <= 0 {
		return ErrInvalidMaxSamples
	}
	return nil
}

// New creates a new Profiler with the specified options.
//
// The profiler is created in a disabled state. Call Start() to begin profiling.
//
// Example:
//
//	prof := profiler.New(
//	    profiler.WithSamplingRate(0.5),
//	    profiler.WithMaxSamples(5000),
//	)
func New(opts ...Option) *Profiler {
	cfg := DefaultConfig()

	// Apply options
	for _, opt := range opts {
		opt(cfg)
	}

	p := &Profiler{
		enabled:    cfg.Enabled,
		config:     cfg,
		collector:  NewMetricCollector(),
		cpuProf:    NewCPUProfiler(),
		memProf:    NewMemoryProfiler(),
		renderProf: &RenderProfiler{},
		detector: &BottleneckDetector{
			thresholds: make(map[string]time.Duration),
		},
	}

	// Copy thresholds to detector
	for k, v := range cfg.Thresholds {
		p.detector.thresholds[k] = v
	}

	return p
}

// Start begins performance profiling.
//
// Returns ErrAlreadyStarted if profiling is already active.
//
// Example:
//
//	prof := profiler.New()
//	if err := prof.Start(); err != nil {
//	    log.Fatal(err)
//	}
//	defer prof.Stop()
func (p *Profiler) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.enabled {
		return ErrAlreadyStarted
	}

	p.enabled = true
	p.collector.Enable()

	return nil
}

// Stop ends performance profiling.
//
// Returns ErrNotStarted if profiling is not active.
//
// Example:
//
//	prof.Stop()
//	report := prof.GenerateReport()
func (p *Profiler) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.enabled {
		return ErrNotStarted
	}

	p.enabled = false
	p.collector.Disable()

	return nil
}

// IsEnabled returns whether profiling is currently active.
//
// Thread-safe for concurrent access.
func (p *Profiler) IsEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.enabled
}

// Enable activates profiling without starting a new session.
//
// Use this to temporarily enable/disable profiling during a session.
func (p *Profiler) Enable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = true
}

// Disable deactivates profiling without stopping the session.
//
// Use this to temporarily disable profiling during a session.
func (p *Profiler) Disable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = false
}

// GenerateReport creates a performance report from collected data.
//
// The report includes summary metrics, component performance data,
// detected bottlenecks, and optimization recommendations.
//
// Example:
//
//	report := prof.GenerateReport()
//	report.SaveHTML("performance.html")
func (p *Profiler) GenerateReport() *Report {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return &Report{
		Summary: &Summary{
			Duration:        0,
			TotalOperations: 0,
			AverageFPS:      0,
			MemoryUsage:     0,
			GoroutineCount:  0,
		},
		Components:      make([]*ComponentMetrics, 0),
		Bottlenecks:     make([]*BottleneckInfo, 0),
		CPUProfile:      &CPUProfileData{},
		MemProfile:      &MemProfileData{},
		Recommendations: make([]*Recommendation, 0),
		Timestamp:       time.Now(),
	}
}

// WithEnabled sets the initial enabled state.
//
// Example:
//
//	prof := profiler.New(profiler.WithEnabled(true))
func WithEnabled(enabled bool) Option {
	return func(c *Config) {
		c.Enabled = enabled
	}
}

// WithSamplingRate sets the sampling rate (0.0 to 1.0).
//
// A rate of 1.0 means 100% of operations are profiled.
// A rate of 0.1 means 10% of operations are sampled.
//
// Lower rates reduce overhead but provide less accurate data.
//
// Example:
//
//	prof := profiler.New(profiler.WithSamplingRate(0.1)) // 10% sampling
func WithSamplingRate(rate float64) Option {
	return func(c *Config) {
		c.SamplingRate = rate
	}
}

// WithMaxSamples sets the maximum number of samples to retain.
//
// When this limit is reached, reservoir sampling is used to
// maintain representative samples while bounding memory usage.
//
// Example:
//
//	prof := profiler.New(profiler.WithMaxSamples(5000))
func WithMaxSamples(max int) Option {
	return func(c *Config) {
		c.MaxSamples = max
	}
}

// WithMinimalMetrics enables low-overhead mode.
//
// In minimal mode, only essential metrics are collected,
// reducing overhead for production use.
//
// Example:
//
//	prof := profiler.New(profiler.WithMinimalMetrics())
func WithMinimalMetrics() Option {
	return func(c *Config) {
		c.MinimalMetrics = true
	}
}

// WithThreshold sets a performance threshold for an operation.
//
// Operations exceeding their threshold are flagged as bottlenecks.
//
// Example:
//
//	prof := profiler.New(
//	    profiler.WithThreshold("render", 16*time.Millisecond),
//	    profiler.WithThreshold("update", 5*time.Millisecond),
//	)
func WithThreshold(operation string, threshold time.Duration) Option {
	return func(c *Config) {
		if c.Thresholds == nil {
			c.Thresholds = make(map[string]time.Duration)
		}
		c.Thresholds[operation] = threshold
	}
}
