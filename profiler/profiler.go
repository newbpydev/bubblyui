// Package profiler provides comprehensive performance profiling for BubblyUI applications.
//
// The profiler enables runtime performance analysis, CPU and memory profiling,
// rendering benchmarks, component performance tracking, and optimization
// recommendations. It integrates with Go's built-in pprof tools and operates with
// minimal overhead (< 3% when enabled, < 0.1% when disabled).
//
// This package is an alias for github.com/newbpydev/bubblyui/pkg/bubbly/profiler,
// providing a cleaner import path for users.
//
// # Quick Start
//
//	import "github.com/newbpydev/bubblyui/profiler"
//
//	func main() {
//	    prof := profiler.New(profiler.WithEnabled(true))
//	    prof.Start()
//	    defer prof.Stop()
//
//	    // Run your BubblyUI application
//	    tea.NewProgram(app).Run()
//
//	    // Generate performance report
//	    report := prof.GenerateReport()
//	    exporter := profiler.NewExporter()
//	    exporter.ExportHTML(report, "performance-report.html")
//	}
//
// # Features
//
//   - CPU and memory profiling with pprof integration
//   - FPS and render timing tracking
//   - Component performance metrics
//   - Memory leak detection
//   - Bottleneck detection and recommendations
//   - Flame graph generation
//   - Timeline visualization
//   - HTTP handlers for remote profiling
package profiler

import (
	"html/template"
	"net/http"
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/profiler"
)

// =============================================================================
// Environment Variables
// =============================================================================

// Environment variable names for configuration.
const (
	EnvEnabled      = profiler.EnvEnabled
	EnvSamplingRate = profiler.EnvSamplingRate
	EnvMaxSamples   = profiler.EnvMaxSamples
)

// =============================================================================
// Default Values
// =============================================================================

// Default configuration values.
const (
	DefaultSamplingRate   = profiler.DefaultSamplingRate
	DefaultMaxSamples     = profiler.DefaultMaxSamples
	DefaultUpdateInterval = profiler.DefaultUpdateInterval
	DefaultFPSWindowSize  = profiler.DefaultFPSWindowSize
)

// Default dimensions.
const (
	DefaultFlameGraphWidth = profiler.DefaultFlameGraphWidth
	DefaultTimelineWidth   = profiler.DefaultTimelineWidth
)

// Default limits.
const (
	DefaultMaxCPUProfileDuration = profiler.DefaultMaxCPUProfileDuration
)

// =============================================================================
// Errors
// =============================================================================

// Common errors.
var (
	ErrAlreadyStarted   = profiler.ErrAlreadyStarted
	ErrCPUProfileActive = profiler.ErrCPUProfileActive
	ErrInvalidDuration  = profiler.ErrInvalidDuration
	ErrEmptyPanelName   = profiler.ErrEmptyPanelName
	ErrNilBenchmark     = profiler.ErrNilBenchmark
)

// =============================================================================
// Core Profiler
// =============================================================================

// Profiler is the main performance profiler.
type Profiler = profiler.Profiler

// New creates a new Profiler with the given options.
var New = profiler.New

// Option configures a Profiler.
type Option = profiler.Option

// WithEnabled enables or disables the profiler.
var WithEnabled = profiler.WithEnabled

// WithSamplingRate sets the sampling rate (0.0-1.0).
var WithSamplingRate = profiler.WithSamplingRate

// WithMaxSamples sets the maximum number of samples to store.
var WithMaxSamples = profiler.WithMaxSamples

// WithMinimalMetrics enables minimal metrics mode for production.
var WithMinimalMetrics = profiler.WithMinimalMetrics

// WithThreshold sets a threshold for an operation.
func WithThreshold(operation string, threshold time.Duration) Option {
	return profiler.WithThreshold(operation, threshold)
}

// ApplyOptions applies options to a configuration.
var ApplyOptions = profiler.ApplyOptions

// =============================================================================
// Configuration
// =============================================================================

// Config holds profiler configuration.
type Config = profiler.Config

// DefaultConfig returns the default configuration.
var DefaultConfig = profiler.DefaultConfig

// ConfigFromEnv loads configuration from environment variables.
var ConfigFromEnv = profiler.ConfigFromEnv

// =============================================================================
// CPU Profiling
// =============================================================================

// CPUProfiler manages CPU profiling.
type CPUProfiler = profiler.CPUProfiler

// NewCPUProfiler creates a new CPU profiler.
var NewCPUProfiler = profiler.NewCPUProfiler

// CPUProfileData contains CPU profile results.
type CPUProfileData = profiler.CPUProfileData

// HotFunction represents a function with high CPU usage.
type HotFunction = profiler.HotFunction

// =============================================================================
// Memory Profiling
// =============================================================================

// MemoryProfiler manages memory profiling.
type MemoryProfiler = profiler.MemoryProfiler

// NewMemoryProfiler creates a new memory profiler.
var NewMemoryProfiler = profiler.NewMemoryProfiler

// MemProfileData contains memory profile results.
type MemProfileData = profiler.MemProfileData

// MemoryTracker tracks memory allocations over time.
type MemoryTracker = profiler.MemoryTracker

// NewMemoryTracker creates a new memory tracker.
var NewMemoryTracker = profiler.NewMemoryTracker

// =============================================================================
// Leak Detection
// =============================================================================

// LeakDetector detects memory and goroutine leaks.
type LeakDetector = profiler.LeakDetector

// NewLeakDetector creates a new leak detector.
var NewLeakDetector = profiler.NewLeakDetector

// NewLeakDetectorWithThresholds creates a leak detector with custom thresholds.
var NewLeakDetectorWithThresholds = profiler.NewLeakDetectorWithThresholds

// LeakThresholds configures leak detection sensitivity.
type LeakThresholds = profiler.LeakThresholds

// DefaultLeakThresholds returns default leak detection thresholds.
var DefaultLeakThresholds = profiler.DefaultLeakThresholds

// LeakInfo describes a detected leak.
type LeakInfo = profiler.LeakInfo

// =============================================================================
// Render Profiling
// =============================================================================

// RenderProfiler tracks render performance and FPS.
type RenderProfiler = profiler.RenderProfiler

// NewRenderProfiler creates a new render profiler.
var NewRenderProfiler = profiler.NewRenderProfiler

// NewRenderProfilerWithConfig creates a render profiler with custom config.
var NewRenderProfilerWithConfig = profiler.NewRenderProfilerWithConfig

// RenderConfig configures render profiling.
type RenderConfig = profiler.RenderConfig

// DefaultRenderConfig returns the default render configuration.
var DefaultRenderConfig = profiler.DefaultRenderConfig

// FPSCalculator calculates frames per second.
type FPSCalculator = profiler.FPSCalculator

// NewFPSCalculator creates a new FPS calculator.
var NewFPSCalculator = profiler.NewFPSCalculator

// NewFPSCalculatorWithWindowSize creates an FPS calculator with custom window.
var NewFPSCalculatorWithWindowSize = profiler.NewFPSCalculatorWithWindowSize

// FrameInfo contains information about a rendered frame.
type FrameInfo = profiler.FrameInfo

// =============================================================================
// Component Tracking
// =============================================================================

// ComponentTracker tracks component performance metrics.
type ComponentTracker = profiler.ComponentTracker

// NewComponentTracker creates a new component tracker.
var NewComponentTracker = profiler.NewComponentTracker

// ComponentMetrics contains metrics for a component.
type ComponentMetrics = profiler.ComponentMetrics

// ComponentSortField specifies how to sort components.
type ComponentSortField = profiler.ComponentSortField

// Component interface for trackable components.
type Component = profiler.Component

// =============================================================================
// Timing and Metrics
// =============================================================================

// TimingTracker tracks operation timing statistics.
type TimingTracker = profiler.TimingTracker

// NewTimingTracker creates a new timing tracker.
var NewTimingTracker = profiler.NewTimingTracker

// NewTimingTrackerWithMaxSamples creates a timing tracker with custom max samples.
var NewTimingTrackerWithMaxSamples = profiler.NewTimingTrackerWithMaxSamples

// TimingStats contains timing statistics.
type TimingStats = profiler.TimingStats

// TimingSnapshot is a point-in-time timing capture.
type TimingSnapshot = profiler.TimingSnapshot

// MetricCollector collects various metrics.
type MetricCollector = profiler.MetricCollector

// NewMetricCollector creates a new metric collector.
var NewMetricCollector = profiler.NewMetricCollector

// MetricsSnapshot contains a snapshot of all metrics.
type MetricsSnapshot = profiler.MetricsSnapshot

// MetricsUpdateCallback is called when metrics are updated.
type MetricsUpdateCallback = profiler.MetricsUpdateCallback

// PerformanceMetrics contains overall performance data.
type PerformanceMetrics = profiler.PerformanceMetrics

// =============================================================================
// Bottleneck Detection
// =============================================================================

// BottleneckDetector identifies performance bottlenecks.
type BottleneckDetector = profiler.BottleneckDetector

// NewBottleneckDetector creates a new bottleneck detector.
var NewBottleneckDetector = profiler.NewBottleneckDetector

// NewBottleneckDetectorWithThresholds creates a detector with custom thresholds.
var NewBottleneckDetectorWithThresholds = profiler.NewBottleneckDetectorWithThresholds

// BottleneckThresholds configures bottleneck detection sensitivity.
type BottleneckThresholds = profiler.BottleneckThresholds

// DefaultBottleneckThresholds returns default bottleneck thresholds.
var DefaultBottleneckThresholds = profiler.DefaultBottleneckThresholds

// BottleneckInfo describes a detected bottleneck.
type BottleneckInfo = profiler.BottleneckInfo

// BottleneckType categorizes the bottleneck.
type BottleneckType = profiler.BottleneckType

// =============================================================================
// Recommendations
// =============================================================================

// RecommendationEngine generates optimization recommendations.
type RecommendationEngine = profiler.RecommendationEngine

// NewRecommendationEngine creates a new recommendation engine.
var NewRecommendationEngine = profiler.NewRecommendationEngine

// NewRecommendationEngineWithRules creates an engine with custom rules.
var NewRecommendationEngineWithRules = profiler.NewRecommendationEngineWithRules

// Recommendation describes an optimization suggestion.
type Recommendation = profiler.Recommendation

// RecommendationRule defines a rule for generating recommendations.
type RecommendationRule = profiler.RecommendationRule

// Priority levels for recommendations.
type Priority = profiler.Priority

// Category categorizes recommendations.
type Category = profiler.Category

// =============================================================================
// Threshold Monitoring
// =============================================================================

// ThresholdMonitor monitors for threshold violations.
type ThresholdMonitor = profiler.ThresholdMonitor

// NewThresholdMonitor creates a new threshold monitor.
var NewThresholdMonitor = profiler.NewThresholdMonitor

// NewThresholdMonitorWithConfig creates a monitor with custom config.
var NewThresholdMonitorWithConfig = profiler.NewThresholdMonitorWithConfig

// ThresholdConfig configures threshold monitoring.
type ThresholdConfig = profiler.ThresholdConfig

// Alert represents a threshold violation alert.
type Alert = profiler.Alert

// AlertHandler handles threshold violation alerts.
type AlertHandler = profiler.AlertHandler

// =============================================================================
// Pattern Analysis
// =============================================================================

// PatternAnalyzer analyzes performance patterns.
type PatternAnalyzer = profiler.PatternAnalyzer

// NewPatternAnalyzer creates a new pattern analyzer.
var NewPatternAnalyzer = profiler.NewPatternAnalyzer

// NewPatternAnalyzerWithPatterns creates an analyzer with custom patterns.
var NewPatternAnalyzerWithPatterns = profiler.NewPatternAnalyzerWithPatterns

// Pattern defines a performance pattern to detect.
type Pattern = profiler.Pattern

// =============================================================================
// Stack Analysis
// =============================================================================

// StackAnalyzer analyzes call stacks.
type StackAnalyzer = profiler.StackAnalyzer

// NewStackAnalyzer creates a new stack analyzer.
var NewStackAnalyzer = profiler.NewStackAnalyzer

// CallNode represents a node in a call tree.
type CallNode = profiler.CallNode

// =============================================================================
// Visualization - Flame Graphs
// =============================================================================

// FlameGraphGenerator generates flame graph visualizations.
type FlameGraphGenerator = profiler.FlameGraphGenerator

// NewFlameGraphGenerator creates a new flame graph generator.
var NewFlameGraphGenerator = profiler.NewFlameGraphGenerator

// NewFlameGraphGeneratorWithDimensions creates a generator with custom dimensions.
var NewFlameGraphGeneratorWithDimensions = profiler.NewFlameGraphGeneratorWithDimensions

// =============================================================================
// Visualization - Timeline
// =============================================================================

// TimelineGenerator generates timeline visualizations.
type TimelineGenerator = profiler.TimelineGenerator

// NewTimelineGenerator creates a new timeline generator.
var NewTimelineGenerator = profiler.NewTimelineGenerator

// NewTimelineGeneratorWithDimensions creates a generator with custom dimensions.
var NewTimelineGeneratorWithDimensions = profiler.NewTimelineGeneratorWithDimensions

// TimelineData contains data for timeline rendering.
type TimelineData = profiler.TimelineData

// TimedEvent represents an event on the timeline.
type TimedEvent = profiler.TimedEvent

// EventType categorizes timeline events.
type EventType = profiler.EventType

// =============================================================================
// Reports and Export
// =============================================================================

// Report contains a complete performance report.
type Report = profiler.Report

// ReportGenerator generates performance reports.
type ReportGenerator = profiler.ReportGenerator

// NewReportGenerator creates a new report generator.
var NewReportGenerator = profiler.NewReportGenerator

// NewReportGeneratorWithTemplate creates a generator with a custom template.
func NewReportGeneratorWithTemplate(tmpl *template.Template) *ReportGenerator {
	return profiler.NewReportGeneratorWithTemplate(tmpl)
}

// Summary contains a summary of profiling data.
type Summary = profiler.Summary

// Exporter exports profiling data to various formats.
type Exporter = profiler.Exporter

// NewExporter creates a new exporter.
var NewExporter = profiler.NewExporter

// ExportFormat specifies the export format.
type ExportFormat = profiler.ExportFormat

// =============================================================================
// Data Aggregation
// =============================================================================

// DataAggregator aggregates profiling data.
type DataAggregator = profiler.DataAggregator

// NewDataAggregator creates a new data aggregator.
var NewDataAggregator = profiler.NewDataAggregator

// AggregatedData contains aggregated profiling data.
type AggregatedData = profiler.AggregatedData

// AggregatedTiming contains aggregated timing data.
type AggregatedTiming = profiler.AggregatedTiming

// AggregatedCounter contains aggregated counter data.
type AggregatedCounter = profiler.AggregatedCounter

// AggregatedAllocation contains aggregated allocation data.
type AggregatedAllocation = profiler.AggregatedAllocation

// =============================================================================
// Baseline Comparison
// =============================================================================

// Baseline contains baseline performance data for comparison.
type Baseline = profiler.Baseline

// LoadBaseline loads a baseline from a file.
var LoadBaseline = profiler.LoadBaseline

// RegressionInfo describes a performance regression.
type RegressionInfo = profiler.RegressionInfo

// Severity indicates the severity of an issue.
type Severity = profiler.Severity

// ImpactLevel indicates the impact level.
type ImpactLevel = profiler.ImpactLevel

// =============================================================================
// Benchmark Integration
// =============================================================================

// BenchmarkProfiler integrates with Go's testing.B.
type BenchmarkProfiler = profiler.BenchmarkProfiler

// NewBenchmarkProfiler creates a benchmark profiler.
func NewBenchmarkProfiler(b *testing.B) *BenchmarkProfiler {
	return profiler.NewBenchmarkProfiler(b)
}

// BenchmarkStats contains benchmark statistics.
type BenchmarkStats = profiler.BenchmarkStats

// =============================================================================
// HTTP Handlers
// =============================================================================

// HTTPHandler provides HTTP endpoints for profiling.
type HTTPHandler = profiler.HTTPHandler

// NewHTTPHandler creates a new HTTP handler for the profiler.
var NewHTTPHandler = profiler.NewHTTPHandler

// RegisterHandlers registers profiler handlers with an HTTP mux.
func RegisterHandlers(mux *http.ServeMux, prof *Profiler) {
	profiler.RegisterHandlers(mux, prof)
}

// ServeCPUProfile is an HTTP handler for CPU profiles.
var ServeCPUProfile = profiler.ServeCPUProfile

// ServeHeapProfile is an HTTP handler for heap profiles.
var ServeHeapProfile = profiler.ServeHeapProfile

// =============================================================================
// DevTools Integration
// =============================================================================

// DevToolsIntegration provides DevTools integration for the profiler.
type DevToolsIntegration = profiler.DevToolsIntegration

// NewDevToolsIntegration creates a new DevTools integration.
var NewDevToolsIntegration = profiler.NewDevToolsIntegration

// =============================================================================
// Instrumentation
// =============================================================================

// Instrumentor instruments code for profiling.
type Instrumentor = profiler.Instrumentor

// NewInstrumentor creates a new instrumentor.
var NewInstrumentor = profiler.NewInstrumentor

// KeyBinding defines a key binding for profiler controls.
type KeyBinding = profiler.KeyBinding

// ProfileData contains raw profile data.
type ProfileData = profiler.ProfileData

// =============================================================================
// Counter Tracking
// =============================================================================

// CounterTracker tracks counter metrics.
type CounterTracker = profiler.CounterTracker

// CounterStats contains counter statistics.
type CounterStats = profiler.CounterStats

// AllocationStats contains allocation statistics.
type AllocationStats = profiler.AllocationStats

// =============================================================================
// Hook Integration
// =============================================================================

// HookAdapter implements bubbly.FrameworkHook to collect profiling data.
// It tracks component render times and other metrics via the framework hook system.
type HookAdapter = profiler.HookAdapter

// NewHookAdapter creates a new profiler hook adapter.
// Use this to integrate the profiler with the framework's hook system.
//
// Example:
//
//	prof := profiler.New(profiler.WithEnabled(true))
//	hookAdapter := profiler.NewHookAdapter(prof)
//	prof.SetHookAdapter(hookAdapter)
var NewHookAdapter = profiler.NewHookAdapter

// CompositeHook multiplexes framework events to multiple hook implementations.
// This allows both DevTools and Profiler to receive events simultaneously.
//
// Example:
//
//	// Get existing DevTools hook
//	devtoolsHook := bubbly.GetRegisteredHook()
//
//	// Create profiler hook
//	profilerHook := profiler.NewHookAdapter(prof)
//
//	// Combine them
//	composite := profiler.NewCompositeHook(devtoolsHook, profilerHook)
//	bubbly.RegisterHook(composite)
type CompositeHook = profiler.CompositeHook

// NewCompositeHook creates a new composite hook that forwards to multiple hooks.
var NewCompositeHook = profiler.NewCompositeHook
