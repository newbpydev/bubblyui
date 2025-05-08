package bubble

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/core"
	"github.com/stretchr/testify/assert"
)

// TestStateSynchronization verifies that state is properly synchronized between
// components and the bubble model.
func TestStateSynchronization(t *testing.T) {
	// Set a shorter timeout for tests to avoid hanging
	t.Parallel() // Run tests in parallel
	t.Run("Basic State Sharing", func(t *testing.T) {
		// Create just the root component - no model
		root := core.NewComponentManager("root")

		// Create a basic synchronizer with nil model for unit testing
		stateSync := NewStateSynchronizer(nil)

		// Register the component
		stateSync.RegisterComponent(root)

		// Create a shared state directly
		_, err := CreateSharedState[string](stateSync, "sharedTitle", "Initial Title")
		assert.NoError(t, err)

		// Set a value using shared state API
		err = stateSync.SetSharedState("sharedTitle", "Updated Title")
		assert.NoError(t, err)

		// Verify component received the update
		val, err := GetComponentState[string](stateSync, root, "sharedTitle")
		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", val)

		// Set a value from the component side
		err = stateSync.SetComponentState(root, "sharedTitle", "Component Updated Title")
		assert.NoError(t, err)

		// Verify the shared state was updated
		modelVal, err := GetSharedState[string](stateSync, "sharedTitle")
		assert.NoError(t, err)
		assert.Equal(t, "Component Updated Title", modelVal)
	})

	t.Run("Component Tree State Propagation", func(t *testing.T) {
		// Create a component tree
		root := core.NewComponentManager("root")
		child1 := core.NewComponentManager("child1")
		child2 := core.NewComponentManager("child2")
		grandchild := core.NewComponentManager("grandchild")

		// Build the component tree
		root.AddChild(child1)
		root.AddChild(child2)
		child1.AddChild(grandchild)

		// Create a simple synchronizer for testing without model dependency
		stateSync := NewStateSynchronizer(nil)

		// Register components
		stateSync.RegisterComponent(root)
		stateSync.RegisterComponent(child1)
		stateSync.RegisterComponent(child2)
		stateSync.RegisterComponent(grandchild)

		// Create a shared state that should propagate through the tree
		_, err := CreateSharedState[int](stateSync, "counter", 0)
		assert.NoError(t, err)

		// Update from a leaf component
		err = stateSync.SetComponentState(grandchild, "counter", 42)
		assert.NoError(t, err)

		// State should propagate to all components
		// Check in root
		rootVal, err := GetComponentState[int](stateSync, root, "counter")
		assert.NoError(t, err)
		assert.Equal(t, 42, rootVal)

		// Check in other branch
		child2Val, err := GetComponentState[int](stateSync, child2, "counter")
		assert.NoError(t, err)
		assert.Equal(t, 42, child2Val)
	})

	t.Run("State Persistence", func(t *testing.T) {
		// Add timeout to prevent hanging test
		done := make(chan bool)
		go func() {
			defer func() { done <- true }()

			// Create a simple state synchronizer without model dependency
			root := core.NewComponentManager("root")
			stateSync := NewStateSynchronizer(nil)
			stateSync.RegisterComponent(root)

			// Create shared persistent state
			_, err := CreatePersistentState[map[string]interface{}](stateSync,
				"userPreferences",
				map[string]interface{}{
					"theme":    "dark",
					"fontSize": 12,
				},
			)
			assert.NoError(t, err)

			// Modify state
			prefs, err := GetSharedState[map[string]interface{}](stateSync, "userPreferences")
			assert.NoError(t, err)

			prefs["theme"] = "light"
			prefs["fontSize"] = 14

			err = stateSync.SetSharedState("userPreferences", prefs)
			assert.NoError(t, err)

			// Create a snapshot
			snapshot, err := stateSync.CreateStateSnapshot()
			assert.NoError(t, err)
			assert.NotEmpty(t, snapshot)

			// Create a new state synchronizer for restoration
			newRoot := core.NewComponentManager("newRoot")
			newStateSync := NewStateSynchronizer(nil)
			newStateSync.RegisterComponent(newRoot)

			// Create the same state schema
			_, err = CreatePersistentState[map[string]interface{}](newStateSync,
				"userPreferences",
				map[string]interface{}{
					"theme":    "dark", // Default value, should be overwritten by snapshot
					"fontSize": 12,
				},
			)
			assert.NoError(t, err)

			// Restore from snapshot
			err = newStateSync.RestoreFromSnapshot(snapshot)
			assert.NoError(t, err)

			// Verify state was restored
			restoredPrefs, err := GetSharedState[map[string]interface{}](newStateSync, "userPreferences")
			assert.NoError(t, err)
			assert.Equal(t, "light", restoredPrefs["theme"])
			assert.Equal(t, float64(14), restoredPrefs["fontSize"]) // JSON unmarshalling will convert to float64
		}()

		// Wait for test to complete or timeout
		select {
		case <-done:
			// Test completed successfully
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out after 5 seconds")
		}
	})

	t.Run("State Versioning and Migration", func(t *testing.T) {
		// Add timeout to prevent hanging test
		done := make(chan bool)
		go func() {
			defer func() { done <- true }()

			// Create a simple state synchronizer without model dependency
			root := core.NewComponentManager("root")
			stateSync := NewStateSynchronizer(nil)
			stateSync.RegisterComponent(root)

			// Register a migration from v1 to v2 using maps for flexibility
			stateSync.RegisterMigration("userSettings", 1, 2, func(oldData json.RawMessage) (json.RawMessage, error) {
				// v1 structure as map
				var v1Data map[string]interface{}

				// Unmarshal v1 data
				if err := json.Unmarshal(oldData, &v1Data); err != nil {
					return nil, err
				}

				// v2 structure with nested maps
				v2Data := map[string]interface{}{
					"appearance": map[string]interface{}{
						"theme":    v1Data["theme"],
						"fontSize": v1Data["fontSize"],
					},
					"version": 2,
				}

				// Marshal v2 data
				return json.Marshal(v2Data)
			})

			// Create v1 format state as map for simpler serialization
			v1Settings := map[string]interface{}{
				"theme":    "dark",
				"fontSize": 12,
				"version":  1,
			}

			// Serialize to JSON
			v1Data, err := json.Marshal(v1Settings)
			assert.NoError(t, err)

			// Create a snapshot with v1 data
			snapshot := &StateSnapshot{
				Version: 1,
				States: map[string]StateEntry{
					"userSettings": {
						Type:    "map",
						Version: 1,
						Data:    v1Data,
					},
				},
			}

			// Serialize snapshot
			snapshotData, err := json.Marshal(snapshot)
			assert.NoError(t, err)

			// Use maps for consistency with what we registered
			// Define expected v2 structure as map
			v2Settings := map[string]interface{}{
				"appearance": map[string]interface{}{
					"theme":    "light",
					"fontSize": 14,
				},
				"version": 2,
			}

			// Register expected state with map type
			_, err = CreatePersistentStateWithVersion[map[string]interface{}](stateSync,
				"userSettings",
				v2Settings,
				2, // Current version is 2
			)
			assert.NoError(t, err)

			// Restore from v1 snapshot, should trigger migration
			err = stateSync.RestoreFromSnapshot(snapshotData)
			assert.NoError(t, err)

			// Verify migration worked using map instead of struct to avoid type conversion issues
			migratedSettings, err := GetSharedState[map[string]interface{}](stateSync, "userSettings")
			assert.NoError(t, err)

			// Check version
			assert.Equal(t, float64(2), migratedSettings["version"])

			// Check the nested appearance map
			appearance, ok := migratedSettings["appearance"].(map[string]interface{})
			assert.True(t, ok, "Expected appearance to be a map")

			// Check migrated values
			assert.Equal(t, "dark", appearance["theme"])
			assert.Equal(t, float64(12), appearance["fontSize"]) // JSON unmarshals numbers as float64
		}()

		// Wait for test to complete or timeout
		select {
		case <-done:
			// Test completed successfully
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out after 5 seconds")
		}
	})
}

// TestPersistenceWithConcurrentUpdates tests that state persistence works correctly
// with concurrent updates from multiple components.
func TestPersistenceWithConcurrentUpdates(t *testing.T) {
	// Add timeout protection
	t.Parallel()
	testDone := make(chan bool)
	go func() {
		defer func() { testDone <- true }()

		// Create a component tree
		root := core.NewComponentManager("root")
		child1 := core.NewComponentManager("child1")
		child2 := core.NewComponentManager("child2")

		root.AddChild(child1)
		root.AddChild(child2)

		// Create a simple synchronizer without model dependency
		stateSync := NewStateSynchronizer(nil)

		// Register components
		stateSync.RegisterComponent(root)
		stateSync.RegisterComponent(child1)
		stateSync.RegisterComponent(child2)

		// Create shared persistent state
		_, err := CreatePersistentState[map[string]int](stateSync,
			"counters",
			map[string]int{
				"value1": 0,
				"value2": 0,
			},
		)
		assert.NoError(t, err)

		// Create done channel to coordinate goroutines
		done := make(chan bool)

		// Concurrent update from child1
		go func() {
			for i := 0; i < 100; i++ {
				counters, err := GetComponentState[map[string]int](stateSync, child1, "counters")
				assert.NoError(t, err)

				counters["value1"]++

				err = stateSync.SetComponentState(child1, "counters", counters)
				assert.NoError(t, err)

				time.Sleep(1 * time.Millisecond) // Small delay to simulate real-world usage
			}
			done <- true
		}()

		// Concurrent update from child2
		go func() {
			for i := 0; i < 100; i++ {
				counters, err := GetComponentState[map[string]int](stateSync, child2, "counters")
				assert.NoError(t, err)

				counters["value2"]++

				err = stateSync.SetComponentState(child2, "counters", counters)
				assert.NoError(t, err)

				time.Sleep(1 * time.Millisecond) // Small delay
			}
			done <- true
		}()

		// Wait for both goroutines to complete
		<-done
		<-done

		// Create a snapshot
		snapshot, err := stateSync.CreateStateSnapshot()
		assert.NoError(t, err)

		// Check final values
		counters, err := GetSharedState[map[string]int](stateSync, "counters")
		assert.NoError(t, err)
		assert.Equal(t, 100, counters["value1"])
		assert.Equal(t, 100, counters["value2"])

		// Verify state can be restored
		newRoot := core.NewComponentManager("newRoot")
		newStateSync := NewStateSynchronizer(nil)
		newStateSync.RegisterComponent(newRoot)

		// Create the same state schema
		_, err = CreatePersistentState[map[string]int](newStateSync,
			"counters",
			map[string]int{
				"value1": 0,
				"value2": 0,
			},
		)
		assert.NoError(t, err)

		// Restore from snapshot
		err = newStateSync.RestoreFromSnapshot(snapshot)
		assert.NoError(t, err)

		// Verify restored values
		restoredCounters, err := GetSharedState[map[string]int](newStateSync, "counters")
		assert.NoError(t, err)
		assert.Equal(t, 100, restoredCounters["value1"])
		assert.Equal(t, 100, restoredCounters["value2"])
	}()

	// Wait for test to complete or timeout
	select {
	case <-testDone:
		// Test completed successfully
	case <-time.After(10 * time.Second): // Longer timeout due to concurrent operations
		t.Fatal("Test timed out after 10 seconds")
	}
}

// TestComponentModelConsistency tests that state updates remain consistent
// across component tree restructuring and complex updates.
func TestComponentModelConsistency(t *testing.T) {
	// Add timeout protection
	t.Parallel()
	done := make(chan bool)
	go func() {
		defer func() { done <- true }()

		// Create a more complex component tree
		root := core.NewComponentManager("root")
		left := core.NewComponentManager("left")
		right := core.NewComponentManager("right")
		leftChild1 := core.NewComponentManager("leftChild1")
		leftChild2 := core.NewComponentManager("leftChild2")
		rightChild := core.NewComponentManager("rightChild")

		// Build initial tree
		root.AddChild(left)
		root.AddChild(right)
		left.AddChild(leftChild1)
		left.AddChild(leftChild2)
		right.AddChild(rightChild)

		// Create a simple synchronizer without model dependency
		stateSync := NewStateSynchronizer(nil)

		// Register all components
		stateSync.RegisterComponent(root)
		stateSync.RegisterComponent(left)
		stateSync.RegisterComponent(right)
		stateSync.RegisterComponent(leftChild1)
		stateSync.RegisterComponent(leftChild2)
		stateSync.RegisterComponent(rightChild)

		// Create several shared states
		_, err := CreateSharedState[int](stateSync, "counter", 0)
		assert.NoError(t, err)

		_, err = CreateSharedState[map[string]string](stateSync, "settings", map[string]string{
			"mode": "default",
			"view": "compact",
		})
		assert.NoError(t, err)

		// Make some updates from different components
		err = stateSync.SetComponentState(leftChild1, "counter", 5)
		assert.NoError(t, err)

		settings, err := GetComponentState[map[string]string](stateSync, rightChild, "settings")
		assert.NoError(t, err)
		settings["mode"] = "advanced"
		err = stateSync.SetComponentState(rightChild, "settings", settings)
		assert.NoError(t, err)

		// Now restructure the component tree
		left.RemoveChild(leftChild2)
		right.AddChild(leftChild2)

		// Update state again from moved components
		err = stateSync.SetComponentState(leftChild2, "counter", 10)
		assert.NoError(t, err)

		// Verify state is consistent across the restructured tree
		values := make(map[string]int)
		components := []*core.ComponentManager{root, left, right, leftChild1, leftChild2, rightChild}

		for _, comp := range components {
			val, err := GetComponentState[int](stateSync, comp, "counter")
			assert.NoError(t, err)
			values[comp.GetName()] = val

			// All components should see the same value
			assert.Equal(t, 10, val, "Component %s has incorrect counter value", comp.GetName())

			// Also check settings
			settings, err := GetComponentState[map[string]string](stateSync, comp, "settings")
			assert.NoError(t, err)
			assert.Equal(t, "advanced", settings["mode"], "Component %s has incorrect settings", comp.GetName())
		}
	}()

	// Wait for test to complete or timeout
	select {
	case <-done:
		// Test completed successfully
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out after 5 seconds")
	}
}
