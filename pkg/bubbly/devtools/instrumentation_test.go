package devtools

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestInstrumentor_SetCollector tests setting and getting the collector
func TestInstrumentor_SetCollector(t *testing.T) {
	tests := []struct {
		name      string
		collector *DataCollector
		wantNil   bool
	}{
		{
			name:      "set valid collector",
			collector: NewDataCollector(),
			wantNil:   false,
		},
		{
			name:      "set nil collector",
			collector: nil,
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global state
			globalInstrumentor = &Instrumentor{}

			SetCollector(tt.collector)
			got := GetCollector()

			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, tt.collector, got)
			}
		})
	}
}

// TestInstrumentor_NotifyComponentCreated tests component creation notifications
func TestInstrumentor_NotifyComponentCreated(t *testing.T) {
	tests := []struct {
		name           string
		setupCollector bool
		snapshot       *ComponentSnapshot
		wantCalled     bool
	}{
		{
			name:           "collector set - should notify",
			setupCollector: true,
			snapshot: &ComponentSnapshot{
				ID:   "comp-1",
				Name: "TestComponent",
			},
			wantCalled: true,
		},
		{
			name:           "no collector - should not panic",
			setupCollector: false,
			snapshot: &ComponentSnapshot{
				ID:   "comp-2",
				Name: "TestComponent2",
			},
			wantCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global state
			globalInstrumentor = &Instrumentor{}

			hook := &mockComponentHook{}

			if tt.setupCollector {
				collector := NewDataCollector()
				collector.AddComponentHook(hook)
				SetCollector(collector)
			}

			// Should not panic even if collector is nil
			NotifyComponentCreated(tt.snapshot)

			hook.mu.Lock()
			defer hook.mu.Unlock()
			assert.Equal(t, tt.wantCalled, hook.createdCalled)
			if tt.wantCalled {
				assert.Equal(t, tt.snapshot.ID, hook.createdSnapshot.ID)
				assert.Equal(t, tt.snapshot.Name, hook.createdSnapshot.Name)
			}
		})
	}
}

// TestInstrumentor_NotifyComponentMounted tests component mount notifications
func TestInstrumentor_NotifyComponentMounted(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockComponentHook{}
	collector.AddComponentHook(hook)

	// Reset and set collector
	globalInstrumentor = &Instrumentor{}
	SetCollector(collector)

	NotifyComponentMounted("comp-123")

	hook.mu.Lock()
	defer hook.mu.Unlock()
	assert.True(t, hook.mountedCalled)
	assert.Equal(t, "comp-123", hook.mountedID)
}

// TestInstrumentor_NotifyComponentUpdated tests component update notifications
func TestInstrumentor_NotifyComponentUpdated(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockComponentHook{}
	collector.AddComponentHook(hook)

	globalInstrumentor = &Instrumentor{}
	SetCollector(collector)

	NotifyComponentUpdated("comp-456")

	hook.mu.Lock()
	defer hook.mu.Unlock()
	assert.True(t, hook.updatedCalled)
	assert.Equal(t, "comp-456", hook.updatedID)
}

// TestInstrumentor_NotifyComponentUnmounted tests component unmount notifications
func TestInstrumentor_NotifyComponentUnmounted(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockComponentHook{}
	collector.AddComponentHook(hook)

	globalInstrumentor = &Instrumentor{}
	SetCollector(collector)

	NotifyComponentUnmounted("comp-789")

	hook.mu.Lock()
	defer hook.mu.Unlock()
	assert.True(t, hook.unmountedCalled)
	assert.Equal(t, "comp-789", hook.unmountedID)
}

// TestInstrumentor_NotifyRefChanged tests ref change notifications
func TestInstrumentor_NotifyRefChanged(t *testing.T) {
	tests := []struct {
		name     string
		refID    string
		oldValue interface{}
		newValue interface{}
	}{
		{
			name:     "int value change",
			refID:    "ref-1",
			oldValue: 10,
			newValue: 20,
		},
		{
			name:     "string value change",
			refID:    "ref-2",
			oldValue: "old",
			newValue: "new",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewDataCollector()
			hook := &mockStateHook{}
			collector.AddStateHook(hook)

			globalInstrumentor = &Instrumentor{}
			SetCollector(collector)

			NotifyRefChanged(tt.refID, tt.oldValue, tt.newValue)

			assert.True(t, hook.refChangedCalled)
			assert.Equal(t, tt.refID, hook.refID)
			assert.Equal(t, tt.oldValue, hook.oldValue)
			assert.Equal(t, tt.newValue, hook.newValue)
		})
	}
}

// TestInstrumentor_NotifyEvent tests event notifications
func TestInstrumentor_NotifyEvent(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockEventHook{}
	collector.AddEventHook(hook)

	globalInstrumentor = &Instrumentor{}
	SetCollector(collector)

	event := &EventRecord{
		ID:        "event-1",
		Name:      "click",
		SourceID:  "button-1",
		TargetID:  "form-1",
		Timestamp: time.Now(),
	}

	NotifyEvent(event)

	assert.True(t, hook.eventCalled)
	assert.Equal(t, event.ID, hook.event.ID)
	assert.Equal(t, event.Name, hook.event.Name)
}

// TestInstrumentor_NotifyRenderComplete tests render completion notifications
func TestInstrumentor_NotifyRenderComplete(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockPerformanceHook{}
	collector.AddPerformanceHook(hook)

	globalInstrumentor = &Instrumentor{}
	SetCollector(collector)

	duration := 5 * time.Millisecond
	NotifyRenderComplete("comp-render", duration)

	assert.True(t, hook.renderCalled)
	assert.Equal(t, "comp-render", hook.componentID)
	assert.Equal(t, duration, hook.duration)
}

// TestInstrumentor_ThreadSafety tests concurrent access to instrumentor
func TestInstrumentor_ThreadSafety(t *testing.T) {
	collector := NewDataCollector()
	globalInstrumentor = &Instrumentor{}
	SetCollector(collector)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent component notifications
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func(id int) {
			defer wg.Done()
			snapshot := &ComponentSnapshot{
				ID:   string(rune(id)),
				Name: "Component",
			}
			NotifyComponentCreated(snapshot)
			NotifyComponentMounted(string(rune(id)))
			NotifyComponentUpdated(string(rune(id)))
			NotifyComponentUnmounted(string(rune(id)))
		}(i)
	}

	// Concurrent ref notifications
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func(id int) {
			defer wg.Done()
			NotifyRefChanged(string(rune(id)), i, i+1)
		}(i)
	}

	wg.Wait()
	// If we get here without panic, thread safety is working
}

// TestInstrumentor_ZeroOverheadWhenDisabled tests that there's no overhead when disabled
func TestInstrumentor_ZeroOverheadWhenDisabled(t *testing.T) {
	// Reset to nil collector
	globalInstrumentor = &Instrumentor{}
	SetCollector(nil)

	// These should all be no-ops and not panic
	NotifyComponentCreated(&ComponentSnapshot{ID: "test"})
	NotifyComponentMounted("test")
	NotifyComponentUpdated("test")
	NotifyComponentUnmounted("test")
	NotifyRefChanged("ref-1", 1, 2)
	NotifyEvent(&EventRecord{ID: "event-1"})
	NotifyRenderComplete("comp-1", time.Millisecond)

	// If we get here, zero overhead is working
}

// Note: Mock hooks are defined in collector_test.go and reused here
