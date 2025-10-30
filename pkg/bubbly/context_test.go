package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContext_Ref tests that Context.Ref creates reactive references
func TestContext_Ref(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  interface{}
		expectedValue interface{}
	}{
		{
			name:          "create ref with int",
			initialValue:  42,
			expectedValue: 42,
		},
		{
			name:          "create ref with string",
			initialValue:  "hello",
			expectedValue: "hello",
		},
		{
			name:          "create ref with bool",
			initialValue:  true,
			expectedValue: true,
		},
		{
			name:          "create ref with nil",
			initialValue:  nil,
			expectedValue: nil,
		},
		{
			name:          "create ref with struct",
			initialValue:  struct{ X int }{X: 10},
			expectedValue: struct{ X int }{X: 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := &componentImpl{
				name:  "TestComponent",
				state: make(map[string]interface{}),
			}
			ctx := &Context{component: c}

			// Act
			ref := ctx.Ref(tt.initialValue)

			// Assert
			require.NotNil(t, ref, "Ref should not be nil")
			assert.Equal(t, tt.expectedValue, ref.Get(), "Ref value should match initial value")
		})
	}
}

// TestContext_Computed tests that Context.Computed creates computed values
func TestContext_Computed(t *testing.T) {
	tests := []struct {
		name          string
		computeFn     func() interface{}
		expectedValue interface{}
	}{
		{
			name: "computed returns constant",
			computeFn: func() interface{} {
				return 42
			},
			expectedValue: 42,
		},
		{
			name: "computed returns string",
			computeFn: func() interface{} {
				return "computed"
			},
			expectedValue: "computed",
		},
		{
			name: "computed returns calculation",
			computeFn: func() interface{} {
				return 10 * 2
			},
			expectedValue: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := &componentImpl{
				name:  "TestComponent",
				state: make(map[string]interface{}),
			}
			ctx := &Context{component: c}

			// Act
			computed := ctx.Computed(tt.computeFn)

			// Assert
			require.NotNil(t, computed, "Computed should not be nil")
			assert.Equal(t, tt.expectedValue, computed.Get(), "Computed value should match expected")
		})
	}
}

// TestContext_Watch tests that Context.Watch registers watchers
func TestContext_Watch(t *testing.T) {
	t.Run("watch ref changes", func(t *testing.T) {
		// Arrange
		c := &componentImpl{
			name:  "TestComponent",
			state: make(map[string]interface{}),
		}
		ctx := &Context{component: c}
		ref := ctx.Ref(0)

		callCount := 0
		var oldVal, newVal interface{}

		// Act
		ctx.Watch(ref, func(new, old interface{}) {
			callCount++
			newVal = new
			oldVal = old
		})

		ref.Set(42)

		// Assert
		assert.Equal(t, 1, callCount, "Watch callback should be called once")
		assert.Equal(t, 0, oldVal, "Old value should be 0")
		assert.Equal(t, 42, newVal, "New value should be 42")
	})

	t.Run("watch multiple changes", func(t *testing.T) {
		// Arrange
		c := &componentImpl{
			name:  "TestComponent",
			state: make(map[string]interface{}),
		}
		ctx := &Context{component: c}
		ref := ctx.Ref("initial")

		callCount := 0

		// Act
		ctx.Watch(ref, func(new, old interface{}) {
			callCount++
		})

		ref.Set("first")
		ref.Set("second")
		ref.Set("third")

		// Assert
		assert.Equal(t, 3, callCount, "Watch callback should be called three times")
	})
}

// TestContext_Expose tests that Context.Expose stores values in component state
func TestContext_Expose(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{
			name:  "expose ref",
			key:   "count",
			value: NewRef(42),
		},
		{
			name:  "expose string",
			key:   "name",
			value: "test",
		},
		{
			name:  "expose computed",
			key:   "doubled",
			value: NewComputed(func() interface{} { return 84 }),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := &componentImpl{
				name:  "TestComponent",
				state: make(map[string]interface{}),
			}
			ctx := &Context{component: c}

			// Act
			ctx.Expose(tt.key, tt.value)

			// Assert
			assert.Contains(t, c.state, tt.key, "State should contain exposed key")
			assert.Equal(t, tt.value, c.state[tt.key], "Exposed value should match")
		})
	}
}

// TestContext_Get tests that Context.Get retrieves exposed values
func TestContext_Get(t *testing.T) {
	tests := []struct {
		name          string
		setupState    map[string]interface{}
		key           string
		expectedValue interface{}
	}{
		{
			name: "get existing value",
			setupState: map[string]interface{}{
				"count": NewRef(42),
			},
			key:           "count",
			expectedValue: NewRef(42),
		},
		{
			name:          "get non-existent value",
			setupState:    map[string]interface{}{},
			key:           "missing",
			expectedValue: nil,
		},
		{
			name: "get string value",
			setupState: map[string]interface{}{
				"name": "test",
			},
			key:           "name",
			expectedValue: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := &componentImpl{
				name:  "TestComponent",
				state: tt.setupState,
			}
			ctx := &Context{component: c}

			// Act
			value := ctx.Get(tt.key)

			// Assert
			if tt.expectedValue == nil {
				assert.Nil(t, value, "Get should return nil for non-existent key")
			} else {
				// For Ref and Computed, compare the actual values
				if ref, ok := tt.expectedValue.(*Ref[interface{}]); ok {
					gotRef, ok := value.(*Ref[interface{}])
					require.True(t, ok, "Value should be a Ref")
					assert.Equal(t, ref.Get(), gotRef.Get(), "Ref values should match")
				} else {
					assert.Equal(t, tt.expectedValue, value, "Get should return exposed value")
				}
			}
		})
	}
}

// TestContext_On tests that Context.On registers event handlers
func TestContext_On(t *testing.T) {
	t.Run("register event handler", func(t *testing.T) {
		// Arrange
		c := &componentImpl{
			name:     "TestComponent",
			state:    make(map[string]interface{}),
			handlers: make(map[string][]EventHandler),
		}
		ctx := &Context{component: c}

		called := false
		var receivedData interface{}

		// Act
		ctx.On("test-event", func(data interface{}) {
			called = true
			// Handlers receive data payload directly
			receivedData = data
		})

		// Trigger the event
		c.Emit("test-event", "test-data")

		// Assert
		assert.True(t, called, "Event handler should be called")
		assert.Equal(t, "test-data", receivedData, "Handler should receive event data")
	})

	t.Run("register multiple handlers", func(t *testing.T) {
		// Arrange
		c := &componentImpl{
			name:     "TestComponent",
			state:    make(map[string]interface{}),
			handlers: make(map[string][]EventHandler),
		}
		ctx := &Context{component: c}

		callCount := 0

		// Act
		ctx.On("event", func(data interface{}) { callCount++ })
		ctx.On("event", func(data interface{}) { callCount++ })
		ctx.On("event", func(data interface{}) { callCount++ })

		c.Emit("event", nil)

		// Assert
		assert.Equal(t, 3, callCount, "All handlers should be called")
	})
}

// TestContext_Emit tests that Context.Emit triggers event handlers
func TestContext_Emit(t *testing.T) {
	t.Run("emit event with data", func(t *testing.T) {
		// Arrange
		c := &componentImpl{
			name:     "TestComponent",
			state:    make(map[string]interface{}),
			handlers: make(map[string][]EventHandler),
		}
		ctx := &Context{component: c}

		var receivedData interface{}
		ctx.On("submit", func(data interface{}) {
			// Handlers receive data payload directly
			receivedData = data
		})

		// Act
		ctx.Emit("submit", map[string]string{"key": "value"})

		// Assert
		expected := map[string]string{"key": "value"}
		assert.Equal(t, expected, receivedData, "Emitted data should be received")
	})

	t.Run("emit event without handlers", func(t *testing.T) {
		// Arrange
		c := &componentImpl{
			name:     "TestComponent",
			state:    make(map[string]interface{}),
			handlers: make(map[string][]EventHandler),
		}
		ctx := &Context{component: c}

		// Act & Assert (should not panic)
		assert.NotPanics(t, func() {
			ctx.Emit("no-handler", "data")
		}, "Emit should not panic when no handlers registered")
	})
}

// TestContext_Props tests that Context.Props returns component props
func TestContext_Props(t *testing.T) {
	tests := []struct {
		name          string
		props         interface{}
		expectedProps interface{}
	}{
		{
			name: "get struct props",
			props: struct {
				Label string
			}{Label: "Button"},
			expectedProps: struct {
				Label string
			}{Label: "Button"},
		},
		{
			name:          "get string props",
			props:         "simple-props",
			expectedProps: "simple-props",
		},
		{
			name:          "get nil props",
			props:         nil,
			expectedProps: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := &componentImpl{
				name:  "TestComponent",
				props: tt.props,
				state: make(map[string]interface{}),
			}
			ctx := &Context{component: c}

			// Act
			props := ctx.Props()

			// Assert
			assert.Equal(t, tt.expectedProps, props, "Props should match component props")
		})
	}
}

// TestContext_Children tests that Context.Children returns child components
func TestContext_Children(t *testing.T) {
	t.Run("get children", func(t *testing.T) {
		// Arrange
		child1 := &componentImpl{name: "Child1", state: make(map[string]interface{})}
		child2 := &componentImpl{name: "Child2", state: make(map[string]interface{})}

		c := &componentImpl{
			name:     "Parent",
			state:    make(map[string]interface{}),
			children: []Component{child1, child2},
		}
		ctx := &Context{component: c}

		// Act
		children := ctx.Children()

		// Assert
		require.Len(t, children, 2, "Should have 2 children")
		assert.Equal(t, "Child1", children[0].Name(), "First child name should match")
		assert.Equal(t, "Child2", children[1].Name(), "Second child name should match")
	})

	t.Run("get empty children", func(t *testing.T) {
		// Arrange
		c := &componentImpl{
			name:     "Parent",
			state:    make(map[string]interface{}),
			children: []Component{},
		}
		ctx := &Context{component: c}

		// Act
		children := ctx.Children()

		// Assert
		assert.Empty(t, children, "Children should be empty")
	})
}

// TestContext_Integration tests full workflow with Context
func TestContext_Integration(t *testing.T) {
	t.Run("complete setup workflow", func(t *testing.T) {
		// Arrange
		c := &componentImpl{
			name:     "Counter",
			state:    make(map[string]interface{}),
			handlers: make(map[string][]EventHandler),
		}
		ctx := &Context{component: c}

		// Act - Simulate setup function
		count := ctx.Ref(0)
		doubled := ctx.Computed(func() interface{} {
			return count.Get().(int) * 2
		})

		ctx.Expose("count", count)
		ctx.Expose("doubled", doubled)

		var watchOldVal, watchNewVal interface{}
		ctx.Watch(count, func(new, old interface{}) {
			watchNewVal = new
			watchOldVal = old
		})

		incrementCalled := false
		ctx.On("increment", func(data interface{}) {
			incrementCalled = true
			current := count.Get().(int)
			count.Set(current + 1)
		})

		// Assert initial state
		assert.Equal(t, 0, count.Get(), "Initial count should be 0")
		assert.Equal(t, 0, doubled.Get(), "Initial doubled should be 0")

		// Trigger event
		ctx.Emit("increment", nil)

		// Assert after event
		assert.True(t, incrementCalled, "Increment handler should be called")
		assert.Equal(t, 1, count.Get(), "Count should be 1 after increment")
		assert.Equal(t, 2, doubled.Get(), "Doubled should be 2 after increment")

		// Verify watcher was called
		assert.Equal(t, 0, watchOldVal, "Watch should receive old value 0")
		assert.Equal(t, 1, watchNewVal, "Watch should receive new value 1")

		// Verify exposed values
		exposedCount := ctx.Get("count").(*Ref[interface{}])
		assert.Equal(t, 1, exposedCount.Get(), "Exposed count should match")
	})
}
