package testutil

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStateInspector_GetRef tests retrieving a ref by name
func TestStateInspector_GetRef(t *testing.T) {
	tests := []struct {
		name    string
		refName string
		setup   func() *StateInspector
		wantNil bool
	}{
		{
			name:    "existing ref returns ref",
			refName: "count",
			setup: func() *StateInspector {
				refs := map[string]*bubbly.Ref[interface{}]{
					"count": bubbly.NewRef[interface{}](42),
				}
				return NewStateInspector(refs, nil, nil)
			},
			wantNil: false,
		},
		{
			name:    "non-existent ref returns nil",
			refName: "missing",
			setup: func() *StateInspector {
				refs := map[string]*bubbly.Ref[interface{}]{}
				return NewStateInspector(refs, nil, nil)
			},
			wantNil: true,
		},
		{
			name:    "empty refs map returns nil",
			refName: "any",
			setup: func() *StateInspector {
				return NewStateInspector(map[string]*bubbly.Ref[interface{}]{}, nil, nil)
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := tt.setup()
			ref := si.GetRef(tt.refName)

			if tt.wantNil {
				assert.Nil(t, ref, "expected nil ref")
			} else {
				assert.NotNil(t, ref, "expected non-nil ref")
			}
		})
	}
}

// TestStateInspector_GetRefValue tests getting ref values
func TestStateInspector_GetRefValue(t *testing.T) {
	tests := []struct {
		name     string
		refName  string
		setup    func() *StateInspector
		expected interface{}
		wantErr  bool
	}{
		{
			name:    "get integer value",
			refName: "count",
			setup: func() *StateInspector {
				refs := map[string]*bubbly.Ref[interface{}]{
					"count": bubbly.NewRef[interface{}](42),
				}
				return NewStateInspector(refs, nil, nil)
			},
			expected: 42,
			wantErr:  false,
		},
		{
			name:    "get string value",
			refName: "name",
			setup: func() *StateInspector {
				refs := map[string]*bubbly.Ref[interface{}]{
					"name": bubbly.NewRef[interface{}]("test"),
				}
				return NewStateInspector(refs, nil, nil)
			},
			expected: "test",
			wantErr:  false,
		},
		{
			name:    "get boolean value",
			refName: "enabled",
			setup: func() *StateInspector {
				refs := map[string]*bubbly.Ref[interface{}]{
					"enabled": bubbly.NewRef[interface{}](true),
				}
				return NewStateInspector(refs, nil, nil)
			},
			expected: true,
			wantErr:  false,
		},
		{
			name:    "non-existent ref panics",
			refName: "missing",
			setup: func() *StateInspector {
				return NewStateInspector(map[string]*bubbly.Ref[interface{}]{}, nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := tt.setup()

			if tt.wantErr {
				assert.Panics(t, func() {
					si.GetRefValue(tt.refName)
				}, "expected panic for missing ref")
			} else {
				value := si.GetRefValue(tt.refName)
				assert.Equal(t, tt.expected, value, "value should match")
			}
		})
	}
}

// TestStateInspector_SetRefValue tests setting ref values
func TestStateInspector_SetRefValue(t *testing.T) {
	tests := []struct {
		name     string
		refName  string
		newValue interface{}
		setup    func() *StateInspector
		wantErr  bool
	}{
		{
			name:     "set integer value",
			refName:  "count",
			newValue: 100,
			setup: func() *StateInspector {
				refs := map[string]*bubbly.Ref[interface{}]{
					"count": bubbly.NewRef[interface{}](42),
				}
				return NewStateInspector(refs, nil, nil)
			},
			wantErr: false,
		},
		{
			name:     "set string value",
			refName:  "name",
			newValue: "updated",
			setup: func() *StateInspector {
				refs := map[string]*bubbly.Ref[interface{}]{
					"name": bubbly.NewRef[interface{}]("original"),
				}
				return NewStateInspector(refs, nil, nil)
			},
			wantErr: false,
		},
		{
			name:     "set nil value",
			refName:  "data",
			newValue: nil,
			setup: func() *StateInspector {
				refs := map[string]*bubbly.Ref[interface{}]{
					"data": bubbly.NewRef[interface{}]("something"),
				}
				return NewStateInspector(refs, nil, nil)
			},
			wantErr: false,
		},
		{
			name:     "non-existent ref panics",
			refName:  "missing",
			newValue: 42,
			setup: func() *StateInspector {
				return NewStateInspector(map[string]*bubbly.Ref[interface{}]{}, nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := tt.setup()

			if tt.wantErr {
				assert.Panics(t, func() {
					si.SetRefValue(tt.refName, tt.newValue)
				}, "expected panic for missing ref")
			} else {
				si.SetRefValue(tt.refName, tt.newValue)
				// Verify the value was actually set
				actual := si.GetRefValue(tt.refName)
				assert.Equal(t, tt.newValue, actual, "value should be updated")
			}
		})
	}
}

// TestStateInspector_MultipleRefs tests working with multiple refs
func TestStateInspector_MultipleRefs(t *testing.T) {
	refs := map[string]*bubbly.Ref[interface{}]{
		"count":   bubbly.NewRef[interface{}](0),
		"name":    bubbly.NewRef[interface{}]("test"),
		"enabled": bubbly.NewRef[interface{}](false),
	}
	si := NewStateInspector(refs, nil, nil)

	// Get all refs
	countRef := si.GetRef("count")
	nameRef := si.GetRef("name")
	enabledRef := si.GetRef("enabled")

	require.NotNil(t, countRef)
	require.NotNil(t, nameRef)
	require.NotNil(t, enabledRef)

	// Verify initial values
	assert.Equal(t, 0, si.GetRefValue("count"))
	assert.Equal(t, "test", si.GetRefValue("name"))
	assert.Equal(t, false, si.GetRefValue("enabled"))

	// Update values
	si.SetRefValue("count", 42)
	si.SetRefValue("name", "updated")
	si.SetRefValue("enabled", true)

	// Verify updated values
	assert.Equal(t, 42, si.GetRefValue("count"))
	assert.Equal(t, "updated", si.GetRefValue("name"))
	assert.Equal(t, true, si.GetRefValue("enabled"))
}

// TestStateInspector_EmptyRefs tests behavior with empty refs map
func TestStateInspector_EmptyRefs(t *testing.T) {
	si := NewStateInspector(map[string]*bubbly.Ref[interface{}]{}, nil, nil)

	// All operations should handle empty map gracefully
	assert.Nil(t, si.GetRef("any"))

	assert.Panics(t, func() {
		si.GetRefValue("any")
	}, "should panic on missing ref")

	assert.Panics(t, func() {
		si.SetRefValue("any", 42)
	}, "should panic on missing ref")
}

// TestStateInspector_NilRefsMap tests behavior with nil refs map
func TestStateInspector_NilRefsMap(t *testing.T) {
	si := NewStateInspector(nil, nil, nil)

	// Should handle nil map gracefully
	assert.Nil(t, si.GetRef("any"))

	assert.Panics(t, func() {
		si.GetRefValue("any")
	}, "should panic on missing ref")

	assert.Panics(t, func() {
		si.SetRefValue("any", 42)
	}, "should panic on missing ref")
}

// TestStateInspector_ThreadSafety tests concurrent access
func TestStateInspector_ThreadSafety(t *testing.T) {
	refs := map[string]*bubbly.Ref[interface{}]{
		"count": bubbly.NewRef[interface{}](0),
	}
	si := NewStateInspector(refs, nil, nil)

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				si.GetRef("count")
				si.GetRefValue("count")
				si.SetRefValue("count", j)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic or deadlock
	assert.NotNil(t, si.GetRef("count"))
}

// TestStateInspector_Integration tests integration with ComponentTest
func TestStateInspector_Integration(t *testing.T) {
	harness := NewHarness(t)

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			name := ctx.Ref("test")
			ctx.Expose("count", count)
			ctx.Expose("name", name)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)

	ct := harness.Mount(component)
	require.NotNil(t, ct)
	require.NotNil(t, ct.state)

	// Note: Currently refs are not automatically extracted to harness.refs
	// This is expected behavior - full extraction will be implemented
	// when we integrate with component state inspection in future tasks
}

// TestStateInspector_GetComputed tests retrieving computed values
func TestStateInspector_GetComputed(t *testing.T) {
	count := bubbly.NewRef[interface{}](5)
	doubled := bubbly.NewComputed(func() interface{} {
		return count.Get().(int) * 2
	})

	computed := map[string]*bubbly.Computed[interface{}]{
		"doubled": doubled,
	}
	si := NewStateInspector(nil, computed, nil)

	// Get existing computed
	result := si.GetComputed("doubled")
	assert.NotNil(t, result)
	assert.Equal(t, 10, result.Get())

	// Get non-existent computed
	assert.Nil(t, si.GetComputed("missing"))

	// Nil computed map
	siNil := NewStateInspector(nil, nil, nil)
	assert.Nil(t, siNil.GetComputed("any"))
}

// TestStateInspector_GetComputedValue tests getting computed values
func TestStateInspector_GetComputedValue(t *testing.T) {
	count := bubbly.NewRef[interface{}](5)
	doubled := bubbly.NewComputed(func() interface{} {
		return count.Get().(int) * 2
	})

	computed := map[string]*bubbly.Computed[interface{}]{
		"doubled": doubled,
	}
	si := NewStateInspector(nil, computed, nil)

	// Get value
	value := si.GetComputedValue("doubled")
	assert.Equal(t, 10, value)

	// Update ref and verify computed updates
	count.Set(10)
	value = si.GetComputedValue("doubled")
	assert.Equal(t, 20, value)

	// Panic on missing computed
	assert.Panics(t, func() {
		si.GetComputedValue("missing")
	}, "should panic on missing computed")
}

// TestStateInspector_GetWatcher tests retrieving watchers
func TestStateInspector_GetWatcher(t *testing.T) {
	ref := bubbly.NewRef[interface{}](0)
	callCount := 0

	cleanup := bubbly.Watch(ref, func(newVal, oldVal interface{}) {
		callCount++
	})

	watchers := map[string]bubbly.WatchCleanup{
		"countWatcher": cleanup,
	}
	si := NewStateInspector(nil, nil, watchers)

	// Get existing watcher
	result := si.GetWatcher("countWatcher")
	assert.NotNil(t, result)

	// Trigger watcher
	ref.Set(1)
	assert.Equal(t, 1, callCount)

	// Clean up watcher
	result()
	ref.Set(2)
	assert.Equal(t, 1, callCount, "watcher should be cleaned up")

	// Get non-existent watcher
	assert.Nil(t, si.GetWatcher("missing"))

	// Nil watchers map
	siNil := NewStateInspector(nil, nil, nil)
	assert.Nil(t, siNil.GetWatcher("any"))
}

// TestStateInspector_HasRef tests checking ref existence
func TestStateInspector_HasRef(t *testing.T) {
	refs := map[string]*bubbly.Ref[interface{}]{
		"count": bubbly.NewRef[interface{}](0),
	}
	si := NewStateInspector(refs, nil, nil)

	assert.True(t, si.HasRef("count"))
	assert.False(t, si.HasRef("missing"))

	// Nil refs
	siNil := NewStateInspector(nil, nil, nil)
	assert.False(t, siNil.HasRef("any"))
}

// TestStateInspector_HasComputed tests checking computed existence
func TestStateInspector_HasComputed(t *testing.T) {
	computed := map[string]*bubbly.Computed[interface{}]{
		"doubled": bubbly.NewComputed(func() interface{} { return 10 }),
	}
	si := NewStateInspector(nil, computed, nil)

	assert.True(t, si.HasComputed("doubled"))
	assert.False(t, si.HasComputed("missing"))

	// Nil computed
	siNil := NewStateInspector(nil, nil, nil)
	assert.False(t, siNil.HasComputed("any"))
}

// TestStateInspector_HasWatcher tests checking watcher existence
func TestStateInspector_HasWatcher(t *testing.T) {
	ref := bubbly.NewRef[interface{}](0)
	cleanup := bubbly.Watch(ref, func(newVal, oldVal interface{}) {})

	watchers := map[string]bubbly.WatchCleanup{
		"countWatcher": cleanup,
	}
	si := NewStateInspector(nil, nil, watchers)

	assert.True(t, si.HasWatcher("countWatcher"))
	assert.False(t, si.HasWatcher("missing"))

	// Nil watchers
	siNil := NewStateInspector(nil, nil, nil)
	assert.False(t, siNil.HasWatcher("any"))

	// Cleanup
	cleanup()
}

// TestStateInspector_AllFeatures tests all features together
func TestStateInspector_AllFeatures(t *testing.T) {
	// Create refs
	count := bubbly.NewRef[interface{}](5)
	name := bubbly.NewRef[interface{}]("test")

	// Create computed
	doubled := bubbly.NewComputed(func() interface{} {
		return count.Get().(int) * 2
	})

	// Create watcher
	watcherCalls := 0
	cleanup := bubbly.Watch(count, func(newVal, oldVal interface{}) {
		watcherCalls++
	})

	// Create inspector
	si := NewStateInspector(
		map[string]*bubbly.Ref[interface{}]{
			"count": count,
			"name":  name,
		},
		map[string]*bubbly.Computed[interface{}]{
			"doubled": doubled,
		},
		map[string]bubbly.WatchCleanup{
			"countWatcher": cleanup,
		},
	)

	// Test refs
	assert.True(t, si.HasRef("count"))
	assert.Equal(t, 5, si.GetRefValue("count"))

	// Test computed
	assert.True(t, si.HasComputed("doubled"))
	assert.Equal(t, 10, si.GetComputedValue("doubled"))

	// Test watcher
	assert.True(t, si.HasWatcher("countWatcher"))

	// Update ref and verify reactivity
	si.SetRefValue("count", 10)
	assert.Equal(t, 10, si.GetRefValue("count"))
	assert.Equal(t, 20, si.GetComputedValue("doubled"))
	assert.Equal(t, 1, watcherCalls)

	// Cleanup
	si.GetWatcher("countWatcher")()
}
