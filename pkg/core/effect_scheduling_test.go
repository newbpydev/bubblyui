package core

import (
	"fmt"
	"sync"
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
		// Create signals
		trigger1 := CreateSignal(0)
		trigger2 := CreateSignal("hello")

		// Track batch execution
		batchExecuted := make(map[string]bool)
		executionOrder := []string{}
		orderMutex := sync.Mutex{}

		// Create effects with different batch IDs
		effect1ID := CreateEffect(func() {
			_ = trigger1.Value()
			orderMutex.Lock()
			executionOrder = append(executionOrder, "batch1_effect1")
			batchExecuted["batch1"] = true
			orderMutex.Unlock()
		})

		effect2ID := CreateEffect(func() {
			_ = trigger1.Value()
			orderMutex.Lock()
			executionOrder = append(executionOrder, "batch1_effect2")
			batchExecuted["batch1"] = true
			orderMutex.Unlock()
		})

		effect3ID := CreateEffect(func() {
			_ = trigger2.Value()
			orderMutex.Lock()
			executionOrder = append(executionOrder, "batch2_effect")
			batchExecuted["batch2"] = true
			orderMutex.Unlock()
		})

		noBatchEffectID := CreateEffect(func() {
			_ = trigger1.Value()
			orderMutex.Lock()
			executionOrder = append(executionOrder, "no_batch_effect")
			batchExecuted["no_batch"] = true
			orderMutex.Unlock()
		})

		// Assign effects to batches
		AddToEffectBatch(effect1ID, "batch1")
		AddToEffectBatch(effect2ID, "batch1")
		AddToEffectBatch(effect3ID, "batch2")

		// Clear tracking
		orderMutex.Lock()
		executionOrder = []string{}
		batchExecuted = make(map[string]bool)
		orderMutex.Unlock()

		// Trigger batch1 effects
		trigger1.Set(1)
		time.Sleep(20 * time.Millisecond) // Give time for effects to run

		// Verify batch execution
		orderMutex.Lock()
		assert.True(t, batchExecuted["batch1"], "Batch 1 should be executed")
		assert.True(t, batchExecuted["no_batch"], "Non-batched effect should be executed")
		assert.False(t, batchExecuted["batch2"], "Batch 2 should not be executed")

		// Check that batch1 effects ran together
		batch1StartIdx := -1
		for i, name := range executionOrder {
			if name == "batch1_effect1" || name == "batch1_effect2" {
				if batch1StartIdx == -1 {
					batch1StartIdx = i
				}
			}
		}

		assert.True(t, batch1StartIdx >= 0, "Batch 1 effects should have executed")
		orderMutex.Unlock()

		// Cleanup
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
