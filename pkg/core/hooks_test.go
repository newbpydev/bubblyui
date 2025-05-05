package core

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHookManager(t *testing.T) {
	hm := NewHookManager("TestComponent")
	assert.NotNil(t, hm)
	assert.Equal(t, "TestComponent", hm.componentName)
	assert.Equal(t, 0, hm.nextHookID)
	assert.Empty(t, hm.mountHooks)
	assert.Empty(t, hm.updateHooks)
	assert.Empty(t, hm.unmountHooks)
}

func TestOnMount(t *testing.T) {
	hm := NewHookManager("TestComponent")
	executionCount := 0
	
	hookID := hm.OnMount(func() error {
		executionCount++
		return nil
	})
	
	assert.NotEmpty(t, hookID)
	assert.Contains(t, hm.mountHooks, hookID)
	assert.False(t, hm.mountHooks[hookID].executed)
	
	// Test execution
	err := hm.ExecuteMountHooks()
	assert.NoError(t, err)
	assert.Equal(t, 1, executionCount)
	assert.True(t, hm.mountHooks[hookID].executed)
	
	// Test that it doesn't execute again
	err = hm.ExecuteMountHooks()
	assert.NoError(t, err)
	assert.Equal(t, 1, executionCount)
}

func TestOnMountWithError(t *testing.T) {
	hm := NewHookManager("TestComponent")
	expectedErr := errors.New("mount error")
	
	hookID := hm.OnMount(func() error {
		return expectedErr
	})
	
	err := hm.ExecuteMountHooks()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr.Error())
	assert.True(t, hm.mountHooks[hookID].executed)
}

func TestOnUpdate(t *testing.T) {
	hm := NewHookManager("TestComponent")
	executionCount := 0
	
	deps := []interface{}{1, "test", true}
	hookID := hm.OnUpdate(func(prevDeps []interface{}) error {
		executionCount++
		return nil
	}, deps)
	
	assert.NotEmpty(t, hookID)
	assert.Contains(t, hm.updateHooks, hookID)
	assert.Equal(t, deps, hm.updateHooks[hookID].deps)
	assert.False(t, hm.updateHooks[hookID].executed)
	
	// First execution should always run
	err := hm.ExecuteUpdateHooks()
	assert.NoError(t, err)
	assert.Equal(t, 1, executionCount)
	assert.True(t, hm.updateHooks[hookID].executed)
	
	// Second execution with same deps should not run
	err = hm.ExecuteUpdateHooks()
	assert.NoError(t, err)
	assert.Equal(t, 1, executionCount)
	
	// Update dependencies
	newDeps := []interface{}{2, "test", true}
	err = hm.UpdateHookDependencies(hookID, newDeps)
	assert.NoError(t, err)
	
	// Should execute again with changed deps
	err = hm.ExecuteUpdateHooks()
	assert.NoError(t, err)
	assert.Equal(t, 2, executionCount)
}

func TestOnUpdateWithCustomEquals(t *testing.T) {
	hm := NewHookManager("TestComponent")
	executionCount := 0
	
	type customType struct {
		ID   int
		Name string
	}
	
	deps := []interface{}{customType{ID: 1, Name: "test"}}
	
	// Custom equality function that only compares ID
	equals := func(a, b interface{}) bool {
		objA, okA := a.(customType)
		objB, okB := b.(customType)
		if !okA || !okB {
			return false
		}
		return objA.ID == objB.ID
	}
	
	hookID := hm.OnUpdateWithEquals(func(prevDeps []interface{}) error {
		executionCount++
		return nil
	}, deps, equals)
	
	// First execution
	err := hm.ExecuteUpdateHooks()
	assert.NoError(t, err)
	assert.Equal(t, 1, executionCount)
	
	// Update with same ID but different name
	newDeps := []interface{}{customType{ID: 1, Name: "changed"}}
	err = hm.UpdateHookDependencies(hookID, newDeps)
	assert.NoError(t, err)
	
	// Should not execute because our custom equals only checks ID
	err = hm.ExecuteUpdateHooks()
	assert.NoError(t, err)
	assert.Equal(t, 1, executionCount)
	
	// Update with different ID
	newDeps = []interface{}{customType{ID: 2, Name: "changed"}}
	err = hm.UpdateHookDependencies(hookID, newDeps)
	assert.NoError(t, err)
	
	// Should execute
	err = hm.ExecuteUpdateHooks()
	assert.NoError(t, err)
	assert.Equal(t, 2, executionCount)
}

func TestOnUnmount(t *testing.T) {
	hm := NewHookManager("TestComponent")
	executionCount := 0
	
	hookID := hm.OnUnmount(func() error {
		executionCount++
		return nil
	})
	
	assert.NotEmpty(t, hookID)
	assert.Contains(t, hm.unmountHooks, hookID)
	
	// Test execution
	err := hm.ExecuteUnmountHooks()
	assert.NoError(t, err)
	assert.Equal(t, 1, executionCount)
	
	// Test that it executes again (unlike mount hooks)
	err = hm.ExecuteUnmountHooks()
	assert.NoError(t, err)
	assert.Equal(t, 2, executionCount)
}

func TestOnUnmountWithError(t *testing.T) {
	hm := NewHookManager("TestComponent")
	expectedErr := errors.New("unmount error")
	
	hm.OnUnmount(func() error {
		return expectedErr
	})
	
	err := hm.ExecuteUnmountHooks()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr.Error())
}

func TestRemoveHook(t *testing.T) {
	hm := NewHookManager("TestComponent")
	
	// Add hooks of each type
	mountID := hm.OnMount(func() error { return nil })
	updateID := hm.OnUpdate(func(prevDeps []interface{}) error { return nil }, []interface{}{1})
	unmountID := hm.OnUnmount(func() error { return nil })
	
	// Remove each type of hook
	err := hm.RemoveHook(mountID)
	assert.NoError(t, err)
	assert.NotContains(t, hm.mountHooks, mountID)
	
	err = hm.RemoveHook(updateID)
	assert.NoError(t, err)
	assert.NotContains(t, hm.updateHooks, updateID)
	
	err = hm.RemoveHook(unmountID)
	assert.NoError(t, err)
	assert.NotContains(t, hm.unmountHooks, unmountID)
	
	// Try to remove non-existent hook
	err = hm.RemoveHook("non-existent")
	assert.Error(t, err)
}

func TestConcurrentHookOperations(t *testing.T) {
	hm := NewHookManager("TestComponent")
	
	var wg sync.WaitGroup
	numGoroutines := 50
	hookIDs := make([]HookID, 0, numGoroutines*3)
	
	// Concurrent hook registration
	for i := 0; i < numGoroutines; i++ {
		wg.Add(3) // One for each hook type
		
		go func(idx int) {
			defer wg.Done()
			id := hm.OnMount(func() error { return nil })
			hookIDs = append(hookIDs, id)
		}(i)
		
		go func(idx int) {
			defer wg.Done()
			id := hm.OnUpdate(func(prevDeps []interface{}) error { return nil }, []interface{}{idx})
			hookIDs = append(hookIDs, id)
		}(i)
		
		go func(idx int) {
			defer wg.Done()
			id := hm.OnUnmount(func() error { return nil })
			hookIDs = append(hookIDs, id)
		}(i)
	}
	
	wg.Wait()
	
	// Verify all hooks were registered
	assert.Equal(t, numGoroutines, len(hm.mountHooks))
	assert.Equal(t, numGoroutines, len(hm.updateHooks))
	assert.Equal(t, numGoroutines, len(hm.unmountHooks))
	
	// Test concurrent execution
	var execWg sync.WaitGroup
	execWg.Add(3)
	
	go func() {
		defer execWg.Done()
		err := hm.ExecuteMountHooks()
		assert.NoError(t, err)
	}()
	
	go func() {
		defer execWg.Done()
		err := hm.ExecuteUpdateHooks()
		assert.NoError(t, err)
	}()
	
	go func() {
		defer execWg.Done()
		err := hm.ExecuteUnmountHooks()
		assert.NoError(t, err)
	}()
	
	execWg.Wait()
}

func TestHookManagerStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}
	
	hm := NewHookManager("StressTestComponent")
	numHooks := 1000
	
	// Register many hooks
	for i := 0; i < numHooks; i++ {
		hm.OnMount(func() error { return nil })
		hm.OnUpdate(func(prevDeps []interface{}) error { return nil }, []interface{}{i})
		hm.OnUnmount(func() error { return nil })
	}
	
	// Verify hook counts
	assert.Equal(t, numHooks, len(hm.mountHooks))
	assert.Equal(t, numHooks, len(hm.updateHooks))
	assert.Equal(t, numHooks, len(hm.unmountHooks))
	
	// Execute all hooks
	err := hm.ExecuteMountHooks()
	assert.NoError(t, err)
	
	err = hm.ExecuteUpdateHooks()
	assert.NoError(t, err)
	
	err = hm.ExecuteUnmountHooks()
	assert.NoError(t, err)
}

func TestUpdateHookOrderPreservation(t *testing.T) {
	hm := NewHookManager("TestComponent")
	executionOrder := make([]int, 0, 3)
	
	// Register hooks that track execution order
	hm.OnUpdate(func(prevDeps []interface{}) error {
		executionOrder = append(executionOrder, 1)
		return nil
	}, []interface{}{1})
	
	hm.OnUpdate(func(prevDeps []interface{}) error {
		executionOrder = append(executionOrder, 2)
		return nil
	}, []interface{}{2})
	
	hm.OnUpdate(func(prevDeps []interface{}) error {
		executionOrder = append(executionOrder, 3)
		return nil
	}, []interface{}{3})
	
	// Execute hooks
	err := hm.ExecuteUpdateHooks()
	assert.NoError(t, err)
	
	// Verify execution order is preserved
	assert.Equal(t, []int{1, 2, 3}, executionOrder)
}

func TestHookPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	// Create a large number of components with hooks
	numComponents := 100
	components := make([]*HookManager, numComponents)
	
	for i := 0; i < numComponents; i++ {
		components[i] = NewHookManager(fmt.Sprintf("Component%d", i))
		
		// Add hooks to each component
		components[i].OnMount(func() error { return nil })
		
		// Add multiple update hooks with different dependencies
		for j := 0; j < 5; j++ {
			components[i].OnUpdate(func(prevDeps []interface{}) error { return nil }, []interface{}{j})
		}
		
		components[i].OnUnmount(func() error { return nil })
	}
	
	// Measure time to execute all mount hooks
	for i := 0; i < numComponents; i++ {
		err := components[i].ExecuteMountHooks()
		assert.NoError(t, err)
	}
	
	// Measure time to execute all update hooks
	for i := 0; i < numComponents; i++ {
		err := components[i].ExecuteUpdateHooks()
		assert.NoError(t, err)
	}
	
	// This test doesn't assert on timing, but could be extended to do so
	// if benchmarks are desired
}

func TestErrorCollectionBehavior(t *testing.T) {
	hm := NewHookManager("TestComponent")
	
	// Register multiple hooks that return errors
	hm.OnMount(func() error {
		return errors.New("error 1")
	})
	
	hm.OnMount(func() error {
		return errors.New("error 2")
	})
	
	hm.OnMount(func() error {
		return nil // This one succeeds
	})
	
	// Execute mount hooks
	err := hm.ExecuteMountHooks()
	
	// Verify we get at least one error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error")
	
	// Register multiple update hooks that return errors
	hm.OnUpdate(func(prevDeps []interface{}) error {
		return errors.New("update error 1")
	}, []interface{}{1})
	
	hm.OnUpdate(func(prevDeps []interface{}) error {
		return errors.New("update error 2")
	}, []interface{}{2})
	
	// Execute update hooks
	err = hm.ExecuteUpdateHooks()
	
	// Verify we get at least one error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}
