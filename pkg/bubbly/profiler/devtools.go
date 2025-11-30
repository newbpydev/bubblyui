// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// DefaultUpdateInterval is the default interval for sending metrics updates.
const DefaultUpdateInterval = 100 * time.Millisecond

// Common errors for DevToolsIntegration.
var (
	// ErrEmptyPanelName is returned when registering a panel with empty name.
	ErrEmptyPanelName = errors.New("panel name cannot be empty")

	// ErrNegativeInterval is returned when setting a negative update interval.
	ErrNegativeInterval = errors.New("update interval cannot be negative")
)

// MetricsSnapshot captures a point-in-time view of profiler metrics.
//
// This snapshot is sent to dev tools for visualization and is designed
// to be lightweight and thread-safe for concurrent access.
//
// Thread Safety:
//
//	Snapshots are immutable after creation and safe to share across goroutines.
type MetricsSnapshot struct {
	// Timings contains timing statistics for operations
	Timings map[string]*TimingSnapshot

	// Components contains per-component metrics
	Components []*ComponentMetrics

	// FPS is the current frames per second
	FPS float64

	// DroppedFrames is the percentage of dropped frames
	DroppedFrames float64

	// MemoryUsage is the current heap allocation in bytes
	MemoryUsage uint64

	// GoroutineCount is the number of active goroutines
	GoroutineCount int

	// Timestamp is when this snapshot was created
	Timestamp time.Time
}

// TimingSnapshot captures timing statistics for an operation.
//
// Thread Safety:
//
//	Snapshots are immutable after creation and safe to share across goroutines.
type TimingSnapshot struct {
	// Name is the operation name
	Name string

	// Count is the number of times this operation was recorded
	Count int64

	// Total is the cumulative duration
	Total time.Duration

	// Min is the minimum duration
	Min time.Duration

	// Max is the maximum duration
	Max time.Duration

	// Mean is the average duration
	Mean time.Duration

	// P50 is the 50th percentile (median)
	P50 time.Duration

	// P95 is the 95th percentile
	P95 time.Duration

	// P99 is the 99th percentile
	P99 time.Duration
}

// MetricsUpdateCallback is called when new metrics are available.
type MetricsUpdateCallback func(snapshot *MetricsSnapshot)

// DevToolsIntegration provides integration between the profiler and dev tools.
//
// It bridges the profiler's metrics collection with the dev tools visualization
// system, enabling real-time performance monitoring in the dev tools UI.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	prof := profiler.New()
//	dti := NewDevToolsIntegration(prof)
//	dti.Enable()
//
//	// Register a panel in dev tools
//	dti.RegisterPanel("Performance")
//
//	// Send metrics periodically
//	dti.SendMetrics()
type DevToolsIntegration struct {
	// profiler is the parent profiler instance
	profiler *Profiler

	// enabled indicates whether integration is active
	enabled atomic.Bool

	// metricsBuffer stores recent metrics snapshots
	metricsBuffer []*MetricsSnapshot

	// panels stores registered panel names
	panels map[string]bool

	// callbacks stores registered update callbacks
	callbacks []MetricsUpdateCallback

	// updateInterval is the interval between metric updates
	updateInterval time.Duration

	// mu protects concurrent access to internal state
	mu sync.RWMutex
}

// NewDevToolsIntegration creates a new dev tools integration.
//
// If profiler is nil, a new profiler with default settings is created.
// The integration starts in a disabled state; call Enable() to begin.
//
// Example:
//
//	dti := NewDevToolsIntegration(profiler)
//	dti.Enable()
func NewDevToolsIntegration(profiler *Profiler) *DevToolsIntegration {
	if profiler == nil {
		profiler = New()
	}

	return &DevToolsIntegration{
		profiler:       profiler,
		metricsBuffer:  make([]*MetricsSnapshot, 0),
		panels:         make(map[string]bool),
		callbacks:      make([]MetricsUpdateCallback, 0),
		updateInterval: DefaultUpdateInterval,
	}
}

// Enable activates the dev tools integration.
//
// When enabled, SendMetrics() will collect and buffer metrics.
// When disabled, SendMetrics() is a no-op with minimal overhead.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (dti *DevToolsIntegration) Enable() {
	dti.enabled.Store(true)
}

// Disable deactivates the dev tools integration.
//
// When disabled, SendMetrics() returns immediately without collecting data.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (dti *DevToolsIntegration) Disable() {
	dti.enabled.Store(false)
}

// IsEnabled returns whether the integration is currently active.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (dti *DevToolsIntegration) IsEnabled() bool {
	return dti.enabled.Load()
}

// SendMetrics collects current profiler metrics and sends them to dev tools.
//
// If the integration is disabled, this is a no-op with minimal overhead.
// Metrics are buffered and callbacks are notified.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	dti.Enable()
//	dti.SendMetrics() // Collects and buffers current metrics
func (dti *DevToolsIntegration) SendMetrics() {
	// Fast path when disabled
	if !dti.enabled.Load() {
		return
	}

	// Collect metrics snapshot
	snapshot := dti.collectSnapshot()
	if snapshot == nil {
		return
	}

	// Buffer the snapshot
	dti.mu.Lock()
	dti.metricsBuffer = append(dti.metricsBuffer, snapshot)
	// Keep buffer bounded (last 1000 snapshots)
	if len(dti.metricsBuffer) > 1000 {
		dti.metricsBuffer = dti.metricsBuffer[1:]
	}
	callbacks := make([]MetricsUpdateCallback, len(dti.callbacks))
	copy(callbacks, dti.callbacks)
	dti.mu.Unlock()

	// Notify callbacks (outside lock to prevent deadlock)
	for _, cb := range callbacks {
		if cb != nil {
			cb(snapshot)
		}
	}
}

// collectSnapshot creates a metrics snapshot from the profiler.
func (dti *DevToolsIntegration) collectSnapshot() *MetricsSnapshot {
	dti.mu.RLock()
	profiler := dti.profiler
	dti.mu.RUnlock()

	if profiler == nil {
		return nil
	}

	snapshot := &MetricsSnapshot{
		Timings:        make(map[string]*TimingSnapshot),
		Components:     make([]*ComponentMetrics, 0),
		Timestamp:      time.Now(),
		GoroutineCount: runtime.NumGoroutine(),
	}

	// Collect timing stats
	if profiler.collector != nil {
		timings := profiler.collector.GetTimings()
		if timings != nil {
			for _, name := range timings.GetOperationNames() {
				stats := timings.GetStats(name)
				if stats != nil {
					snapshot.Timings[name] = &TimingSnapshot{
						Name:  name,
						Count: stats.Count,
						Total: stats.Total,
						Min:   stats.Min,
						Max:   stats.Max,
						Mean:  stats.Mean,
						P50:   stats.P50,
						P95:   stats.P95,
						P99:   stats.P99,
					}
				}
			}
		}
	}

	// Collect render stats
	if profiler.renderProf != nil {
		snapshot.FPS = profiler.renderProf.GetFPS()
		snapshot.DroppedFrames = profiler.renderProf.GetDroppedFramePercent()
	}

	// Collect memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	snapshot.MemoryUsage = memStats.HeapAlloc

	return snapshot
}

// RegisterPanel registers a panel with the dev tools.
//
// The panel name must be non-empty. Duplicate registrations are allowed
// and will update the existing panel.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	err := dti.RegisterPanel("Performance")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (dti *DevToolsIntegration) RegisterPanel(name string) error {
	if name == "" {
		return ErrEmptyPanelName
	}

	dti.mu.Lock()
	defer dti.mu.Unlock()

	dti.panels[name] = true
	return nil
}

// UnregisterPanel removes a panel from dev tools.
//
// If the panel doesn't exist, this is a no-op.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (dti *DevToolsIntegration) UnregisterPanel(name string) {
	dti.mu.Lock()
	defer dti.mu.Unlock()

	delete(dti.panels, name)
}

// PanelExists checks if a panel is registered.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (dti *DevToolsIntegration) PanelExists(name string) bool {
	dti.mu.RLock()
	defer dti.mu.RUnlock()

	return dti.panels[name]
}

// GetPanelCount returns the number of registered panels.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (dti *DevToolsIntegration) GetPanelCount() int {
	dti.mu.RLock()
	defer dti.mu.RUnlock()

	return len(dti.panels)
}

// GetPanelNames returns a list of all registered panel names.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (dti *DevToolsIntegration) GetPanelNames() []string {
	dti.mu.RLock()
	defer dti.mu.RUnlock()

	names := make([]string, 0, len(dti.panels))
	for name := range dti.panels {
		names = append(names, name)
	}
	return names
}

// GetMetricsSnapshot returns the most recent metrics snapshot.
//
// Returns nil if no metrics have been collected yet.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (dti *DevToolsIntegration) GetMetricsSnapshot() *MetricsSnapshot {
	dti.mu.RLock()
	defer dti.mu.RUnlock()

	if len(dti.metricsBuffer) == 0 {
		return nil
	}

	return dti.metricsBuffer[len(dti.metricsBuffer)-1]
}

// GetMetricsCount returns the number of buffered metrics snapshots.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (dti *DevToolsIntegration) GetMetricsCount() int {
	dti.mu.RLock()
	defer dti.mu.RUnlock()

	return len(dti.metricsBuffer)
}

// ClearMetrics clears all buffered metrics.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (dti *DevToolsIntegration) ClearMetrics() {
	dti.mu.Lock()
	defer dti.mu.Unlock()

	dti.metricsBuffer = make([]*MetricsSnapshot, 0)
}

// SetUpdateInterval sets the interval between metric updates.
//
// If interval is 0, the default interval is used.
// Returns ErrNegativeInterval if interval is negative.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (dti *DevToolsIntegration) SetUpdateInterval(interval time.Duration) error {
	if interval < 0 {
		return ErrNegativeInterval
	}

	dti.mu.Lock()
	defer dti.mu.Unlock()

	if interval == 0 {
		dti.updateInterval = DefaultUpdateInterval
	} else {
		dti.updateInterval = interval
	}

	return nil
}

// GetUpdateInterval returns the current update interval.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (dti *DevToolsIntegration) GetUpdateInterval() time.Duration {
	dti.mu.RLock()
	defer dti.mu.RUnlock()

	return dti.updateInterval
}

// OnMetricsUpdate registers a callback for metrics updates.
//
// The callback is called each time SendMetrics() is called with new data.
// Multiple callbacks can be registered and will be called in order.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	dti.OnMetricsUpdate(func(snapshot *MetricsSnapshot) {
//	    fmt.Printf("FPS: %.1f\n", snapshot.FPS)
//	})
func (dti *DevToolsIntegration) OnMetricsUpdate(callback MetricsUpdateCallback) {
	if callback == nil {
		return
	}

	dti.mu.Lock()
	defer dti.mu.Unlock()

	dti.callbacks = append(dti.callbacks, callback)
}

// GetProfiler returns the underlying profiler instance.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (dti *DevToolsIntegration) GetProfiler() *Profiler {
	dti.mu.RLock()
	defer dti.mu.RUnlock()

	return dti.profiler
}

// Reset clears all state and returns to initial configuration.
//
// This clears metrics buffer, panels, and callbacks but preserves
// the profiler reference and enabled state.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (dti *DevToolsIntegration) Reset() {
	dti.mu.Lock()
	defer dti.mu.Unlock()

	dti.metricsBuffer = make([]*MetricsSnapshot, 0)
	dti.panels = make(map[string]bool)
	dti.callbacks = make([]MetricsUpdateCallback, 0)
	dti.updateInterval = DefaultUpdateInterval
}
