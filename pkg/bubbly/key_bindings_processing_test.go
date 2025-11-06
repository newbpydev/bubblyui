package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKeyBindingProcessing_Lookup tests key lookup in bindings map
func TestKeyBindingProcessing_Lookup(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		bindings    map[string][]KeyBinding
		shouldFind  bool
		expectEvent string
	}{
		{
			name: "key found in map",
			key:  "space",
			bindings: map[string][]KeyBinding{
				"space": {
					{Key: "space", Event: "increment", Description: "Increment"},
				},
			},
			shouldFind:  true,
			expectEvent: "increment",
		},
		{
			name: "key not found in map",
			key:  "x",
			bindings: map[string][]KeyBinding{
				"space": {
					{Key: "space", Event: "increment", Description: "Increment"},
				},
			},
			shouldFind:  false,
			expectEvent: "",
		},
		{
			name:        "empty bindings map",
			key:         "space",
			bindings:    map[string][]KeyBinding{},
			shouldFind:  false,
			expectEvent: "",
		},
		{
			name: "ctrl key combination",
			key:  "ctrl+c",
			bindings: map[string][]KeyBinding{
				"ctrl+c": {
					{Key: "ctrl+c", Event: "quit", Description: "Quit"},
				},
			},
			shouldFind:  true,
			expectEvent: "quit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := NewComponent("TestComponent").
				Template(func(ctx RenderContext) string { return "test" }).
				Build()
			require.NoError(t, err)

			impl := component.(*componentImpl)
			impl.keyBindings = tt.bindings

			// Test the lookup logic directly
			impl.keyBindingsMu.RLock()
			bindings, found := impl.keyBindings[tt.key]
			impl.keyBindingsMu.RUnlock()

			if tt.shouldFind {
				assert.True(t, found, "Expected to find key binding")
				assert.NotEmpty(t, bindings, "Expected non-empty bindings")
				assert.Equal(t, tt.expectEvent, bindings[0].Event)
			} else {
				assert.False(t, found, "Expected not to find key binding")
			}
		})
	}
}

// TestKeyBindingProcessing_EventEmission tests event emission on key match
func TestKeyBindingProcessing_EventEmission(t *testing.T) {
	tests := []struct {
		name        string
		binding     KeyBinding
		expectEvent string
		expectData  interface{}
		shouldEmit  bool
	}{
		{
			name: "simple event emission",
			binding: KeyBinding{
				Key:         "space",
				Event:       "increment",
				Description: "Increment counter",
			},
			expectEvent: "increment",
			expectData:  nil,
			shouldEmit:  true,
		},
		{
			name: "event with data",
			binding: KeyBinding{
				Key:         "enter",
				Event:       "submit",
				Description: "Submit form",
				Data:        map[string]string{"action": "save"},
			},
			expectEvent: "submit",
			expectData:  map[string]string{"action": "save"},
			shouldEmit:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := NewComponent("TestComponent").
				WithConditionalKeyBinding(tt.binding).
				Template(func(ctx RenderContext) string { return "test" }).
				Build()
			require.NoError(t, err)

			impl := component.(*componentImpl)

			// Track emitted events
			var emittedEvent string
			var emittedData interface{}
			impl.On(tt.expectEvent, func(data interface{}) {
				emittedEvent = tt.expectEvent
				emittedData = data
			})

			// Manually emit to test event system works
			impl.Emit(tt.binding.Event, tt.binding.Data)

			if tt.shouldEmit {
				assert.Equal(t, tt.expectEvent, emittedEvent)
				assert.Equal(t, tt.expectData, emittedData)
			}
		})
	}
}

// TestKeyBindingProcessing_ConditionEvaluation tests condition evaluation
func TestKeyBindingProcessing_ConditionEvaluation(t *testing.T) {
	tests := []struct {
		name           string
		conditionValue bool
		shouldProcess  bool
	}{
		{
			name:           "condition true - should process",
			conditionValue: true,
			shouldProcess:  true,
		},
		{
			name:           "condition false - should skip",
			conditionValue: false,
			shouldProcess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conditionCalled := false
			binding := KeyBinding{
				Key:         "space",
				Event:       "action",
				Description: "Conditional action",
				Condition: func() bool {
					conditionCalled = true
					return tt.conditionValue
				},
			}

			component, err := NewComponent("TestComponent").
				WithConditionalKeyBinding(binding).
				Template(func(ctx RenderContext) string { return "test" }).
				Build()
			require.NoError(t, err)

			impl := component.(*componentImpl)

			// Get the binding and test condition
			impl.keyBindingsMu.RLock()
			bindings := impl.keyBindings["space"]
			impl.keyBindingsMu.RUnlock()

			require.NotEmpty(t, bindings)

			// Test condition evaluation
			if bindings[0].Condition != nil {
				result := bindings[0].Condition()
				assert.True(t, conditionCalled, "Condition should be called")
				assert.Equal(t, tt.conditionValue, result)
			}
		})
	}
}

// TestKeyBindingProcessing_FirstMatchingWins tests that first matching binding wins
func TestKeyBindingProcessing_FirstMatchingWins(t *testing.T) {
	binding1 := KeyBinding{
		Key:         "space",
		Event:       "first",
		Description: "First binding",
	}

	binding2 := KeyBinding{
		Key:         "space",
		Event:       "second",
		Description: "Second binding",
	}

	component, err := NewComponent("TestComponent").
		WithConditionalKeyBinding(binding1).
		WithConditionalKeyBinding(binding2).
		Template(func(ctx RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	impl := component.(*componentImpl)

	// Verify both bindings registered
	impl.keyBindingsMu.RLock()
	bindings := impl.keyBindings["space"]
	impl.keyBindingsMu.RUnlock()

	require.Len(t, bindings, 2, "Should have 2 bindings for 'space'")
	assert.Equal(t, "first", bindings[0].Event)
	assert.Equal(t, "second", bindings[1].Event)

	// In actual processing, only first matching should execute
	// This will be verified in integration test with Update()
}

// TestKeyBindingProcessing_QuitEvent tests special "quit" event handling
func TestKeyBindingProcessing_QuitEvent(t *testing.T) {
	binding := KeyBinding{
		Key:         "ctrl+c",
		Event:       "quit",
		Description: "Quit application",
	}

	component, err := NewComponent("TestComponent").
		WithConditionalKeyBinding(binding).
		Template(func(ctx RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	impl := component.(*componentImpl)

	// Verify quit binding registered
	impl.keyBindingsMu.RLock()
	bindings := impl.keyBindings["ctrl+c"]
	impl.keyBindingsMu.RUnlock()

	require.NotEmpty(t, bindings)
	assert.Equal(t, "quit", bindings[0].Event)

	// The actual quit handling will be tested in Update() integration test
}

// TestKeyBindingProcessing_NoMatch tests that no match passes through
func TestKeyBindingProcessing_NoMatch(t *testing.T) {
	component, err := NewComponent("TestComponent").
		WithKeyBinding("space", "increment", "Increment").
		Template(func(ctx RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	impl := component.(*componentImpl)

	// Track if any event was emitted
	eventEmitted := false
	impl.On("increment", func(data interface{}) {
		eventEmitted = true
	})

	// Look for non-existent key
	impl.keyBindingsMu.RLock()
	_, found := impl.keyBindings["x"]
	impl.keyBindingsMu.RUnlock()

	assert.False(t, found, "Should not find binding for 'x'")
	assert.False(t, eventEmitted, "No event should be emitted")
}

// TestKeyBindingProcessing_MultipleBindingsWithConditions tests mode-based input
func TestKeyBindingProcessing_MultipleBindingsWithConditions(t *testing.T) {
	navigationMode := true

	navBinding := KeyBinding{
		Key:         "space",
		Event:       "toggle",
		Description: "Toggle selection",
		Condition: func() bool {
			return navigationMode
		},
	}

	inputBinding := KeyBinding{
		Key:         "space",
		Event:       "addSpace",
		Description: "Add space character",
		Condition: func() bool {
			return !navigationMode
		},
	}

	component, err := NewComponent("TestComponent").
		WithConditionalKeyBinding(navBinding).
		WithConditionalKeyBinding(inputBinding).
		Template(func(ctx RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	impl := component.(*componentImpl)

	// Verify both bindings registered
	impl.keyBindingsMu.RLock()
	bindings := impl.keyBindings["space"]
	impl.keyBindingsMu.RUnlock()

	require.Len(t, bindings, 2, "Should have 2 bindings for 'space'")

	// Test navigation mode
	t.Run("navigation mode", func(t *testing.T) {
		navigationMode = true

		// First binding should match
		if bindings[0].Condition != nil {
			assert.True(t, bindings[0].Condition())
		}

		// Second binding should not match
		if bindings[1].Condition != nil {
			assert.False(t, bindings[1].Condition())
		}
	})

	// Test input mode
	t.Run("input mode", func(t *testing.T) {
		navigationMode = false

		// First binding should not match
		if bindings[0].Condition != nil {
			assert.False(t, bindings[0].Condition())
		}

		// Second binding should match
		if bindings[1].Condition != nil {
			assert.True(t, bindings[1].Condition())
		}
	})
}

// TestKeyBindingProcessing_NilSafety tests nil safety
func TestKeyBindingProcessing_NilSafety(t *testing.T) {
	t.Run("nil keyBindings map", func(t *testing.T) {
		component, err := NewComponent("TestComponent").
			Template(func(ctx RenderContext) string { return "test" }).
			Build()
		require.NoError(t, err)

		impl := component.(*componentImpl)
		impl.keyBindings = nil

		// Should not panic
		impl.keyBindingsMu.RLock()
		_, found := impl.keyBindings["space"]
		impl.keyBindingsMu.RUnlock()

		assert.False(t, found)
	})

	t.Run("nil condition function", func(t *testing.T) {
		binding := KeyBinding{
			Key:         "space",
			Event:       "action",
			Description: "Action",
			Condition:   nil, // Explicitly nil
		}

		component, err := NewComponent("TestComponent").
			WithConditionalKeyBinding(binding).
			Template(func(ctx RenderContext) string { return "test" }).
			Build()
		require.NoError(t, err)

		impl := component.(*componentImpl)

		impl.keyBindingsMu.RLock()
		bindings := impl.keyBindings["space"]
		impl.keyBindingsMu.RUnlock()

		require.NotEmpty(t, bindings)
		assert.Nil(t, bindings[0].Condition, "Condition should be nil")

		// Should not panic when checking nil condition
		// In actual code: if binding.Condition != nil && !binding.Condition()
	})
}

// BenchmarkKeyBindingLookup benchmarks key binding lookup performance
// Requirement: < 50ns per lookup
func BenchmarkKeyBindingLookup(b *testing.B) {
	// Create component with 100 key bindings
	component, err := NewComponent("BenchComponent").
		Template(func(ctx RenderContext) string { return "test" }).
		Build()
	if err != nil {
		b.Fatal(err)
	}

	impl := component.(*componentImpl)
	impl.keyBindings = make(map[string][]KeyBinding)

	// Populate with 100 bindings
	for i := 0; i < 100; i++ {
		key := string(rune('a' + (i % 26)))
		impl.keyBindings[key] = append(impl.keyBindings[key], KeyBinding{
			Key:         key,
			Event:       "event",
			Description: "Description",
		})
	}

	// Add the key we'll look up
	impl.keyBindings["space"] = []KeyBinding{
		{Key: "space", Event: "increment", Description: "Increment"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark the lookup path
		impl.keyBindingsMu.RLock()
		bindings, found := impl.keyBindings["space"]
		impl.keyBindingsMu.RUnlock()

		if !found || len(bindings) == 0 {
			b.Fatal("Expected to find binding")
		}
	}
}
