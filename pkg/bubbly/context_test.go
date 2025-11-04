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
			assert.Equal(t, tt.expectedValue, ref.GetTyped(), "Ref value should match initial value")
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
			assert.Equal(t, tt.expectedValue, computed.GetTyped(), "Computed value should match expected")
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
					assert.Equal(t, ref.GetTyped(), gotRef.GetTyped(), "Ref values should match")
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
			return count.GetTyped().(int) * 2
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
			current := count.GetTyped().(int)
			count.Set(current + 1)
		})

		// Assert initial state
		assert.Equal(t, 0, count.GetTyped(), "Initial count should be 0")
		assert.Equal(t, 0, doubled.GetTyped(), "Initial doubled should be 0")

		// Trigger event
		ctx.Emit("increment", nil)

		// Assert after event
		assert.True(t, incrementCalled, "Increment handler should be called")
		assert.Equal(t, 1, count.GetTyped(), "Count should be 1 after increment")
		assert.Equal(t, 2, doubled.GetTyped(), "Doubled should be 2 after increment")

		// Verify watcher was called
		assert.Equal(t, 0, watchOldVal, "Watch should receive old value 0")
		assert.Equal(t, 1, watchNewVal, "Watch should receive new value 1")

		// Verify exposed values
		exposedCount := ctx.Get("count").(*Ref[interface{}])
		assert.Equal(t, 1, exposedCount.GetTyped(), "Exposed count should match")
	})
}

// TestContext_Provide tests that Context.Provide stores values
func TestContext_Provide(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{
			name:  "provide string value",
			key:   "theme",
			value: "dark",
		},
		{
			name:  "provide int value",
			key:   "count",
			value: 42,
		},
		{
			name:  "provide ref value",
			key:   "user",
			value: NewRef[interface{}]("John"),
		},
		{
			name:  "provide nil value",
			key:   "optional",
			value: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := newComponentImpl("TestComponent")
			ctx := &Context{component: c}

			// Act
			ctx.Provide(tt.key, tt.value)

			// Assert
			c.providesMu.RLock()
			stored, exists := c.provides[tt.key]
			c.providesMu.RUnlock()

			assert.True(t, exists, "Key should exist in provides map")
			assert.Equal(t, tt.value, stored, "Stored value should match provided value")
		})
	}
}

// TestContext_Provide_Overwrite tests that providing same key overwrites
func TestContext_Provide_Overwrite(t *testing.T) {
	// Arrange
	c := newComponentImpl("TestComponent")
	ctx := &Context{component: c}

	// Act
	ctx.Provide("theme", "dark")
	ctx.Provide("theme", "light") // Overwrite

	// Assert
	c.providesMu.RLock()
	value := c.provides["theme"]
	c.providesMu.RUnlock()

	assert.Equal(t, "light", value, "Second provide should overwrite first")
}

// TestContext_Inject_FromSelf tests injecting from same component
func TestContext_Inject_FromSelf(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		providedVal interface{}
		defaultVal  interface{}
		expectedVal interface{}
	}{
		{
			name:        "inject provided string",
			key:         "theme",
			providedVal: "dark",
			defaultVal:  "light",
			expectedVal: "dark",
		},
		{
			name:        "inject provided ref",
			key:         "count",
			providedVal: NewRef[interface{}](42),
			defaultVal:  NewRef[interface{}](0),
			expectedVal: NewRef[interface{}](42),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := newComponentImpl("TestComponent")
			ctx := &Context{component: c}
			ctx.Provide(tt.key, tt.providedVal)

			// Act
			result := ctx.Inject(tt.key, tt.defaultVal)

			// Assert
			assert.Equal(t, tt.providedVal, result, "Should return provided value")
		})
	}
}

// TestContext_Inject_NotFound tests inject with no provider
func TestContext_Inject_NotFound(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		defaultVal interface{}
	}{
		{
			name:       "inject with string default",
			key:        "theme",
			defaultVal: "light",
		},
		{
			name:       "inject with int default",
			key:        "count",
			defaultVal: 0,
		},
		{
			name:       "inject with nil default",
			key:        "optional",
			defaultVal: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := newComponentImpl("TestComponent")
			ctx := &Context{component: c}

			// Act
			result := ctx.Inject(tt.key, tt.defaultVal)

			// Assert
			assert.Equal(t, tt.defaultVal, result, "Should return default when not found")
		})
	}
}

// TestContext_Inject_FromParent tests inject from parent component
func TestContext_Inject_FromParent(t *testing.T) {
	// Arrange
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")
	child.parent = parent

	parentCtx := &Context{component: parent}
	childCtx := &Context{component: child}

	// Parent provides value
	parentCtx.Provide("theme", "dark")

	// Act
	result := childCtx.Inject("theme", "light")

	// Assert
	assert.Equal(t, "dark", result, "Child should receive parent's provided value")
}

// TestContext_Inject_NearestWins tests that nearest provider wins
func TestContext_Inject_NearestWins(t *testing.T) {
	// Arrange
	grandparent := newComponentImpl("Grandparent")
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")

	parent.parent = grandparent
	child.parent = parent

	grandparentCtx := &Context{component: grandparent}
	parentCtx := &Context{component: parent}
	childCtx := &Context{component: child}

	// Both provide same key with different values
	grandparentCtx.Provide("theme", "dark")
	parentCtx.Provide("theme", "light")

	// Act
	result := childCtx.Inject("theme", "default")

	// Assert
	assert.Equal(t, "light", result, "Nearest provider (parent) should win over grandparent")
}

// TestContext_Inject_DeepTree tests inject across multiple levels
func TestContext_Inject_DeepTree(t *testing.T) {
	// Arrange - Create 4-level tree
	root := newComponentImpl("Root")
	level1 := newComponentImpl("Level1")
	level2 := newComponentImpl("Level2")
	level3 := newComponentImpl("Level3")

	level1.parent = root
	level2.parent = level1
	level3.parent = level2

	rootCtx := &Context{component: root}
	level3Ctx := &Context{component: level3}

	// Root provides value
	rootCtx.Provide("config", "production")

	// Act - Level3 injects
	result := level3Ctx.Inject("config", "development")

	// Assert
	assert.Equal(t, "production", result, "Should find value from root through deep tree")
}

// TestContext_Inject_MultipleKeys tests multiple provide/inject keys
func TestContext_Inject_MultipleKeys(t *testing.T) {
	// Arrange
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")
	child.parent = parent

	parentCtx := &Context{component: parent}
	childCtx := &Context{component: child}

	// Parent provides multiple values
	parentCtx.Provide("theme", "dark")
	parentCtx.Provide("user", "John")
	parentCtx.Provide("count", 42)

	// Act & Assert
	assert.Equal(t, "dark", childCtx.Inject("theme", "light"))
	assert.Equal(t, "John", childCtx.Inject("user", "Guest"))
	assert.Equal(t, 42, childCtx.Inject("count", 0))
	assert.Equal(t, "default", childCtx.Inject("missing", "default"))
}

// TestContext_Inject_ReactiveValues tests providing/injecting reactive refs
func TestContext_Inject_ReactiveValues(t *testing.T) {
	// Arrange
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")
	child.parent = parent

	parentCtx := &Context{component: parent}
	childCtx := &Context{component: child}

	// Parent provides a Ref
	themeRef := NewRef[interface{}]("dark")
	parentCtx.Provide("theme", themeRef)

	// Act
	injectedRef := childCtx.Inject("theme", NewRef[interface{}]("light")).(*Ref[interface{}])

	// Assert
	assert.Equal(t, themeRef, injectedRef, "Should inject same Ref instance")
	assert.Equal(t, "dark", injectedRef.GetTyped(), "Injected ref should have correct value")

	// Modify parent's ref
	themeRef.Set("light")

	// Child should see the change (same ref instance)
	assert.Equal(t, "light", injectedRef.GetTyped(), "Child should see reactive changes")
}

// TestContext_Ref_AutoCommandsDisabled tests that Context.Ref creates standard Ref when auto commands disabled
func TestContext_Ref_AutoCommandsDisabled(t *testing.T) {
	tests := []struct {
		name         string
		initialValue interface{}
		newValue     interface{}
	}{
		{
			name:         "int ref without auto commands",
			initialValue: 0,
			newValue:     42,
		},
		{
			name:         "string ref without auto commands",
			initialValue: "hello",
			newValue:     "world",
		},
		{
			name:         "bool ref without auto commands",
			initialValue: false,
			newValue:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := newComponentImpl("TestComponent")
			c.autoCommands = false // Explicitly disabled (default)
			ctx := &Context{component: c}

			// Act
			ref := ctx.Ref(tt.initialValue)
			ref.Set(tt.newValue)

			// Assert
			assert.Equal(t, tt.newValue, ref.Get(), "Ref should update value")
			// Command queue should be nil when auto commands disabled (Task 2.5)
			assert.Nil(t, c.commandQueue, "Command queue should be nil when auto mode disabled")
		})
	}
}

// TestContext_Ref_AutoCommandsEnabled tests that Context.Ref creates CommandRef when auto commands enabled
func TestContext_Ref_AutoCommandsEnabled(t *testing.T) {
	tests := []struct {
		name         string
		initialValue interface{}
		newValue     interface{}
	}{
		{
			name:         "int ref with auto commands",
			initialValue: 0,
			newValue:     42,
		},
		{
			name:         "string ref with auto commands",
			initialValue: "hello",
			newValue:     "world",
		},
		{
			name:         "bool ref with auto commands",
			initialValue: false,
			newValue:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := newComponentImpl("TestComponent")
			c.autoCommands = true // Enable auto commands
			// Initialize command infrastructure (Task 2.5)
			c.commandQueue = NewCommandQueue()
			c.commandGen = &defaultCommandGenerator{}
			ctx := &Context{component: c}

			// Act
			ref := ctx.Ref(tt.initialValue)
			ref.Set(tt.newValue)

			// Assert
			assert.Equal(t, tt.newValue, ref.Get(), "Ref should update value synchronously")
			assert.Equal(t, 1, c.commandQueue.Len(), "Command should be generated when auto mode enabled")

			// Verify command generates correct message
			cmds := c.commandQueue.DrainAll()
			require.Len(t, cmds, 1, "Should have exactly one command")

			msg := cmds[0]()
			stateMsg, ok := msg.(StateChangedMsg)
			require.True(t, ok, "Command should return StateChangedMsg")
			assert.Equal(t, c.id, stateMsg.ComponentID, "Message should have correct component ID")
			assert.Equal(t, tt.initialValue, stateMsg.OldValue, "Message should have correct old value")
			assert.Equal(t, tt.newValue, stateMsg.NewValue, "Message should have correct new value")
		})
	}
}

// TestContext_Ref_AutoCommandsEnabled_MultipleChanges tests command batching with multiple Set() calls
func TestContext_Ref_AutoCommandsEnabled_MultipleChanges(t *testing.T) {
	// Arrange
	c := newComponentImpl("TestComponent")
	c.autoCommands = true
	// Initialize command infrastructure (Task 2.5)
	c.commandQueue = NewCommandQueue()
	c.commandGen = &defaultCommandGenerator{}
	ctx := &Context{component: c}

	// Act
	ref1 := ctx.Ref(0)
	ref2 := ctx.Ref("hello")
	ref3 := ctx.Ref(false)

	ref1.Set(1)
	ref2.Set("world")
	ref3.Set(true)

	// Assert
	assert.Equal(t, 3, c.commandQueue.Len(), "Should have 3 commands queued")

	// Verify all commands
	cmds := c.commandQueue.DrainAll()
	require.Len(t, cmds, 3, "Should have exactly 3 commands")

	// Execute commands and verify messages
	for i, cmd := range cmds {
		msg := cmd()
		stateMsg, ok := msg.(StateChangedMsg)
		require.True(t, ok, "Command %d should return StateChangedMsg", i)
		assert.Equal(t, c.id, stateMsg.ComponentID, "Message %d should have correct component ID", i)
	}
}

// TestContext_Ref_AutoCommandsEnabled_ThreadSafe tests concurrent Ref creation and updates
func TestContext_Ref_AutoCommandsEnabled_ThreadSafe(t *testing.T) {
	// Arrange
	c := newComponentImpl("TestComponent")
	c.autoCommands = true
	// Initialize command infrastructure (Task 2.5)
	c.commandQueue = NewCommandQueue()
	c.commandGen = &defaultCommandGenerator{}
	ctx := &Context{component: c}

	// Act - Create and update refs concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(val int) {
			ref := ctx.Ref(val)
			ref.Set(val * 2)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Assert
	assert.Equal(t, 10, c.commandQueue.Len(), "Should have 10 commands queued")
	assert.NotPanics(t, func() {
		c.commandQueue.DrainAll()
	}, "Should handle concurrent access without panicking")
}

// TestContext_Ref_BackwardCompatibility tests that existing code works without auto commands
func TestContext_Ref_BackwardCompatibility(t *testing.T) {
	// Arrange - Component without auto commands (default)
	c := newComponentImpl("TestComponent")
	ctx := &Context{component: c}

	// Act - Use Ref as before
	count := ctx.Ref(0)
	count.Set(1)
	count.Set(2)
	count.Set(3)

	// Assert
	assert.Equal(t, 3, count.Get(), "Ref should work normally")
	// Command queue should be nil when auto commands disabled (Task 2.5)
	assert.Nil(t, c.commandQueue, "Command queue should be nil by default")
}

// TestContext_Ref_AutoCommandsEnabled_CommandExecution tests that commands execute correctly
func TestContext_Ref_AutoCommandsEnabled_CommandExecution(t *testing.T) {
	// Arrange
	c := newComponentImpl("TestComponent")
	c.autoCommands = true
	// Initialize command infrastructure (Task 2.5)
	c.commandQueue = NewCommandQueue()
	c.commandGen = &defaultCommandGenerator{}
	ctx := &Context{component: c}

	// Act
	ref := ctx.Ref(10)
	ref.Set(20)

	// Get command
	cmds := c.commandQueue.DrainAll()
	require.Len(t, cmds, 1, "Should have one command")

	// Execute command
	msg := cmds[0]()

	// Assert
	stateMsg, ok := msg.(StateChangedMsg)
	require.True(t, ok, "Should return StateChangedMsg")
	assert.Equal(t, c.id, stateMsg.ComponentID)
	assert.Equal(t, 10, stateMsg.OldValue)
	assert.Equal(t, 20, stateMsg.NewValue)
	assert.NotZero(t, stateMsg.Timestamp, "Timestamp should be set")
}

// TestContext_Ref_AutoCommandsEnabled_RefIDUnique tests that each ref gets unique ID
func TestContext_Ref_AutoCommandsEnabled_RefIDUnique(t *testing.T) {
	// Arrange
	c := newComponentImpl("TestComponent")
	c.autoCommands = true
	// Initialize command infrastructure (Task 2.5)
	c.commandQueue = NewCommandQueue()
	c.commandGen = &defaultCommandGenerator{}
	ctx := &Context{component: c}

	// Act
	ref1 := ctx.Ref(1)
	ref2 := ctx.Ref(2)
	ref3 := ctx.Ref(3)

	ref1.Set(10)
	ref2.Set(20)
	ref3.Set(30)

	// Assert
	cmds := c.commandQueue.DrainAll()
	require.Len(t, cmds, 3, "Should have 3 commands")

	// Collect ref IDs
	refIDs := make(map[string]bool)
	for _, cmd := range cmds {
		msg := cmd().(StateChangedMsg)
		refIDs[msg.RefID] = true
	}

	assert.Len(t, refIDs, 3, "Each ref should have unique ID")
}
