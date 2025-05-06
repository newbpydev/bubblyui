package core

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEffectDependencyChangeDetection(t *testing.T) {
	t.Run("Skip when dependencies don't change", func(t *testing.T) {
		// Create a signal
		counter := CreateSignal(0)

		// Track effect executions
		var effectRunCount atomic.Int32

		// Create an effect that reads the counter
		effectID := CreateEffect(func() {
			_ = counter.Value() // Read the signal to create dependency
			effectRunCount.Add(1)
		})

		// Effect should have run once on creation
		assert.Equal(t, int32(1), effectRunCount.Load(), "Effect should run on creation")

		// Change the value to trigger the effect
		counter.Set(1)
		time.Sleep(10 * time.Millisecond) // Give time for async processing
		assert.Equal(t, int32(2), effectRunCount.Load(), "Effect should run when dependency changes")

		// Change to same value - should NOT run the effect again
		counter.Set(1) // No change in actual value
		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, int32(2), effectRunCount.Load(), "Effect should not run when value doesn't change")

		// Change to different value - should run the effect
		counter.Set(2)
		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, int32(3), effectRunCount.Load(), "Effect should run when dependency changes again")

		// Cleanup
		RemoveEffect(effectID)
	})

	t.Run("Various Dependency Patterns", func(t *testing.T) {
		// Create signals with different data types to test varied dependency patterns
		counter := CreateSignal(0)
		message := CreateSignal("Hello")
		active := CreateSignal(true)

		// Create tracking variables
		var effectRunCount atomic.Int32

		// Create computed signal that depends on multiple signals
		computed := CreateComputed(func() string {
			count := counter.Value()
			msg := message.Value()
			isActive := active.Value()

			if isActive {
				return fmt.Sprintf("%s x%d", msg, count)
			}
			return "Inactive"
		})

		// Create effect that depends on the computed signal
		effectID := CreateEffect(func() {
			_ = computed.Value()
			effectRunCount.Add(1)
		})

		// Effect should have run once on creation
		assert.Equal(t, int32(1), effectRunCount.Load(), "Effect should run on creation")

		// Change counter - should trigger effect
		counter.Set(5)
		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, int32(2), effectRunCount.Load(), "Effect should run when nested dependency changes")

		// Change message - should trigger effect
		message.Set("Greetings")
		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, int32(3), effectRunCount.Load(), "Effect should run when string dependency changes")

		// Change active to false - should trigger effect
		active.Set(false)
		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, int32(4), effectRunCount.Load(), "Effect should run when boolean dependency changes")

		// Now active is false, changing counter or message shouldn't matter
		// because the computed value doesn't use them anymore
		active.Set(false)          // should not run effect again (value didn't change)
		counter.Set(10)            // should not run effect (active is false)
		message.Set("New Message") // should not run effect (active is false)
		time.Sleep(10 * time.Millisecond)

		// The effect should still have only run 4 times
		assert.Equal(t, int32(4), effectRunCount.Load(), "Effect should not run when dependencies change but aren't used")

		// Now set active back to true, should trigger effect
		active.Set(true)
		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, int32(5), effectRunCount.Load(), "Effect should run when dependencies become visible again")

		// Cleanup
		RemoveEffect(effectID)
	})
}

func TestEffectScheduling(t *testing.T) {
	t.Run("Priority-based Scheduling", func(t *testing.T) {
		// Reset global state
		globalMutex.Lock()
		processingQueue = false
		batchMode = false
		effectQueue = []string{}
		batchedSignals = make(map[string]any)
		pendingEffects = make(map[string]bool)
		globalMutex.Unlock()

		// Track execution order directly
		executionOrder := []string{}

		// Create effect functions
		highFn := func() { executionOrder = append(executionOrder, "high") }
		normalFn := func() { executionOrder = append(executionOrder, "normal") }
		lowFn := func() { executionOrder = append(executionOrder, "low") }

		// Manually create effect IDs
		highID := fmt.Sprintf("high_%d", time.Now().UnixNano())
		normalID := fmt.Sprintf("normal_%d", time.Now().UnixNano()+1)
		lowID := fmt.Sprintf("low_%d", time.Now().UnixNano()+2)

		// Register effects directly
		globalMutex.Lock()
		effectsRegistry[highID] = &Effect{fn: highFn, debugInfo: "high_priority"}
		effectsRegistry[normalID] = &Effect{fn: normalFn, debugInfo: "normal_priority"}
		effectsRegistry[lowID] = &Effect{fn: lowFn, debugInfo: "low_priority"}

		// Set priorities explicitly
		effectInfos[highID] = EffectInfo{Priority: PriorityHigh}
		effectInfos[normalID] = EffectInfo{Priority: PriorityNormal}
		effectInfos[lowID] = EffectInfo{Priority: PriorityLow}

		// Add to queue in reverse order
		effectQueue = []string{lowID, normalID, highID}
		globalMutex.Unlock()

		// Process the queue directly
		processEffectQueue()

		// Verify execution order
		expectedOrder := []string{"high", "normal", "low"}
		assert.Equal(t, expectedOrder, executionOrder, "Effects should run in priority order")

		// Cleanup
		globalMutex.Lock()
		delete(effectsRegistry, highID)
		delete(effectsRegistry, normalID)
		delete(effectsRegistry, lowID)
		globalMutex.Unlock()
	})

	t.Run("Effect Batching", func(t *testing.T) {
		// Enable debug to see what's happening
		EnableDebugMode()
		defer DisableDebugMode()

		// Create a timeout for the test
		testTimeout := time.After(2 * time.Second)

		// Create channels for thread-safe communication between effects and test
		// Use larger buffer sizes to prevent deadlocks
		batch1Channel := make(chan string, 10)
		noBatchChannel := make(chan string, 10)
		batch2Channel := make(chan string, 10)

		// Create signals - use constants for initial values to avoid confusion
		trigger1 := CreateSignal(42)
		trigger2 := CreateSignal("test")

		// Create a function to drain channels before the real test
		drainChannels := func() {
			fmt.Println("Draining channels from initial effects...")
			// Drain with timeout to prevent hanging
			drainTimeout := time.After(100 * time.Millisecond)
			draining := true
			for draining {
				select {
				case msg := <-batch1Channel:
					fmt.Printf("Drained from batch1: %s\n", msg)
				case msg := <-noBatchChannel:
					fmt.Printf("Drained from noBatch: %s\n", msg)
				case msg := <-batch2Channel:
					fmt.Printf("Drained from batch2: %s\n", msg)
				case <-drainTimeout:
					fmt.Println("Drain timeout reached")
					draining = false
				default:
					// All channels empty
					draining = false
				}
			}
		}

		// Create helper for effect creation with debug info
		createTestEffect := func(name string, ch chan string, triggerSignal interface{}) string {
			return CreateEffect(func() {
				// Access the signal to create a dependency
				// The Value will be accessed differently based on the signal type
				var signalValue interface{}

				// Type switch to handle different signal types
				switch s := triggerSignal.(type) {
				case *Signal[int]:
					signalValue = s.Value()
				case *Signal[string]:
					signalValue = s.Value()
				default:
					fmt.Printf("Unexpected signal type: %T\n", triggerSignal)
				}

				msg := fmt.Sprintf("%s executed with triggerSignal value: %v", name, signalValue)
				fmt.Println(msg)

				// Send to channel without blocking
				select {
				case ch <- msg:
					// Successfully sent
				default:
					// Channel full or closed
					fmt.Printf("Warning: Could not send to channel for %s\n", name)
				}
			})
		}

		// Reset global state for clean test
		fmt.Println("Resetting global state...")
		globalMutex.Lock()
		processingQueue = false
		batchMode = false
		effectQueue = []string{}
		batchedSignals = make(map[string]any)
		pendingEffects = make(map[string]bool)
		globalMutex.Unlock()

		// Create effects with different batch IDs
		fmt.Println("Creating test effects...")
		effect1ID := createTestEffect("batch1_effect1", batch1Channel, trigger1)
		effect2ID := createTestEffect("batch1_effect2", batch1Channel, trigger1)
		effect3ID := createTestEffect("batch2_effect", batch2Channel, trigger2)
		noBatchEffectID := createTestEffect("no_batch_effect", noBatchChannel, trigger1)

		// Register effects with batches
		fmt.Println("Assigning effects to batches...")
		AddToEffectBatch(effect1ID, "batch1")
		AddToEffectBatch(effect2ID, "batch1")
		AddToEffectBatch(effect3ID, "batch2")
		// noBatchEffectID is intentionally not added to any batch

		// Initial effects execution will have triggered the channels, drain them
		drainChannels()

		// Update trigger1 in batch mode - this should affect batch1 and noBatch effects only
		fmt.Println("==== TEST STARTS HERE ====")
		fmt.Println("Triggering batch operation on trigger1...")
		Batch(func() {
			trigger1.Set(100) // Set to new value
		})

		// Wait and collect results with timeout
		batch1Count := 0
		noBatchRan := false
		batch2Ran := false

		fmt.Println("Waiting for effect execution results...")
		// Use timeout for test safety
		waitTimeout := time.After(500 * time.Millisecond)
		checkingResults := true

		for checkingResults {
			select {
			case msg := <-batch1Channel:
				batch1Count++
				fmt.Printf("Batch1 effect executed: %s\n", msg)

			case msg := <-noBatchChannel:
				noBatchRan = true
				fmt.Printf("Non-batch effect executed: %s\n", msg)

			case msg := <-batch2Channel:
				batch2Ran = true
				fmt.Printf("Batch2 effect executed (unexpected): %s\n", msg)

			case <-waitTimeout:
				fmt.Println("Test wait timeout reached")
				checkingResults = false

			case <-testTimeout:
				// Safety timeout for the entire test
				fmt.Println("Overall test timeout reached")
				checkingResults = false
			}
		}

		// Assert results
		fmt.Printf("\nTest Results: batch1Count=%d, noBatchRan=%v, batch2Ran=%v\n",
			batch1Count, noBatchRan, batch2Ran)

		// For full batching to work, batch1 effects must run
		assert.True(t, batch1Count > 0, "Batch1 effects should have executed")

		// Non-batched effects should also run even in batch mode
		assert.True(t, noBatchRan, "Non-batched effect should have executed")

		// Batch2 effects should NOT run since trigger2 wasn't updated
		assert.False(t, batch2Ran, "Batch2 effect should NOT have executed")

		// Cleanup
		fmt.Println("Cleaning up effects...")
		RemoveEffect(effect1ID)
		RemoveEffect(effect2ID)
		RemoveEffect(effect3ID)
		RemoveEffect(noBatchEffectID)
	})

	t.Run("Effect Cancellation", func(t *testing.T) {
		// Reset global state
		globalMutex.Lock()
		processingQueue = false
		batchMode = false
		effectQueue = []string{}
		batchedSignals = make(map[string]any)
		pendingEffects = make(map[string]bool)
		globalMutex.Unlock()

		// Track effect executions directly
		effect1Ran := false
		effect2Ran := false

		// Create effect functions
		effect1Fn := func() { effect1Ran = true }
		effect2Fn := func() { effect2Ran = true }

		// Manually create effect IDs
		effect1ID := fmt.Sprintf("effect1_%d", time.Now().UnixNano())
		effect2ID := fmt.Sprintf("effect2_%d", time.Now().UnixNano()+1)

		// Register effects directly
		globalMutex.Lock()
		effectsRegistry[effect1ID] = &Effect{fn: effect1Fn, debugInfo: "effect1"}
		effectsRegistry[effect2ID] = &Effect{fn: effect2Fn, debugInfo: "effect2"}

		// Mark effect1 as cancelled
		effectInfos[effect1ID] = EffectInfo{Priority: PriorityNormal, IsCancelled: true}
		effectInfos[effect2ID] = EffectInfo{Priority: PriorityNormal}

		// Add both to the queue
		effectQueue = []string{effect1ID, effect2ID}
		globalMutex.Unlock()

		// Process the queue
		processEffectQueue()

		// Verify only effect2 ran
		assert.False(t, effect1Ran, "Cancelled effect should not run")
		assert.True(t, effect2Ran, "Non-cancelled effect should run")

		// Cleanup
		globalMutex.Lock()
		delete(effectsRegistry, effect1ID)
		delete(effectsRegistry, effect2ID)
		globalMutex.Unlock()
	})

	t.Run("Deferred Effects", func(t *testing.T) {
		// Reset global state
		globalMutex.Lock()
		processingQueue = false
		batchMode = false
		effectQueue = []string{}
		batchedSignals = make(map[string]any)
		pendingEffects = make(map[string]bool)
		globalMutex.Unlock()

		// Track execution order directly
		executionOrder := []string{}

		// Create effect functions
		normalFn := func() { executionOrder = append(executionOrder, "normal") }
		deferredFn := func() { executionOrder = append(executionOrder, "deferred") }

		// Manually create effect IDs
		normalID := fmt.Sprintf("normal_%d", time.Now().UnixNano())
		deferredID := fmt.Sprintf("deferred_%d", time.Now().UnixNano()+1)

		// Register effects directly
		globalMutex.Lock()
		effectsRegistry[normalID] = &Effect{fn: normalFn, debugInfo: "normal_effect"}
		effectsRegistry[deferredID] = &Effect{fn: deferredFn, debugInfo: "deferred_effect"}

		// Set effect info - one is deferred, one is normal
		effectInfos[normalID] = EffectInfo{Priority: PriorityNormal, IsDeferred: false}
		effectInfos[deferredID] = EffectInfo{Priority: PriorityNormal, IsDeferred: true}

		// Add to queue in incorrect order (deferred first) to test sorting
		effectQueue = []string{deferredID, normalID}
		globalMutex.Unlock()

		// Process the queue directly
		processEffectQueue()

		// Verify execution order
		assert.Equal(t, 2, len(executionOrder), "Both effects should have run")
		assert.Equal(t, "normal", executionOrder[0], "Normal effect should run first")
		assert.Equal(t, "deferred", executionOrder[1], "Deferred effect should run second")

		// Cleanup
		globalMutex.Lock()
		delete(effectsRegistry, normalID)
		delete(effectsRegistry, deferredID)
		globalMutex.Unlock()
	})
}
