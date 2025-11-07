package devtools

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewStateViewer tests the StateViewer constructor
func TestNewStateViewer(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates new state viewer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewDevToolsStore(1000, 5000)
			sv := NewStateViewer(store)

			require.NotNil(t, sv)
			assert.NotNil(t, sv.store)
			assert.Nil(t, sv.selected)
			assert.Equal(t, "", sv.filter)
		})
	}
}

// TestStateViewer_Render tests the Render method
func TestStateViewer_Render(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*DevToolsStore)
		contains []string
	}{
		{
			name:     "empty state",
			setup:    func(s *DevToolsStore) {},
			contains: []string{"Reactive State", "No components"},
		},
		{
			name: "single component with refs",
			setup: func(s *DevToolsStore) {
				s.AddComponent(&ComponentSnapshot{
					ID:   "comp-1",
					Name: "Counter",
					Refs: []*RefSnapshot{
						{ID: "ref-1", Name: "count", Type: "int", Value: 42, Watchers: 0},
					},
				})
			},
			contains: []string{"Counter", "count", "42", "int"},
		},
		{
			name: "multiple components",
			setup: func(s *DevToolsStore) {
				s.AddComponent(&ComponentSnapshot{
					ID:   "comp-1",
					Name: "Counter",
					Refs: []*RefSnapshot{
						{ID: "ref-1", Name: "count", Type: "int", Value: 42},
					},
				})
				s.AddComponent(&ComponentSnapshot{
					ID:   "comp-2",
					Name: "Input",
					Refs: []*RefSnapshot{
						{ID: "ref-2", Name: "text", Type: "string", Value: "hello"},
					},
				})
			},
			contains: []string{"Counter", "Input", "count", "text"},
		},
		{
			name: "selected ref highlighted",
			setup: func(s *DevToolsStore) {
				s.AddComponent(&ComponentSnapshot{
					ID:   "comp-1",
					Name: "Counter",
					Refs: []*RefSnapshot{
						{ID: "ref-1", Name: "count", Type: "int", Value: 42},
					},
				})
			},
			contains: []string{"►", "count"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewDevToolsStore(1000, 5000)
			tt.setup(store)

			sv := NewStateViewer(store)
			if tt.name == "selected ref highlighted" {
				sv.SelectRef("ref-1")
			}

			output := sv.Render()

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected, "output should contain %q", expected)
			}
		})
	}
}

// TestStateViewer_SelectRef tests ref selection
func TestStateViewer_SelectRef(t *testing.T) {
	tests := []struct {
		name     string
		refID    string
		expected bool
	}{
		{name: "select existing ref", refID: "ref-1", expected: true},
		{name: "select non-existent ref", refID: "ref-999", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewDevToolsStore(1000, 5000)
			store.AddComponent(&ComponentSnapshot{
				ID:   "comp-1",
				Name: "Counter",
				Refs: []*RefSnapshot{
					{ID: "ref-1", Name: "count", Type: "int", Value: 42},
				},
			})

			sv := NewStateViewer(store)
			result := sv.SelectRef(tt.refID)

			assert.Equal(t, tt.expected, result)
			if tt.expected {
				assert.NotNil(t, sv.GetSelected())
				assert.Equal(t, tt.refID, sv.GetSelected().ID)
			} else {
				assert.Nil(t, sv.GetSelected())
			}
		})
	}
}

// TestStateViewer_Filter tests filtering functionality
func TestStateViewer_Filter(t *testing.T) {
	tests := []struct {
		name     string
		filter   string
		contains []string
		excludes []string
	}{
		{
			name:     "no filter shows all",
			filter:   "",
			contains: []string{"count", "text", "enabled"},
			excludes: []string{},
		},
		{
			name:     "filter by name",
			filter:   "count",
			contains: []string{"count"},
			excludes: []string{"text", "enabled"},
		},
		{
			name:     "case insensitive filter",
			filter:   "TEXT",
			contains: []string{"text"},
			excludes: []string{"count", "enabled"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewDevToolsStore(1000, 5000)
			store.AddComponent(&ComponentSnapshot{
				ID:   "comp-1",
				Name: "Counter",
				Refs: []*RefSnapshot{
					{ID: "ref-1", Name: "count", Type: "int", Value: 42},
					{ID: "ref-2", Name: "text", Type: "string", Value: "hello"},
					{ID: "ref-3", Name: "enabled", Type: "bool", Value: true},
				},
			})

			sv := NewStateViewer(store)
			sv.SetFilter(tt.filter)

			output := sv.Render()

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
			for _, excluded := range tt.excludes {
				assert.NotContains(t, output, excluded)
			}
		})
	}
}

// TestStateViewer_EditValue tests value editing
func TestStateViewer_EditValue(t *testing.T) {
	tests := []struct {
		name      string
		refID     string
		newValue  interface{}
		expectErr bool
	}{
		{name: "edit int value", refID: "ref-1", newValue: 100, expectErr: false},
		{name: "edit string value", refID: "ref-2", newValue: "world", expectErr: false},
		{name: "edit non-existent ref", refID: "ref-999", newValue: 42, expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewDevToolsStore(1000, 5000)
			store.AddComponent(&ComponentSnapshot{
				ID:   "comp-1",
				Name: "Counter",
				Refs: []*RefSnapshot{
					{ID: "ref-1", Name: "count", Type: "int", Value: 42},
					{ID: "ref-2", Name: "text", Type: "string", Value: "hello"},
				},
			})

			sv := NewStateViewer(store)
			err := sv.EditValue(tt.refID, tt.newValue)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify value was updated in store
				comp := store.GetComponent("comp-1")
				for _, ref := range comp.Refs {
					if ref.ID == tt.refID {
						assert.Equal(t, tt.newValue, ref.Value)
					}
				}
			}
		})
	}
}

// TestStateViewer_GetFilter tests filter getter
func TestStateViewer_GetFilter(t *testing.T) {
	store := NewDevToolsStore(1000, 5000)
	sv := NewStateViewer(store)

	sv.SetFilter("test")
	assert.Equal(t, "test", sv.GetFilter())
}

// TestStateViewer_GetSelected tests selected ref getter
func TestStateViewer_GetSelected(t *testing.T) {
	store := NewDevToolsStore(1000, 5000)
	store.AddComponent(&ComponentSnapshot{
		ID:   "comp-1",
		Name: "Counter",
		Refs: []*RefSnapshot{
			{ID: "ref-1", Name: "count", Type: "int", Value: 42},
		},
	})

	sv := NewStateViewer(store)

	// Initially no selection
	assert.Nil(t, sv.GetSelected())

	// After selection
	sv.SelectRef("ref-1")
	selected := sv.GetSelected()
	require.NotNil(t, selected)
	assert.Equal(t, "ref-1", selected.ID)
}

// TestStateViewer_ClearSelection tests clearing selection
func TestStateViewer_ClearSelection(t *testing.T) {
	store := NewDevToolsStore(1000, 5000)
	store.AddComponent(&ComponentSnapshot{
		ID:   "comp-1",
		Name: "Counter",
		Refs: []*RefSnapshot{
			{ID: "ref-1", Name: "count", Type: "int", Value: 42},
		},
	})

	sv := NewStateViewer(store)
	sv.SelectRef("ref-1")
	assert.NotNil(t, sv.GetSelected())

	sv.ClearSelection()
	assert.Nil(t, sv.GetSelected())
}

// TestStateViewer_ThreadSafety tests concurrent access
func TestStateViewer_ThreadSafety(t *testing.T) {
	store := NewDevToolsStore(1000, 5000)
	store.AddComponent(&ComponentSnapshot{
		ID:   "comp-1",
		Name: "Counter",
		Refs: []*RefSnapshot{
			{ID: "ref-1", Name: "count", Type: "int", Value: 42},
			{ID: "ref-2", Name: "text", Type: "string", Value: "hello"},
		},
	})

	sv := NewStateViewer(store)

	var wg sync.WaitGroup
	operations := 100

	// Concurrent reads
	for i := 0; i < operations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = sv.Render()
			_ = sv.GetSelected()
			_ = sv.GetFilter()
		}()
	}

	// Concurrent writes
	for i := 0; i < operations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx%2 == 0 {
				sv.SelectRef("ref-1")
			} else {
				sv.SetFilter("test")
			}
		}(i)
	}

	wg.Wait()
	// If we get here without deadlock or race, test passes
}

// TestStateViewer_ComplexValues tests rendering complex value types
func TestStateViewer_ComplexValues(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		contains string
	}{
		{name: "nil value", value: nil, contains: "<nil>"},
		{name: "slice value", value: []int{1, 2, 3}, contains: "[1 2 3]"},
		{name: "map value", value: map[string]int{"a": 1}, contains: "map["},
		{name: "struct value", value: struct{ X int }{X: 42}, contains: "{"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewDevToolsStore(1000, 5000)
			store.AddComponent(&ComponentSnapshot{
				ID:   "comp-1",
				Name: "Test",
				Refs: []*RefSnapshot{
					{ID: "ref-1", Name: "value", Type: "interface{}", Value: tt.value},
				},
			})

			sv := NewStateViewer(store)
			output := sv.Render()

			assert.Contains(t, output, tt.contains)
		})
	}
}

// TestStateViewer_EmptyComponents tests rendering with no refs
func TestStateViewer_EmptyComponents(t *testing.T) {
	store := NewDevToolsStore(1000, 5000)
	store.AddComponent(&ComponentSnapshot{
		ID:   "comp-1",
		Name: "Empty",
		Refs: []*RefSnapshot{},
	})

	sv := NewStateViewer(store)
	output := sv.Render()

	assert.Contains(t, output, "Empty")
	assert.Contains(t, output, "no refs")
}
