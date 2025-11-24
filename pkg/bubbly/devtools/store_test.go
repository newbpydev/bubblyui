package devtools

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStateHistory_Record tests recording state changes
func TestStateHistory_Record(t *testing.T) {
	tests := []struct {
		name      string
		maxSize   int
		changes   []StateChange
		wantCount int
	}{
		{
			name:    "single change",
			maxSize: 10,
			changes: []StateChange{
				{RefID: "ref-1", RefName: "counter", OldValue: 0, NewValue: 1},
			},
			wantCount: 1,
		},
		{
			name:    "multiple changes",
			maxSize: 10,
			changes: []StateChange{
				{RefID: "ref-1", RefName: "counter", OldValue: 0, NewValue: 1},
				{RefID: "ref-1", RefName: "counter", OldValue: 1, NewValue: 2},
				{RefID: "ref-2", RefName: "name", OldValue: "Alice", NewValue: "Bob"},
			},
			wantCount: 3,
		},
		{
			name:    "exceeds max size",
			maxSize: 2,
			changes: []StateChange{
				{RefID: "ref-1", RefName: "counter", OldValue: 0, NewValue: 1},
				{RefID: "ref-1", RefName: "counter", OldValue: 1, NewValue: 2},
				{RefID: "ref-1", RefName: "counter", OldValue: 2, NewValue: 3},
			},
			wantCount: 2, // Only last 2 kept
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := NewStateHistory(tt.maxSize)

			for _, change := range tt.changes {
				history.Record(change)
			}

			all := history.GetAll()
			assert.Equal(t, tt.wantCount, len(all))
		})
	}
}

// TestStateHistory_GetHistory tests retrieving history for a specific ref
func TestStateHistory_GetHistory(t *testing.T) {
	history := NewStateHistory(100)

	// Record changes for multiple refs
	history.Record(StateChange{RefID: "ref-1", RefName: "counter", OldValue: 0, NewValue: 1})
	history.Record(StateChange{RefID: "ref-2", RefName: "name", OldValue: "Alice", NewValue: "Bob"})
	history.Record(StateChange{RefID: "ref-1", RefName: "counter", OldValue: 1, NewValue: 2})

	tests := []struct {
		name      string
		refID     string
		wantCount int
	}{
		{
			name:      "ref with multiple changes",
			refID:     "ref-1",
			wantCount: 2,
		},
		{
			name:      "ref with single change",
			refID:     "ref-2",
			wantCount: 1,
		},
		{
			name:      "non-existent ref",
			refID:     "ref-999",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := history.GetHistory(tt.refID)
			assert.Equal(t, tt.wantCount, len(changes))

			// Verify all changes are for the correct ref
			for _, change := range changes {
				assert.Equal(t, tt.refID, change.RefID)
			}
		})
	}
}

// TestStateHistory_Clear tests clearing the history
func TestStateHistory_Clear(t *testing.T) {
	history := NewStateHistory(100)

	// Add some changes
	history.Record(StateChange{RefID: "ref-1", OldValue: 0, NewValue: 1})
	history.Record(StateChange{RefID: "ref-2", OldValue: "a", NewValue: "b"})

	require.Equal(t, 2, len(history.GetAll()))

	// Clear
	history.Clear()

	assert.Equal(t, 0, len(history.GetAll()))
}

// TestStateHistory_Concurrent tests concurrent access to state history
func TestStateHistory_Concurrent(t *testing.T) {
	history := NewStateHistory(1000)
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				history.Record(StateChange{
					RefID:    "ref-1",
					OldValue: j,
					NewValue: j + 1,
				})
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = history.GetHistory("ref-1")
				_ = history.GetAll()
			}
		}()
	}

	wg.Wait()

	// Should have exactly 1000 changes (maxSize)
	all := history.GetAll()
	assert.Equal(t, 1000, len(all))
}

// TestEventLog_Append tests appending events
func TestEventLog_Append(t *testing.T) {
	tests := []struct {
		name      string
		maxSize   int
		events    []EventRecord
		wantCount int
	}{
		{
			name:    "single event",
			maxSize: 10,
			events: []EventRecord{
				{ID: "evt-1", Name: "click", SourceID: "btn-1"},
			},
			wantCount: 1,
		},
		{
			name:    "multiple events",
			maxSize: 10,
			events: []EventRecord{
				{ID: "evt-1", Name: "click", SourceID: "btn-1"},
				{ID: "evt-2", Name: "submit", SourceID: "form-1"},
				{ID: "evt-3", Name: "change", SourceID: "input-1"},
			},
			wantCount: 3,
		},
		{
			name:    "exceeds max size",
			maxSize: 2,
			events: []EventRecord{
				{ID: "evt-1", Name: "click"},
				{ID: "evt-2", Name: "submit"},
				{ID: "evt-3", Name: "change"},
			},
			wantCount: 2, // Only last 2 kept
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := NewEventLog(tt.maxSize)

			for _, event := range tt.events {
				log.Append(event)
			}

			assert.Equal(t, tt.wantCount, log.Len())
		})
	}
}

// TestEventLog_GetRecent tests retrieving recent events
func TestEventLog_GetRecent(t *testing.T) {
	log := NewEventLog(100)

	// Add 10 events
	for i := 0; i < 10; i++ {
		log.Append(EventRecord{
			ID:   string(rune('a' + i)),
			Name: "event",
		})
	}

	tests := []struct {
		name      string
		n         int
		wantCount int
	}{
		{
			name:      "get 5 recent",
			n:         5,
			wantCount: 5,
		},
		{
			name:      "get all",
			n:         10,
			wantCount: 10,
		},
		{
			name:      "request more than available",
			n:         20,
			wantCount: 10,
		},
		{
			name:      "get zero",
			n:         0,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recent := log.GetRecent(tt.n)
			assert.Equal(t, tt.wantCount, len(recent))
		})
	}
}

// TestEventLog_Clear tests clearing the event log
func TestEventLog_Clear(t *testing.T) {
	log := NewEventLog(100)

	// Add some events
	log.Append(EventRecord{ID: "evt-1", Name: "click"})
	log.Append(EventRecord{ID: "evt-2", Name: "submit"})

	require.Equal(t, 2, log.Len())

	// Clear
	log.Clear()

	assert.Equal(t, 0, log.Len())
}

// TestEventLog_Concurrent tests concurrent access to event log
func TestEventLog_Concurrent(t *testing.T) {
	log := NewEventLog(1000)
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				log.Append(EventRecord{
					ID:   "evt",
					Name: "event",
				})
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = log.GetRecent(10)
				_ = log.Len()
			}
		}()
	}

	wg.Wait()

	// Should have exactly 1000 events (maxSize)
	assert.Equal(t, 1000, log.Len())
}

// TestPerformanceData_RecordRender tests recording render performance
func TestPerformanceData_RecordRender(t *testing.T) {
	perf := NewPerformanceData()

	// Record first render
	perf.RecordRender("comp-1", "Counter", 10*time.Millisecond)

	comp := perf.GetComponent("comp-1")
	require.NotNil(t, comp)
	assert.Equal(t, "comp-1", comp.ComponentID)
	assert.Equal(t, "Counter", comp.ComponentName)
	assert.Equal(t, int64(1), comp.RenderCount)
	assert.Equal(t, 10*time.Millisecond, comp.AvgRenderTime)
	assert.Equal(t, 10*time.Millisecond, comp.MaxRenderTime)
	assert.Equal(t, 10*time.Millisecond, comp.MinRenderTime)

	// Record second render (slower)
	perf.RecordRender("comp-1", "Counter", 20*time.Millisecond)

	comp = perf.GetComponent("comp-1")
	require.NotNil(t, comp)
	assert.Equal(t, int64(2), comp.RenderCount)
	assert.Equal(t, 15*time.Millisecond, comp.AvgRenderTime) // (10+20)/2
	assert.Equal(t, 20*time.Millisecond, comp.MaxRenderTime)
	assert.Equal(t, 10*time.Millisecond, comp.MinRenderTime)

	// Record third render (faster)
	perf.RecordRender("comp-1", "Counter", 6*time.Millisecond)

	comp = perf.GetComponent("comp-1")
	require.NotNil(t, comp)
	assert.Equal(t, int64(3), comp.RenderCount)
	assert.Equal(t, 12*time.Millisecond, comp.AvgRenderTime) // (10+20+6)/3 = 12
	assert.Equal(t, 20*time.Millisecond, comp.MaxRenderTime)
	assert.Equal(t, 6*time.Millisecond, comp.MinRenderTime)
}

// TestPerformanceData_GetAll tests retrieving all performance data
func TestPerformanceData_GetAll(t *testing.T) {
	perf := NewPerformanceData()

	// Record for multiple components
	perf.RecordRender("comp-1", "Counter", 10*time.Millisecond)
	perf.RecordRender("comp-2", "Form", 20*time.Millisecond)
	perf.RecordRender("comp-3", "List", 15*time.Millisecond)

	all := perf.GetAll()
	assert.Equal(t, 3, len(all))

	// Verify all components present
	assert.NotNil(t, all["comp-1"])
	assert.NotNil(t, all["comp-2"])
	assert.NotNil(t, all["comp-3"])
}

// TestPerformanceData_Clear tests clearing performance data
func TestPerformanceData_Clear(t *testing.T) {
	perf := NewPerformanceData()

	// Add some data
	perf.RecordRender("comp-1", "Counter", 10*time.Millisecond)
	perf.RecordRender("comp-2", "Form", 20*time.Millisecond)

	require.Equal(t, 2, len(perf.GetAll()))

	// Clear
	perf.Clear()

	assert.Equal(t, 0, len(perf.GetAll()))
	assert.Nil(t, perf.GetComponent("comp-1"))
}

// TestPerformanceData_Concurrent tests concurrent access to performance data
func TestPerformanceData_Concurrent(t *testing.T) {
	perf := NewPerformanceData()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				perf.RecordRender("comp-1", "Counter", time.Duration(j)*time.Millisecond)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = perf.GetComponent("comp-1")
				_ = perf.GetAll()
			}
		}()
	}

	wg.Wait()

	// Should have recorded 1000 renders
	comp := perf.GetComponent("comp-1")
	require.NotNil(t, comp)
	assert.Equal(t, int64(1000), comp.RenderCount)
}

// TestDevToolsStore_AddComponent tests adding components
func TestDevToolsStore_AddComponent(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	snapshot := &ComponentSnapshot{
		ID:        "comp-1",
		Name:      "Counter",
		Type:      "bubbly.Component",
		Timestamp: time.Now(),
	}

	store.AddComponent(snapshot)

	// Retrieve component
	retrieved := store.GetComponent("comp-1")
	require.NotNil(t, retrieved)
	assert.Equal(t, "comp-1", retrieved.ID)
	assert.Equal(t, "Counter", retrieved.Name)
}

// TestDevToolsStore_GetAllComponents tests retrieving all components
func TestDevToolsStore_GetAllComponents(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	// Add multiple components
	store.AddComponent(&ComponentSnapshot{ID: "comp-1", Name: "Counter"})
	store.AddComponent(&ComponentSnapshot{ID: "comp-2", Name: "Form"})
	store.AddComponent(&ComponentSnapshot{ID: "comp-3", Name: "List"})

	all := store.GetAllComponents()
	assert.Equal(t, 3, len(all))
}

// TestDevToolsStore_RemoveComponent tests removing components
func TestDevToolsStore_RemoveComponent(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	// Add component
	store.AddComponent(&ComponentSnapshot{ID: "comp-1", Name: "Counter"})
	require.NotNil(t, store.GetComponent("comp-1"))

	// Remove component
	store.RemoveComponent("comp-1")

	assert.Nil(t, store.GetComponent("comp-1"))
}

// TestDevToolsStore_UpdateComponent tests updating existing components
func TestDevToolsStore_UpdateComponent(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	// Add initial component
	store.AddComponent(&ComponentSnapshot{
		ID:   "comp-1",
		Name: "Counter",
		State: map[string]interface{}{
			"count": 0,
		},
	})

	// Update component
	store.AddComponent(&ComponentSnapshot{
		ID:   "comp-1",
		Name: "Counter",
		State: map[string]interface{}{
			"count": 42,
		},
	})

	// Verify updated
	comp := store.GetComponent("comp-1")
	require.NotNil(t, comp)
	assert.Equal(t, 42, comp.State["count"])
}

// TestDevToolsStore_StateHistory tests state history integration
func TestDevToolsStore_StateHistory(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	history := store.GetStateHistory()
	require.NotNil(t, history)

	// Record change
	history.Record(StateChange{
		RefID:    "ref-1",
		OldValue: 0,
		NewValue: 1,
	})

	changes := history.GetHistory("ref-1")
	assert.Equal(t, 1, len(changes))
}

// TestDevToolsStore_EventLog tests event log integration
func TestDevToolsStore_EventLog(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	log := store.GetEventLog()
	require.NotNil(t, log)

	// Append event
	log.Append(EventRecord{
		ID:   "evt-1",
		Name: "click",
	})

	assert.Equal(t, 1, log.Len())
}

// TestDevToolsStore_PerformanceData tests performance data integration
func TestDevToolsStore_PerformanceData(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	perf := store.GetPerformanceData()
	require.NotNil(t, perf)

	// Record render
	perf.RecordRender("comp-1", "Counter", 10*time.Millisecond)

	comp := perf.GetComponent("comp-1")
	require.NotNil(t, comp)
	assert.Equal(t, int64(1), comp.RenderCount)
}

// TestDevToolsStore_Clear tests clearing all data
func TestDevToolsStore_Clear(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	// Add data
	store.AddComponent(&ComponentSnapshot{ID: "comp-1", Name: "Counter"})
	store.GetStateHistory().Record(StateChange{RefID: "ref-1"})
	store.GetEventLog().Append(EventRecord{ID: "evt-1"})
	store.GetPerformanceData().RecordRender("comp-1", "Counter", 10*time.Millisecond)

	// Verify data exists
	require.Equal(t, 1, len(store.GetAllComponents()))
	require.Equal(t, 1, len(store.GetStateHistory().GetAll()))
	require.Equal(t, 1, store.GetEventLog().Len())
	require.NotNil(t, store.GetPerformanceData().GetComponent("comp-1"))

	// Clear
	store.Clear()

	// Verify all cleared
	assert.Equal(t, 0, len(store.GetAllComponents()))
	assert.Equal(t, 0, len(store.GetStateHistory().GetAll()))
	assert.Equal(t, 0, store.GetEventLog().Len())
	assert.Nil(t, store.GetPerformanceData().GetComponent("comp-1"))
}

// TestDevToolsStore_Concurrent tests concurrent access to store
func TestDevToolsStore_Concurrent(t *testing.T) {
	store := NewDevToolsStore(1000, 1000, 1000)
	var wg sync.WaitGroup

	// Concurrent component operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				snapshot := &ComponentSnapshot{
					ID:   "comp-" + string(rune('a'+id)),
					Name: "Component",
				}
				store.AddComponent(snapshot)
				_ = store.GetComponent(snapshot.ID)
				_ = store.GetAllComponents()
			}
		}(i)
	}

	// Concurrent state history operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				store.GetStateHistory().Record(StateChange{RefID: "ref-1"})
			}
		}()
	}

	// Concurrent event log operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				store.GetEventLog().Append(EventRecord{ID: "evt"})
			}
		}()
	}

	// Concurrent performance operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				store.GetPerformanceData().RecordRender("comp-1", "Counter", time.Millisecond)
			}
		}()
	}

	wg.Wait()

	// Verify no data corruption
	assert.Equal(t, 10, len(store.GetAllComponents()))
	assert.Equal(t, 500, len(store.GetStateHistory().GetAll()))
	assert.Equal(t, 500, store.GetEventLog().Len())
	comp := store.GetPerformanceData().GetComponent("comp-1")
	require.NotNil(t, comp)
	assert.Equal(t, int64(500), comp.RenderCount)
}

// BenchmarkStateHistory_Record benchmarks recording state changes
func BenchmarkStateHistory_Record(b *testing.B) {
	history := NewStateHistory(10000)
	change := StateChange{
		RefID:     "ref-1",
		RefName:   "counter",
		OldValue:  0,
		NewValue:  1,
		Timestamp: time.Now(),
		Source:    "test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		history.Record(change)
	}
}

// BenchmarkStateHistory_GetHistory benchmarks retrieving history for a ref
func BenchmarkStateHistory_GetHistory(b *testing.B) {
	history := NewStateHistory(10000)

	// Pre-populate with changes for multiple refs
	for i := 0; i < 1000; i++ {
		history.Record(StateChange{
			RefID:     "ref-1",
			RefName:   "counter",
			OldValue:  i,
			NewValue:  i + 1,
			Timestamp: time.Now(),
			Source:    "test",
		})
		history.Record(StateChange{
			RefID:     "ref-2",
			RefName:   "name",
			OldValue:  "old",
			NewValue:  "new",
			Timestamp: time.Now(),
			Source:    "test",
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = history.GetHistory("ref-1")
	}
}

// BenchmarkStateHistory_GetAll benchmarks retrieving all history
func BenchmarkStateHistory_GetAll(b *testing.B) {
	history := NewStateHistory(10000)

	// Pre-populate with changes
	for i := 0; i < 1000; i++ {
		history.Record(StateChange{
			RefID:     "ref-1",
			OldValue:  i,
			NewValue:  i + 1,
			Timestamp: time.Now(),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = history.GetAll()
	}
}

// BenchmarkStateHistory_Concurrent benchmarks concurrent operations
func BenchmarkStateHistory_Concurrent(b *testing.B) {
	history := NewStateHistory(10000)

	b.RunParallel(func(pb *testing.PB) {
		change := StateChange{
			RefID:     "ref-1",
			OldValue:  0,
			NewValue:  1,
			Timestamp: time.Now(),
		}

		for pb.Next() {
			// Mix of reads and writes
			if b.N%2 == 0 {
				history.Record(change)
			} else {
				_ = history.GetHistory("ref-1")
			}
		}
	})
}

// TestStore_RemoveComponentChild tests removing a child from a component
func TestStore_RemoveComponentChild(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	// Add parent and children
	store.AddComponent(&ComponentSnapshot{ID: "parent-1", Name: "Parent"})
	store.AddComponent(&ComponentSnapshot{ID: "child-1", Name: "Child1"})
	store.AddComponent(&ComponentSnapshot{ID: "child-2", Name: "Child2"})

	// Add children to parent
	store.AddComponentChild("parent-1", "child-1")
	store.AddComponentChild("parent-1", "child-2")

	// Verify children are added
	children := store.GetComponentChildren("parent-1")
	assert.Equal(t, 2, len(children))
	assert.Contains(t, children, "child-1")
	assert.Contains(t, children, "child-2")

	// Remove one child
	store.RemoveComponentChild("parent-1", "child-1")

	// Verify child was removed
	children = store.GetComponentChildren("parent-1")
	assert.Equal(t, 1, len(children))
	assert.NotContains(t, children, "child-1")
	assert.Contains(t, children, "child-2")
}

// TestStore_RemoveComponentChild_NonExistentParent tests removing child from non-existent parent
func TestStore_RemoveComponentChild_NonExistentParent(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	// Should not panic when removing child from non-existent parent
	assert.NotPanics(t, func() {
		store.RemoveComponentChild("non-existent", "child-1")
	})
}

// TestStore_RemoveComponentChild_NonExistentChild tests removing non-existent child
func TestStore_RemoveComponentChild_NonExistentChild(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	// Add parent
	store.AddComponent(&ComponentSnapshot{ID: "parent-1", Name: "Parent"})
	store.AddComponentChild("parent-1", "child-1")

	// Should not panic when removing non-existent child
	assert.NotPanics(t, func() {
		store.RemoveComponentChild("parent-1", "non-existent-child")
	})

	// Original child should still be there
	children := store.GetComponentChildren("parent-1")
	assert.Contains(t, children, "child-1")
}

// TestStore_RemoveComponentChild_UpdatesChildrenSnapshot tests that parent's Children field is updated
func TestStore_RemoveComponentChild_UpdatesChildrenSnapshot(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	// Add parent with children in snapshot
	childSnapshot := &ComponentSnapshot{ID: "child-1", Name: "Child1"}
	store.AddComponent(&ComponentSnapshot{
		ID:       "parent-1",
		Name:     "Parent",
		Children: []*ComponentSnapshot{childSnapshot},
	})

	store.AddComponentChild("parent-1", "child-1")

	// Remove child
	store.RemoveComponentChild("parent-1", "child-1")

	// Verify Children field in snapshot is also updated
	parent := store.GetComponent("parent-1")
	assert.NotNil(t, parent)
	for _, child := range parent.Children {
		assert.NotEqual(t, "child-1", child.ID, "Child should be removed from parent's Children")
	}
}

// TestStore_GetComponentChildren tests getting children of a component
func TestStore_GetComponentChildren(t *testing.T) {
	tests := []struct {
		name          string
		setupChildren []string
		wantCount     int
	}{
		{
			name:          "no children",
			setupChildren: []string{},
			wantCount:     0,
		},
		{
			name:          "one child",
			setupChildren: []string{"child-1"},
			wantCount:     1,
		},
		{
			name:          "multiple children",
			setupChildren: []string{"child-1", "child-2", "child-3"},
			wantCount:     3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewDevToolsStore(100, 100, 1000)

			store.AddComponent(&ComponentSnapshot{ID: "parent-1", Name: "Parent"})
			for _, childID := range tt.setupChildren {
				store.AddComponentChild("parent-1", childID)
			}

			children := store.GetComponentChildren("parent-1")
			assert.Equal(t, tt.wantCount, len(children))

			for _, childID := range tt.setupChildren {
				assert.Contains(t, children, childID)
			}
		})
	}
}

// TestStore_GetComponentChildren_ReturnsCopy tests that returned slice is a copy
func TestStore_GetComponentChildren_ReturnsCopy(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	store.AddComponent(&ComponentSnapshot{ID: "parent-1", Name: "Parent"})
	store.AddComponentChild("parent-1", "child-1")

	children := store.GetComponentChildren("parent-1")

	// Modify the returned slice
	children[0] = "modified"

	// Original should be unchanged
	originalChildren := store.GetComponentChildren("parent-1")
	assert.Equal(t, "child-1", originalChildren[0])
}

// TestStore_GetComponentChildren_NonExistentParent tests getting children of non-existent parent
func TestStore_GetComponentChildren_NonExistentParent(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	children := store.GetComponentChildren("non-existent")
	assert.Equal(t, 0, len(children))
	assert.NotNil(t, children) // Should return empty slice, not nil
}

// TestExtractRefName tests extracting simple name from ref ID
func TestExtractRefName(t *testing.T) {
	tests := []struct {
		name     string
		refID    string
		wantName string
	}{
		{
			name:     "ref with -ref- pattern",
			refID:    "counter-ref-0x123abc",
			wantName: "counter",
		},
		{
			name:     "ref with -0x pattern",
			refID:    "count-0x456def",
			wantName: "count",
		},
		{
			name:     "ref without pattern",
			refID:    "simpleRef",
			wantName: "simpleRef",
		},
		{
			name:     "complex ref name with -ref-",
			refID:    "my-complex-name-ref-0x789",
			wantName: "my-complex-name",
		},
		{
			name:     "short ref name",
			refID:    "x-0x1",
			wantName: "x",
		},
		{
			name:     "empty string",
			refID:    "",
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRefName(tt.refID)
			assert.Equal(t, tt.wantName, result)
		})
	}
}

// TestStore_RegisterRefOwner tests registering ref ownership
func TestStore_RegisterRefOwner(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	// Add a component
	store.AddComponent(&ComponentSnapshot{
		ID:   "comp-1",
		Name: "TestComponent",
		Refs: []*RefSnapshot{},
	})

	// Register ref owner
	store.RegisterRefOwner("comp-1", "ref-1")

	// Verify by updating the ref
	ownerID, updated := store.UpdateRefValue("ref-1", 42)
	assert.True(t, updated)
	assert.Equal(t, "comp-1", ownerID)

	// Verify ref was added to component's Refs
	comp := store.GetComponent("comp-1")
	assert.NotNil(t, comp)
	assert.GreaterOrEqual(t, len(comp.Refs), 1)
}

// TestStore_RegisterRefOwner_DuplicateRef tests registering same ref twice
func TestStore_RegisterRefOwner_DuplicateRef(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	store.AddComponent(&ComponentSnapshot{
		ID:   "comp-1",
		Name: "TestComponent",
		Refs: []*RefSnapshot{},
	})

	// Register same ref twice
	store.RegisterRefOwner("comp-1", "ref-1")
	store.RegisterRefOwner("comp-1", "ref-1")

	// Should not add duplicate
	comp := store.GetComponent("comp-1")
	refCount := 0
	for _, ref := range comp.Refs {
		if ref.ID == "ref-1" {
			refCount++
		}
	}
	assert.Equal(t, 1, refCount, "Should not add duplicate ref")
}

// TestStore_RegisterRefOwner_NonExistentComponent tests registering ref for non-existent component
func TestStore_RegisterRefOwner_NonExistentComponent(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	// Should not panic
	assert.NotPanics(t, func() {
		store.RegisterRefOwner("non-existent", "ref-1")
	})
}

// TestStore_UpdateRefValue tests updating ref values
func TestStore_UpdateRefValue(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	// Add component and register ref
	store.AddComponent(&ComponentSnapshot{
		ID:   "comp-1",
		Name: "TestComponent",
		Refs: []*RefSnapshot{},
	})
	store.RegisterRefOwner("comp-1", "ref-1")

	// Update ref value
	ownerID, updated := store.UpdateRefValue("ref-1", "new-value")

	assert.True(t, updated)
	assert.Equal(t, "comp-1", ownerID)

	// Verify the ref was updated in component's Refs
	comp := store.GetComponent("comp-1")
	assert.NotNil(t, comp)
	for _, ref := range comp.Refs {
		if ref.ID == "ref-1" {
			assert.Equal(t, "new-value", ref.Value)
			break
		}
	}
}

// TestStore_UpdateRefValue_UnregisteredRef tests updating unregistered ref
func TestStore_UpdateRefValue_UnregisteredRef(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	// Try to update unregistered ref
	ownerID, updated := store.UpdateRefValue("unregistered-ref", "value")

	assert.False(t, updated)
	assert.Empty(t, ownerID)
}

// TestStore_GetSince tests getting data since a checkpoint
func TestStore_GetSince(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	// Add component
	store.AddComponent(&ComponentSnapshot{
		ID:        "comp-1",
		Name:      "TestComponent",
		Timestamp: time.Now(),
	})

	// Record state change
	store.GetStateHistory().Record(StateChange{
		RefID:     "ref-1",
		OldValue:  0,
		NewValue:  1,
		Timestamp: time.Now(),
	})

	// Record event
	store.GetEventLog().Append(EventRecord{
		ID:        "evt-1",
		Name:      "click",
		Timestamp: time.Now(),
	})

	// Get data since checkpoint with ID 0
	checkpoint := &ExportCheckpoint{
		LastEventID: 0,
		LastStateID: 0,
		Timestamp:   time.Now().Add(-1 * time.Hour),
	}
	data, err := store.GetSince(checkpoint)

	assert.NoError(t, err)
	assert.NotNil(t, data)
}

// TestStore_GetSince_NilCheckpoint tests GetSince with nil checkpoint
func TestStore_GetSince_NilCheckpoint(t *testing.T) {
	store := NewDevToolsStore(100, 100, 1000)

	data, err := store.GetSince(nil)

	assert.Error(t, err)
	assert.Nil(t, data)
}
