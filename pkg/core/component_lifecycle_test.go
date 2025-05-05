package core

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestLifecycleMethods tests the lifecycle methods of components
func TestLifecycleMethods(t *testing.T) {
	t.Run("Lifecycle Method Order", func(t *testing.T) {
		// Test that lifecycle methods are called in the correct order
		executionOrder := []string{}

		// Create a component with custom lifecycle hooks to track execution order
		component := CreateComponent(
			"lifecycle-test",
			WithInit(func(c Component) error {
				executionOrder = append(executionOrder, "init")
				return nil
			}),
			WithDispose(func(c Component) error {
				executionOrder = append(executionOrder, "dispose")
				return nil
			}),
		)

		// Execute lifecycle methods
		err := component.Initialize()
		assert.NoError(t, err, "Initialize should not return error")

		// Execute a few updates
		_, err = component.Update(tea.KeyMsg{Type: tea.KeyEnter})
		assert.NoError(t, err, "Update should not return error")

		// Render the component
		component.Render()

		// Dispose the component
		err = component.Dispose()
		assert.NoError(t, err, "Dispose should not return error")

		// Check the execution order
		expected := []string{"init", "dispose"}
		assert.Equal(t, expected, executionOrder, "Lifecycle methods should be called in the correct order")
	})

	t.Run("Lifecycle Error Propagation", func(t *testing.T) {
		// Test that errors in lifecycle methods are properly propagated
		initErr := errors.New("init error")
		disposeErr := errors.New("dispose error")

		// Create a component with lifecycle methods that return errors
		component := CreateComponent(
			"error-test",
			WithInit(func(c Component) error {
				return initErr
			}),
			WithDispose(func(c Component) error {
				return disposeErr
			}),
		)

		// Test Init error propagation
		err := component.Initialize()
		assert.Equal(t, initErr, err, "Initialize should propagate the error")

		// Test Dispose error propagation
		err = component.Dispose()
		assert.Equal(t, disposeErr, err, "Dispose should propagate the error")
	})
}

// TestStatefulLifecycleMethods tests the lifecycle methods of stateful components
func TestStatefulLifecycleMethods(t *testing.T) {
	t.Run("Stateful Component Lifecycle", func(t *testing.T) {
		// Create a stateful component
		component := CreateStatefulComponent(
			"stateful-lifecycle",
			"Stateful Lifecycle Test",
		)

		// Initialize the component
		assert.False(t, component.IsMounted(), "Component should not be mounted before initialization")
		err := component.Initialize()
		assert.NoError(t, err, "Initialize should not return error")
		assert.True(t, component.IsMounted(), "Component should be mounted after initialization")

		// Cleanup and test unmounting
		err = component.Dispose()
		assert.NoError(t, err, "Dispose should not return error")
		assert.False(t, component.IsMounted(), "Component should be unmounted after disposal")
	})

	t.Run("Lifecycle Method with State Changes", func(t *testing.T) {
		// Create a stateful component with state
		component := CreateStatefulComponent(
			"lifecycle-with-state",
			"Lifecycle With State Test",
		)

		// Add a state value
		counter, setCounter, _ := WithState(component, "counter", 0)
		assert.Equal(t, 0, counter.Get(), "Initial counter value should be 0")

		// Use direct mount hook rather than effect
		effectExecuted := false
		component.GetState().GetHookManager().OnMount(func() error {
			effectExecuted = true
			setCounter(counter.Get() + 1)
			return nil
		})

		// Initialize component to trigger mount hooks
		err := component.Initialize()
		assert.NoError(t, err, "Initialize should not return error")

		// No need to execute hooks manually, Initialize should have done it
		assert.True(t, effectExecuted, "Mount effect should be executed")
		assert.Equal(t, 1, counter.Get(), "Counter should be incremented by effect")

		// Test update with update hook
		updateCalled := false
		component.GetState().GetHookManager().OnUpdate(func(prevDeps []interface{}) error {
			updateCalled = true
			setCounter(counter.Get() + 1)
			return nil
		}, []interface{}{"trigger-update"})

		// Execute update hooks
		err = component.ExecuteEffect()
		assert.NoError(t, err, "ExecuteEffect should not return error")
		assert.True(t, updateCalled, "Update effect should be executed")
		assert.Equal(t, 2, counter.Get(), "Counter should be incremented by update effect")
	})
}

// TestNestedLifecycle tests lifecycle method propagation in nested components
func TestNestedLifecycle(t *testing.T) {
	t.Run("Nested Component Lifecycle", func(t *testing.T) {
		// Create a tree of components
		parentEvents := []string{}
		child1Events := []string{}
		child2Events := []string{}

		// Create parent with event recording
		parent := CreateComponent(
			"parent",
			WithInit(func(c Component) error {
				parentEvents = append(parentEvents, "init")
				return nil
			}),
			WithDispose(func(c Component) error {
				parentEvents = append(parentEvents, "dispose")
				return nil
			}),
		)

		// Create first child
		child1 := CreateComponent(
			"child1",
			WithInit(func(c Component) error {
				child1Events = append(child1Events, "init")
				return nil
			}),
			WithDispose(func(c Component) error {
				child1Events = append(child1Events, "dispose")
				return nil
			}),
		)

		// Create second child
		child2 := CreateComponent(
			"child2",
			WithInit(func(c Component) error {
				child2Events = append(child2Events, "init")
				return nil
			}),
			WithDispose(func(c Component) error {
				child2Events = append(child2Events, "dispose")
				return nil
			}),
		)

		// Build component tree
		parent.AddChild(child1)
		parent.AddChild(child2)

		// Initialize entire tree
		err := parent.Initialize()
		assert.NoError(t, err, "Initialize should not return error")

		// Check init order
		assert.Equal(t, []string{"init"}, parentEvents, "Parent init should be called")
		assert.Equal(t, []string{"init"}, child1Events, "Child1 init should be called")
		assert.Equal(t, []string{"init"}, child2Events, "Child2 init should be called")

		// Dispose entire tree
		err = parent.Dispose()
		assert.NoError(t, err, "Dispose should not return error")

		// Check dispose order
		assert.Equal(t, []string{"init", "dispose"}, parentEvents, "Parent dispose should be called")
		assert.Equal(t, []string{"init", "dispose"}, child1Events, "Child1 dispose should be called")
		assert.Equal(t, []string{"init", "dispose"}, child2Events, "Child2 dispose should be called")
	})

	t.Run("Child Error Bubbling", func(t *testing.T) {
		// Test that errors in child components bubble up to parent
		expectedError := errors.New("child error")

		// Create parent component
		parent := CreateComponent("parent")

		// Create child with error in init
		child := CreateComponent(
			"child-with-error",
			WithInit(func(c Component) error {
				return expectedError
			}),
		)

		// Add child to parent
		parent.AddChild(child)

		// Initialize should propagate child error
		err := parent.Initialize()
		assert.Error(t, err, "Error should propagate from child to parent")
		assert.True(t, errors.Is(err, expectedError) || err.Error() == "error initializing child child-with-error: child error", "Should contain original error")
	})
}

// TestComponentLifecycleConsistency tests that lifecycle methods maintain consistent state
func TestComponentLifecycleConsistency(t *testing.T) {
	t.Run("State Consistency During Lifecycle", func(t *testing.T) {
		// Create a stateful component
		component := CreateStatefulComponent(
			"consistency-test",
			"Consistency Test",
		)

		// Create a state to track component status
		status, setStatus, _ := WithState(component, "status", "created")

		// Track when cleanup runs and status values during different phases
		cleanupRun := false
		statusAfterInit := ""
		statusBeforeCleanup := ""
		statusAfterCleanup := ""

		// Setup direct unmount hook rather than an effect with cleanup, which is more reliable
		component.GetState().GetHookManager().OnUnmount(func() error {
			cleanupRun = true
			statusBeforeCleanup = status.Get()
			setStatus("disposed")
			statusAfterCleanup = status.Get()
			return nil
		})

		// Initialize component - this triggers mount hooks
		err := component.Initialize()
		assert.NoError(t, err, "Initialize should not return error")

		// Update status after initialization
		setStatus("initialized")
		statusAfterInit = status.Get()

		// Verify the status was updated
		assert.Equal(t, "initialized", statusAfterInit, "Status should be updated after initialization")

		// Dispose component - this should trigger the unmount hook
		err = component.Dispose()
		assert.NoError(t, err, "Dispose should not return error")

		// Verify the cleanup ran and updated the status
		assert.True(t, cleanupRun, "Unmount hook should have run")
		assert.Equal(t, "initialized", statusBeforeCleanup, "Status should be 'initialized' before cleanup")
		assert.Equal(t, "disposed", statusAfterCleanup, "Status should be 'disposed' after cleanup")
	})
}
