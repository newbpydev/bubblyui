// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockComponent implements a minimal Component interface for testing.
type mockComponent struct {
	id           string
	name         string
	initCalled   int
	viewCalled   int
	updateCalled int
	viewDelay    time.Duration
	updateDelay  time.Duration
	mu           sync.Mutex
}

func newMockComponent(id, name string) *mockComponent {
	return &mockComponent{
		id:   id,
		name: name,
	}
}

func (m *mockComponent) Init() tea.Cmd {
	m.mu.Lock()
	m.initCalled++
	m.mu.Unlock()
	return nil
}

func (m *mockComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.mu.Lock()
	m.updateCalled++
	delay := m.updateDelay
	m.mu.Unlock()

	if delay > 0 {
		time.Sleep(delay)
	}
	return m, nil
}

func (m *mockComponent) View() string {
	m.mu.Lock()
	m.viewCalled++
	delay := m.viewDelay
	m.mu.Unlock()

	if delay > 0 {
		time.Sleep(delay)
	}
	return "mock view"
}

func (m *mockComponent) Name() string {
	return m.name
}

func (m *mockComponent) ID() string {
	return m.id
}

func (m *mockComponent) Props() interface{} {
	return nil
}

func (m *mockComponent) Emit(event string, data interface{}) {}

func (m *mockComponent) On(event string, handler func(interface{})) {}

func (m *mockComponent) KeyBindings() map[string][]KeyBinding {
	return nil
}

func (m *mockComponent) HelpText() string {
	return ""
}

func (m *mockComponent) IsInitialized() bool {
	return true
}

// --- Tests ---

func TestNewInstrumentor(t *testing.T) {
	tests := []struct {
		name     string
		profiler *Profiler
		wantNil  bool
	}{
		{
			name:     "with valid profiler",
			profiler: New(),
			wantNil:  false,
		},
		{
			name:     "with nil profiler",
			profiler: nil,
			wantNil:  false, // Should still create instrumentor with defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := NewInstrumentor(tt.profiler)
			if tt.wantNil {
				assert.Nil(t, inst)
			} else {
				assert.NotNil(t, inst)
				assert.NotNil(t, inst.componentTracker)
				assert.NotNil(t, inst.collector)
			}
		})
	}
}

func TestInstrumentor_EnableDisable(t *testing.T) {
	inst := NewInstrumentor(New())

	// Initially disabled
	assert.False(t, inst.IsEnabled())

	// Enable
	inst.Enable()
	assert.True(t, inst.IsEnabled())

	// Disable
	inst.Disable()
	assert.False(t, inst.IsEnabled())
}

func TestInstrumentor_InstrumentRender_NilComponent(t *testing.T) {
	inst := NewInstrumentor(New())
	inst.Enable()

	// Should not panic with nil component
	stop := inst.InstrumentRender(nil)
	assert.NotNil(t, stop)
	stop() // Should be safe to call
}

func TestInstrumentor_InstrumentRender_RecordsMetrics(t *testing.T) {
	inst := NewInstrumentor(New())
	inst.Enable()

	comp := newMockComponent("comp-1", "TestComponent")
	comp.viewDelay = 5 * time.Millisecond

	// Start timing
	stop := inst.InstrumentRender(comp)

	// Simulate render
	_ = comp.View()

	// Stop timing
	stop()

	// Check metrics were recorded
	metrics := inst.GetComponentTracker().GetMetrics("comp-1")
	require.NotNil(t, metrics)
	assert.Equal(t, int64(1), metrics.RenderCount)
	assert.GreaterOrEqual(t, metrics.TotalRenderTime, 5*time.Millisecond)
}

func TestInstrumentor_InstrumentUpdate_NilComponent(t *testing.T) {
	inst := NewInstrumentor(New())
	inst.Enable()

	// Should not panic with nil component
	stop := inst.InstrumentUpdate(nil)
	assert.NotNil(t, stop)
	stop() // Should be safe to call
}

func TestInstrumentor_InstrumentUpdate_RecordsMetrics(t *testing.T) {
	inst := NewInstrumentor(New())
	inst.Enable()

	comp := newMockComponent("comp-2", "TestComponent")
	comp.updateDelay = 5 * time.Millisecond

	// Start timing
	stop := inst.InstrumentUpdate(comp)

	// Simulate update
	_, _ = comp.Update(nil)

	// Stop timing
	stop()

	// Check timing was recorded in collector
	timings := inst.GetCollector().GetTimings()
	stats := timings.GetStats("update.comp-2")
	require.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.Count)
	assert.GreaterOrEqual(t, stats.Total, 5*time.Millisecond)
}

func TestInstrumentor_Disabled_MinimalOverhead(t *testing.T) {
	inst := NewInstrumentor(New())
	// Keep disabled

	comp := newMockComponent("comp-3", "TestComponent")

	// Measure overhead when disabled
	iterations := 10000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		stop := inst.InstrumentRender(comp)
		stop()
	}
	elapsed := time.Since(start)

	// Average should be very low (< 1μs per call when disabled)
	avgNs := elapsed.Nanoseconds() / int64(iterations)
	t.Logf("Disabled overhead: %d ns/op", avgNs)
	assert.Less(t, avgNs, int64(1000), "Overhead should be < 1μs when disabled")
}

func TestInstrumentor_NoBreakingChanges(t *testing.T) {
	inst := NewInstrumentor(New())
	inst.Enable()

	comp := newMockComponent("comp-4", "TestComponent")

	// Instrument and use component
	stop := inst.InstrumentRender(comp)
	result := comp.View()
	stop()

	// Component should still work normally
	assert.Equal(t, "mock view", result)
	assert.Equal(t, 1, comp.viewCalled)
}

func TestInstrumentor_ThreadSafe(t *testing.T) {
	inst := NewInstrumentor(New())
	inst.Enable()

	var wg sync.WaitGroup
	goroutines := 50
	iterations := 100

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			comp := newMockComponent(
				"comp-concurrent",
				"ConcurrentComponent",
			)

			for j := 0; j < iterations; j++ {
				stop := inst.InstrumentRender(comp)
				_ = comp.View()
				stop()

				stopUpdate := inst.InstrumentUpdate(comp)
				_, _ = comp.Update(nil)
				stopUpdate()
			}
		}()
	}

	wg.Wait()

	// Should have recorded all renders
	metrics := inst.GetComponentTracker().GetMetrics("comp-concurrent")
	require.NotNil(t, metrics)
	expectedRenders := int64(goroutines * iterations)
	assert.Equal(t, expectedRenders, metrics.RenderCount)
}

func TestInstrumentedComponent_View(t *testing.T) {
	inst := NewInstrumentor(New())
	inst.Enable()

	original := newMockComponent("comp-5", "OriginalComponent")
	original.viewDelay = 2 * time.Millisecond

	// Wrap the component
	wrapped := inst.InstrumentComponent(original)
	require.NotNil(t, wrapped)

	// Call View on wrapped component
	result := wrapped.View()

	// Should return original result
	assert.Equal(t, "mock view", result)

	// Should have recorded timing
	metrics := inst.GetComponentTracker().GetMetrics("comp-5")
	require.NotNil(t, metrics)
	assert.Equal(t, int64(1), metrics.RenderCount)
	assert.GreaterOrEqual(t, metrics.TotalRenderTime, 2*time.Millisecond)
}

func TestInstrumentedComponent_Update(t *testing.T) {
	inst := NewInstrumentor(New())
	inst.Enable()

	original := newMockComponent("comp-6", "OriginalComponent")
	original.updateDelay = 2 * time.Millisecond

	// Wrap the component
	wrapped := inst.InstrumentComponent(original)
	require.NotNil(t, wrapped)

	// Call Update on wrapped component
	model, cmd := wrapped.Update(nil)

	// Should return original model
	assert.NotNil(t, model)
	assert.Nil(t, cmd)

	// Should have recorded timing
	timings := inst.GetCollector().GetTimings()
	stats := timings.GetStats("update.comp-6")
	require.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.Count)
}

func TestInstrumentedComponent_Init(t *testing.T) {
	inst := NewInstrumentor(New())
	inst.Enable()

	original := newMockComponent("comp-7", "OriginalComponent")

	// Wrap the component
	wrapped := inst.InstrumentComponent(original)
	require.NotNil(t, wrapped)

	// Call Init on wrapped component
	cmd := wrapped.Init()

	// Should work correctly
	assert.Nil(t, cmd)
	assert.Equal(t, 1, original.initCalled)
}

func TestInstrumentedComponent_DelegatesMethods(t *testing.T) {
	inst := NewInstrumentor(New())
	inst.Enable()

	original := newMockComponent("comp-8", "OriginalComponent")

	// Wrap the component
	wrapped := inst.InstrumentComponent(original)

	// All methods should delegate to original
	assert.Equal(t, "comp-8", wrapped.ID())
	assert.Equal(t, "OriginalComponent", wrapped.Name())
	assert.Nil(t, wrapped.Props())
	assert.True(t, wrapped.IsInitialized())
}

func TestInstrumentor_InstrumentComponent_NilComponent(t *testing.T) {
	inst := NewInstrumentor(New())
	inst.Enable()

	// Should return nil for nil component
	wrapped := inst.InstrumentComponent(nil)
	assert.Nil(t, wrapped)
}

func TestInstrumentor_GetComponentTracker(t *testing.T) {
	inst := NewInstrumentor(New())
	tracker := inst.GetComponentTracker()
	assert.NotNil(t, tracker)
}

func TestInstrumentor_GetCollector(t *testing.T) {
	inst := NewInstrumentor(New())
	collector := inst.GetCollector()
	assert.NotNil(t, collector)
}

func TestInstrumentor_Reset(t *testing.T) {
	inst := NewInstrumentor(New())
	inst.Enable()

	comp := newMockComponent("comp-9", "TestComponent")

	// Record some metrics
	stop := inst.InstrumentRender(comp)
	_ = comp.View()
	stop()

	// Verify metrics exist
	metrics := inst.GetComponentTracker().GetMetrics("comp-9")
	require.NotNil(t, metrics)

	// Reset
	inst.Reset()

	// Metrics should be cleared
	metrics = inst.GetComponentTracker().GetMetrics("comp-9")
	assert.Nil(t, metrics)
}

func TestInstrumentedComponent_MultipleRenders(t *testing.T) {
	inst := NewInstrumentor(New())
	inst.Enable()

	original := newMockComponent("comp-10", "MultiRenderComponent")

	wrapped := inst.InstrumentComponent(original)

	// Multiple renders
	for i := 0; i < 10; i++ {
		_ = wrapped.View()
	}

	// Should have recorded all renders
	metrics := inst.GetComponentTracker().GetMetrics("comp-10")
	require.NotNil(t, metrics)
	assert.Equal(t, int64(10), metrics.RenderCount)
}

func TestInstrumentor_Disabled_NoMetrics(t *testing.T) {
	inst := NewInstrumentor(New())
	// Keep disabled

	comp := newMockComponent("comp-11", "TestComponent")

	// Try to record metrics while disabled
	stop := inst.InstrumentRender(comp)
	_ = comp.View()
	stop()

	// No metrics should be recorded
	metrics := inst.GetComponentTracker().GetMetrics("comp-11")
	assert.Nil(t, metrics)
}

func TestInstrumentedComponent_Disabled_NoMetrics(t *testing.T) {
	inst := NewInstrumentor(New())
	// Keep disabled

	original := newMockComponent("comp-12", "TestComponent")
	wrapped := inst.InstrumentComponent(original)

	// Use wrapped component while disabled
	_ = wrapped.View()

	// No metrics should be recorded
	metrics := inst.GetComponentTracker().GetMetrics("comp-12")
	assert.Nil(t, metrics)
}

// Benchmark tests for overhead measurement
func BenchmarkInstrumentor_Disabled(b *testing.B) {
	inst := NewInstrumentor(New())
	comp := newMockComponent("bench-1", "BenchComponent")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stop := inst.InstrumentRender(comp)
		stop()
	}
}

func BenchmarkInstrumentor_Enabled(b *testing.B) {
	inst := NewInstrumentor(New())
	inst.Enable()
	comp := newMockComponent("bench-2", "BenchComponent")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stop := inst.InstrumentRender(comp)
		stop()
	}
}

func BenchmarkInstrumentedComponent_View(b *testing.B) {
	inst := NewInstrumentor(New())
	inst.Enable()
	original := newMockComponent("bench-3", "BenchComponent")
	wrapped := inst.InstrumentComponent(original)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapped.View()
	}
}

// Test for overhead percentage
func TestInstrumentor_OverheadPercentage(t *testing.T) {
	inst := NewInstrumentor(New())
	inst.Enable()

	comp := newMockComponent("overhead-test", "OverheadComponent")
	// Use 5ms to simulate realistic component rendering time
	// Typical TUI components take 5-50ms to render, making this a reasonable baseline
	comp.viewDelay = 5 * time.Millisecond

	iterations := 100

	// Measure without instrumentation
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_ = comp.View()
	}
	baselineTime := time.Since(start)

	// Reset view count
	comp.viewCalled = 0

	// Measure with instrumentation
	start = time.Now()
	for i := 0; i < iterations; i++ {
		stop := inst.InstrumentRender(comp)
		_ = comp.View()
		stop()
	}
	instrumentedTime := time.Since(start)

	// Calculate overhead percentage
	overhead := float64(instrumentedTime-baselineTime) / float64(baselineTime) * 100
	t.Logf("Instrumentation overhead: %.2f%%", overhead)
	t.Logf("Baseline time: %v, Instrumented time: %v", baselineTime, instrumentedTime)

	// Overhead should be < 3% as per spec
	assert.Less(t, overhead, 3.0, "Overhead should be < 3%")
}
